package handlers

import (
	"ender/services"
	"ender/services/payment"
	"log"
	"net/http"

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
func RegisterUtilRoutes(se *core.ServeEvent) {
	// GET /api/health
	se.Router.GET("/api/health", func(e *core.RequestEvent) error {
		return e.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// GET /api/utils/app-settings
	se.Router.GET("/api/utils/app-settings", func(e *core.RequestEvent) error {
		settings := services.GetAppSettings(e.App)
		return e.JSON(http.StatusOK, settings)
	})
}

func handlePaymentWebhook(e *core.RequestEvent, providerName string) error {
	var payload map[string]any
	if err := e.BindBody(&payload); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"detail": "Invalid payload"})
	}

	return processWebhookPayload(e, providerName, payload)
}

func processWebhookPayload(e *core.RequestEvent, providerName string, payload map[string]any) error {
	provider := payment.GetProvider()
	if provider == nil || provider.Name() != providerName {
		return e.JSON(http.StatusBadRequest, map[string]string{"detail": "Unknown payment provider"})
	}

	event := provider.ParseWebhook(payload)
	if event == nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"detail": "Unrecognized webhook payload"})
	}

	switch event.EventType {
	case payment.EventPaymentCompleted:
		sub, err := services.CompleteInvoicePayment(e.App, event.RemoteID, event.TransactionID)
		if err != nil {
			log.Printf("ERROR: CompleteInvoicePayment failed: %v", err)
			return e.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
		}
		return e.JSON(http.StatusOK, map[string]any{
			"status":          "ok",
			"subscription_id": sub.Id,
		})

	case payment.EventAuthorizationCompleted:
		sub, err := services.CompleteAuthorization(e.App, event.RemoteID, event.UserUUID)
		if err != nil {
			log.Printf("ERROR: CompleteAuthorization failed: %v", err)
			return e.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
		}
		return e.JSON(http.StatusOK, map[string]any{
			"status":          "ok",
			"subscription_id": sub.Id,
		})

	default:
		return e.JSON(http.StatusOK, map[string]string{"status": "ignored"})
	}
}
