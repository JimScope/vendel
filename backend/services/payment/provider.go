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

// AuthorizationRequest represents a request for payment authorization.
type AuthorizationRequest struct {
	RemoteID    string
	CallbackURL string
	SuccessURL  string
	ErrorURL    string
}

// AuthorizationResult is the result of requesting authorization.
type AuthorizationResult struct {
	AuthorizationURL string
}

// ChargeRequest represents a request to charge an authorized user.
type ChargeRequest struct {
	UserUUID    string
	Amount      float64
	Currency    string
	Description string
	RemoteID    string
}

// ChargeResult is the result of charging a user.
type ChargeResult struct {
	TransactionID string
	Amount        float64
}

// WebhookEventType represents types of payment webhook events.
type WebhookEventType string

const (
	EventPaymentCompleted       WebhookEventType = "payment_completed"
	EventAuthorizationCompleted WebhookEventType = "authorization_completed"
	EventPaymentFailed          WebhookEventType = "payment_failed"
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
	UserUUID      string
	Amount        float64
	RawPayload    map[string]any
}

// Provider is the interface all payment providers must implement.
type Provider interface {
	Name() string
	DisplayName() string
	IsConfigured() bool
	CreateInvoice(req InvoiceRequest) (*InvoiceResult, error)
	GetAuthorizationURL(req AuthorizationRequest) (*AuthorizationResult, error)
	ChargeAuthorizedUser(req ChargeRequest) (*ChargeResult, error)
	ParseWebhook(req WebhookRequest) (*WebhookEvent, error)
}

// GetProviders returns all configured payment providers.
func GetProviders() []Provider {
	all := []Provider{
		NewQvaPayProvider(),
		NewStripeProvider(),
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
