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

		// ── contact_groups ──────────────────────────────────────────
		groups := core.NewBaseCollection("contact_groups")
		groups.Fields.Add(
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
			&core.RelationField{
				Name:         "user",
				CollectionId: users.Id,
				Required:     true,
				MaxSelect:    1,
			},
			&core.TextField{
				Name:     "name",
				Required: true,
				Max:      100,
			},
		)
		groups.ListRule = ptrStr("user = @request.auth.id")
		groups.ViewRule = ptrStr("user = @request.auth.id")
		groups.CreateRule = ptrStr("@request.auth.id != ''")
		groups.UpdateRule = ptrStr("user = @request.auth.id")
		groups.DeleteRule = ptrStr("user = @request.auth.id")
		groups.AddIndex("idx_contact_groups_user_name", true, "user, name", "")

		if err := app.Save(groups); err != nil {
			return err
		}

		// ── contacts ────────────────────────────────────────────────
		contacts := core.NewBaseCollection("contacts")
		contacts.Fields.Add(
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
			&core.RelationField{
				Name:         "user",
				CollectionId: users.Id,
				Required:     true,
				MaxSelect:    1,
			},
			&core.TextField{
				Name:     "name",
				Required: true,
				Max:      200,
			},
			&core.TextField{
				Name:     "phone_number",
				Required: true,
				Max:      20,
				Pattern:  `^\+[1-9]\d{1,14}$`,
			},
			&core.RelationField{
				Name:         "groups",
				CollectionId: groups.Id,
				MaxSelect:    50,
			},
			&core.TextField{
				Name: "notes",
				Max:  500,
			},
		)
		contacts.ListRule = ptrStr("user = @request.auth.id")
		contacts.ViewRule = ptrStr("user = @request.auth.id")
		contacts.CreateRule = ptrStr("@request.auth.id != ''")
		contacts.UpdateRule = ptrStr("user = @request.auth.id")
		contacts.DeleteRule = ptrStr("user = @request.auth.id")
		contacts.AddIndex("idx_contacts_user_phone", true, "user, phone_number", "")
		contacts.AddIndex("idx_contacts_user", false, "user", "")

		return app.Save(contacts)
	}, func(app core.App) error {
		// Rollback
		if col, err := app.FindCollectionByNameOrId("contacts"); err == nil {
			_ = app.Delete(col)
		}
		if col, err := app.FindCollectionByNameOrId("contact_groups"); err == nil {
			_ = app.Delete(col)
		}
		return nil
	})
}
