package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const stripeBaseURL = "https://api.stripe.com"

// StripeProvider implements the Provider interface for Stripe using raw HTTP calls.
type StripeProvider struct {
	SecretKey     string
	WebhookSecret string
	client        *http.Client
}

// NewStripeProvider creates a new Stripe provider from environment.
func NewStripeProvider() *StripeProvider {
	return &StripeProvider{
		SecretKey:     os.Getenv("STRIPE_SECRET_KEY"),
		WebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
		client:        &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *StripeProvider) Name() string        { return "stripe" }
func (p *StripeProvider) DisplayName() string { return "Stripe" }
func (p *StripeProvider) IsConfigured() bool  { return p.SecretKey != "" && p.WebhookSecret != "" }

func (p *StripeProvider) CreateInvoice(req InvoiceRequest) (*InvoiceResult, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("Stripe is not configured")
	}

	// Amount in cents
	amountCents := int64(req.Amount * 100)

	params := url.Values{}
	params.Set("mode", "payment")
	params.Set("line_items[0][price_data][currency]", strings.ToLower(req.Currency))
	params.Set("line_items[0][price_data][product_data][name]", req.Description)
	params.Set("line_items[0][price_data][unit_amount]", strconv.FormatInt(amountCents, 10))
	params.Set("line_items[0][quantity]", "1")
	params.Set("metadata[remote_id]", req.RemoteID)
	if req.SuccessURL != "" {
		params.Set("success_url", req.SuccessURL)
	}
	if req.ErrorURL != "" {
		params.Set("cancel_url", req.ErrorURL)
	}

	data, err := p.post("/v1/checkout/sessions", params)
	if err != nil {
		return nil, err
	}

	sessionURL, _ := data["url"].(string)
	sessionID, _ := data["id"].(string)
	if sessionURL == "" {
		return nil, fmt.Errorf("Stripe checkout session did not return URL")
	}

	return &InvoiceResult{
		InvoiceID:  sessionID,
		PaymentURL: sessionURL,
	}, nil
}

func (p *StripeProvider) GetAuthorizationURL(req AuthorizationRequest) (*AuthorizationResult, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("Stripe is not configured")
	}

	params := url.Values{}
	params.Set("mode", "setup")
	params.Set("customer_creation", "always")
	params.Set("metadata[remote_id]", req.RemoteID)
	// Setup mode requires a dummy payment_method_types or Stripe auto-detects
	params.Set("payment_method_types[0]", "card")
	if req.SuccessURL != "" {
		params.Set("success_url", req.SuccessURL)
	}
	if req.ErrorURL != "" {
		params.Set("cancel_url", req.ErrorURL)
	}

	data, err := p.post("/v1/checkout/sessions", params)
	if err != nil {
		return nil, err
	}

	sessionURL, _ := data["url"].(string)
	if sessionURL == "" {
		return nil, fmt.Errorf("Stripe setup session did not return URL")
	}

	return &AuthorizationResult{AuthorizationURL: sessionURL}, nil
}

func (p *StripeProvider) ChargeAuthorizedUser(req ChargeRequest) (*ChargeResult, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("Stripe is not configured")
	}

	// UserUUID stores "cus_xxx:pm_xxx"
	parts := strings.SplitN(req.UserUUID, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid Stripe customer/payment method: %s", req.UserUUID)
	}
	customerID := parts[0]
	paymentMethodID := parts[1]

	amountCents := int64(req.Amount * 100)

	params := url.Values{}
	params.Set("amount", strconv.FormatInt(amountCents, 10))
	params.Set("currency", strings.ToLower(req.Currency))
	params.Set("customer", customerID)
	params.Set("payment_method", paymentMethodID)
	params.Set("off_session", "true")
	params.Set("confirm", "true")
	params.Set("description", req.Description)
	params.Set("metadata[remote_id]", req.RemoteID)

	data, err := p.post("/v1/payment_intents", params)
	if err != nil {
		return nil, err
	}

	piID, _ := data["id"].(string)
	status, _ := data["status"].(string)
	if status != "succeeded" {
		return nil, fmt.Errorf("Stripe payment intent status: %s", status)
	}

	return &ChargeResult{
		TransactionID: piID,
		Amount:        req.Amount,
	}, nil
}

func (p *StripeProvider) ParseWebhook(req WebhookRequest) (*WebhookEvent, error) {
	// Verify signature
	sigHeader := req.Headers["Stripe-Signature"]
	if sigHeader == "" {
		sigHeader = req.Headers["stripe-signature"]
	}
	if sigHeader == "" {
		return nil, fmt.Errorf("missing Stripe-Signature header")
	}

	if err := p.verifySignature(sigHeader, req.RawBody); err != nil {
		return nil, fmt.Errorf("Stripe signature verification failed: %w", err)
	}

	// Parse event from raw body
	var event struct {
		Type string         `json:"type"`
		Data struct {
			Object json.RawMessage `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(req.RawBody, &event); err != nil {
		return nil, fmt.Errorf("failed to parse Stripe event: %w", err)
	}

	var obj map[string]any
	if err := json.Unmarshal(event.Data.Object, &obj); err != nil {
		return nil, fmt.Errorf("failed to parse Stripe event object: %w", err)
	}

	switch event.Type {
	case "checkout.session.completed":
		return p.handleCheckoutCompleted(obj)
	case "payment_intent.payment_failed":
		return p.handlePaymentFailed(obj)
	default:
		return nil, fmt.Errorf("unhandled Stripe event type: %s", event.Type)
	}
}

func (p *StripeProvider) handleCheckoutCompleted(obj map[string]any) (*WebhookEvent, error) {
	mode, _ := obj["mode"].(string)
	metadata, _ := obj["metadata"].(map[string]any)
	remoteID, _ := metadata["remote_id"].(string)

	switch mode {
	case "payment":
		piID, _ := obj["payment_intent"].(string)
		return &WebhookEvent{
			EventType:     EventPaymentCompleted,
			RemoteID:      remoteID,
			TransactionID: piID,
			RawPayload:    obj,
		}, nil

	case "setup":
		customerID, _ := obj["customer"].(string)
		setupIntentID, _ := obj["setup_intent"].(string)

		// Fetch setup intent to get payment method
		paymentMethodID := ""
		if setupIntentID != "" {
			si, err := p.get("/v1/setup_intents/" + setupIntentID)
			if err == nil {
				paymentMethodID, _ = si["payment_method"].(string)
			}
		}

		userUUID := customerID
		if paymentMethodID != "" {
			userUUID = customerID + ":" + paymentMethodID
		}

		return &WebhookEvent{
			EventType:  EventAuthorizationCompleted,
			RemoteID:   remoteID,
			UserUUID:   userUUID,
			RawPayload: obj,
		}, nil

	default:
		return nil, fmt.Errorf("unknown checkout session mode: %s", mode)
	}
}

func (p *StripeProvider) handlePaymentFailed(obj map[string]any) (*WebhookEvent, error) {
	metadata, _ := obj["metadata"].(map[string]any)
	remoteID, _ := metadata["remote_id"].(string)
	piID, _ := obj["id"].(string)

	return &WebhookEvent{
		EventType:     EventPaymentFailed,
		RemoteID:      remoteID,
		TransactionID: piID,
		RawPayload:    obj,
	}, nil
}

// verifySignature verifies the Stripe-Signature header using HMAC-SHA256.
func (p *StripeProvider) verifySignature(sigHeader string, body []byte) error {
	// Parse "t=timestamp,v1=signature"
	var timestamp, sig string
	for _, part := range strings.Split(sigHeader, ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			timestamp = kv[1]
		case "v1":
			sig = kv[1]
		}
	}

	if timestamp == "" || sig == "" {
		return fmt.Errorf("invalid signature header format")
	}

	// Compute expected signature: HMAC-SHA256(timestamp + "." + body, webhookSecret)
	payload := timestamp + "." + string(body)
	mac := hmac.New(sha256.New, []byte(p.WebhookSecret))
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return fmt.Errorf("signature mismatch")
	}

	// Reject stale webhooks to prevent replay attacks
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid signature timestamp")
	}
	if time.Since(time.Unix(ts, 0)) > 5*time.Minute {
		return fmt.Errorf("webhook timestamp too old")
	}

	return nil
}

// ── HTTP helpers ─────────────────────────────────────────────────────

func (p *StripeProvider) post(path string, params url.Values) (map[string]any, error) {
	req, err := http.NewRequest("POST", stripeBaseURL+path, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.SecretKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	var data map[string]any
	if err := json.Unmarshal(respBody, &data); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %s", string(respBody))
	}

	if resp.StatusCode >= 400 {
		errObj, _ := data["error"].(map[string]any)
		msg, _ := errObj["message"].(string)
		if msg == "" {
			msg = fmt.Sprintf("Stripe API error (HTTP %d)", resp.StatusCode)
		}
		return nil, fmt.Errorf("Stripe: %s", msg)
	}

	return data, nil
}

func (p *StripeProvider) get(path string) (map[string]any, error) {
	req, err := http.NewRequest("GET", stripeBaseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.SecretKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	var data map[string]any
	if err := json.Unmarshal(respBody, &data); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %s", string(respBody))
	}

	if resp.StatusCode >= 400 {
		errObj, _ := data["error"].(map[string]any)
		msg, _ := errObj["message"].(string)
		if msg == "" {
			msg = fmt.Sprintf("Stripe API error (HTTP %d)", resp.StatusCode)
		}
		return nil, fmt.Errorf("Stripe: %s", msg)
	}

	return data, nil
}
