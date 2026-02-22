package payment

import (
	"fmt"
	"os"
)

const (
	tropipaySandboxURL    = "https://tropipay-dev.herokuapp.com/api/v2"
	tropipayProductionURL = "https://www.tropipay.com/api/v2"
)

// TropipayProvider implements the Provider interface for Tropipay (placeholder).
type TropipayProvider struct {
	ClientID     string
	ClientSecret string
	Environment  string
}

// NewTropipayProvider creates a new Tropipay provider from environment.
func NewTropipayProvider() *TropipayProvider {
	env := os.Getenv("TROPIPAY_ENVIRONMENT")
	if env == "" {
		env = "sandbox"
	}
	return &TropipayProvider{
		ClientID:     os.Getenv("TROPIPAY_CLIENT_ID"),
		ClientSecret: os.Getenv("TROPIPAY_CLIENT_SECRET"),
		Environment:  env,
	}
}

func (p *TropipayProvider) Name() string      { return "tropipay" }
func (p *TropipayProvider) IsConfigured() bool { return p.ClientID != "" && p.ClientSecret != "" }

func (p *TropipayProvider) baseURL() string {
	if p.Environment == "production" {
		return tropipayProductionURL
	}
	return tropipaySandboxURL
}

func (p *TropipayProvider) CreateInvoice(req InvoiceRequest) (*InvoiceResult, error) {
	// TODO: Implement OAuth2 token flow + POST /api/v2/paymentcards
	return nil, fmt.Errorf("Tropipay provider not yet implemented")
}

func (p *TropipayProvider) GetAuthorizationURL(req AuthorizationRequest) (*AuthorizationResult, error) {
	return nil, fmt.Errorf("Tropipay provider does not support authorized payments")
}

func (p *TropipayProvider) ChargeAuthorizedUser(req ChargeRequest) (*ChargeResult, error) {
	return nil, fmt.Errorf("Tropipay provider does not support authorized payments")
}

func (p *TropipayProvider) ParseWebhook(payload map[string]any) *WebhookEvent {
	// TODO: Implement Tropipay webhook parsing
	return nil
}
