package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		subs, err := app.FindCollectionByNameOrId("subscriptions")
		if err != nil {
			return err
		}

		subs.Fields.Add(&core.TextField{
			Name:    "provider",
			Max:     50,
		})

		return app.Save(subs)
	}, func(app core.App) error {
		subs, err := app.FindCollectionByNameOrId("subscriptions")
		if err != nil {
			return err
		}

		subs.Fields.RemoveByName("provider")
		return app.Save(subs)
	})
}
