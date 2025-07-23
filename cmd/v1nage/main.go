package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"regexp"
	"strings"

	"v1nage/pkg/config"
	"v1nage/pkg/eurocore"
	"v1nage/pkg/ns"
	"v1nage/pkg/sse"
	"v1nage/pkg/webhook"

	gsse "github.com/tmaxmax/go-sse"
)

func main() {
	conf, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	waRegex := regexp.MustCompile(`^@@(.*)@@ was admitted to the World Assembly.?$`)
	updateRegex := regexp.MustCompile(fmt.Sprintf(`^%%%%%s%%%% updated.?$`, conf.Region))
	moveRegex := regexp.MustCompile(fmt.Sprintf(`^@@(.+)@@ relocated from %%%%.+%%%% to %%%%%s%%%%.?$`, conf.Region))

	nsClient := ns.New(conf.User, int(conf.Limit))

	eurocoreClient := eurocore.New(conf.Eurocore.Url, conf.Eurocore.Username, conf.Eurocore.Password)

	webhookClient, err := webhook.New(conf.Webhook.Id, conf.Webhook.Token)
	if err != nil {
		slog.Error("unable to build webhook client", slog.Any("error", err))
		return
	}
	defer webhookClient.Close()

	sseClient := sse.New()

	happeningsUrl := fmt.Sprintf("https://www.nationstates.net/api/region:%s", conf.Region)

	err = sseClient.Subscribe(happeningsUrl, func(e gsse.Event) {
		event := sse.Event{}

		err = json.Unmarshal([]byte(e.Data), &event)
		if err != nil {
			slog.Error("unable to unmarshal event", slog.Any("error", err))
			return
		}

		if updateRegex.Match([]byte(event.Text)) {
			go func() {
				err = webhookClient.Send(fmt.Sprintf("[%s](https://www.nationstates.net/region=%s) updated!", conf.Region, conf.Region))
				if err != nil {
					slog.Error("unable to send webhook", slog.Any("error", err))
				}
			}()

			return
		}

		matches := waRegex.FindStringSubmatch(event.Text)

		if len(matches) > 0 {
			nationName := matches[1]

			nationLink := fmt.Sprintf("[%s](https://www.nationstates.net/nation=%s#composebutton)", nationName, nationName)
			msg := strings.ReplaceAll(conf.JoinMessage, "$nation", nationLink)

			go func() {
				err = webhookClient.Send(msg)
				if err != nil {
					slog.Error("unable to send webhook", slog.Any("error", err))
				}
			}()

			if conf.JoinTelegram.Secret != "" {
				telegram := eurocore.Telegram{
					Recipient: nationName,
					Sender:    conf.JoinTelegram.Author,
					Id:        conf.JoinTelegram.Id,
					Secret:    conf.JoinTelegram.Secret,
					Type:      "standard",
				}

				go eurocoreClient.SendTelegram(telegram)
			} else {
				slog.Warn("join telegram not set, skipping")
			}

			return
		}

		matches = moveRegex.FindStringSubmatch(event.Text)

		if len(matches) > 0 {
			nationName := matches[1]

			nation, err := nsClient.GetNation(nationName)
			if err != nil {
				slog.Error("unable to retrieve nation details", slog.Any("error", err))
				return
			}

			nationLink := fmt.Sprintf("[%s](https://www.nationstates.net/nation=%s#composebutton)", nationName, nationName)
			msg := strings.ReplaceAll(conf.MoveMessage, "$nation", nationLink)

			if nation.WAStatus == "WA Member" {
				go func() {
					err = webhookClient.Send(msg)
					if err != nil {
						slog.Error("unable to send webhook", slog.Any("error", err))
					}
				}()

				if conf.MoveTelegram.Secret != "" {
					telegram := eurocore.Telegram{
						Recipient: nationName,
						Sender:    conf.MoveTelegram.Author,
						Id:        conf.MoveTelegram.Id,
						Secret:    conf.MoveTelegram.Secret,
						Type:      "standard",
					}

					go eurocoreClient.SendTelegram(telegram)
				} else {
					slog.Warn("move telegram not set, skipping")
				}
			}

			return
		}
	})
	if err != nil {
		slog.Error("unable to subscribe to happenings", slog.Any("error", err))
	}
}
