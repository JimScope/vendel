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
	IsConfigured() bool
	CreateInvoice(req InvoiceRequest) (*InvoiceResult, error)
	GetAuthorizationURL(req AuthorizationRequest) (*AuthorizationResult, error)
	ChargeAuthorizedUser(req ChargeRequest) (*ChargeResult, error)
	ParseWebhook(payload map[string]any) *WebhookEvent
}

// GetProvider returns the configured payment provider.
func GetProvider() Provider {
	p := NewQvaPayProvider()
	if p.IsConfigured() {
		return p
	}
	return nil
}
