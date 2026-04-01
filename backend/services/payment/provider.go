package payment

import (
	"net/http"
	"os"
	"time"
)

// InvoiceRequest represents a request to create a payment invoice.
type InvoiceRequest struct {
	Amount      float64
	Currency    string
	Description string
	RemoteID    string
	WebhookURL  string
	SuccessURL  string
	ErrorURL    string
}

// InvoiceResult is the result of creating an invoice.
type InvoiceResult struct {
	InvoiceID  string
	PaymentURL string
}

// WebhookEventType represents types of payment webhook events.
type WebhookEventType string

const (
	EventPaymentCompleted WebhookEventType = "payment_completed"
	EventPaymentFailed    WebhookEventType = "payment_failed"
	EventDepositReceived  WebhookEventType = "deposit_received"
)

// WebhookRequest carries the raw HTTP data needed by providers for webhook parsing.
type WebhookRequest struct {
	RawBody []byte
	Headers map[string]string
	Payload map[string]any // pre-parsed JSON, used by QvaPay
}

// WebhookEvent is a parsed webhook event from a payment provider.
type WebhookEvent struct {
	EventType     WebhookEventType
	RemoteID      string
	TransactionID string
	Amount        float64
	Asset         string // e.g. "USDT", "USDC" — set by deposit-based providers
	RawPayload    map[string]any
}

// Provider is the interface all payment providers must implement.
type Provider interface {
	Name() string
	DisplayName() string
	PaymentMethod() string // always "balance" — all providers feed user balance
	IsConfigured() bool
	CreateInvoice(req InvoiceRequest) (*InvoiceResult, error)
	ParseWebhook(req WebhookRequest) (*WebhookEvent, error)
}

// ConfigResolver reads config values (used for system_config DB fallback).
type ConfigResolver func(key string) string

// GetProviders returns all configured payment providers using env vars only.
func GetProviders() []Provider {
	return getProviders(nil)
}

// GetProvidersWithConfig returns all configured providers, using the resolver
// as fallback when env vars are not set. This allows admin UI configuration.
func GetProvidersWithConfig(resolve ConfigResolver) []Provider {
	return getProviders(resolve)
}

func getProviders(resolve ConfigResolver) []Provider {
	r := func(envKey, configKey string) string {
		v := os.Getenv(envKey)
		if v == "" && resolve != nil {
			v = resolve(configKey)
		}
		return v
	}

	all := []Provider{
		&QvaPayProvider{
			AppID:     r("QVAPAY_APP_ID", "qvapay_app_id"),
			AppSecret: r("QVAPAY_APP_SECRET", "qvapay_app_secret"),
			client:    &http.Client{Timeout: 30 * time.Second},
		},
		&StripeProvider{
			SecretKey:     r("STRIPE_SECRET_KEY", "stripe_secret_key"),
			WebhookSecret: r("STRIPE_WEBHOOK_SECRET", "stripe_webhook_secret"),
			client:        &http.Client{Timeout: 30 * time.Second},
		},
		func() *TronDealerProvider {
			apiURL := r("TRONDEALER_API_URL", "trondealer_api_url")
			if apiURL == "" {
				apiURL = defaultTronDealerBaseURL
			}
			return &TronDealerProvider{
				APIKey:        r("TRONDEALER_API_KEY", "trondealer_api_key"),
				APIURL:        apiURL,
				WebhookSecret: r("TRONDEALER_WEBHOOK_SECRET", "trondealer_webhook_secret"),
				client:        &http.Client{Timeout: 30 * time.Second},
			}
		}(),
	}

	var configured []Provider
	for _, p := range all {
		if p.IsConfigured() {
			configured = append(configured, p)
		}
	}
	return configured
}

// GetProvider returns a configured provider by name, or nil if not found/configured.
func GetProvider(name string) Provider {
	for _, p := range GetProviders() {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

// GetProviderWithConfig returns a provider by name using config fallback.
func GetProviderWithConfig(name string, resolve ConfigResolver) Provider {
	for _, p := range GetProvidersWithConfig(resolve) {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

// GetDefaultProvider returns the first configured provider.
func GetDefaultProvider() Provider {
	providers := GetProviders()
	if len(providers) > 0 {
		return providers[0]
	}
	return nil
}
