package chess

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Clock struct {
	Left      [2]time.Duration
	Increment time.Duration
	Active    Color
	Running   bool

	last time.Time
}

func NewClock(base, inc time.Duration) Clock {
	return Clock{Left: [2]time.Duration{base, base}, Increment: inc}
}

func (c *Clock) Update(now time.Time) bool {
	if !c.Running {
		return false
	}
	c.Left[c.Active] -= now.Sub(c.last)
	c.last = now
	if c.Left[c.Active] <= 0 {
		c.Left[c.Active] = 0
		c.Running = false
		return true
	}
	return false
}

func (c *Clock) Press(now time.Time, mover Color) {
	if c.Running {
		c.Update(now)
		c.Left[mover] += c.Increment
	}
	c.Active = mover.Other()
	c.Running = true
	c.last = now
}

func (c *Clock) Stop() { c.Running = false }

func ParseTimeControl(s string) (base, inc time.Duration, err error) {
	minStr, incStr, _ := strings.Cut(s, "+")
	m, err := strconv.ParseFloat(minStr, 64)
	if err != nil || m <= 0 {
		return 0, 0, fmt.Errorf("bad time control %q (want e.g 10+0 or 3+2)", s)
	}
	var i float64
	if incStr != "" {
		if i, err = strconv.ParseFloat(incStr, 64); err != nil || i < 0 {
			return 0, 0, fmt.Errorf("bad increment in time control %q", s)
		}
	}
	return time.Duration(m * float64(time.Minute)), time.Duration(i * float64(time.Second)), nil
}

func (c *Clock) Set(white, black time.Duration, active Color, running bool, now time.Time) {
	c.Left = [2]time.Duration{white, black}
	c.Active = active
	c.Running = running
	c.last = now
}
