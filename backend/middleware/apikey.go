package middleware

import (
	"fmt"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// AuthenticateDevice validates the X-API-Key header against sms_devices.
// Returns the device record.
func AuthenticateDevice(e *core.RequestEvent) (*core.Record, error) {
	apiKey := e.Request.Header.Get("X-API-Key")
	if apiKey == "" {
		return nil, fmt.Errorf("missing X-API-Key header")
	}

	records, err := e.App.FindRecordsByFilter(
		"sms_devices",
		"api_key = {:key}",
		"", 1, 0,
		dbx.Params{"key": apiKey},
	)
	if err != nil || len(records) == 0 {
		return nil, fmt.Errorf("invalid device API key")
	}

	return records[0], nil
}

// AuthenticateIntegrationAPIKey validates the X-API-Key header against api_keys.
// Returns the user ID associated with the API key.
func AuthenticateIntegrationAPIKey(e *core.RequestEvent) (string, error) {
	apiKey := e.Request.Header.Get("X-API-Key")
	if apiKey == "" {
		return "", fmt.Errorf("missing X-API-Key header")
	}

	records, err := e.App.FindRecordsByFilter(
		"api_keys",
		"key = {:key} && is_active = true",
		"", 1, 0,
		dbx.Params{"key": apiKey},
	)
	if err != nil || len(records) == 0 {
		return "", fmt.Errorf("invalid integration API key")
	}

	record := records[0]

	// Update last_used_at in background
	go func() {
		record.Set("last_used_at", time.Now().UTC().Format(time.RFC3339))
		_ = e.App.Save(record)
	}()

	return record.GetString("user"), nil
}

// ResolveAuthOrAPIKey tries JWT auth first, falls back to integration API key.
// Returns the user ID.
func ResolveAuthOrAPIKey(e *core.RequestEvent) (string, error) {
	// Try PocketBase JWT auth (set by PB middleware)
	info, _ := e.RequestInfo()
	if info != nil && info.Auth != nil && info.Auth.Id != "" {
		return info.Auth.Id, nil
	}

	// Fall back to integration API key
	return AuthenticateIntegrationAPIKey(e)
}
