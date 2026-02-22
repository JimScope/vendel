package handlers

import (
	"ender/middleware"
	"ender/services"
	"fmt"
	"net/http"
	"os"

	"github.com/pocketbase/pocketbase/core"
)

// RegisterPlanRoutes registers plan/subscription-related routes.
func RegisterPlanRoutes(se *core.ServeEvent) {
	// GET /api/plans/quota — Get user quota info
	se.Router.GET("/api/plans/quota", func(e *core.RequestEvent) error {
		userId, err := middleware.ResolveAuthOrAPIKey(e)
		if err != nil {
			return e.JSON(http.StatusUnauthorized, map[string]string{"detail": "Authentication required"})
		}

		quota, err := services.GetUserQuota(e.App, userId)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
		}

		return e.JSON(http.StatusOK, quota)
	})

	// PUT /api/plans/upgrade — Start subscription
	se.Router.PUT("/api/plans/upgrade", func(e *core.RequestEvent) error {
		info, _ := e.RequestInfo()
		userId := info.Auth.Id
		if userId == "" {
			return e.JSON(http.StatusUnauthorized, map[string]string{"detail": "Authentication required"})
		}

		var body struct {
			PlanID        string `json:"plan_id"`
			BillingCycle  string `json:"billing_cycle"`
			PaymentMethod string `json:"payment_method"`
		}
		if err := e.BindBody(&body); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{"detail": "Invalid request body"})
		}

		if body.BillingCycle == "" {
			body.BillingCycle = "monthly"
		}
		if body.PaymentMethod == "" {
			body.PaymentMethod = services.GetSystemConfigValue(e.App, "default_payment_method")
			if body.PaymentMethod == "" {
				body.PaymentMethod = "invoice"
			}
		}

		// Build callback URLs
		baseURL := os.Getenv("SERVER_BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:8090"
		}
		provider := os.Getenv("PAYMENT_PROVIDER")
		if provider == "" {
			provider = "qvapay"
		}

		webhookURL := fmt.Sprintf("%s/api/webhooks/%s", baseURL, provider)
		frontendHost := os.Getenv("FRONTEND_HOST")
		if frontendHost == "" {
			frontendHost = "http://localhost:5173"
		}
		successURL := frontendHost + "/settings"
		errorURL := frontendHost + "/settings"

		sub, redirectURL, err := services.StartSubscription(
			e.App, userId, body.PlanID, body.BillingCycle, body.PaymentMethod,
			webhookURL, successURL, errorURL,
		)
		if err != nil {
			if qe, ok := err.(*services.QuotaError); ok {
				return e.JSON(qe.StatusCode, qe.Body)
			}
			return e.JSON(http.StatusBadRequest, map[string]string{"detail": err.Error()})
		}

		result := map[string]any{
			"subscription_id": sub.Id,
			"status":          sub.GetString("status"),
		}
		if redirectURL != "" {
			result["payment_url"] = redirectURL
			result["message"] = "Redirect user to payment URL"
		} else {
			result["message"] = "Subscription activated"
		}

		return e.JSON(http.StatusOK, result)
	})

	// POST /api/plans/cancel — Cancel subscription
	se.Router.POST("/api/plans/cancel", func(e *core.RequestEvent) error {
		info, _ := e.RequestInfo()
		userId := info.Auth.Id
		if userId == "" {
			return e.JSON(http.StatusUnauthorized, map[string]string{"detail": "Authentication required"})
		}

		var body struct {
			Immediate bool `json:"immediate"`
		}
		_ = e.BindBody(&body)

		sub, err := services.CancelSubscription(e.App, userId, body.Immediate)
		if err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{"detail": err.Error()})
		}

		msg := "Subscription will cancel at period end"
		if body.Immediate {
			msg = "Subscription canceled immediately"
		}

		return e.JSON(http.StatusOK, map[string]any{
			"message":         msg,
			"subscription_id": sub.Id,
			"status":          sub.GetString("status"),
		})
	})
}
