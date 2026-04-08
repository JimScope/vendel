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

		// ── promo_codes ─────────────────────────────────────────────
		promos := core.NewBaseCollection("promo_codes")
		promos.Fields.Add(
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
			&core.TextField{
				Name:     "code",
				Required: true,
				Min:      3,
				Max:      50,
			},
			&core.NumberField{
				Name:     "amount",
				Required: true,
				Min:      ptrFloat(0.01),
			},
			&core.NumberField{
				Name: "max_redemptions",
				Min:  ptrFloat(0),
			},
			&core.NumberField{
				Name: "times_redeemed",
				Min:  ptrFloat(0),
			},
			&core.DateField{Name: "expires_at"},
			&core.BoolField{Name: "active"},
		)
		// Superuser-only management
		promos.ListRule = nil
		promos.ViewRule = nil
		promos.CreateRule = nil
		promos.UpdateRule = nil
		promos.DeleteRule = nil
		promos.AddIndex("idx_promo_codes_code", true, "code", "")

		if err := app.Save(promos); err != nil {
			return err
		}

		// ── promo_redemptions ───────────────────────────────────────
		redemptions := core.NewBaseCollection("promo_redemptions")
		redemptions.Fields.Add(
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.RelationField{
				Name:         "user",
				CollectionId: users.Id,
				Required:     true,
				MaxSelect:    1,
			},
			&core.RelationField{
				Name:         "promo_code",
				CollectionId: promos.Id,
				Required:     true,
				MaxSelect:    1,
			},
			&core.NumberField{
				Name: "amount_credited",
				Min:  ptrFloat(0),
			},
		)
		redemptions.ListRule = ptrStr("user = @request.auth.id")
		redemptions.ViewRule = ptrStr("user = @request.auth.id")
		redemptions.CreateRule = nil
		redemptions.UpdateRule = nil
		redemptions.DeleteRule = nil
		redemptions.AddIndex("idx_promo_redemptions_user_code", true, "user, promo_code", "")

		return app.Save(redemptions)
	}, func(app core.App) error {
		// Rollback
		if c, err := app.FindCollectionByNameOrId("promo_redemptions"); err == nil {
			_ = app.Delete(c)
		}
		if c, err := app.FindCollectionByNameOrId("promo_codes"); err == nil {
			_ = app.Delete(c)
		}
		return nil
	})
}
