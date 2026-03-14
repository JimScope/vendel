package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// sms_messages: composite index for retry cron + general status queries
		messages, err := app.FindCollectionByNameOrId("sms_messages")
		if err != nil {
			return err
		}
		messages.AddIndex("idx_sms_messages_status_type_created", false, "status, message_type, created", "")
		messages.AddIndex("idx_sms_messages_user", false, "user", "")
		if err := app.Save(messages); err != nil {
			return err
		}

		// webhook_configs: active webhooks per user (used on every SMS event)
		webhooks, err := app.FindCollectionByNameOrId("webhook_configs")
		if err != nil {
			return err
		}
		webhooks.AddIndex("idx_webhook_configs_user_active", false, "user, active", "")
		if err := app.Save(webhooks); err != nil {
			return err
		}

		// webhook_delivery_logs: retry cron filters on delivery_status + next_retry_at
		logs, err := app.FindCollectionByNameOrId("webhook_delivery_logs")
		if err != nil {
			return err
		}
		logs.AddIndex("idx_webhook_logs_retry", false, "delivery_status, next_retry_at", "")
		if err := app.Save(logs); err != nil {
			return err
		}

		// subscriptions: renewal cron + general status queries
		subs, err := app.FindCollectionByNameOrId("subscriptions")
		if err != nil {
			return err
		}
		subs.AddIndex("idx_subscriptions_status", false, "status", "")
		if err := app.Save(subs); err != nil {
			return err
		}

		// Seed retention config defaults into system_config
		sysConfig, err := app.FindCollectionByNameOrId("system_config")
		if err != nil {
			return err
		}
		seeds := []struct{ key, value string }{
			{"message_retention_days", "90"},
			{"webhook_log_retention_days", "30"},
			{"incoming_retention_days", "7"},
		}
		for _, s := range seeds {
			existing, _ := app.FindFirstRecordByFilter("system_config", "key = {:key}", dbx.Params{"key": s.key})
			if existing != nil {
				continue // don't overwrite existing config
			}
			record := core.NewRecord(sysConfig)
			record.Set("key", s.key)
			record.Set("value", s.value)
			if err := app.Save(record); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		// Remove indexes (reverse order)
		if subs, err := app.FindCollectionByNameOrId("subscriptions"); err == nil {
			subs.RemoveIndex("idx_subscriptions_status")
			_ = app.Save(subs)
		}
		if logs, err := app.FindCollectionByNameOrId("webhook_delivery_logs"); err == nil {
			logs.RemoveIndex("idx_webhook_logs_retry")
			_ = app.Save(logs)
		}
		if webhooks, err := app.FindCollectionByNameOrId("webhook_configs"); err == nil {
			webhooks.RemoveIndex("idx_webhook_configs_user_active")
			_ = app.Save(webhooks)
		}
		if messages, err := app.FindCollectionByNameOrId("sms_messages"); err == nil {
			messages.RemoveIndex("idx_sms_messages_status_type_created")
			messages.RemoveIndex("idx_sms_messages_user")
			_ = app.Save(messages)
		}

		// Remove seeded config records
		seeds := []string{"message_retention_days", "webhook_log_retention_days", "incoming_retention_days"}
		for _, key := range seeds {
			record, _ := app.FindFirstRecordByFilter("system_config", "key = {:key}", dbx.Params{"key": key})
			if record != nil {
				_ = app.Delete(record)
			}
		}

		return nil
	})
}
