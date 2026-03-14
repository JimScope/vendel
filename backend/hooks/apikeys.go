package hooks

import (
	"time"
	"vendel/services"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterApiKeyHooks generates a secure key and sets a default 1-year
// expiration on API key creation.
func RegisterApiKeyHooks(app *pocketbase.PocketBase) {
	app.OnRecordCreate("api_keys").BindFunc(func(e *core.RecordEvent) error {
		e.Record.Set("key", services.GenerateSecureKey("vk_", 32))

		// Default expiration to 1 year if not explicitly set
		if e.Record.GetDateTime("expires_at").IsZero() {
			e.Record.Set("expires_at", time.Now().AddDate(1, 0, 0).UTC().Format(time.RFC3339))
		}

		// Unhide so the key is returned in the create response (only shown once)
		e.Record.Unhide("key")

		return e.Next()
	})
}
