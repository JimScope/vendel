package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
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

// NewTronDealerProvider creates a new TronDealer provider from environment.
func NewTronDealerProvider() *TronDealerProvider {
	apiURL := os.Getenv("TRONDEALER_API_URL")
	if apiURL == "" {
		apiURL = defaultTronDealerBaseURL
	}
	return &TronDealerProvider{
		APIKey:        os.Getenv("TRONDEALER_API_KEY"),
		APIURL:        strings.TrimRight(apiURL, "/"),
		WebhookSecret: os.Getenv("TRONDEALER_WEBHOOK_SECRET"),
		client:        &http.Client{Timeout: 30 * time.Second},
	}
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
// webhook payload into a WebhookEvent. TronDealer sends transaction data
// when a deposit is confirmed on-chain.
//
// Expected payload:
//
//	{
//	  "event": "transaction.confirmed",
//	  "transaction": {
//	    "tx_hash": "0x...",
//	    "from_address": "0x...",
//	    "to_address": "0x...",
//	    "asset": "USDT",
//	    "amount": "100.00",
//	    "status": "confirmed",
//	    ...
//	  },
//	  "wallet": {
//	    "id": "uuid",
//	    "address": "0x...",
//	    "label": "remote-id-here"
//	  }
//	}
func (p *TronDealerProvider) ParseWebhook(req WebhookRequest) (*WebhookEvent, error) {
	// Verify HMAC-SHA256 signature
	signature := req.Headers["X-Webhook-Signature"]
	if signature == "" {
		signature = req.Headers["x-webhook-signature"]
	}
	if signature == "" {
		return nil, fmt.Errorf("missing X-Webhook-Signature header")
	}

	if err := p.verifySignature(signature, req.RawBody); err != nil {
		return nil, fmt.Errorf("TronDealer signature verification failed: %w", err)
	}

	// Parse payload from raw body
	var payload map[string]any
	if err := json.Unmarshal(req.RawBody, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse TronDealer webhook: %w", err)
	}

	eventType, _ := payload["event"].(string)

	switch eventType {
	case "transaction.confirmed":
		return p.handleTransactionConfirmed(payload)
	case "transaction.failed":
		return p.handleTransactionFailed(payload)
	default:
		return nil, fmt.Errorf("unrecognized TronDealer event type: %s", eventType)
	}
}

func (p *TronDealerProvider) handleTransactionConfirmed(payload map[string]any) (*WebhookEvent, error) {
	tx, _ := payload["transaction"].(map[string]any)
	wallet, _ := payload["wallet"].(map[string]any)
	if tx == nil || wallet == nil {
		return nil, fmt.Errorf("TronDealer webhook missing transaction or wallet data")
	}

	// RemoteID is the wallet address — the handler uses it to look up the
	// user via user_balances.wallet_address.
	walletAddress, _ := wallet["address"].(string)
	if walletAddress == "" {
		return nil, fmt.Errorf("TronDealer webhook: wallet has no address")
	}

	txHash, _ := tx["tx_hash"].(string)
	asset, _ := tx["asset"].(string)
	amountStr, _ := tx["amount"].(string)
	var amount float64
	fmt.Sscanf(amountStr, "%f", &amount)

	return &WebhookEvent{
		EventType:     EventDepositReceived,
		RemoteID:      walletAddress,
		TransactionID: txHash,
		Amount:        amount,
		Asset:         asset,
		RawPayload:    payload,
	}, nil
}

func (p *TronDealerProvider) handleTransactionFailed(payload map[string]any) (*WebhookEvent, error) {
	wallet, _ := payload["wallet"].(map[string]any)
	remoteID := ""
	if wallet != nil {
		remoteID, _ = wallet["label"].(string)
	}

	tx, _ := payload["transaction"].(map[string]any)
	txHash := ""
	if tx != nil {
		txHash, _ = tx["tx_hash"].(string)
	}

	return &WebhookEvent{
		EventType:     EventPaymentFailed,
		RemoteID:      remoteID,
		TransactionID: txHash,
		RawPayload:    payload,
	}, nil
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
