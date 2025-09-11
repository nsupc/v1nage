package utils

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"v1nage/pkg/webhook"

	"github.com/nsupc/eurogo/client"
	"github.com/nsupc/eurogo/telegrams"
)

func HandleUpdate(client *webhook.WebhookClient, region string) {
	msg := fmt.Sprintf("[%s](https://www.nationstates.net/region=%s) updated!", region, region)

	err := client.Send(msg)
	if err != nil {
		slog.Error("failed to send region update notification", slog.Any("error", err))
	}
}

func HandleWa(wClient *webhook.WebhookClient, eClient client.Client, msg string, nation string, tmplId string) {
	link := fmt.Sprintf("[%s](https://www.nationstates.net/nation=%s#endorse)", nation, nation)
	msg = strings.ReplaceAll(msg, "$nation", link)

	err := wClient.Send(msg)
	if err != nil {
		slog.Error("failed to send wa event notification", slog.Any("error", err))
	}

	if tmplId == "" {
		slog.Warn("telegram template not set, skipping")
		return
	}

	tmpl, err := eClient.GetTemplate(tmplId)
	if err != nil {
		slog.Error("failed to retrieve telegram template", slog.Any("error", err), slog.String("template", tmplId))
	}

	telegram := telegrams.New(tmpl.Nation, nation, strconv.Itoa(tmpl.Tgid), tmpl.Key, telegrams.Standard)
	err = eClient.SendTelegram(telegram)
	if err != nil {
		slog.Error("failed to send telegram", slog.Any("error", err))
	}

	slog.Info("telegram sent", slog.String("recipient", nation))
}
