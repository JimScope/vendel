package services

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// ExportUserData returns all data belonging to a user in a structured format.
// Hidden fields (api_key, fcm_token, secret_key, key, password) are excluded.
func ExportUserData(app core.App, userId string) (map[string]any, error) {
	user, err := app.FindRecordById("users", userId)
	if err != nil {
		return nil, err
	}

	export := map[string]any{
		"exported_at": "",
		"user": map[string]any{
			"id":         user.Id,
			"email":      user.Email(),
			"full_name":  user.GetString("full_name"),
			"verified":   user.Verified(),
			"created":    user.GetString("created"),
			"updated":    user.GetString("updated"),
		},
	}

	// Direct user-owned collections (field: "user")
	directCollections := []struct {
		key    string
		name   string
		fields []string
	}{
		{"devices", "sms_devices", []string{"id", "created", "name", "phone_number", "device_type"}},
		{"messages", "sms_messages", []string{"id", "created", "to", "from_number", "body", "status", "message_type", "batch_id", "error_message", "sent_at", "delivered_at"}},
		{"webhooks", "webhook_configs", []string{"id", "created", "url", "events", "active", "include_body"}},
		{"api_keys", "api_keys", []string{"id", "created", "name", "is_active", "expires_at", "last_used_at"}},
		{"templates", "sms_templates", []string{"id", "created", "name", "body"}},
		{"scheduled_sms", "scheduled_sms", []string{"id", "created", "name", "recipients", "body", "schedule_type", "scheduled_at", "cron_expression", "timezone", "status"}},
	}

	for _, col := range directCollections {
		records, err := app.FindRecordsByFilter(col.name, "user = {:uid}", "-created", 0, 0, dbx.Params{"uid": userId})
		if err != nil {
			export[col.key] = []any{}
			continue
		}
		export[col.key] = pickFields(records, col.fields)
	}

	// Quota (single record)
	quota, err := app.FindFirstRecordByFilter("user_quotas", "user = {:uid}", dbx.Params{"uid": userId})
	if err == nil {
		plan, _ := app.FindRecordById("user_plans", quota.GetString("plan"))
		planName := ""
		if plan != nil {
			planName = plan.GetString("name")
		}
		export["quota"] = map[string]any{
			"plan":               planName,
			"sms_sent_this_month": quota.GetInt("sms_sent_this_month"),
			"devices_registered": quota.GetInt("devices_registered"),
			"last_reset_date":    quota.GetString("last_reset_date"),
		}
	}

	// Subscriptions + payments (indirect relation)
	subs, err := app.FindRecordsByFilter("subscriptions", "user = {:uid}", "-created", 0, 0, dbx.Params{"uid": userId})
	if err != nil {
		export["subscriptions"] = []any{}
		export["payments"] = []any{}
	} else {
		subFields := []string{"id", "created", "billing_cycle", "status", "payment_method", "current_period_start", "current_period_end", "cancel_at_period_end", "canceled_at"}
		export["subscriptions"] = pickFields(subs, subFields)

		// Payments via subscriptions
		var allPayments []map[string]any
		for _, sub := range subs {
			payments, err := app.FindRecordsByFilter("payments", "subscription = {:sid}", "-created", 0, 0, dbx.Params{"sid": sub.Id})
			if err != nil {
				continue
			}
			payFields := []string{"id", "created", "amount", "currency", "status", "provider", "period_start", "period_end", "paid_at"}
			for _, p := range pickFields(payments, payFields) {
				allPayments = append(allPayments, p)
			}
		}
		if allPayments == nil {
			allPayments = []map[string]any{}
		}
		export["payments"] = allPayments
	}

	return export, nil
}

// pickFields extracts only the specified fields from records.
func pickFields(records []*core.Record, fields []string) []map[string]any {
	result := make([]map[string]any, 0, len(records))
	for _, r := range records {
		item := make(map[string]any, len(fields))
		for _, f := range fields {
			item[f] = r.Get(f)
		}
		result = append(result, item)
	}
	return result
}
