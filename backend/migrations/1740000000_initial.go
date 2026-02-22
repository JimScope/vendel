package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// ── 1. users (auth collection — already exists, just customize) ─
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			// Fallback: create if somehow missing
			users = core.NewAuthCollection("users")
		}
		users.Fields.Add(
			&core.TextField{Name: "full_name", Max: 255},
			&core.BoolField{Name: "is_superuser"},
		)
		users.AuthRule = ptrStr("")
		users.ListRule = ptrStr("id = @request.auth.id || @request.auth.is_superuser = true")
		users.ViewRule = ptrStr("id = @request.auth.id || @request.auth.is_superuser = true")
		users.UpdateRule = ptrStr("id = @request.auth.id")
		users.DeleteRule = ptrStr("id = @request.auth.id")
		users.CreateRule = nil // only via API or superuser
		users.OAuth2.Enabled = true
		users.PasswordAuth.Enabled = true

		if err := app.Save(users); err != nil {
			return err
		}

		// ── 2. user_plans ────────────────────────────────────────────
		plans := core.NewBaseCollection("user_plans")
		plans.Fields.Add(
			&core.TextField{Name: "name", Required: true, Max: 100},
			&core.NumberField{Name: "max_sms_per_month", Required: true},
			&core.NumberField{Name: "max_devices", Required: true},
			&core.NumberField{Name: "price"},
			&core.NumberField{Name: "price_yearly"},
			&core.BoolField{Name: "is_public"},
		)
		plans.ListRule = ptrStr("is_public = true || @request.auth.is_superuser = true")
		plans.ViewRule = ptrStr("is_public = true || @request.auth.is_superuser = true")
		plans.CreateRule = ptrStr("@request.auth.is_superuser = true")
		plans.UpdateRule = ptrStr("@request.auth.is_superuser = true")
		plans.DeleteRule = ptrStr("@request.auth.is_superuser = true")
		plans.AddIndex("idx_user_plans_name", true, "name", "")

		if err := app.Save(plans); err != nil {
			return err
		}

		// ── 3. user_quotas ───────────────────────────────────────────
		quotas := core.NewBaseCollection("user_quotas")
		quotas.Fields.Add(
			&core.NumberField{Name: "sms_sent_this_month"},
			&core.NumberField{Name: "devices_registered"},
			&core.AutodateField{Name: "last_reset_date", OnCreate: true},
			&core.RelationField{
				Name:         "user",
				CollectionId: users.Id,
				Required:     true,
				MaxSelect:    1,
			},
			&core.RelationField{
				Name:         "plan",
				CollectionId: plans.Id,
				Required:     true,
				MaxSelect:    1,
			},
		)
		quotas.ListRule = ptrStr("user = @request.auth.id || @request.auth.is_superuser = true")
		quotas.ViewRule = ptrStr("user = @request.auth.id || @request.auth.is_superuser = true")
		quotas.CreateRule = ptrStr("@request.auth.is_superuser = true")
		quotas.UpdateRule = ptrStr("@request.auth.is_superuser = true")
		quotas.DeleteRule = ptrStr("@request.auth.is_superuser = true")
		quotas.AddIndex("idx_user_quotas_user", true, "user", "")

		if err := app.Save(quotas); err != nil {
			return err
		}

		// ── 4. sms_devices ───────────────────────────────────────────
		devices := core.NewBaseCollection("sms_devices")
		devices.Fields.Add(
			&core.TextField{Name: "name", Required: true, Max: 255},
			&core.TextField{Name: "phone_number", Required: true, Max: 20},
			&core.TextField{Name: "api_key", Max: 255, Hidden: true},
			&core.TextField{Name: "fcm_token", Max: 500, Hidden: true},
			&core.RelationField{
				Name:         "user",
				CollectionId: users.Id,
				Required:     true,
				MaxSelect:    1,
			},
		)
		devices.ListRule = ptrStr("user = @request.auth.id || @request.auth.is_superuser = true")
		devices.ViewRule = ptrStr("user = @request.auth.id || @request.auth.is_superuser = true")
		devices.CreateRule = ptrStr("user = @request.auth.id")
		devices.UpdateRule = ptrStr("user = @request.auth.id")
		devices.DeleteRule = ptrStr("user = @request.auth.id")
		devices.AddIndex("idx_sms_devices_api_key", true, "api_key", "")

		if err := app.Save(devices); err != nil {
			return err
		}

		// ── 5. sms_messages ──────────────────────────────────────────
		messages := core.NewBaseCollection("sms_messages")
		messages.Fields.Add(
			&core.TextField{Name: "to", Max: 20},
			&core.TextField{Name: "from_number", Max: 20},
			&core.TextField{Name: "body", Required: true, Max: 1600},
			&core.SelectField{
				Name:      "status",
				Values:    []string{"pending", "assigned", "sending", "sent", "delivered", "failed", "received"},
				MaxSelect: 1,
			},
			&core.SelectField{
				Name:      "message_type",
				Values:    []string{"outgoing", "incoming"},
				MaxSelect: 1,
			},
			&core.TextField{Name: "batch_id", Max: 36},
			&core.RelationField{
				Name:         "device",
				CollectionId: devices.Id,
				MaxSelect:    1,
			},
			&core.RelationField{
				Name:         "user",
				CollectionId: users.Id,
				Required:     true,
				MaxSelect:    1,
			},
			&core.BoolField{Name: "webhook_sent"},
			&core.TextField{Name: "error_message", Max: 500},
			&core.DateField{Name: "sent_at"},
			&core.DateField{Name: "delivered_at"},
		)
		messages.ListRule = ptrStr("user = @request.auth.id || @request.auth.is_superuser = true")
		messages.ViewRule = ptrStr("user = @request.auth.id || @request.auth.is_superuser = true")
		messages.CreateRule = ptrStr("@request.auth.is_superuser = true")
		messages.UpdateRule = ptrStr("@request.auth.is_superuser = true")
		messages.DeleteRule = ptrStr("@request.auth.is_superuser = true")
		messages.AddIndex("idx_sms_messages_batch_id", false, "batch_id", "")

		if err := app.Save(messages); err != nil {
			return err
		}

		// ── 6. webhook_configs ───────────────────────────────────────
		webhooks := core.NewBaseCollection("webhook_configs")
		webhooks.Fields.Add(
			&core.URLField{Name: "url", Required: true},
			&core.TextField{Name: "secret_key", Max: 255, Hidden: true},
			&core.JSONField{Name: "events", MaxSize: 500},
			&core.BoolField{Name: "active"},
			&core.RelationField{
				Name:         "user",
				CollectionId: users.Id,
				Required:     true,
				MaxSelect:    1,
			},
		)
		webhooks.ListRule = ptrStr("user = @request.auth.id || @request.auth.is_superuser = true")
		webhooks.ViewRule = ptrStr("user = @request.auth.id || @request.auth.is_superuser = true")
		webhooks.CreateRule = ptrStr("user = @request.auth.id")
		webhooks.UpdateRule = ptrStr("user = @request.auth.id")
		webhooks.DeleteRule = ptrStr("user = @request.auth.id")

		if err := app.Save(webhooks); err != nil {
			return err
		}

		// ── 7. api_keys ─────────────────────────────────────────────
		apiKeys := core.NewBaseCollection("api_keys")
		apiKeys.Fields.Add(
			&core.TextField{Name: "name", Required: true, Max: 255},
			&core.TextField{Name: "key", Max: 255, Hidden: true},
			&core.BoolField{Name: "is_active"},
			&core.RelationField{
				Name:         "user",
				CollectionId: users.Id,
				Required:     true,
				MaxSelect:    1,
			},
			&core.DateField{Name: "last_used_at"},
		)
		apiKeys.ListRule = ptrStr("user = @request.auth.id || @request.auth.is_superuser = true")
		apiKeys.ViewRule = ptrStr("user = @request.auth.id || @request.auth.is_superuser = true")
		apiKeys.CreateRule = ptrStr("user = @request.auth.id")
		apiKeys.UpdateRule = ptrStr("user = @request.auth.id")
		apiKeys.DeleteRule = ptrStr("user = @request.auth.id")
		apiKeys.AddIndex("idx_api_keys_key", true, "key", "")

		if err := app.Save(apiKeys); err != nil {
			return err
		}

		// ── 8. subscriptions ─────────────────────────────────────────
		subs := core.NewBaseCollection("subscriptions")
		subs.Fields.Add(
			&core.RelationField{
				Name:         "user",
				CollectionId: users.Id,
				Required:     true,
				MaxSelect:    1,
			},
			&core.RelationField{
				Name:         "plan",
				CollectionId: plans.Id,
				Required:     true,
				MaxSelect:    1,
			},
			&core.SelectField{
				Name:      "billing_cycle",
				Values:    []string{"monthly", "yearly"},
				MaxSelect: 1,
			},
			&core.SelectField{
				Name:      "status",
				Values:    []string{"pending", "active", "past_due", "canceled", "expired"},
				MaxSelect: 1,
			},
			&core.SelectField{
				Name:      "payment_method",
				Values:    []string{"invoice", "authorized"},
				MaxSelect: 1,
			},
			&core.DateField{Name: "current_period_start"},
			&core.DateField{Name: "current_period_end"},
			&core.TextField{Name: "provider_user_uuid", Max: 255},
			&core.BoolField{Name: "cancel_at_period_end"},
			&core.DateField{Name: "canceled_at"},
		)
		subs.ListRule = ptrStr("user = @request.auth.id || @request.auth.is_superuser = true")
		subs.ViewRule = ptrStr("user = @request.auth.id || @request.auth.is_superuser = true")
		subs.CreateRule = ptrStr("@request.auth.is_superuser = true")
		subs.UpdateRule = ptrStr("@request.auth.is_superuser = true")
		subs.DeleteRule = ptrStr("@request.auth.is_superuser = true")
		subs.AddIndex("idx_subscriptions_user", true, "user", "")

		if err := app.Save(subs); err != nil {
			return err
		}

		// ── 9. payments ──────────────────────────────────────────────
		payments := core.NewBaseCollection("payments")
		payments.Fields.Add(
			&core.RelationField{
				Name:         "subscription",
				CollectionId: subs.Id,
				Required:     true,
				MaxSelect:    1,
			},
			&core.NumberField{Name: "amount", Required: true},
			&core.TextField{Name: "currency", Max: 10},
			&core.SelectField{
				Name:      "status",
				Values:    []string{"pending", "completed", "failed", "refunded"},
				MaxSelect: 1,
			},
			&core.TextField{Name: "provider", Max: 50},
			&core.TextField{Name: "provider_transaction_id", Max: 255},
			&core.TextField{Name: "provider_invoice_id", Max: 255},
			&core.URLField{Name: "provider_invoice_url"},
			&core.DateField{Name: "period_start"},
			&core.DateField{Name: "period_end"},
			&core.DateField{Name: "paid_at"},
		)
		payments.ListRule = ptrStr("subscription.user = @request.auth.id || @request.auth.is_superuser = true")
		payments.ViewRule = ptrStr("subscription.user = @request.auth.id || @request.auth.is_superuser = true")
		payments.CreateRule = ptrStr("@request.auth.is_superuser = true")
		payments.UpdateRule = ptrStr("@request.auth.is_superuser = true")
		payments.DeleteRule = ptrStr("@request.auth.is_superuser = true")
		payments.AddIndex("idx_payments_provider_tx", true, "provider_transaction_id", "provider_transaction_id != ''")

		if err := app.Save(payments); err != nil {
			return err
		}

		// ── 10. system_config ────────────────────────────────────────
		sysConfig := core.NewBaseCollection("system_config")
		sysConfig.Fields.Add(
			&core.TextField{Name: "key", Required: true, Max: 100},
			&core.TextField{Name: "value", Max: 1000},
			&core.TextField{Name: "description", Max: 500},
		)
		sysConfig.ListRule = ptrStr("@request.auth.is_superuser = true")
		sysConfig.ViewRule = ptrStr("@request.auth.is_superuser = true")
		sysConfig.CreateRule = ptrStr("@request.auth.is_superuser = true")
		sysConfig.UpdateRule = ptrStr("@request.auth.is_superuser = true")
		sysConfig.DeleteRule = ptrStr("@request.auth.is_superuser = true")
		sysConfig.AddIndex("idx_system_config_key", true, "key", "")

		if err := app.Save(sysConfig); err != nil {
			return err
		}

		// ── Seed data ────────────────────────────────────────────────

		// Free plan
		freePlan := core.NewRecord(plans)
		freePlan.Set("name", "Free")
		freePlan.Set("max_sms_per_month", 50)
		freePlan.Set("max_devices", 1)
		freePlan.Set("price", 0)
		freePlan.Set("price_yearly", 0)
		freePlan.Set("is_public", true)
		if err := app.Save(freePlan); err != nil {
			return err
		}

		// System config defaults
		configs := []struct{ key, value, desc string }{
			{"app_name", "Ender", "Application display name"},
			{"support_email", "support@ender.app", "Support email address"},
			{"maintenance_mode", "false", "Enable/disable maintenance mode"},
			{"default_payment_method", "invoice", "Default payment method (invoice or authorized)"},
			{"webhook_timeout", "10", "Webhook delivery timeout in seconds"},
		}
		for _, c := range configs {
			rec := core.NewRecord(sysConfig)
			rec.Set("key", c.key)
			rec.Set("value", c.value)
			rec.Set("description", c.desc)
			if err := app.Save(rec); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		// Rollback: delete collections in reverse order
		names := []string{
			"system_config", "payments", "subscriptions", "api_keys",
			"webhook_configs", "sms_messages", "sms_devices",
			"user_quotas", "user_plans", "users",
		}
		for _, name := range names {
			col, err := app.FindCollectionByNameOrId(name)
			if err == nil {
				if err := app.Delete(col); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func ptrStr(s string) *string {
	return &s
}
