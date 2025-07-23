package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"v1nage/pkg/config"
	"v1nage/pkg/ns"
	"v1nage/pkg/sse"
	"v1nage/pkg/webhook"

	"github.com/nsupc/eurogo/client"
	"github.com/nsupc/eurogo/telegrams"
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

	eurocoreClient := client.New(conf.Eurocore.Username, conf.Eurocore.Password, conf.Eurocore.Url)

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

			go func() {
				nationLink := fmt.Sprintf("[%s](https://www.nationstates.net/nation=%s#composebutton)", nationName, nationName)
				msg := strings.ReplaceAll(conf.JoinMessage, "$nation", nationLink)

				err = webhookClient.Send(msg)
				if err != nil {
					slog.Error("unable to send webhook", slog.Any("error", err))
				}
			}()

			go func() {
				var telegram telegrams.NewTelegram

				if conf.JoinTelegram.Template != "" {
					tmpl, err := eurocoreClient.GetTemplate(conf.JoinTelegram.Template)
					if err != nil {
						slog.Error("unable to retrieve template", slog.Any("error", err))
						return
					}

					telegram = telegrams.New(tmpl.Nation, nationName, strconv.Itoa(tmpl.Tgid), tmpl.Key, telegrams.Standard)
				} else if conf.JoinTelegram.Id != "" {
					telegram = telegrams.New(conf.JoinTelegram.Author, nationName, conf.JoinTelegram.Id, conf.JoinTelegram.Secret, telegrams.Standard)
				} else {
					slog.Warn("join telegram not set, skipping")
					return
				}

				err := eurocoreClient.SendTelegram(telegram)
				if err != nil {
					slog.Error("unable to send join telegram", slog.Any("error", err))
				} else {
					slog.Info("join telegram sent", slog.String("recipient", nationName))
				}
			}()

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

			if nation.WAStatus == "WA Member" {
				go func() {
					nationLink := fmt.Sprintf("[%s](https://www.nationstates.net/nation=%s#composebutton)", nationName, nationName)
					msg := strings.ReplaceAll(conf.MoveMessage, "$nation", nationLink)

					err = webhookClient.Send(msg)
					if err != nil {
						slog.Error("unable to send webhook", slog.Any("error", err))
					}
				}()

				go func() {
					var telegram telegrams.NewTelegram

					if conf.MoveTelegram.Template != "" {
						tmpl, err := eurocoreClient.GetTemplate(conf.MoveTelegram.Template)
						if err != nil {
							slog.Error("unable to retrieve template", slog.Any("error", err))
							return
						}

						telegram = telegrams.New(tmpl.Nation, nationName, strconv.Itoa(tmpl.Tgid), tmpl.Key, telegrams.Standard)
					} else if conf.MoveTelegram.Id != "" {
						telegram = telegrams.New(conf.MoveTelegram.Author, nationName, conf.JoinTelegram.Id, conf.JoinTelegram.Secret, telegrams.Standard)
					} else {
						slog.Warn("move telegram not set, skipping")
						return
					}

					err := eurocoreClient.SendTelegram(telegram)
					if err != nil {
						slog.Error("unable to send move telegram", slog.Any("error", err))
					} else {
						slog.Info("move telegram sent", slog.String("recipient", nationName))
					}
				}()
			}

			return
		}
	})

	if err != nil {
		slog.Error("unable to subscribe to happenings", slog.Any("error", err))
	}
}
