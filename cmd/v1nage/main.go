package main

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"v1nage/pkg/config"
	"v1nage/pkg/ns"
	"v1nage/pkg/sse"
	"v1nage/pkg/utils"
	"v1nage/pkg/webhook"

	"github.com/nsupc/eurogo/client"
)

func main() {
	conf, err := config.Read()
	if err != nil {
		slog.Error("failed to read config", slog.Any("error", err))
		os.Exit(1)
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

	happeningsUrl := fmt.Sprintf("https://www.nationstates.net/api/region:%s", conf.Region)

	sseClient := sse.New(happeningsUrl)
	sseClient.Subscribe(func(e sse.Event) {
		if updateRegex.Match([]byte(e.Text)) {
			go utils.HandleUpdate(webhookClient, conf.Region)
			return
		}

		matches := joinRegex.FindStringSubmatch(e.Text)

		if len(matches) > 0 {
			nationName := matches[1]

			go utils.HandleWa(webhookClient, *eurocoreClient, conf.JoinMessage, nationName, conf.JoinTelegram.Template)

			return
		}

		matches = moveRegex.FindStringSubmatch(e.Text)

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
}
