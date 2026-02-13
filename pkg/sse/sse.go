package sse

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/r3labs/sse/v2"
	"gopkg.in/cenkalti/backoff.v1"
)

type SSEClient struct {
	client *sse.Client
}

func (s *SSEClient) Subscribe(cb func(Event)) {
	s.client.Subscribe("messages", func(msg *sse.Event) {
		event := Event{}

		err := json.Unmarshal(msg.Data, &event)
		if err != nil {
			slog.Error("unable to unmarshal event", slog.Any("error", err))
			return
		}

		cb(event)
	})
}

func New(url string) *SSEClient {
	backoffStrategy := backoff.NewExponentialBackOff()
	backoffStrategy.MaxElapsedTime = 0
	backoffStrategy.MaxInterval = 15 * time.Minute

	client := sse.NewClient(url)
	client.ReconnectStrategy = backoffStrategy

	return &SSEClient{
		client,
	}
}

type Event struct {
	Id   string `json:"id"`
	Time string `json:"time"`
	Text string `json:"str"`
}
