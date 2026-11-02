package lichess

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const TokenURL = "https://lichess.org/account/oauth/token/create?scopes[]=board:play&description=kingsh"

type Client struct {
	Token   string
	BaseURL string
	HTTP    *http.Client
}

func NewClient(token string) *Client {
	return &Client{Token: token, BaseURL: "https://lichess.org", HTTP: &http.Client{}}
}

type Event struct {
	Type string `json:"type"`
	Game struct {
		ID    string `json:"gameId"`
		Color string `json:"color"`
	} `json:"game"`
}

type Player struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Rating  int    `json:"rating"`
	AILevel int    `json:"aiLevel"`
}

type GameState struct {
	Moves  string `json:"moves"`
	WTime  int64  `json:"wtime"`
	BTime  int64  `json:"btime"`
	Status string `json:"status"`
	Winner string `json:"winner"`
	WDraw  bool   `json:"wdraw"`
	BDraw  bool   `json:"bdraw"`
}

type GameEvent struct {
	Type string `json:"type"`

	White Player    `json:"white"`
	Black Player    `json:"black"`
	Rated bool      `json:"rated"`
	State GameState `json:"state"`
	Clock *struct {
		Initial   int64 `json:"initial"`
		Increment int64 `json:"increment"`
	} `json:"clock"`

	GameState

	Gone bool `json:"gone"`
}

func (e GameEvent) CurrentState() GameState {
	if e.Type == "gameFull" {
		return e.State
	}
	return e.GameState
}

type Seek struct {
	Time      time.Duration
	Increment time.Duration
	Rated     bool
}

func IsRapidOrSlower(base, inc time.Duration) bool {
	return base+40*inc >= 8*time.Minute
}

func (c *Client) StreamEvents(ctx context.Context, fn func(Event)) error {
	return c.stream(ctx, "GET", "/api/stream/event", nil, func(line []byte) error {
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			return err
		}
		fn(e)
		return nil
	})
}

func (c *Client) SeekGame(ctx context.Context, s Seek) error {
	form := url.Values{
		"rated":     {strconv.FormatBool(s.Rated)},
		"time":      {strconv.FormatFloat(s.Time.Minutes(), 'f', -1, 64)},
		"increment": {strconv.Itoa(int(s.Increment / time.Second))},
		"variant":   {"standard"},
		"color":     {"random"},
	}
	return c.stream(ctx, "POST", "/api/board/seek", form, func([]byte) error { return nil })
}

func (c *Client) StreamGame(ctx context.Context, gameID string, fn func(GameEvent)) error {
	return c.stream(ctx, "GET", "/api/board/game/stream/"+url.PathEscape(gameID), nil, func(line []byte) error {
		var e GameEvent
		if err := json.Unmarshal(line, &e); err != nil {
			return err
		}
		fn(e)
		return nil
	})
}

func (c *Client) Move(ctx context.Context, gameID, uci string) error {
	return c.post(ctx, "/api/board/game/"+url.PathEscape(gameID)+"/move/"+uci)
}

func (c *Client) Resign(ctx context.Context, gameID string) error {
	return c.post(ctx, "/api/board/game/"+url.PathEscape(gameID)+"/resign")
}

func (c *Client) Abort(ctx context.Context, gameID string) error {
	return c.post(ctx, "/api/board/game/"+url.PathEscape(gameID)+"/abort")
}

func (c *Client) OfferDraw(ctx context.Context, gameID string) error {
	return c.post(ctx, "/api/board/game/"+url.PathEscape(gameID)+"/draw/yes")
}

func (c *Client) post(ctx context.Context, path string) error {
	resp, err := c.request(ctx, "POST", path, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) stream(ctx context.Context, method, path string, form url.Values, fn func([]byte) error) error {
	resp, err := c.request(ctx, method, path, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		if err := fn(line); err != nil {
			return fmt.Errorf("lichess: reading %s: %w", path, err)
		}
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return sc.Err()
}

func (c *Client) request(ctx context.Context, method, path string, form url.Values) (*http.Response, error) {
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, apiError(resp)
	}
	return resp, nil
}

func apiError(resp *http.Response) error {
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var e struct {
		Error any `json:"error"`
	}
	msg := strings.TrimSpace(string(data))
	if json.Unmarshal(data, &e) == nil && e.Error != nil {
		msg = fmt.Sprint(e.Error)
	}
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("lichess rejected the token (%s); create one with the board:play scope at %s", msg, TokenURL)
	case http.StatusTooManyRequests:
		return fmt.Errorf("lichess: too many requests, wait a minute and try again")
	}
	if msg == "" {
		msg = resp.Status
	}
	return fmt.Errorf("lichess: %s", msg)
}
