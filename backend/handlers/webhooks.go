package handlers

import (
	"ender/services"
	"ender/services/payment"
	"log/slog"
	"net/http"

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

		// Parse query params as payload for authorization callbacks
		payload := make(map[string]any)
		for key, values := range e.Request.URL.Query() {
			if len(values) > 0 {
				payload[key] = values[0]
			}
		}

		return processWebhookPayload(e, providerName, payload)
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
}

func handlePaymentWebhook(e *core.RequestEvent, providerName string) error {
	var payload map[string]any
	if err := e.BindBody(&payload); err != nil {
		return apis.NewBadRequestError("Invalid payload", nil)
	}

	return processWebhookPayload(e, providerName, payload)
}

func processWebhookPayload(e *core.RequestEvent, providerName string, payload map[string]any) error {
	provider := payment.GetProvider()
	if provider == nil || provider.Name() != providerName {
		return apis.NewBadRequestError("Unknown payment provider", nil)
	}

	event := provider.ParseWebhook(payload)
	if event == nil {
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

	default:
		return e.JSON(http.StatusOK, map[string]string{"status": "ignored"})
	}
}
