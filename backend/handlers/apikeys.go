package handlers

import (
	"ender/services"
	"net/http"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterApiKeyRoutes registers custom API key management routes.
func RegisterApiKeyRoutes(se *core.ServeEvent) {
	// POST /api/api-keys/{id}/rotate — Rotate an API key (auth: JWT)
	se.Router.POST("/api/api-keys/{id}/rotate", func(e *core.RequestEvent) error {
		info, _ := e.RequestInfo()
		if info == nil || info.Auth == nil || info.Auth.Id == "" {
			return apis.NewUnauthorizedError("Authentication required", nil)
		}
		userId := info.Auth.Id

		keyId := e.Request.PathValue("id")
		if keyId == "" {
			return apis.NewBadRequestError("API key ID required", nil)
		}

		oldKey, err := e.App.FindRecordById("api_keys", keyId)
		if err != nil {
			return apis.NewNotFoundError("API key not found", nil)
		}
		if oldKey.GetString("user") != userId {
			return apis.NewForbiddenError("Not your API key", nil)
		}
		if !oldKey.GetBool("is_active") {
			return apis.NewBadRequestError("Cannot rotate a revoked key", nil)
		}

		var body struct {
			ExpiresAt string `json:"expires_at"`
		}
		_ = e.BindBody(&body) // optional fields

		// Deactivate old key
		oldKey.Set("is_active", false)
		if err := e.App.Save(oldKey); err != nil {
			return apis.NewApiError(http.StatusInternalServerError, "Failed to deactivate old key", nil)
		}

		// Create new key (the OnRecordCreate hook generates the actual key value)
		col, err := e.App.FindCollectionByNameOrId("api_keys")
		if err != nil {
			return apis.NewApiError(http.StatusInternalServerError, "Collection not found", nil)
		}

		newKey := core.NewRecord(col)
		newKey.Set("name", oldKey.GetString("name")+" (rotated)")
		newKey.Set("user", userId)
		newKey.Set("is_active", true)
		if body.ExpiresAt != "" {
			newKey.Set("expires_at", body.ExpiresAt)
		}

		if err := e.App.Save(newKey); err != nil {
			return apis.NewApiError(http.StatusInternalServerError, "Failed to create new key", nil)
		}

		// Unhide key so it's returned in the response (shown once)
		newKey.Unhide("key")

		return e.JSON(http.StatusOK, map[string]any{
			"id":         newKey.Id,
			"name":       newKey.GetString("name"),
			"key":        newKey.GetString("key"),
			"key_prefix": services.GenerateKeyPrefix(newKey.GetString("key")),
			"is_active":  true,
			"expires_at": newKey.GetString("expires_at"),
			"created":    newKey.GetString("created"),
		})
	})
}
