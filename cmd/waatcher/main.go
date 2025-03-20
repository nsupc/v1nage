package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"regexp"

	"waatcher/pkg/sse"
	"waatcher/pkg/webhook"

	gsse "github.com/tmaxmax/go-sse"
)

var waRegex = regexp.MustCompile(`^@@(.*)@@ was admitted to the World Assembly.?$`)
var updateRegex = regexp.MustCompile(`^%%europeia%% updated.?$`)

func main() {
	webhookClient, err := webhook.NewClient()
	if err != nil {
		slog.Error("unable to build webhook client", slog.Any("error", err))
		os.Exit(1)
	}
	defer webhookClient.Close()

	sseClient := sse.NewClient()

	sseClient.Subscribe("https://www.nationstates.net/api/region:europeia", func(e gsse.Event) {
		event := sse.Event{}

		err = json.Unmarshal([]byte(e.Data), &event)
		if err != nil {
			log.Printf("Error: %v\n", err)
			return
		}

		if updateRegex.Match([]byte(event.Text)) {
			webhookClient.Send("region updated!")
			return
		}

		matches := waRegex.FindStringSubmatch(event.Text)

		if len(matches) > 0 {
			nationName := matches[1]

			go webhookClient.Send(fmt.Sprintf("[%s](https://www.nationstates.net/nation=%s)", nationName, nationName))
		}
	})
}
