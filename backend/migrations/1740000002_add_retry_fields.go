package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		messages, err := app.FindCollectionByNameOrId("sms_messages")
		if err != nil {
			return err
		}

		messages.Fields.Add(&core.NumberField{
			Name: "retry_count",
			Min:  floatPtr(0),
		})

		messages.Fields.Add(&core.DateField{
			Name: "last_retry_at",
		})

		return app.Save(messages)
	}, func(app core.App) error {
		messages, err := app.FindCollectionByNameOrId("sms_messages")
		if err != nil {
			return err
		}

		messages.Fields.RemoveByName("retry_count")
		messages.Fields.RemoveByName("last_retry_at")
		return app.Save(messages)
	})
}

func floatPtr(f float64) *float64 {
	return &f
}
