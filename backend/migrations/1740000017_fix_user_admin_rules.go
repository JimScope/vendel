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

		// Allow superusers to list, view, and delete any user
		users.ListRule = ptrStr("id = @request.auth.id || @request.auth.is_superuser = true")
		users.ViewRule = ptrStr("id = @request.auth.id || @request.auth.is_superuser = true")
		users.DeleteRule = ptrStr("id = @request.auth.id || @request.auth.is_superuser = true")

		return app.Save(users)
	}, func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		users.ListRule = ptrStr("id = @request.auth.id")
		users.ViewRule = ptrStr("id = @request.auth.id")
		users.DeleteRule = ptrStr("id = @request.auth.id")

		return app.Save(users)
	})
}
