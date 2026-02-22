package middleware

import (
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// MaintenanceMiddleware returns a hook function that checks maintenance mode.
func MaintenanceMiddleware(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		path := e.Request.URL.Path

		// Always allow health, admin, and API collections endpoints
		if path == "/api/health" ||
			strings.HasPrefix(path, "/_/") ||
			strings.HasPrefix(path, "/api/admins") {
			return e.Next()
		}

		// Check maintenance mode in system_config
		records, err := app.FindRecordsByFilter(
			"system_config",
			"key = 'maintenance_mode'",
			"", 1, 0,
		)
		if err == nil && len(records) > 0 {
			if strings.ToLower(records[0].GetString("value")) == "true" {
				return e.JSON(http.StatusServiceUnavailable, map[string]string{
					"detail": "Service is under maintenance. Please try again later.",
				})
			}
		}

		return e.Next()
	}
}
