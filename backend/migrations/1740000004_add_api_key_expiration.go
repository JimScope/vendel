package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		apiKeys, err := app.FindCollectionByNameOrId("api_keys")
		if err != nil {
			return err
		}

		apiKeys.Fields.Add(&core.DateField{
			Name: "expires_at",
		})

		return app.Save(apiKeys)
	}, func(app core.App) error {
		apiKeys, err := app.FindCollectionByNameOrId("api_keys")
		if err != nil {
			return err
		}

		apiKeys.Fields.RemoveByName("expires_at")
		return app.Save(apiKeys)
	})
}
