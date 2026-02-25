package middleware

import (
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// MaintenanceMiddleware returns a hook function that checks maintenance mode.
func MaintenanceMiddleware(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		path := e.Request.URL.Path

		// Always allow: health, PocketBase admin, auth, app-settings, system-config
		if path == "/api/health" ||
			strings.HasPrefix(path, "/_/") ||
			strings.HasPrefix(path, "/api/admins") ||
			strings.HasPrefix(path, "/api/collections/users/auth-") ||
			strings.HasPrefix(path, "/api/system-config") ||
			path == "/api/utils/app-settings" {
			return e.Next()
		}

		// Check maintenance mode in system_config
		record, err := app.FindFirstRecordByFilter(
			"system_config",
			"key = 'maintenance_mode'",
		)
		if err == nil && record != nil {
			if strings.ToLower(record.GetString("value")) == "true" {
				return apis.NewApiError(http.StatusServiceUnavailable, "Service is under maintenance. Please try again later.", nil)
			}
		}

		return e.Next()
	}
}
