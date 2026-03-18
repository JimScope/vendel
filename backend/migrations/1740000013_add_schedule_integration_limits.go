package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		plans, err := app.FindCollectionByNameOrId("user_plans")
		if err != nil {
			return err
		}

		plans.Fields.Add(
			&core.NumberField{Name: "max_scheduled_sms"},
			&core.NumberField{Name: "max_integrations"},
		)
		if err := app.Save(plans); err != nil {
			return err
		}

		// Update existing Free plan with defaults
		freePlan, err := app.FindFirstRecordByFilter("user_plans", "name = 'Free'")
		if err == nil && freePlan != nil {
			freePlan.Set("max_scheduled_sms", 1)
			freePlan.Set("max_integrations", 1)
			if err := app.Save(freePlan); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		plans, err := app.FindCollectionByNameOrId("user_plans")
		if err != nil {
			return err
		}

		plans.Fields.RemoveByName("max_scheduled_sms")
		plans.Fields.RemoveByName("max_integrations")
		return app.Save(plans)
	})
}
