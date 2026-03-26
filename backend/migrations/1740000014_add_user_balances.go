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

		// ── user_balances ────────────────────────────────────────────
		bal := core.NewBaseCollection("user_balances")
		bal.Fields.Add(
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
			&core.RelationField{
				Name:         "user",
				CollectionId: users.Id,
				Required:     true,
				MaxSelect:    1,
			},
			&core.NumberField{
				Name: "balance",
				Min:  ptrFloat(0),
			},
			&core.TextField{
				Name:    "currency",
				Max:     10,
				Pattern: "^[A-Z]{3,10}$",
			},
			&core.TextField{Name: "wallet_address", Max: 255},
			&core.TextField{Name: "wallet_id", Max: 255},
		)
		bal.ListRule = ptrStr("user = @request.auth.id")
		bal.ViewRule = ptrStr("user = @request.auth.id")
		bal.CreateRule = nil
		bal.UpdateRule = nil
		bal.DeleteRule = nil
		bal.AddIndex("idx_user_balances_user", true, "user", "")

		if err := app.Save(bal); err != nil {
			return err
		}

		// ── Add "balance" to subscription payment_method options ─────
		subs, err := app.FindCollectionByNameOrId("subscriptions")
		if err != nil {
			return err
		}

		pmField := subs.Fields.GetByName("payment_method").(*core.SelectField)
		pmField.Values = append(pmField.Values, "balance")

		return app.Save(subs)
	}, func(app core.App) error {
		// Rollback: remove "balance" from payment_method, drop user_balances
		subs, err := app.FindCollectionByNameOrId("subscriptions")
		if err == nil {
			pmField := subs.Fields.GetByName("payment_method").(*core.SelectField)
			filtered := make([]string, 0, len(pmField.Values))
			for _, v := range pmField.Values {
				if v != "balance" {
					filtered = append(filtered, v)
				}
			}
			pmField.Values = filtered
			_ = app.Save(subs)
		}

		bal, err := app.FindCollectionByNameOrId("user_balances")
		if err == nil {
			return app.Delete(bal)
		}
		return nil
	})
}

func ptrFloat(f float64) *float64 {
	return &f
}
