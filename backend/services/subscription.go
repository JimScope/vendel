package services

import (
	"fmt"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	"vendel/services/payment"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/mailer"
	"github.com/pocketbase/pocketbase/tools/types"
)

// StartSubscription begins the subscription process.
// Returns (subscription record, redirect URL or empty, error).
func StartSubscription(
	app core.App,
	userId, planId string,
	billingCycle, paymentMethod string,
	providerName string,
	webhookURL, successURL, errorURL string,
) (*core.Record, string, error) {
	// Check for existing subscription
	existing, _ := findSubscriptionByUser(app, userId)
	if existing != nil {
		switch existing.GetString("status") {
		case "pending":
			// Delete pending to allow new attempt
			_ = app.Delete(existing)
			existing = nil
		case "active":
			if existing.GetString("plan") == planId {
				return nil, "", fmt.Errorf("already subscribed to this plan")
			}
			return nil, "", fmt.Errorf("cancel current subscription before changing plans")
		}
	}

	plan, err := app.FindRecordById("user_plans", planId)
	if err != nil {
		return nil, "", fmt.Errorf("plan not found")
	}
	if !plan.GetBool("is_public") {
		return nil, "", fmt.Errorf("plan not available")
	}

	// Calculate amount and period
	var amount float64
	var periodDays int
	if billingCycle == "monthly" {
		amount = plan.GetFloat("price")
		periodDays = 30
	} else {
		amount = plan.GetFloat("price_yearly")
		if amount <= 0 {
			amount = plan.GetFloat("price") * 12
		}
		periodDays = 365
	}

	now := time.Now().UTC()
	periodEnd := now.Add(time.Duration(periodDays) * 24 * time.Hour)

	// Free plan — activate immediately
	if amount <= 0 {
		sub, err := createSubscriptionRecord(app, userId, planId, billingCycle, paymentMethod, providerName, "active", now, periodEnd)
		if err != nil {
			return nil, "", err
		}
		if err := updateUserQuota(app, userId, planId); err != nil {
			app.Logger().Warn("failed to update quota", slog.Any("error", err))
		}
		return sub, "", nil
	}

	// Delete existing expired/canceled subscription
	if existing != nil {
		_ = app.Delete(existing)
	}

	// Create pending subscription
	sub, err := createSubscriptionRecord(app, userId, planId, billingCycle, paymentMethod, providerName, "pending", now, now)
	if err != nil {
		return nil, "", err
	}

	provider := payment.GetProvider(providerName)
	if provider == nil {
		provider = payment.GetDefaultProvider()
	}
	if provider == nil {
		return nil, "", fmt.Errorf("no payment provider configured")
	}

	if paymentMethod == "invoice" {
		return startInvoiceSubscription(app, sub, plan, provider, amount, periodDays, webhookURL)
	}
	return startAuthorizedSubscription(app, sub, provider, userId, webhookURL, successURL, errorURL)
}

func startInvoiceSubscription(
	app core.App, sub, plan *core.Record,
	provider payment.Provider,
	amount float64, periodDays int,
	webhookURL string,
) (*core.Record, string, error) {
	now := time.Now().UTC()
	periodEnd := now.Add(time.Duration(periodDays) * 24 * time.Hour)

	// Create payment record
	pay, err := createPaymentRecord(app, sub.Id, amount, "USD", provider.Name(),
		now, periodEnd, "pending")
	if err != nil {
		return nil, "", err
	}

	// Create invoice with provider
	result, err := provider.CreateInvoice(payment.InvoiceRequest{
		Amount:      amount,
		Currency:    "USD",
		Description: fmt.Sprintf("Subscription: %s (%s)", plan.GetString("name"), sub.GetString("billing_cycle")),
		RemoteID:    pay.Id,
		WebhookURL:  webhookURL,
	})
	if err != nil {
		return nil, "", fmt.Errorf("payment provider error: %w", err)
	}

	if result.PaymentURL == "" {
		return nil, "", fmt.Errorf("payment provider did not return payment URL")
	}

	// Save invoice details
	pay.Set("provider_invoice_id", result.InvoiceID)
	pay.Set("provider_invoice_url", result.PaymentURL)
	_ = app.Save(pay)

	return sub, result.PaymentURL, nil
}

func startAuthorizedSubscription(
	app core.App, sub *core.Record,
	provider payment.Provider,
	userId, webhookURL, successURL, errorURL string,
) (*core.Record, string, error) {
	// Append signed state token to callback URL to prevent authorization poisoning
	stateToken, err := GenerateCallbackState(userId)
	if err != nil {
		return nil, "", fmt.Errorf("generate callback state: %w", err)
	}
	sep := "?"
	if strings.Contains(webhookURL, "?") {
		sep = "&"
	}
	signedCallbackURL := webhookURL + sep + "state=" + stateToken

	result, err := provider.GetAuthorizationURL(payment.AuthorizationRequest{
		RemoteID:    userId,
		CallbackURL: signedCallbackURL,
		SuccessURL:  successURL,
		ErrorURL:    errorURL,
	})
	if err != nil {
		return nil, "", fmt.Errorf("payment provider error: %w", err)
	}

	if result.AuthorizationURL == "" {
		return nil, "", fmt.Errorf("payment provider did not return authorization URL")
	}

	return sub, result.AuthorizationURL, nil
}

// CompleteInvoicePayment activates a subscription after invoice payment.
func CompleteInvoicePayment(app core.App, paymentId, transactionId string) (*core.Record, error) {
	pay, err := app.FindRecordById("payments", paymentId)
	if err != nil {
		return nil, fmt.Errorf("payment not found")
	}

	if pay.GetString("status") == "completed" {
		sub, _ := app.FindRecordById("subscriptions", pay.GetString("subscription"))
		return sub, nil
	}

	sub, err := app.FindRecordById("subscriptions", pay.GetString("subscription"))
	if err != nil {
		return nil, fmt.Errorf("subscription not found")
	}

	// Update payment + subscription atomically
	err = app.RunInTransaction(func(txApp core.App) error {
		pay.Set("status", "completed")
		pay.Set("provider_transaction_id", transactionId)
		pay.Set("paid_at", types.NowDateTime())
		if err := txApp.Save(pay); err != nil {
			return err
		}

		status := sub.GetString("status")
		if status == "pending" || status == "past_due" {
			sub.Set("status", "active")
			sub.Set("current_period_start", pay.GetString("period_start"))
			sub.Set("current_period_end", pay.GetString("period_end"))
			if err := txApp.Save(sub); err != nil {
				return err
			}
			if status == "pending" {
				return updateUserQuota(txApp, sub.GetString("user"), sub.GetString("plan"))
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return sub, nil
}

// CompleteAuthorization handles the callback after user authorizes recurring payments.
func CompleteAuthorization(app core.App, userId, providerUserUUID string) (*core.Record, error) {
	sub, err := findSubscriptionByUser(app, userId)
	if err != nil || sub == nil {
		return nil, fmt.Errorf("subscription not found")
	}

	if sub.GetString("status") != "pending" {
		sub.Set("provider_user_uuid", providerUserUUID)
		_ = app.Save(sub)
		return sub, nil
	}

	sub.Set("provider_user_uuid", providerUserUUID)
	_ = app.Save(sub)

	// Charge first payment
	plan, err := app.FindRecordById("user_plans", sub.GetString("plan"))
	if err != nil {
		return nil, fmt.Errorf("plan not found")
	}

	var amount float64
	var periodDays int
	if sub.GetString("billing_cycle") == "monthly" {
		amount = plan.GetFloat("price")
		periodDays = 30
	} else {
		amount = plan.GetFloat("price_yearly")
		if amount <= 0 {
			amount = plan.GetFloat("price") * 12
		}
		periodDays = 365
	}

	now := time.Now().UTC()
	periodEnd := now.Add(time.Duration(periodDays) * 24 * time.Hour)

	providerName := sub.GetString("provider")
	if providerName == "" {
		providerName = "qvapay"
	}
	pay, err := createPaymentRecord(app, sub.Id, amount, "USD",
		providerName, now, periodEnd, "pending")
	if err != nil {
		return nil, err
	}

	provider := payment.GetProvider(providerName)
	if provider == nil {
		provider = payment.GetDefaultProvider()
	}
	if provider == nil {
		return nil, fmt.Errorf("no payment provider configured")
	}
	chargeResult, err := provider.ChargeAuthorizedUser(payment.ChargeRequest{
		UserUUID:    providerUserUUID,
		Amount:      amount,
		Currency:    "USD",
		Description: fmt.Sprintf("Subscription: %s (%s)", plan.GetString("name"), sub.GetString("billing_cycle")),
		RemoteID:    pay.Id,
	})
	if err != nil {
		_ = app.RunInTransaction(func(txApp core.App) error {
			pay.Set("status", "failed")
			sub.Set("status", "expired")
			_ = txApp.Save(pay)
			return txApp.Save(sub)
		})
		return nil, fmt.Errorf("payment failed: %w", err)
	}

	err = app.RunInTransaction(func(txApp core.App) error {
		pay.Set("status", "completed")
		pay.Set("provider_transaction_id", chargeResult.TransactionID)
		pay.Set("paid_at", types.NowDateTime())
		if err := txApp.Save(pay); err != nil {
			return err
		}

		sub.Set("status", "active")
		sub.Set("current_period_start", now)
		sub.Set("current_period_end", periodEnd)
		if err := txApp.Save(sub); err != nil {
			return err
		}

		return updateUserQuota(txApp, userId, sub.GetString("plan"))
	})
	if err != nil {
		return nil, err
	}

	return sub, nil
}

// ProcessRenewal handles automatic subscription renewal.
func ProcessRenewal(app core.App, subscriptionId string) error {
	sub, err := app.FindRecordById("subscriptions", subscriptionId)
	if err != nil {
		return fmt.Errorf("subscription not found")
	}

	if sub.GetString("provider_user_uuid") == "" {
		return fmt.Errorf("no payment authorization")
	}

	if sub.GetString("status") != "active" {
		return fmt.Errorf("subscription not active")
	}

	if sub.GetBool("cancel_at_period_end") {
		sub.Set("status", "canceled")
		return app.Save(sub)
	}

	plan, err := app.FindRecordById("user_plans", sub.GetString("plan"))
	if err != nil {
		return fmt.Errorf("plan not found")
	}

	periodStart := sub.GetDateTime("current_period_end").Time()
	var amount float64
	var periodEnd time.Time
	if sub.GetString("billing_cycle") == "monthly" {
		amount = plan.GetFloat("price")
		periodEnd = periodStart.Add(30 * 24 * time.Hour)
	} else {
		amount = plan.GetFloat("price_yearly")
		if amount <= 0 {
			amount = plan.GetFloat("price") * 12
		}
		periodEnd = periodStart.Add(365 * 24 * time.Hour)
	}

	providerName := sub.GetString("provider")
	if providerName == "" {
		providerName = "qvapay"
	}
	provider := payment.GetProvider(providerName)
	if provider == nil {
		provider = payment.GetDefaultProvider()
	}
	if provider == nil {
		return fmt.Errorf("no payment provider configured")
	}
	pay, err := createPaymentRecord(app, sub.Id, amount, "USD",
		providerName, periodStart, periodEnd, "pending")
	if err != nil {
		return err
	}

	chargeResult, err := provider.ChargeAuthorizedUser(payment.ChargeRequest{
		UserUUID:    sub.GetString("provider_user_uuid"),
		Amount:      amount,
		Currency:    "USD",
		Description: fmt.Sprintf("Renewal: %s (%s)", plan.GetString("name"), sub.GetString("billing_cycle")),
		RemoteID:    pay.Id,
	})
	if err != nil {
		_ = app.RunInTransaction(func(txApp core.App) error {
			pay.Set("status", "failed")
			sub.Set("status", "past_due")
			_ = txApp.Save(pay)
			return txApp.Save(sub)
		})
		notifyRenewalFailure(app, sub, plan, err)
		return fmt.Errorf("renewal charge failed: %w", err)
	}

	return app.RunInTransaction(func(txApp core.App) error {
		pay.Set("status", "completed")
		pay.Set("provider_transaction_id", chargeResult.TransactionID)
		pay.Set("paid_at", types.NowDateTime())
		if err := txApp.Save(pay); err != nil {
			return err
		}

		sub.Set("current_period_start", periodStart)
		sub.Set("current_period_end", periodEnd)
		return txApp.Save(sub)
	})
}

// CancelSubscription cancels a subscription, optionally immediately.
func CancelSubscription(app core.App, userId string, immediate bool) (*core.Record, error) {
	sub, err := findSubscriptionByUser(app, userId)
	if err != nil || sub == nil {
		return nil, fmt.Errorf("subscription not found")
	}

	if sub.GetString("status") == "canceled" {
		return nil, fmt.Errorf("already canceled")
	}

	now := types.NowDateTime()

	if immediate {
		sub.Set("status", "canceled")
		sub.Set("canceled_at", now)
		if err := app.Save(sub); err != nil {
			return nil, err
		}
		_ = downgradeToFreePlan(app, userId)
		return sub, nil
	}

	sub.Set("cancel_at_period_end", true)
	sub.Set("canceled_at", now)
	if err := app.Save(sub); err != nil {
		return nil, err
	}
	return sub, nil
}

// CheckRenewals processes renewals for active authorized subscriptions due for renewal.
func CheckRenewals(app core.App) error {
	now := time.Now().UTC().Format(time.RFC3339)

	subs, err := app.FindRecordsByFilter(
		"subscriptions",
		"status = 'active' && payment_method = 'authorized' && provider_user_uuid != '' && current_period_end <= {:now} && cancel_at_period_end = false",
		"", 0, 0,
		dbx.Params{"now": now},
	)
	if err != nil {
		return err
	}

	for _, sub := range subs {
		if err := ProcessRenewal(app, sub.Id); err != nil {
			app.Logger().Warn("renewal failed", slog.String("subscription", sub.Id), slog.Any("error", err))
		}
	}

	app.Logger().Info("Processed subscription renewals", slog.Int("count", len(subs)))
	return nil
}

// ── Helpers ──────────────────────────────────────────────────────────

// FindPaymentByTransactionID looks up a payment record by provider_transaction_id.
func FindPaymentByTransactionID(app core.App, transactionID string) (*core.Record, error) {
	return app.FindFirstRecordByFilter(
		"payments",
		"provider_transaction_id = {:txId}",
		dbx.Params{"txId": transactionID},
	)
}

func findSubscriptionByUser(app core.App, userId string) (*core.Record, error) {
	return app.FindFirstRecordByFilter(
		"subscriptions",
		"user = {:userId}",
		dbx.Params{"userId": userId},
	)
}

func createSubscriptionRecord(
	app core.App,
	userId, planId, billingCycle, paymentMethod, provider, status string,
	periodStart, periodEnd time.Time,
) (*core.Record, error) {
	collection, err := app.FindCollectionByNameOrId("subscriptions")
	if err != nil {
		return nil, err
	}

	record := core.NewRecord(collection)
	record.Set("user", userId)
	record.Set("plan", planId)
	record.Set("billing_cycle", billingCycle)
	record.Set("payment_method", paymentMethod)
	record.Set("provider", provider)
	record.Set("status", status)
	record.Set("current_period_start", periodStart)
	record.Set("current_period_end", periodEnd)
	record.Set("cancel_at_period_end", false)

	if err := app.Save(record); err != nil {
		return nil, err
	}
	return record, nil
}

func createPaymentRecord(
	app core.App,
	subscriptionId string, amount float64, currency, provider string,
	periodStart, periodEnd time.Time, status string,
) (*core.Record, error) {
	collection, err := app.FindCollectionByNameOrId("payments")
	if err != nil {
		return nil, err
	}

	record := core.NewRecord(collection)
	record.Set("subscription", subscriptionId)
	record.Set("amount", amount)
	record.Set("currency", currency)
	record.Set("provider", provider)
	record.Set("status", status)
	record.Set("period_start", periodStart)
	record.Set("period_end", periodEnd)

	if err := app.Save(record); err != nil {
		return nil, err
	}
	return record, nil
}

func updateUserQuota(app core.App, userId, planId string) error {
	quota, err := app.FindFirstRecordByFilter(
		"user_quotas",
		"user = {:userId}",
		dbx.Params{"userId": userId},
	)
	if err != nil || quota == nil {
		return nil
	}
	quota.Set("plan", planId)
	quota.Set("sms_sent_this_month", 0)
	quota.Set("last_reset_date", types.NowDateTime())
	return app.Save(quota)
}

func downgradeToFreePlan(app core.App, userId string) error {
	plan, err := findFreePlan(app)
	if err != nil {
		return err
	}
	return updateUserQuota(app, userId, plan.Id)
}

func notifyRenewalFailure(app core.App, sub, plan *core.Record, chargeErr error) {
	user, err := app.FindRecordById("users", sub.GetString("user"))
	if err != nil {
		app.Logger().Warn("could not find user for renewal failure email", slog.Any("error", err))
		return
	}

	email := user.GetString("email")
	if email == "" {
		return
	}

	planName := plan.GetString("name")
	msg := &mailer.Message{
		From: mail.Address{
			Address: app.Settings().Meta.SenderAddress,
			Name:    app.Settings().Meta.SenderName,
		},
		To:      []mail.Address{{Address: email}},
		Subject: "Subscription renewal failed",
		HTML: fmt.Sprintf(
			`<p>Hi,</p>
<p>We were unable to renew your <strong>%s</strong> subscription. Your account has been marked as past due.</p>
<p>Please update your payment method to avoid service interruption.</p>
<p>Error: %s</p>`,
			planName, chargeErr.Error(),
		),
	}

	if err := app.NewMailClient().Send(msg); err != nil {
		app.Logger().Warn("failed to send renewal failure email", slog.String("email", email), slog.Any("error", err))
	}
}
