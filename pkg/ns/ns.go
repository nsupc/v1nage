package ns

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Nation struct {
	Name     string `xml:"id,attr"`
	WAStatus string `xml:"UNSTATUS"`
}

type Client struct {
	client             http.Client
	user               string
	requests           []time.Time
	ratelimitLimit     int
	ratelimitRemaining int
	ratelimitResetIn   time.Duration
	maxRequests        int
}

func (c *Client) clearBucket() {
	now := time.Now()

	filtered := []time.Time{}

	for _, instant := range c.requests {
		if now.Sub(instant) <= 30*time.Second {
			filtered = append(filtered, instant)
		}
	}

	c.requests = filtered
}

func (c *Client) acquire() error {
	c.clearBucket()

	if len(c.requests) >= c.maxRequests {
		return errors.New("too many requests")
	}

	now := time.Now()

	c.requests = append(c.requests, now)

	return nil
}

func (c *Client) GetNation(name string) (*Nation, error) {
	nationName := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(name)), " ", "_")

	url := fmt.Sprintf("https://www.nationstates.net/cgi-bin/api.cgi?nation=%s&q=wa", nationName)

	err := c.acquire()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", c.user)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	limit := resp.Header.Get("ratelimit-limit")
	if limit != "" {
		limit, err := strconv.Atoi(limit)
		if err != nil {
			slog.Warn("failed to convert ratelimit-limit to int", slog.Any("error", err))
		} else {
			c.ratelimitLimit = limit
		}

	}

	remaining := resp.Header.Get("ratelimit-remaining")
	if remaining != "" {
		remaining, err := strconv.Atoi(remaining)
		if err != nil {
			slog.Warn("failed to convert ratelimit-remaining to int", slog.Any("error", err))
		} else {
			c.ratelimitRemaining = remaining
		}
	}

	reset := resp.Header.Get("ratelimit-reset")
	if reset != "" {
		reset, err := strconv.Atoi(reset)
		if err != nil {
			slog.Warn("failed to convert ratelimit-reset to int", slog.Any("error", err))
		} else {
			c.ratelimitResetIn = time.Duration(reset)
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	nation := Nation{}

	err = xml.Unmarshal(body, &nation)
	if err != nil {
		return nil, err
	}

	return &nation, nil
}

func New(user string, maxRequests int) *Client {
	client := Client{
		client:             http.Client{},
		user:               user,
		requests:           []time.Time{},
		ratelimitLimit:     50,
		ratelimitRemaining: 50,
		ratelimitResetIn:   30 * time.Second,
		maxRequests:        maxRequests,
	}

	return &client
}
