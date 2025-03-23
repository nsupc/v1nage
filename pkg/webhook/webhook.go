package webhook

import (
	"context"

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

func New(id string, token string) (*WebhookClient, error) {
	webhookId, err := snowflake.Parse(id)
	if err != nil {
		return nil, err
	}

	client := webhook.New(webhookId, token)

	wh := WebhookClient{
		client,
	}

	return &wh, nil
}
