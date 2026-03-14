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

		webhooks.Fields.Add(&core.BoolField{
			Name: "include_body",
		})

		return app.Save(webhooks)
	}, func(app core.App) error {
		webhooks, err := app.FindCollectionByNameOrId("webhook_configs")
		if err != nil {
			return err
		}

		webhooks.Fields.RemoveByName("include_body")

		return app.Save(webhooks)
	})
}
