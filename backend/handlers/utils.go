package handlers

import (
	"vendel/middleware"
	"vendel/services"
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterUtilRoutes registers utility routes.
// Note: /api/health is provided by PocketBase out of the box.
func RegisterUtilRoutes(se *core.ServeEvent) {
	// GET /api/utils/app-settings
	se.Router.GET("/api/utils/app-settings", func(e *core.RequestEvent) error {
		settings := services.GetAppSettings(e.App)
		return e.JSON(http.StatusOK, settings)
	})

	// GET /api/user/export — export all user data (GDPR Art. 20)
	se.Router.GET("/api/user/export", func(e *core.RequestEvent) error {
		data, err := services.ExportUserData(e.App, e.Auth.Id)
		if err != nil {
			return apis.NewApiError(http.StatusInternalServerError, "Failed to export data", nil)
		}

		data["exported_at"] = time.Now().UTC().Format(time.RFC3339)

		e.Response.Header().Set("Content-Disposition", "attachment; filename=\"vendel-export.json\"")
		return e.JSON(http.StatusOK, data)
	}).Bind(apis.RequireAuth("users"))

	// GET /api/system-config — returns all system_config records (app admin only)
	se.Router.GET("/api/system-config", func(e *core.RequestEvent) error {
		if !middleware.IsAppSuperuser(e) {
			return e.ForbiddenError("", nil)
		}
		records, err := e.App.FindAllRecords("system_config")
		if err != nil {
			return apis.NewApiError(http.StatusInternalServerError, "Failed to fetch system config", nil)
		}
		return e.JSON(http.StatusOK, map[string]any{"data": records})
	}).Bind(apis.RequireAuth("users"))

	// PATCH /api/system-config/{key} — update (or create) a config value (app admin only)
	se.Router.PATCH("/api/system-config/{key}", func(e *core.RequestEvent) error {
		if !middleware.IsAppSuperuser(e) {
			return e.ForbiddenError("", nil)
		}
		key := e.Request.PathValue("key")

		var body struct {
			Value string `json:"value"`
		}
		if err := e.BindBody(&body); err != nil {
			return apis.NewBadRequestError("Invalid request body", nil)
		}

		// Try to find existing record by key
		record, _ := e.App.FindFirstRecordByFilter("system_config", "key = {:key}", map[string]any{"key": key})

		if record != nil {
			// Update existing
			record.Set("value", body.Value)
			if err := e.App.Save(record); err != nil {
				return apis.NewApiError(http.StatusInternalServerError, "Failed to update config", nil)
			}
		} else {
			// Create new
			collection, err := e.App.FindCollectionByNameOrId("system_config")
			if err != nil {
				return apis.NewApiError(http.StatusInternalServerError, "system_config collection not found", nil)
			}
			record = core.NewRecord(collection)
			record.Set("key", key)
			record.Set("value", body.Value)
			if err := e.App.Save(record); err != nil {
				return apis.NewApiError(http.StatusInternalServerError, "Failed to create config", nil)
			}
		}

		return e.JSON(http.StatusOK, record)
	}).Bind(apis.RequireAuth("users"))

	// POST /api/sms/validate-cron — Validate a cron expression (auth: JWT)
	se.Router.POST("/api/sms/validate-cron", func(e *core.RequestEvent) error {
		var body struct {
			Expression string `json:"expression"`
			Timezone   string `json:"timezone"`
		}
		if err := e.BindBody(&body); err != nil {
			return apis.NewBadRequestError("Invalid request body", nil)
		}

		if body.Expression == "" {
			return e.JSON(http.StatusOK, map[string]any{
				"valid": false,
				"error": "Expression is required",
			})
		}

		if body.Timezone == "" {
			body.Timezone = "UTC"
		}

		if err := services.ValidateCronExpression(body.Expression); err != nil {
			return e.JSON(http.StatusOK, map[string]any{
				"valid": false,
				"error": err.Error(),
			})
		}

		nextRun, err := services.ComputeNextRun(body.Expression, body.Timezone)
		if err != nil {
			return e.JSON(http.StatusOK, map[string]any{
				"valid": false,
				"error": err.Error(),
			})
		}

		return e.JSON(http.StatusOK, map[string]any{
			"valid":    true,
			"next_run": nextRun,
		})
	}).Bind(apis.RequireAuth("users"))
}
