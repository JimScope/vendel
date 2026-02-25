package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		devices, err := app.FindCollectionByNameOrId("sms_devices")
		if err != nil {
			return err
		}

		// ── sms_templates ───────────────────────────────────────────
		templates := core.NewBaseCollection("sms_templates")
		templates.Fields.Add(
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
			&core.TextField{Name: "name", Required: true, Max: 100},
			&core.TextField{Name: "body", Required: true, Max: 1600},
			&core.RelationField{
				Name:         "user",
				CollectionId: users.Id,
				Required:     true,
				MaxSelect:    1,
			},
		)
		templates.ListRule = ptrStr("user = @request.auth.id")
		templates.ViewRule = ptrStr("user = @request.auth.id")
		templates.CreateRule = ptrStr("user = @request.auth.id")
		templates.UpdateRule = ptrStr("user = @request.auth.id")
		templates.DeleteRule = ptrStr("user = @request.auth.id")
		templates.AddIndex("idx_sms_templates_user", false, "user", "")

		if err := app.Save(templates); err != nil {
			return err
		}

		// ── scheduled_sms ───────────────────────────────────────────
		scheduled := core.NewBaseCollection("scheduled_sms")
		scheduled.Fields.Add(
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
			&core.TextField{Name: "name", Required: true, Max: 100},
			&core.JSONField{Name: "recipients", Required: true, MaxSize: 5000},
			&core.TextField{Name: "body", Required: true, Max: 1600},
			&core.RelationField{
				Name:         "device_id",
				CollectionId: devices.Id,
				MaxSelect:    1,
			},
			&core.RelationField{
				Name:         "user",
				CollectionId: users.Id,
				Required:     true,
				MaxSelect:    1,
			},
			&core.SelectField{
				Name:      "schedule_type",
				Values:    []string{"one_time", "recurring"},
				MaxSelect: 1,
				Required:  true,
			},
			&core.DateField{Name: "scheduled_at"},
			&core.TextField{Name: "cron_expression", Max: 100},
			&core.TextField{Name: "timezone", Max: 50},
			&core.DateField{Name: "next_run_at"},
			&core.DateField{Name: "last_run_at"},
			&core.SelectField{
				Name:      "status",
				Values:    []string{"active", "paused", "completed"},
				MaxSelect: 1,
				Required:  true,
			},
		)
		scheduled.ListRule = ptrStr("user = @request.auth.id")
		scheduled.ViewRule = ptrStr("user = @request.auth.id")
		scheduled.CreateRule = ptrStr("user = @request.auth.id")
		scheduled.UpdateRule = ptrStr("user = @request.auth.id")
		scheduled.DeleteRule = ptrStr("user = @request.auth.id")
		scheduled.AddIndex("idx_scheduled_sms_next_run", false, "next_run_at, status", "")
		scheduled.AddIndex("idx_scheduled_sms_user", false, "user", "")

		return app.Save(scheduled)
	}, func(app core.App) error {
		// Rollback: delete in reverse order
		for _, name := range []string{"scheduled_sms", "sms_templates"} {
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
