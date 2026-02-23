package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		webhooks, err := app.FindCollectionByNameOrId("webhook_configs")
		if err != nil {
			return err
		}

		logs := core.NewBaseCollection("webhook_delivery_logs")
		logs.Fields.Add(
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.RelationField{
				Name:         "webhook",
				CollectionId: webhooks.Id,
				Required:     true,
				MaxSelect:    1,
			},
			&core.TextField{Name: "event", Required: true, Max: 50},
			&core.URLField{Name: "url", Required: true},
			&core.JSONField{Name: "request_body", MaxSize: 10000},
			&core.NumberField{Name: "response_status"},
			&core.TextField{Name: "response_body", Max: 2000},
			&core.SelectField{
				Name:      "delivery_status",
				Values:    []string{"success", "failed"},
				MaxSelect: 1,
				Required:  true,
			},
			&core.TextField{Name: "error_message", Max: 500},
			&core.NumberField{Name: "duration_ms"},
		)
		logs.ListRule = ptrStr("webhook.user = @request.auth.id")
		logs.ViewRule = ptrStr("webhook.user = @request.auth.id")
		logs.CreateRule = nil // service only
		logs.UpdateRule = nil
		logs.DeleteRule = nil
		logs.AddIndex("idx_webhook_logs_webhook_created", false, "webhook, created", "")

		return app.Save(logs)
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("webhook_delivery_logs")
		if err == nil {
			return app.Delete(col)
		}
		return nil
	})
}
