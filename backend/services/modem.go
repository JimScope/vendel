package services

import (
	"encoding/json"
	"log/slog"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/subscriptions"
)

// NotifyModemAgent sends an SSE event to modem agents subscribed to "modem/<deviceId>".
func NotifyModemAgent(app core.App, deviceId string, record *core.Record) {
	topic := "modem/" + deviceId

	data, err := json.Marshal(map[string]string{
		"message_id": record.Id,
		"recipient":  record.GetString("to"),
		"body":       record.GetString("body"),
	})
	if err != nil {
		app.Logger().Error("failed to marshal modem SSE payload", slog.Any("error", err))
		return
	}

	msg := subscriptions.Message{
		Name: topic,
		Data: data,
	}

	sent := 0
	for _, client := range app.SubscriptionsBroker().Clients() {
		if client.HasSubscription(topic) {
			client.Send(msg)
			sent++
		}
	}

	if sent > 0 {
		app.Logger().Info("notified modem agent via SSE",
			slog.String("device", deviceId),
			slog.String("message", record.Id),
			slog.Int("subscribers", sent),
		)
	}
}

// ClaimPendingMessages atomically marks assigned messages as "sending" for a device.
// Used on agent startup to recover any messages assigned while the agent was offline.
func ClaimPendingMessages(app core.App, deviceId string) ([]*core.Record, error) {
	records, err := app.FindRecordsByFilter(
		"sms_messages",
		"device = {:deviceId} && status = 'assigned' && message_type = 'outgoing'",
		"-created", 50, 0,
		dbx.Params{"deviceId": deviceId},
	)
	if err != nil {
		return nil, err
	}

	var claimed []*core.Record
	for _, r := range records {
		r.Set("status", "sending")
		if err := app.Save(r); err == nil {
			claimed = append(claimed, r)
		}
	}
	return claimed, nil
}
