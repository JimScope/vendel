package middleware

import (
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// maintenanceCache holds a cached maintenance mode status with TTL.
var maintenanceCache atomic.Value

const maintenanceCacheTTL = 60 * time.Second

type cachedMaintenance struct {
	enabled bool
	expires time.Time
}

// InvalidateMaintenanceCache forces the next request to re-check the DB.
func InvalidateMaintenanceCache() {
	maintenanceCache.Store((*cachedMaintenance)(nil))
}

// isMaintenanceEnabled checks the cache first, falls back to DB.
func isMaintenanceEnabled(app core.App) bool {
	if cached, ok := maintenanceCache.Load().(*cachedMaintenance); ok && cached != nil {
		if time.Now().Before(cached.expires) {
			return cached.enabled
		}
	}

	enabled := false
	record, err := app.FindFirstRecordByFilter("system_config", "key = 'maintenance_mode'")
	if err == nil && record != nil {
		enabled = strings.ToLower(record.GetString("value")) == "true"
	}

	maintenanceCache.Store(&cachedMaintenance{
		enabled: enabled,
		expires: time.Now().Add(maintenanceCacheTTL),
	})
	return enabled
}

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

		if isMaintenanceEnabled(app) {
			return apis.NewApiError(http.StatusServiceUnavailable, "Service is under maintenance. Please try again later.", nil)
		}

		return e.Next()
	}
}
