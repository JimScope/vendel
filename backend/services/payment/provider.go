package payment

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

// GetProviders returns all configured payment providers.
func GetProviders() []Provider {
	all := []Provider{
		NewQvaPayProvider(),
		NewStripeProvider(),
		NewTronDealerProvider(),
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

// GetDefaultProvider returns the first configured provider (backwards compat).
func GetDefaultProvider() Provider {
	providers := GetProviders()
	if len(providers) > 0 {
		return providers[0]
	}
	return nil
}
