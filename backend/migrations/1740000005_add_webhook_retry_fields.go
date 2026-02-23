package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		logs, err := app.FindCollectionByNameOrId("webhook_delivery_logs")
		if err != nil {
			return err
		}

		logs.Fields.Add(&core.NumberField{
			Name: "retry_count",
			Min:  floatPtr(0),
		})

		logs.Fields.Add(&core.DateField{
			Name: "next_retry_at",
		})

		logs.Fields.Add(&core.RelationField{
			Name:         "original_log",
			CollectionId: logs.Id,
			MaxSelect:    1,
		})

		return app.Save(logs)
	}, func(app core.App) error {
		logs, err := app.FindCollectionByNameOrId("webhook_delivery_logs")
		if err != nil {
			return err
		}

		logs.Fields.RemoveByName("retry_count")
		logs.Fields.RemoveByName("next_retry_at")
		logs.Fields.RemoveByName("original_log")
		return app.Save(logs)
	})
}
