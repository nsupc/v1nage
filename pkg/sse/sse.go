package sse

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/tmaxmax/go-sse"
)

type SSEClient struct {
	client sse.Client
}

func (s *SSEClient) Subscribe(url string, onEvent sse.EventCallback) error {
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, http.NoBody)
	if err != nil {
		return err
	}

	conn := s.client.NewConnection(req)

	_ = conn.SubscribeToAll(onEvent)

	err = retry.Do(
		func() error {
			err = conn.Connect()
			if err != nil {
				return err
			}

			return nil
		},
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			slog.Error("retrying connection", slog.Any("retry", n), slog.Any("error", err))
		}),
		retry.UntilSucceeded(),
	)
	if err != nil {
		return err
	}

	return nil
}

func New() *SSEClient {
	client := SSEClient{
		client: sse.Client{
			HTTPClient: &http.Client{},
			OnRetry:    OnRetry,
			Backoff:    sse.DefaultClient.Backoff,
		},
	}

	return &client
}

func OnRetry(err error, duration time.Duration) {
	slog.Error("disconnect", slog.Any("error", err), slog.Duration("retry", duration))
}

type Event struct {
	Id   string  `json:"id"`
	Time float64 `json:"time"`
	Text string  `json:"str"`
}
