package hooks

import (
	"fmt"
	"vendel/services"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterWebhookHooks registers URL validation, secret encryption,
// and quota checks for webhook_configs on create and update.
func RegisterWebhookHooks(app *pocketbase.PocketBase) {
	app.OnRecordCreate("webhook_configs").BindFunc(func(e *core.RecordEvent) error {
		userId := e.Record.GetString("user")
		if err := services.CheckIntegrationQuota(e.App, userId); err != nil {
			return err
		}
		return validateAndEncryptWebhook(e)
	})

	app.OnRecordUpdate("webhook_configs").BindFunc(validateAndEncryptWebhook)
}

func validateAndEncryptWebhook(e *core.RecordEvent) error {
	if url := e.Record.GetString("url"); url != "" {
		if err := services.ValidateWebhookURL(url); err != nil {
			return fmt.Errorf("invalid webhook URL: %w", err)
		}
	}
	if secret := e.Record.GetString("secret_key"); secret != "" {
		encrypted, err := services.EncryptSecret(secret)
		if err != nil {
			return err
		}
		e.Record.Set("secret_key", encrypted)
	}
	return e.Next()
}
