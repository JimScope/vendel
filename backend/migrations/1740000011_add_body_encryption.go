package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Add body_hash blind index for deduplication on encrypted bodies
		messages, err := app.FindCollectionByNameOrId("sms_messages")
		if err != nil {
			return err
		}
		messages.Fields.Add(
			&core.TextField{Name: "body_hash", Max: 64},
		)
		// Increase body max to accommodate AES-GCM encrypted + base64 encoded values
		if bodyField, ok := messages.Fields.GetByName("body").(*core.TextField); ok {
			bodyField.Max = 3200
		}
		messages.AddIndex("idx_sms_messages_body_hash", false, "body_hash", "")
		if err := app.Save(messages); err != nil {
			return err
		}

		// Increase body max for templates
		templates, err := app.FindCollectionByNameOrId("sms_templates")
		if err != nil {
			return err
		}
		if bodyField, ok := templates.Fields.GetByName("body").(*core.TextField); ok {
			bodyField.Max = 3200
		}
		if err := app.Save(templates); err != nil {
			return err
		}

		// Increase body max for scheduled_sms
		scheduled, err := app.FindCollectionByNameOrId("scheduled_sms")
		if err != nil {
			return err
		}
		if bodyField, ok := scheduled.Fields.GetByName("body").(*core.TextField); ok {
			bodyField.Max = 3200
		}
		return app.Save(scheduled)
	}, func(app core.App) error {
		messages, err := app.FindCollectionByNameOrId("sms_messages")
		if err != nil {
			return err
		}
		messages.Fields.RemoveByName("body_hash")
		messages.RemoveIndex("idx_sms_messages_body_hash")
		if bodyField, ok := messages.Fields.GetByName("body").(*core.TextField); ok {
			bodyField.Max = 1600
		}
		if err := app.Save(messages); err != nil {
			return err
		}

		templates, err := app.FindCollectionByNameOrId("sms_templates")
		if err != nil {
			return err
		}
		if bodyField, ok := templates.Fields.GetByName("body").(*core.TextField); ok {
			bodyField.Max = 1600
		}
		if err := app.Save(templates); err != nil {
			return err
		}

		scheduled, err := app.FindCollectionByNameOrId("scheduled_sms")
		if err != nil {
			return err
		}
		if bodyField, ok := scheduled.Fields.GetByName("body").(*core.TextField); ok {
			bodyField.Max = 1600
		}
		return app.Save(scheduled)
	})
}
