package handlers

import (
	"vendel/middleware"
	"vendel/services"
	"vendel/services/payment"
	"fmt"
	"net/http"
	"os"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterPlanRoutes registers plan/subscription-related routes.
func RegisterPlanRoutes(se *core.ServeEvent) {
	// GET /api/plans/quota — Get user quota info
	se.Router.GET("/api/plans/quota", func(e *core.RequestEvent) error {
		userId, err := middleware.ResolveAuthOrAPIKey(e)
		if err != nil {
			return apis.NewUnauthorizedError("Authentication required", nil)
		}

		quota, err := services.GetUserQuota(e.App, userId)
		if err != nil {
			return apis.NewApiError(http.StatusInternalServerError, err.Error(), nil)
		}

		return e.JSON(http.StatusOK, quota)
	})

	// PUT /api/plans/upgrade — Start subscription
	se.Router.PUT("/api/plans/upgrade", func(e *core.RequestEvent) error {
		userId := e.Auth.Id

		var body struct {
			PlanID       string `json:"plan_id"`
			BillingCycle string `json:"billing_cycle"`
			Provider     string `json:"provider"`
		}
		if err := e.BindBody(&body); err != nil {
			return apis.NewBadRequestError("Invalid request body", nil)
		}

		if body.BillingCycle == "" {
			body.BillingCycle = "monthly"
		}

		// Resolve payment provider — payment method is derived from it
		var provider payment.Provider
		if body.Provider != "" {
			provider = payment.GetProvider(body.Provider)
			if provider == nil {
				return apis.NewBadRequestError("Unknown payment provider: "+body.Provider, nil)
			}
		} else {
			provider = payment.GetDefaultProvider()
		}
		if provider == nil {
			return apis.NewBadRequestError("No payment provider configured", nil)
		}

		providerName := provider.Name()
		paymentMethod := provider.PaymentMethod()

		// Build callback URLs
		baseURL := os.Getenv("APP_URL")
		if baseURL == "" {
			baseURL = "http://localhost:8090"
		}

		webhookURL := fmt.Sprintf("%s/api/webhooks/%s", baseURL, providerName)
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:5173"
		}
		successURL := frontendURL + "/settings"
		errorURL := frontendURL + "/settings"

		sub, redirectURL, err := services.StartSubscription(
			e.App, userId, body.PlanID, body.BillingCycle, paymentMethod,
			providerName, webhookURL, successURL, errorURL,
		)
		if err != nil {
			return handleServiceError(e, err)
		}

		result := map[string]any{
			"subscription_id": sub.Id,
			"status":          sub.GetString("status"),
		}
		if paymentMethod == "balance" && redirectURL != "" {
			// redirectURL is the wallet address for crypto deposits
			result["wallet_address"] = redirectURL
			result["message"] = "Deposit to wallet address to activate subscription"
		} else if redirectURL != "" {
			result["payment_url"] = redirectURL
			result["message"] = "Redirect user to payment URL"
		} else {
			result["message"] = "Subscription activated"
		}

		return e.JSON(http.StatusOK, result)
	}).Bind(apis.RequireAuth("users"))

	// GET /api/plans/balance — Get user crypto balance and wallet info
	se.Router.GET("/api/plans/balance", func(e *core.RequestEvent) error {
		userId := e.Auth.Id

		bal, err := services.GetOrCreateBalance(e.App, userId)
		if err != nil {
			return apis.NewApiError(http.StatusInternalServerError, err.Error(), nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"balance":        bal.GetFloat("balance"),
			"currency":       bal.GetString("currency"),
			"wallet_address": bal.GetString("wallet_address"),
			"wallet_id":      bal.GetString("wallet_id"),
		})
	}).Bind(apis.RequireAuth("users"))

	// POST /api/plans/cancel — Cancel subscription
	se.Router.POST("/api/plans/cancel", func(e *core.RequestEvent) error {
		userId := e.Auth.Id

		var body struct {
			Immediate bool `json:"immediate"`
		}
		_ = e.BindBody(&body)

		sub, err := services.CancelSubscription(e.App, userId, body.Immediate)
		if err != nil {
			return apis.NewBadRequestError(err.Error(), nil)
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
	}).Bind(apis.RequireAuth("users"))
}
