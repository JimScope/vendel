package services

import (
	"vendel/services/payment"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// GetSystemConfigValue reads a value from the system_config collection.
func GetSystemConfigValue(app core.App, key string) string {
	record, err := app.FindFirstRecordByFilter(
		"system_config",
		"key = {:key}",
		dbx.Params{"key": key},
	)
	if err != nil || record == nil {
		return ""
	}
	return record.GetString("value")
}

// GetAppSettings returns public app settings.
func GetAppSettings(app core.App) map[string]any {
	keys := []string{"app_name", "support_email"}
	result := make(map[string]any)
	for _, k := range keys {
		result[k] = GetSystemConfigValue(app, k)
	}
	// Fill defaults
	if result["app_name"] == "" {
		result["app_name"] = "Vendel"
	}
	// Add maintenance status
	if strings.ToLower(GetSystemConfigValue(app, "maintenance_mode")) == "true" {
		result["maintenance_mode"] = "true"
	} else {
		result["maintenance_mode"] = "false"
	}
	// Add configured payment providers
	var providers []map[string]string
	for _, p := range payment.GetProviders() {
		providers = append(providers, map[string]string{
			"name":         p.Name(),
			"display_name": p.DisplayName(),
		})
	}
	result["payment_providers"] = providers
	return result
}
