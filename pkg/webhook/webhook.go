package webhook

import (
	"context"
	"errors"
	"os"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	"github.com/disgoorg/snowflake/v2"
)

type WebhookClient struct {
	client webhook.Client
}

func (c *WebhookClient) Send(message string) error {
	_, err := c.client.CreateMessage(discord.NewWebhookMessageCreateBuilder().SetContent(message).Build())
	if err != nil {
		return err
	}

	return nil
}

func (c *WebhookClient) Close() {
	c.client.Close(context.Background())
}

func NewClient() (*WebhookClient, error) {
	webhookId := snowflake.GetEnv("webhook_id")
	if webhookId == 0 {
		return nil, errors.New("please set the 'webhook_id environment variable")
	}

	webhookToken := os.Getenv("webhook_token")
	if webhookToken == "" {
		return nil, errors.New("please set the 'webhook_token' environment variable")
	}

	client := webhook.New(webhookId, webhookToken)

	wh := WebhookClient{
		client,
	}

	return &wh, nil
}
