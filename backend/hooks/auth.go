package hooks

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterAuthHooks blocks login for unverified users (password auth only; OAuth2 auto-verifies).
func RegisterAuthHooks(app *pocketbase.PocketBase) {
	app.OnRecordAuthWithPasswordRequest("users").BindFunc(func(e *core.RecordAuthWithPasswordRequestEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		if !e.Record.Verified() {
			return e.UnauthorizedError("Please verify your email address before logging in.", nil)
		}
		return nil
	})
}
