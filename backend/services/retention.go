package services

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

const (
	defaultMessageRetentionDays      = 90
	defaultWebhookLogRetentionDays   = 30
	defaultIncomingRetentionDays     = 7
)

// PurgeExpiredData deletes old SMS messages and webhook delivery logs
// based on retention periods configured in system_config.
func PurgeExpiredData(app core.App) error {
	msgDays := getRetentionDays(app, "message_retention_days", defaultMessageRetentionDays)
	incomingDays := getRetentionDays(app, "incoming_message_retention_days", defaultIncomingRetentionDays)
	webhookDays := getRetentionDays(app, "webhook_log_retention_days", defaultWebhookLogRetentionDays)

	// Purge incoming messages (shorter TTL)
	incomingCutoff := time.Now().UTC().AddDate(0, 0, -incomingDays).Format("2006-01-02 15:04:05")
	incomingRes, err := app.DB().NewQuery(
		"DELETE FROM sms_messages WHERE message_type = 'incoming' AND created < {:cutoff}",
	).Bind(dbx.Params{"cutoff": incomingCutoff}).Execute()
	if err != nil {
		return err
	}
	incomingCount, _ := incomingRes.RowsAffected()

	// Purge outgoing messages (standard TTL)
	msgCutoff := time.Now().UTC().AddDate(0, 0, -msgDays).Format("2006-01-02 15:04:05")
	msgRes, err := app.DB().NewQuery(
		"DELETE FROM sms_messages WHERE message_type != 'incoming' AND created < {:cutoff}",
	).Bind(dbx.Params{"cutoff": msgCutoff}).Execute()
	if err != nil {
		return err
	}
	msgCount, _ := msgRes.RowsAffected()

	// Purge webhook delivery logs
	whCutoff := time.Now().UTC().AddDate(0, 0, -webhookDays).Format("2006-01-02 15:04:05")
	whRes, err := app.DB().NewQuery(
		"DELETE FROM webhook_delivery_logs WHERE created < {:cutoff}",
	).Bind(dbx.Params{"cutoff": whCutoff}).Execute()
	if err != nil {
		return err
	}
	whCount, _ := whRes.RowsAffected()

	total := incomingCount + msgCount + whCount
	if total > 0 {
		app.Logger().Info("Purged expired data",
			slog.Int64("incoming_messages", incomingCount),
			slog.Int64("outgoing_messages", msgCount),
			slog.Int64("webhook_logs", whCount),
		)
	}

	return nil
}

func getRetentionDays(app core.App, key string, fallback int) int {
	val := GetSystemConfigValue(app, key)
	if val == "" {
		return fallback
	}
	days, err := strconv.Atoi(val)
	if err != nil || days < 1 {
		return fallback
	}
	return days
}
