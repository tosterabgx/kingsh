package lichess

import "net/http"

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
