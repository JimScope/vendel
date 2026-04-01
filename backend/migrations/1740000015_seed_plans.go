package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("user_plans")
		if err != nil {
			return err
		}

		plans := []struct {
			name             string
			maxSMS           int
			maxDevices       int
			maxScheduled     int
			maxIntegrations  int
			price            float64
			priceYearly      float64
		}{
			{"Hobby", 500, 3, 10, 5, 0.99, 9.99},
			{"Pro", 5000, 10, 100, 20, 2.99, 29.99},
		}

		for _, p := range plans {
			// Skip if already exists
			existing, _ := app.FindFirstRecordByFilter("user_plans", "name = {:name}", map[string]any{"name": p.name})
			if existing != nil {
				continue
			}

			record := core.NewRecord(collection)
			record.Set("name", p.name)
			record.Set("max_sms_per_month", p.maxSMS)
			record.Set("max_devices", p.maxDevices)
			record.Set("max_scheduled_sms", p.maxScheduled)
			record.Set("max_integrations", p.maxIntegrations)
			record.Set("price", p.price)
			record.Set("price_yearly", p.priceYearly)
			record.Set("is_public", true)

			if err := app.Save(record); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		for _, name := range []string{"Hobby", "Pro"} {
			record, err := app.FindFirstRecordByFilter("user_plans", "name = {:name}", map[string]any{"name": name})
			if err == nil && record != nil {
				_ = app.Delete(record)
			}
		}
		return nil
	})
}
