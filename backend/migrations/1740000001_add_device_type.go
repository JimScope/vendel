package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		devices, err := app.FindCollectionByNameOrId("sms_devices")
		if err != nil {
			return err
		}

		// Add device_type select field
		devices.Fields.Add(&core.SelectField{
			Name:      "device_type",
			Values:    []string{"android", "modem"},
			MaxSelect: 1,
		})

		if err := app.Save(devices); err != nil {
			return err
		}

		// Backfill existing devices as "android"
		records, err := app.FindRecordsByFilter(
			"sms_devices",
			"device_type = ''",
			"", 0, 0,
		)
		if err != nil {
			return nil // no records to backfill
		}
		for _, r := range records {
			r.Set("device_type", "android")
			if err := app.Save(r); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		devices, err := app.FindCollectionByNameOrId("sms_devices")
		if err != nil {
			return err
		}

		devices.Fields.RemoveByName("device_type")
		return app.Save(devices)
	})
}
