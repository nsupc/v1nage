package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"v1nage/pkg/eurocore"
	"v1nage/pkg/sse"
	"v1nage/pkg/webhook"

	gsse "github.com/tmaxmax/go-sse"
)

func main() {
	var region, eurocoreUrl, eurocoreUser, eurocorePassword, webhookId, webhookToken, telegramSender, telegramId, telegramSecret string

	flag.StringVar(&region, "region", "", "region to waatch")
	flag.StringVar(&eurocoreUrl, "url", "", "base url for eurocore instance")
	flag.StringVar(&eurocoreUser, "user", "", "eurocore username")
	flag.StringVar(&eurocorePassword, "password", "", "eurocore user password")
	flag.StringVar(&webhookId, "webhook-id", "", "discord webhook id")
	flag.StringVar(&webhookToken, "webhook-token", "", "discord webhook token")
	flag.StringVar(&telegramSender, "telegram-sender", "", "nation sending telegram")
	flag.StringVar(&telegramId, "telegram-id", "", "telegram id")
	flag.StringVar(&telegramSecret, "telegram-secret", "", "telegram secret key")

	flag.Parse()

	region = strings.ReplaceAll(strings.ToLower(region), " ", "_")

	waRegex := regexp.MustCompile(`^@@(.*)@@ was admitted to the World Assembly.?$`)
	updateRegex := regexp.MustCompile(fmt.Sprintf("^%%%s%% updated.?$", region))

	eurocoreClient := eurocore.New(eurocoreUrl, eurocoreUser, eurocorePassword)

	webhookClient, err := webhook.New(webhookId, webhookToken)
	if err != nil {
		slog.Error("unable to build webhook client", slog.Any("error", err))
		os.Exit(1)
	}
	defer webhookClient.Close()

	sseClient := sse.New()

	happeningsUrl := fmt.Sprintf("https://www.nationstates.net/api/region:%s", region)
	sseClient.Subscribe(happeningsUrl, func(e gsse.Event) {
		event := sse.Event{}

		err = json.Unmarshal([]byte(e.Data), &event)
		if err != nil {
			slog.Error("unable to unmarshal event", slog.Any("error", err))
			return
		}

		if updateRegex.Match([]byte(event.Text)) {
			go webhookClient.Send(fmt.Sprintf("[%s](https://www.nationstates.net/region=%s) updated!", region, region))
			return
		}

		matches := waRegex.FindStringSubmatch(event.Text)

		if len(matches) > 0 {
			nationName := matches[1]

			telegram := eurocore.Telegram{
				Recipient: nationName,
				Sender:    telegramSender,
				Id:        telegramId,
				Secret:    telegramSecret,
				Type:      "standard",
			}

			go webhookClient.Send(fmt.Sprintf("[%s](https://www.nationstates.net/nation=%s#composebutton)", nationName, nationName))
			go eurocoreClient.SendTelegram(telegram)
		}
	})
}
