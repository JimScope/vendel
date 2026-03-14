package hooks

import (
	"vendel/services"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterDeviceHooks registers hooks for SMS device lifecycle:
// quota check, API key generation, default device_type, and count management.
func RegisterDeviceHooks(app *pocketbase.PocketBase) {
	// SMS Devices: check quota + generate API key + default device_type on create
	app.OnRecordCreate("sms_devices").BindFunc(func(e *core.RecordEvent) error {
		userId := e.Record.GetString("user")
		if err := services.CheckDeviceQuota(e.App, userId); err != nil {
			return err
		}

		// Increment device count before persisting so quota stays consistent
		if err := services.IncrementDeviceCount(e.App, userId); err != nil {
			return err
		}

		// Generate secure API key
		e.Record.Set("api_key", services.GenerateSecureKey("dk_", 32))

		// Default device_type to "android" if not set
		if e.Record.GetString("device_type") == "" {
			e.Record.Set("device_type", "android")
		}

		// Unhide so the key is returned in the create response (only shown once)
		e.Record.Unhide("api_key")

		if err := e.Next(); err != nil {
			// Rollback the increment if record creation fails
			_ = services.DecrementDeviceCount(e.App, userId)
			return err
		}

		return nil
	})

	// SMS Devices: decrement count on delete
	app.OnRecordDelete("sms_devices").BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		userId := e.Record.GetString("user")
		return services.DecrementDeviceCount(e.App, userId)
	})
}
