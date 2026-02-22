package services

import (
	"fmt"
	"log"
	"net/mail"
	"time"

	"ender/services/payment"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/mailer"
)

// StartSubscription begins the subscription process.
// Returns (subscription record, redirect URL or empty, error).
func StartSubscription(
	app core.App,
	userId, planId string,
	billingCycle, paymentMethod string,
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
		sub, err := createSubscriptionRecord(app, userId, planId, billingCycle, paymentMethod, "active", now, periodEnd)
		if err != nil {
			return nil, "", err
		}
		if err := updateUserQuota(app, userId, planId); err != nil {
			log.Printf("WARNING: failed to update quota: %v", err)
		}
		return sub, "", nil
	}

	// Delete existing expired/canceled subscription
	if existing != nil {
		_ = app.Delete(existing)
	}

	// Create pending subscription
	sub, err := createSubscriptionRecord(app, userId, planId, billingCycle, paymentMethod, "pending", now, now)
	if err != nil {
		return nil, "", err
	}

	provider := payment.GetProvider()
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
	result, err := provider.GetAuthorizationURL(payment.AuthorizationRequest{
		RemoteID:    userId,
		CallbackURL: webhookURL,
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

	// Update payment
	pay.Set("status", "completed")
	pay.Set("provider_transaction_id", transactionId)
	pay.Set("paid_at", time.Now().UTC().Format(time.RFC3339))
	_ = app.Save(pay)

	status := sub.GetString("status")
	if status == "pending" || status == "past_due" {
		sub.Set("status", "active")
		sub.Set("current_period_start", pay.GetString("period_start"))
		sub.Set("current_period_end", pay.GetString("period_end"))
		_ = app.Save(sub)

		if status == "pending" {
			_ = updateUserQuota(app, sub.GetString("user"), sub.GetString("plan"))
		}
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

	pay, err := createPaymentRecord(app, sub.Id, amount, "USD",
		payment.GetProvider().Name(), now, periodEnd, "pending")
	if err != nil {
		return nil, err
	}

	provider := payment.GetProvider()
	chargeResult, err := provider.ChargeAuthorizedUser(payment.ChargeRequest{
		UserUUID:    providerUserUUID,
		Amount:      amount,
		Currency:    "USD",
		Description: fmt.Sprintf("Subscription: %s (%s)", plan.GetString("name"), sub.GetString("billing_cycle")),
		RemoteID:    pay.Id,
	})
	if err != nil {
		pay.Set("status", "failed")
		sub.Set("status", "expired")
		_ = app.Save(pay)
		_ = app.Save(sub)
		return nil, fmt.Errorf("payment failed: %w", err)
	}

	pay.Set("status", "completed")
	pay.Set("provider_transaction_id", chargeResult.TransactionID)
	pay.Set("paid_at", now.Format(time.RFC3339))
	_ = app.Save(pay)

	sub.Set("status", "active")
	sub.Set("current_period_start", now.Format(time.RFC3339))
	sub.Set("current_period_end", periodEnd.Format(time.RFC3339))
	_ = app.Save(sub)

	_ = updateUserQuota(app, userId, sub.GetString("plan"))
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

	provider := payment.GetProvider()
	pay, err := createPaymentRecord(app, sub.Id, amount, "USD",
		provider.Name(), periodStart, periodEnd, "pending")
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
		pay.Set("status", "failed")
		sub.Set("status", "past_due")
		_ = app.Save(pay)
		_ = app.Save(sub)
		notifyRenewalFailure(app, sub, plan, err)
		return fmt.Errorf("renewal charge failed: %w", err)
	}

	pay.Set("status", "completed")
	pay.Set("provider_transaction_id", chargeResult.TransactionID)
	pay.Set("paid_at", time.Now().UTC().Format(time.RFC3339))
	_ = app.Save(pay)

	sub.Set("current_period_start", periodStart.Format(time.RFC3339))
	sub.Set("current_period_end", periodEnd.Format(time.RFC3339))
	return app.Save(sub)
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

	now := time.Now().UTC().Format(time.RFC3339)

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
			log.Printf("WARNING: renewal failed for %s: %v", sub.Id, err)
		}
	}

	log.Printf("Processed %d subscription renewals", len(subs))
	return nil
}

// ── Helpers ──────────────────────────────────────────────────────────

// FindPaymentByTransactionID looks up a payment record by provider_transaction_id.
func FindPaymentByTransactionID(app core.App, transactionID string) (*core.Record, error) {
	records, err := app.FindRecordsByFilter(
		"payments",
		"provider_transaction_id = {:txId}",
		"", 1, 0,
		dbx.Params{"txId": transactionID},
	)
	if err != nil || len(records) == 0 {
		return nil, err
	}
	return records[0], nil
}

func findSubscriptionByUser(app core.App, userId string) (*core.Record, error) {
	records, err := app.FindRecordsByFilter(
		"subscriptions",
		"user = {:userId}",
		"", 1, 0,
		dbx.Params{"userId": userId},
	)
	if err != nil || len(records) == 0 {
		return nil, err
	}
	return records[0], nil
}

func createSubscriptionRecord(
	app core.App,
	userId, planId, billingCycle, paymentMethod, status string,
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
	record.Set("status", status)
	record.Set("current_period_start", periodStart.Format(time.RFC3339))
	record.Set("current_period_end", periodEnd.Format(time.RFC3339))
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
	record.Set("period_start", periodStart.Format(time.RFC3339))
	record.Set("period_end", periodEnd.Format(time.RFC3339))

	if err := app.Save(record); err != nil {
		return nil, err
	}
	return record, nil
}

func updateUserQuota(app core.App, userId, planId string) error {
	records, err := app.FindRecordsByFilter(
		"user_quotas",
		"user = {:userId}",
		"", 1, 0,
		dbx.Params{"userId": userId},
	)
	if err != nil || len(records) == 0 {
		return nil
	}

	quota := records[0]
	quota.Set("plan", planId)
	quota.Set("sms_sent_this_month", 0)
	quota.Set("last_reset_date", time.Now().UTC().Format(time.RFC3339))
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
		log.Printf("WARNING: could not find user for renewal failure email: %v", err)
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
		log.Printf("WARNING: failed to send renewal failure email to %s: %v", email, err)
	}
}
