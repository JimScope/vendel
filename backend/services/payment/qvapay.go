package payment

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const qvaPayBaseURL = "https://api.qvapay.com/v2"

// QvaPayProvider implements the Provider interface for QvaPay API v2.
type QvaPayProvider struct {
	AppID     string
	AppSecret string
	client    *http.Client
}

// NewQvaPayProvider creates a new QvaPay provider from environment.
func NewQvaPayProvider() *QvaPayProvider {
	return &QvaPayProvider{
		AppID:     os.Getenv("QVAPAY_APP_ID"),
		AppSecret: os.Getenv("QVAPAY_APP_SECRET"),
		client:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *QvaPayProvider) Name() string        { return "qvapay" }
func (p *QvaPayProvider) DisplayName() string { return "QvaPay" }
func (p *QvaPayProvider) IsConfigured() bool  { return p.AppID != "" && p.AppSecret != "" }

func (p *QvaPayProvider) headers() map[string]string {
	return map[string]string{
		"app-id":       p.AppID,
		"app-secret":   p.AppSecret,
		"Content-Type": "application/json",
	}
}

func (p *QvaPayProvider) CreateInvoice(req InvoiceRequest) (*InvoiceResult, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("QvaPay is not configured")
	}

	payload := map[string]any{
		"amount":      req.Amount,
		"description": req.Description,
		"remote_id":   req.RemoteID,
	}
	if req.WebhookURL != "" {
		payload["webhook"] = req.WebhookURL
	}

	data, err := p.post("/create_invoice", payload)
	if err != nil {
		return nil, err
	}

	url, _ := data["url"].(string)
	if url == "" {
		errMsg := getStringKey(data, "error", "message")
		return nil, fmt.Errorf("QvaPay create_invoice failed: %s", errMsg)
	}

	invoiceID := getAnyStringKey(data, "transaction_uuid", "transation_uuid", "uuid")
	return &InvoiceResult{
		InvoiceID:  invoiceID,
		PaymentURL: url,
	}, nil
}

func (p *QvaPayProvider) GetAuthorizationURL(req AuthorizationRequest) (*AuthorizationResult, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("QvaPay is not configured")
	}

	data, err := p.post("/authorize_payments", map[string]any{
		"remote_id": req.RemoteID,
		"callback":  req.CallbackURL,
		"success":   req.SuccessURL,
		"error":     req.ErrorURL,
	})
	if err != nil {
		return nil, err
	}

	url := getAnyStringKey(data, "url", "authorization_url")
	if url == "" {
		errMsg := getStringKey(data, "error", "message")
		return nil, fmt.Errorf("QvaPay authorize_payments failed: %s", errMsg)
	}

	return &AuthorizationResult{AuthorizationURL: url}, nil
}

func (p *QvaPayProvider) ChargeAuthorizedUser(req ChargeRequest) (*ChargeResult, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("QvaPay is not configured")
	}

	data, err := p.post("/charge", map[string]any{
		"amount":      req.Amount,
		"user_uuid":   req.UserUUID,
		"description": req.Description,
		"remote_id":   req.RemoteID,
	})
	if err != nil {
		return nil, err
	}

	txID := getAnyStringKey(data, "transaction_uuid", "transation_uuid", "uuid")
	if txID == "" {
		errMsg := getStringKey(data, "error", "message")
		return nil, fmt.Errorf("QvaPay charge failed: %s", errMsg)
	}

	return &ChargeResult{
		TransactionID: txID,
		Amount:        req.Amount,
	}, nil
}

func (p *QvaPayProvider) ParseWebhook(req WebhookRequest) (*WebhookEvent, error) {
	payload := req.Payload

	// Authorization callback: has user_uuid + remote_id
	if userUUID, ok := payload["user_uuid"]; ok {
		remoteID, _ := payload["remote_id"]
		return &WebhookEvent{
			EventType:  EventAuthorizationCompleted,
			RemoteID:   fmt.Sprintf("%v", remoteID),
			UserUUID:   fmt.Sprintf("%v", userUUID),
			RawPayload: payload,
		}, nil
	}

	// Invoice payment webhook: has transaction_uuid
	txUUID := getAnyStringKey(payload, "transaction_uuid", "transation_uuid")
	if txUUID != "" {
		remoteID, _ := payload["remote_id"]
		if remoteID == nil {
			return nil, fmt.Errorf("QvaPay webhook missing remote_id")
		}

		var amount float64
		if a, ok := payload["amount"]; ok {
			switch v := a.(type) {
			case float64:
				amount = v
			case string:
				fmt.Sscanf(v, "%f", &amount)
			}
		}

		return &WebhookEvent{
			EventType:     EventPaymentCompleted,
			RemoteID:      fmt.Sprintf("%v", remoteID),
			TransactionID: txUUID,
			Amount:        amount,
			RawPayload:    payload,
		}, nil
	}

	return nil, fmt.Errorf("unrecognized QvaPay webhook payload")
}

// ── HTTP helpers ─────────────────────────────────────────────────────

func (p *QvaPayProvider) post(path string, payload map[string]any) (map[string]any, error) {
	bodyJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", qvaPayBaseURL+path, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, err
	}

	for k, v := range p.headers() {
		req.Header.Set(k, v)
	}

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

	return data, nil
}

// getAnyStringKey returns the string value of the first matching key.
func getAnyStringKey(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

// getStringKey returns the string value of the first matching key that has a non-empty value.
func getStringKey(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			s := fmt.Sprintf("%v", v)
			if s != "" && s != "<nil>" {
				return s
			}
		}
	}
	return "Unknown error"
}
