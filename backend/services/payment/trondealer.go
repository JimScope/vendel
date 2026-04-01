package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultTronDealerBaseURL = "https://trondealer.com"

// TronDealerProvider implements the Provider interface for TronDealer's
// blockchain wallet-based payment system. Instead of creating traditional
// invoices, it assigns BSC wallets and monitors them for stablecoin deposits.
type TronDealerProvider struct {
	APIKey        string
	APIURL        string
	WebhookSecret string
	client        *http.Client
}


func (p *TronDealerProvider) Name() string          { return "trondealer" }
func (p *TronDealerProvider) DisplayName() string   { return "TronDealer" }
func (p *TronDealerProvider) PaymentMethod() string { return "balance" }
func (p *TronDealerProvider) IsConfigured() bool {
	return p.APIKey != "" && p.WebhookSecret != ""
}

// CreateInvoice assigns a new BSC wallet labeled with the RemoteID.
// The returned PaymentURL is the wallet address where the user should
// send USDT/USDC. The InvoiceID is the wallet UUID from TronDealer.
func (p *TronDealerProvider) CreateInvoice(req InvoiceRequest) (*InvoiceResult, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("TronDealer is not configured")
	}

	data, err := p.post("/api/v2/wallets/assign", map[string]any{
		"label": req.RemoteID,
	})
	if err != nil {
		return nil, err
	}

	success, _ := data["success"].(bool)
	if !success {
		errMsg, _ := data["error"].(string)
		if errMsg == "" {
			errMsg = "unknown error"
		}
		return nil, fmt.Errorf("TronDealer assign wallet failed: %s", errMsg)
	}

	wallet, _ := data["wallet"].(map[string]any)
	if wallet == nil {
		return nil, fmt.Errorf("TronDealer assign wallet: missing wallet in response")
	}

	walletID, _ := wallet["id"].(string)
	address, _ := wallet["address"].(string)
	if address == "" {
		return nil, fmt.Errorf("TronDealer assign wallet: missing address")
	}

	return &InvoiceResult{
		InvoiceID:  walletID,
		PaymentURL: address,
	}, nil
}

// ParseWebhook verifies the HMAC-SHA256 signature and parses a TronDealer
// webhook payload. The signature header is x-signature-256 with format
// "sha256=<hex>". The payload structure is:
//
//	{
//	  "event": "transaction.confirmed",
//	  "timestamp": "...",
//	  "data": {
//	    "tx_hash": "0x...",
//	    "to_address": "0x...",
//	    "asset": "USDT",
//	    "amount": "100.00",
//	    "wallet_label": "...",
//	    "network": "bsc",
//	    ...
//	  }
//	}
func (p *TronDealerProvider) ParseWebhook(req WebhookRequest) (*WebhookEvent, error) {
	// Verify HMAC-SHA256 signature (header: x-signature-256, format: sha256=<hex>)
	signature := req.Headers["X-Signature-256"]
	if signature == "" {
		signature = req.Headers["x-signature-256"]
	}
	if signature == "" {
		return nil, fmt.Errorf("missing x-signature-256 header")
	}

	// Strip "sha256=" prefix
	signature = strings.TrimPrefix(signature, "sha256=")

	if err := p.verifySignature(signature, req.RawBody); err != nil {
		return nil, fmt.Errorf("TronDealer signature verification failed: %w", err)
	}

	// Parse payload: {event, timestamp, data: {...}}
	var payload struct {
		Event string         `json:"event"`
		Data  map[string]any `json:"data"`
	}
	if err := json.Unmarshal(req.RawBody, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse TronDealer webhook: %w", err)
	}

	if payload.Data == nil {
		return nil, fmt.Errorf("TronDealer webhook: missing data field")
	}

	data := payload.Data
	txHash, _ := data["tx_hash"].(string)
	asset, _ := data["asset"].(string)
	toAddress, _ := data["to_address"].(string)

	var amount float64
	switch v := data["amount"].(type) {
	case float64:
		amount = v
	case string:
		fmt.Sscanf(v, "%f", &amount)
	}

	if toAddress == "" {
		return nil, fmt.Errorf("TronDealer webhook: missing to_address")
	}

	switch payload.Event {
	case "transaction.confirmed":
		return &WebhookEvent{
			EventType:     EventDepositReceived,
			RemoteID:      toAddress,
			TransactionID: txHash,
			Amount:        amount,
			Asset:         asset,
			RawPayload:    data,
		}, nil
	default:
		return nil, fmt.Errorf("unhandled TronDealer event: %s", payload.Event)
	}
}

// verifySignature checks the HMAC-SHA256 signature against the raw body.
func (p *TronDealerProvider) verifySignature(signature string, body []byte) error {
	mac := hmac.New(sha256.New, []byte(p.WebhookSecret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}

// ── HTTP helpers ─────────────────────────────────────────────────────

func (p *TronDealerProvider) post(path string, payload map[string]any) (map[string]any, error) {
	bodyJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", p.APIURL+path, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.APIKey)

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
		errMsg, _ := data["error"].(string)
		if errMsg == "" {
			errMsg = fmt.Sprintf("TronDealer API error (HTTP %d)", resp.StatusCode)
		}
		return nil, fmt.Errorf("TronDealer: %s", errMsg)
	}

	return data, nil
}
