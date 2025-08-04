package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"regexp"

	"v1nage/pkg/config"
	"v1nage/pkg/ns"
	"v1nage/pkg/sse"
	"v1nage/pkg/utils"
	"v1nage/pkg/webhook"

	"github.com/nsupc/eurogo/client"
	gsse "github.com/tmaxmax/go-sse"
)

func main() {
	conf, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	joinRegex := regexp.MustCompile(`^@@(.*)@@ was admitted to the World Assembly.?$`)
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
			go utils.HandleUpdate(webhookClient, conf.Region)
			return
		}

		matches := joinRegex.FindStringSubmatch(event.Text)

		if len(matches) > 0 {
			nationName := matches[1]

			go utils.HandleWa(webhookClient, *eurocoreClient, conf.JoinMessage, nationName, conf.JoinTelegram.Template)

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
				go utils.HandleWa(webhookClient, *eurocoreClient, conf.MoveMessage, nationName, conf.MoveTelegram.Template)
			}
		}
	})

	if err != nil {
		slog.Error("unable to subscribe to happenings", slog.Any("error", err))
	}
}
