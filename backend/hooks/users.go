package hooks

import (
	"log/slog"
	"vendel/middleware"
	"vendel/services"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterUserHooks registers hooks for user lifecycle events:
// app_name sync, default quota creation, and cascade deletion.
func RegisterUserHooks(app *pocketbase.PocketBase) {
	// System config: sync app_name to PocketBase settings + invalidate caches
	app.OnRecordAfterUpdateSuccess("system_config").BindFunc(func(e *core.RecordEvent) error {
		key := e.Record.GetString("key")

		if key == "app_name" {
			settings := e.App.Settings()
			settings.Meta.AppName = e.Record.GetString("value")
			settings.Meta.SenderName = e.Record.GetString("value")
			if err := e.App.Save(settings); err != nil {
				e.App.Logger().Warn("could not sync app_name to PocketBase settings", slog.Any("error", err))
			}
		}

		if key == "maintenance_mode" {
			middleware.InvalidateMaintenanceCache()
		}

		return e.Next()
	})

	// Users: create default quota on registration
	app.OnRecordCreate("users").BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		return services.CreateDefaultQuota(e.App, e.Record.Id)
	})

	// Users: cascade delete all related data (GDPR Art. 17)
	// IMPORTANT: cascade MUST run before e.Next() because related
	// collections have Required:true relations to users — PocketBase
	// blocks the user delete if any referencing records still exist.
	app.OnRecordDelete("users").BindFunc(func(e *core.RecordEvent) error {
		userId := e.Record.Id

		// Delete records that reference users directly
		collections := []string{
			"sms_messages", "sms_devices", "webhook_configs",
			"api_keys", "user_quotas", "sms_templates", "scheduled_sms",
		}
		for _, col := range collections {
			records, err := e.App.FindRecordsByFilter(col, "user = {:uid}", "", 0, 0, dbx.Params{"uid": userId})
			if err != nil {
				continue
			}
			for _, r := range records {
				if col == "webhook_configs" {
					logs, _ := e.App.FindRecordsByFilter("webhook_delivery_logs", "webhook = {:wid}", "", 0, 0, dbx.Params{"wid": r.Id})
					for _, l := range logs {
						_ = e.App.Delete(l)
					}
				}
				_ = e.App.Delete(r)
			}
		}

		// Delete subscriptions and their payments (payments link via subscription, not user)
		subscriptions, err := e.App.FindRecordsByFilter("subscriptions", "user = {:uid}", "", 0, 0, dbx.Params{"uid": userId})
		if err == nil {
			for _, sub := range subscriptions {
				payments, _ := e.App.FindRecordsByFilter("payments", "subscription = {:sid}", "", 0, 0, dbx.Params{"sid": sub.Id})
				for _, p := range payments {
					_ = e.App.Delete(p)
				}
				_ = e.App.Delete(sub)
			}
		}

		e.App.Logger().Info("Cascade deleted user data", slog.String("user", userId))
		return e.Next()
	})
}
