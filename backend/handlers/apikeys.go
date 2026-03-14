package handlers

import (
	"vendel/services"
	"net/http"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterApiKeyRoutes registers custom API key management routes.
func RegisterApiKeyRoutes(se *core.ServeEvent) {
	// POST /api/api-keys/{id}/rotate — Rotate an API key (auth: JWT)
	se.Router.POST("/api/api-keys/{id}/rotate", func(e *core.RequestEvent) error {
		keyId := e.Request.PathValue("id")
		if keyId == "" {
			return apis.NewBadRequestError("API key ID required", nil)
		}

		var body struct {
			ExpiresAt string `json:"expires_at"`
		}
		_ = e.BindBody(&body) // optional fields

		result, err := services.RotateAPIKey(e.App, e.Auth.Id, keyId, body.ExpiresAt)
		if err != nil {
			return apis.NewBadRequestError(err.Error(), nil)
		}

		newKey := result.NewKey
		return e.JSON(http.StatusOK, map[string]any{
			"id":         newKey.Id,
			"name":       newKey.GetString("name"),
			"key":        newKey.GetString("key"),
			"key_prefix": services.GenerateKeyPrefix(newKey.GetString("key")),
			"is_active":  true,
			"expires_at": newKey.GetString("expires_at"),
			"created":    newKey.GetString("created"),
		})
	}).Bind(apis.RequireAuth("users"))
}
