package handlers

import (
	"encoding/json"
	"ender/services"
	"ender/services/payment"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterWebhookRoutes registers payment provider webhook endpoints.
func RegisterWebhookRoutes(se *core.ServeEvent) {
	// POST /api/webhooks/{provider} — Payment webhook callback
	se.Router.POST("/api/webhooks/{provider}", func(e *core.RequestEvent) error {
		providerName := e.Request.PathValue("provider")
		return handlePaymentWebhook(e, providerName)
	})

	// GET /api/webhooks/{provider} — Authorization callback (QvaPay sends GET)
	se.Router.GET("/api/webhooks/{provider}", func(e *core.RequestEvent) error {
		providerName := e.Request.PathValue("provider")

		// Verify signed state token to prevent authorization callback poisoning
		stateToken := e.Request.URL.Query().Get("state")
		if stateToken == "" {
			return apis.NewBadRequestError("Missing state parameter", nil)
		}
		verifiedUserId, err := services.VerifyCallbackState(stateToken, 1*time.Hour)
		if err != nil {
			e.App.Logger().Warn("invalid callback state", slog.String("provider", providerName), slog.Any("error", err))
			return apis.NewBadRequestError("Invalid or expired state token", nil)
		}

		// Parse query params as payload for authorization callbacks
		payload := make(map[string]any)
		for key, values := range e.Request.URL.Query() {
			if len(values) > 0 {
				payload[key] = values[0]
			}
		}

		// Override remote_id with the verified userId from the state token
		payload["remote_id"] = verifiedUserId

		return processWebhookPayload(e, providerName, payment.WebhookRequest{
			Payload: payload,
		})
	})
}

// RegisterUtilRoutes registers utility routes.
// Note: /api/health is provided by PocketBase out of the box.
func RegisterUtilRoutes(se *core.ServeEvent) {
	// GET /api/utils/app-settings
	se.Router.GET("/api/utils/app-settings", func(e *core.RequestEvent) error {
		settings := services.GetAppSettings(e.App)
		return e.JSON(http.StatusOK, settings)
	})

	// GET /api/system-config — returns all system_config records (app admin only)
	se.Router.GET("/api/system-config", func(e *core.RequestEvent) error {
		if !isAppSuperuser(e) {
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
		if !isAppSuperuser(e) {
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
}

// isAppSuperuser checks if the authenticated user has the is_superuser flag.
func isAppSuperuser(e *core.RequestEvent) bool {
	record := e.Auth
	return record != nil && record.GetBool("is_superuser")
}

func handlePaymentWebhook(e *core.RequestEvent, providerName string) error {
	// Read raw body for signature verification
	rawBody, err := io.ReadAll(e.Request.Body)
	if err != nil {
		return apis.NewBadRequestError("Failed to read request body", nil)
	}

	// Parse JSON payload
	var payload map[string]any
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return apis.NewBadRequestError("Invalid JSON payload", nil)
	}

	// Extract headers
	headers := make(map[string]string)
	for key := range e.Request.Header {
		headers[key] = e.Request.Header.Get(key)
	}

	return processWebhookPayload(e, providerName, payment.WebhookRequest{
		RawBody: rawBody,
		Headers: headers,
		Payload: payload,
	})
}

func processWebhookPayload(e *core.RequestEvent, providerName string, webhookReq payment.WebhookRequest) error {
	provider := payment.GetProvider(providerName)
	if provider == nil {
		return apis.NewBadRequestError("Unknown payment provider", nil)
	}

	event, err := provider.ParseWebhook(webhookReq)
	if err != nil {
		e.App.Logger().Warn("webhook parse error", slog.String("provider", providerName), slog.Any("error", err))
		return apis.NewBadRequestError("Unrecognized webhook payload", nil)
	}

	// Idempotency: if the payment is already completed, return success without reprocessing
	if event.TransactionID != "" {
		existing, _ := services.FindPaymentByTransactionID(e.App, event.TransactionID)
		if existing != nil && existing.GetString("status") == "completed" {
			sub, _ := e.App.FindRecordById("subscriptions", existing.GetString("subscription"))
			subId := ""
			if sub != nil {
				subId = sub.Id
			}
			return e.JSON(http.StatusOK, map[string]any{
				"status":          "already_processed",
				"subscription_id": subId,
			})
		}
	}

	switch event.EventType {
	case payment.EventPaymentCompleted:
		sub, err := services.CompleteInvoicePayment(e.App, event.RemoteID, event.TransactionID)
		if err != nil {
			e.App.Logger().Error("CompleteInvoicePayment failed", slog.Any("error", err))
			return apis.NewApiError(http.StatusInternalServerError, err.Error(), nil)
		}
		return e.JSON(http.StatusOK, map[string]any{
			"status":          "ok",
			"subscription_id": sub.Id,
		})

	case payment.EventAuthorizationCompleted:
		sub, err := services.CompleteAuthorization(e.App, event.RemoteID, event.UserUUID)
		if err != nil {
			e.App.Logger().Error("CompleteAuthorization failed", slog.Any("error", err))
			return apis.NewApiError(http.StatusInternalServerError, err.Error(), nil)
		}
		return e.JSON(http.StatusOK, map[string]any{
			"status":          "ok",
			"subscription_id": sub.Id,
		})

	case payment.EventPaymentFailed:
		e.App.Logger().Warn("payment failed webhook", slog.String("remote_id", event.RemoteID))
		return e.JSON(http.StatusOK, map[string]string{"status": "noted"})

	default:
		return e.JSON(http.StatusOK, map[string]string{"status": "ignored"})
	}
}
