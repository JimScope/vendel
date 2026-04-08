package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"vendel/services"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterPromoRoutes registers promo code endpoints.
func RegisterPromoRoutes(se *core.ServeEvent) {
	// POST /api/promos/redeem — Redeem a promo code
	se.Router.POST("/api/promos/redeem", func(e *core.RequestEvent) error {
		userId := e.Auth.Id

		var body struct {
			Code string `json:"code"`
		}
		if err := e.BindBody(&body); err != nil {
			return apis.NewBadRequestError("Invalid request body", nil)
		}

		code := strings.TrimSpace(strings.ToUpper(body.Code))
		if code == "" {
			return apis.NewBadRequestError("Code is required", nil)
		}

		// Find the promo code
		promo, err := e.App.FindFirstRecordByFilter(
			"promo_codes",
			"code = {:code}",
			dbx.Params{"code": code},
		)
		if err != nil {
			return apis.NewApiError(http.StatusNotFound, "Invalid promo code", nil)
		}

		// Validate: active
		if !promo.GetBool("active") {
			return apis.NewApiError(http.StatusGone, "This promo code is no longer active", nil)
		}

		// Validate: not expired
		expiresAt := promo.GetDateTime("expires_at")
		if !expiresAt.IsZero() && expiresAt.Time().Before(time.Now().UTC()) {
			return apis.NewApiError(http.StatusGone, "This promo code has expired", nil)
		}

		// Validate: max redemptions not reached
		maxRedemptions := promo.GetInt("max_redemptions")
		timesRedeemed := promo.GetInt("times_redeemed")
		if maxRedemptions > 0 && timesRedeemed >= maxRedemptions {
			return apis.NewApiError(http.StatusGone, "This promo code has been fully redeemed", nil)
		}

		// Validate: user hasn't already redeemed this code
		existing, _ := e.App.FindFirstRecordByFilter(
			"promo_redemptions",
			"user = {:userId} && promo_code = {:promoId}",
			dbx.Params{"userId": userId, "promoId": promo.Id},
		)
		if existing != nil {
			return apis.NewApiError(http.StatusConflict, "You have already redeemed this code", nil)
		}

		amount := promo.GetFloat("amount")

		// Credit balance
		newBalance, err := services.CreditBalance(e.App, userId, amount)
		if err != nil {
			return apis.NewApiError(http.StatusInternalServerError, "Failed to credit balance", nil)
		}

		// Record redemption
		redemptionCol, err := e.App.FindCollectionByNameOrId("promo_redemptions")
		if err != nil {
			return apis.NewApiError(http.StatusInternalServerError, "Internal error", nil)
		}
		redemption := core.NewRecord(redemptionCol)
		redemption.Set("user", userId)
		redemption.Set("promo_code", promo.Id)
		redemption.Set("amount_credited", amount)
		if err := e.App.Save(redemption); err != nil {
			return apis.NewApiError(http.StatusInternalServerError, "Failed to record redemption", nil)
		}

		// Increment times_redeemed
		promo.Set("times_redeemed", timesRedeemed+1)
		if err := e.App.Save(promo); err != nil {
			e.App.Logger().Warn("failed to increment promo redemption count",
				slog.String("promo_id", promo.Id),
				slog.Any("error", err),
			)
		}

		e.App.Logger().Info("promo code redeemed",
			slog.String("user", userId),
			slog.String("code", code),
			slog.Float64("amount", amount),
		)

		return e.JSON(http.StatusOK, map[string]any{
			"message":     fmt.Sprintf("$%.2f credited to your balance", amount),
			"amount":      amount,
			"new_balance": newBalance,
		})
	}).Bind(apis.RequireAuth("users"))
}
