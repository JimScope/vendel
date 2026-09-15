package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"
)

// hmacHex computes the lowercase hex HMAC-SHA256 of body with the given secret.
func hmacHex(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// stripeSigHeader builds a valid "t=<ts>,v1=<sig>" header for the given body/secret.
func stripeSigHeader(secret string, ts int64, body []byte) string {
	signed := fmt.Sprintf("%d.%s", ts, string(body))
	sig := hmacHex(secret, []byte(signed))
	return fmt.Sprintf("t=%d,v1=%s", ts, sig)
}

// ── Stripe signature verification ────────────────────────────────────

func TestStripeVerifySignature(t *testing.T) {
	const secret = "whsec_test_secret"
	body := []byte(`{"type":"checkout.session.completed"}`)
	p := &StripeProvider{WebhookSecret: secret}

	now := time.Now().Unix()

	tests := []struct {
		name    string
		header  string
		body    []byte
		wantErr bool
	}{
		{
			name:    "valid signature accepted",
			header:  stripeSigHeader(secret, now, body),
			body:    body,
			wantErr: false,
		},
		{
			name:    "tampered body rejected",
			header:  stripeSigHeader(secret, now, body),
			body:    []byte(`{"type":"checkout.session.completed","evil":true}`),
			wantErr: true,
		},
		{
			name:    "wrong secret rejected",
			header:  stripeSigHeader("whsec_other_secret", now, body),
			body:    body,
			wantErr: true,
		},
		{
			name:    "replayed old timestamp rejected",
			header:  stripeSigHeader(secret, now-int64(10*time.Minute/time.Second), body),
			body:    body,
			wantErr: true,
		},
		{
			name:    "recent timestamp within window accepted",
			header:  stripeSigHeader(secret, now-60, body),
			body:    body,
			wantErr: false,
		},
		{
			name:    "missing v1 component rejected",
			header:  fmt.Sprintf("t=%d", now),
			body:    body,
			wantErr: true,
		},
		{
			name:    "missing t component rejected",
			header:  "v1=" + hmacHex(secret, body),
			body:    body,
			wantErr: true,
		},
		{
			name:    "non-numeric timestamp rejected",
			header:  "t=notanumber,v1=" + hmacHex(secret, []byte("notanumber."+string(body))),
			body:    body,
			wantErr: true,
		},
		{
			name:    "empty header rejected",
			header:  "",
			body:    body,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.verifySignature(tt.header, tt.body)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

// ── Stripe webhook parsing + cents rounding ──────────────────────────

func TestStripeParseWebhook(t *testing.T) {
	const secret = "whsec_test_secret"
	p := &StripeProvider{WebhookSecret: secret}

	// signWith builds a WebhookRequest whose header matches the body.
	signWith := func(body string) WebhookRequest {
		b := []byte(body)
		return WebhookRequest{
			RawBody: b,
			Headers: map[string]string{
				"Stripe-Signature": stripeSigHeader(secret, time.Now().Unix(), b),
			},
		}
	}

	t.Run("valid checkout completed extracts fields with cents conversion", func(t *testing.T) {
		body := `{"type":"checkout.session.completed","data":{"object":{"mode":"payment","payment_intent":"pi_123","amount_total":1999,"metadata":{"remote_id":"user_abc"}}}}`
		ev, err := p.ParseWebhook(signWith(body))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ev.EventType != EventPaymentCompleted {
			t.Errorf("EventType = %q, want %q", ev.EventType, EventPaymentCompleted)
		}
		if ev.RemoteID != "user_abc" {
			t.Errorf("RemoteID = %q, want user_abc", ev.RemoteID)
		}
		if ev.TransactionID != "pi_123" {
			t.Errorf("TransactionID = %q, want pi_123", ev.TransactionID)
		}
		// amount_total 1999 cents => 19.99
		if ev.Amount != 19.99 {
			t.Errorf("Amount = %v, want 19.99", ev.Amount)
		}
	})

	t.Run("payment failed event parsed", func(t *testing.T) {
		body := `{"type":"payment_intent.payment_failed","data":{"object":{"id":"pi_fail","metadata":{"remote_id":"user_x"}}}}`
		ev, err := p.ParseWebhook(signWith(body))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ev.EventType != EventPaymentFailed {
			t.Errorf("EventType = %q, want %q", ev.EventType, EventPaymentFailed)
		}
		if ev.TransactionID != "pi_fail" {
			t.Errorf("TransactionID = %q, want pi_fail", ev.TransactionID)
		}
	})

	t.Run("missing signature header rejected", func(t *testing.T) {
		body := `{"type":"checkout.session.completed","data":{"object":{"mode":"payment"}}}`
		_, err := p.ParseWebhook(WebhookRequest{RawBody: []byte(body), Headers: map[string]string{}})
		if err == nil {
			t.Fatal("expected error for missing signature")
		}
	})

	t.Run("invalid signature rejected", func(t *testing.T) {
		body := `{"type":"checkout.session.completed","data":{"object":{"mode":"payment"}}}`
		req := WebhookRequest{
			RawBody: []byte(body),
			Headers: map[string]string{"Stripe-Signature": fmt.Sprintf("t=%d,v1=deadbeef", time.Now().Unix())},
		}
		_, err := p.ParseWebhook(req)
		if err == nil {
			t.Fatal("expected error for invalid signature")
		}
	})

	t.Run("unhandled event type rejected", func(t *testing.T) {
		body := `{"type":"customer.created","data":{"object":{}}}`
		_, err := p.ParseWebhook(signWith(body))
		if err == nil {
			t.Fatal("expected error for unhandled event type")
		}
	})

	t.Run("malformed JSON body rejected", func(t *testing.T) {
		body := `{not valid json`
		_, err := p.ParseWebhook(signWith(body))
		if err == nil {
			t.Fatal("expected error for malformed JSON")
		}
	})

	t.Run("non-payment checkout mode rejected", func(t *testing.T) {
		body := `{"type":"checkout.session.completed","data":{"object":{"mode":"subscription"}}}`
		_, err := p.ParseWebhook(signWith(body))
		if err == nil {
			t.Fatal("expected error for non-payment mode")
		}
	})

	t.Run("lowercase signature header key accepted", func(t *testing.T) {
		body := `{"type":"checkout.session.completed","data":{"object":{"mode":"payment","payment_intent":"pi_lc","amount_total":500,"metadata":{"remote_id":"u1"}}}}`
		b := []byte(body)
		req := WebhookRequest{
			RawBody: b,
			Headers: map[string]string{"stripe-signature": stripeSigHeader(secret, time.Now().Unix(), b)},
		}
		ev, err := p.ParseWebhook(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ev.Amount != 5.0 {
			t.Errorf("Amount = %v, want 5.0", ev.Amount)
		}
	})
}

// ── TronDealer signature verification ────────────────────────────────

func TestTronDealerVerifySignature(t *testing.T) {
	const secret = "trondealer_secret"
	p := &TronDealerProvider{WebhookSecret: secret}
	body := []byte(`{"event":"transaction.confirmed","data":{"tx_hash":"0xabc"}}`)

	tests := []struct {
		name      string
		signature string
		body      []byte
		wantErr   bool
	}{
		{"valid signature accepted", hmacHex(secret, body), body, false},
		{"tampered body rejected", hmacHex(secret, body), []byte(`{"event":"x"}`), true},
		{"wrong secret rejected", hmacHex("other", body), body, true},
		{"empty signature rejected", "", body, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.verifySignature(tt.signature, tt.body)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

// ── TronDealer webhook parsing ───────────────────────────────────────

func TestTronDealerParseWebhook(t *testing.T) {
	const secret = "trondealer_secret"
	p := &TronDealerProvider{WebhookSecret: secret}

	sign := func(body string, prefix bool) WebhookRequest {
		b := []byte(body)
		sig := hmacHex(secret, b)
		if prefix {
			sig = "sha256=" + sig
		}
		return WebhookRequest{
			RawBody: b,
			Headers: map[string]string{"X-Signature-256": sig},
		}
	}

	t.Run("valid confirmed deposit extracts fields", func(t *testing.T) {
		body := `{"event":"transaction.confirmed","data":{"tx_hash":"0xdeadbeef","to_address":"0xWALLET","asset":"USDT","amount":"100.50"}}`
		ev, err := p.ParseWebhook(sign(body, true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ev.EventType != EventDepositReceived {
			t.Errorf("EventType = %q, want %q", ev.EventType, EventDepositReceived)
		}
		if ev.RemoteID != "0xWALLET" {
			t.Errorf("RemoteID = %q, want 0xWALLET", ev.RemoteID)
		}
		if ev.TransactionID != "0xdeadbeef" {
			t.Errorf("TransactionID = %q, want 0xdeadbeef", ev.TransactionID)
		}
		if ev.Asset != "USDT" {
			t.Errorf("Asset = %q, want USDT", ev.Asset)
		}
		if ev.Amount != 100.50 {
			t.Errorf("Amount = %v, want 100.50", ev.Amount)
		}
	})

	t.Run("amount as JSON number parsed", func(t *testing.T) {
		body := `{"event":"transaction.confirmed","data":{"tx_hash":"0x1","to_address":"0xW","asset":"USDC","amount":250}}`
		ev, err := p.ParseWebhook(sign(body, true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ev.Amount != 250 {
			t.Errorf("Amount = %v, want 250", ev.Amount)
		}
	})

	t.Run("signature without sha256 prefix accepted", func(t *testing.T) {
		body := `{"event":"transaction.confirmed","data":{"tx_hash":"0x2","to_address":"0xW","amount":"1"}}`
		_, err := p.ParseWebhook(sign(body, false))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing signature header rejected", func(t *testing.T) {
		body := `{"event":"transaction.confirmed","data":{"to_address":"0xW"}}`
		_, err := p.ParseWebhook(WebhookRequest{RawBody: []byte(body), Headers: map[string]string{}})
		if err == nil {
			t.Fatal("expected error for missing signature")
		}
	})

	t.Run("invalid signature rejected", func(t *testing.T) {
		body := `{"event":"transaction.confirmed","data":{"to_address":"0xW"}}`
		req := WebhookRequest{
			RawBody: []byte(body),
			Headers: map[string]string{"X-Signature-256": "sha256=deadbeef"},
		}
		_, err := p.ParseWebhook(req)
		if err == nil {
			t.Fatal("expected error for invalid signature")
		}
	})

	t.Run("missing data field rejected", func(t *testing.T) {
		body := `{"event":"transaction.confirmed"}`
		_, err := p.ParseWebhook(sign(body, true))
		if err == nil {
			t.Fatal("expected error for missing data")
		}
	})

	t.Run("missing to_address rejected", func(t *testing.T) {
		body := `{"event":"transaction.confirmed","data":{"tx_hash":"0x1","amount":"1"}}`
		_, err := p.ParseWebhook(sign(body, true))
		if err == nil {
			t.Fatal("expected error for missing to_address")
		}
	})

	t.Run("unhandled event rejected", func(t *testing.T) {
		body := `{"event":"transaction.pending","data":{"to_address":"0xW"}}`
		_, err := p.ParseWebhook(sign(body, true))
		if err == nil {
			t.Fatal("expected error for unhandled event")
		}
	})

	t.Run("malformed JSON rejected", func(t *testing.T) {
		body := `{bad json`
		_, err := p.ParseWebhook(sign(body, true))
		if err == nil {
			t.Fatal("expected error for malformed JSON")
		}
	})
}

// ── QvaPay webhook parsing (pre-network validation only) ─────────────
//
// QvaPay's ParseWebhook calls verifyTransaction (a live API round-trip)
// once a transaction_uuid and remote_id are present, so only the paths
// that return before that call are unit-testable without a network.

func TestQvaPayParseWebhookValidation(t *testing.T) {
	p := &QvaPayProvider{AppID: "id", AppSecret: "secret"}

	t.Run("payload without transaction uuid rejected", func(t *testing.T) {
		_, err := p.ParseWebhook(WebhookRequest{Payload: map[string]any{"foo": "bar"}})
		if err == nil {
			t.Fatal("expected error for unrecognized payload")
		}
		if !strings.Contains(err.Error(), "unrecognized") {
			t.Errorf("error = %v, want unrecognized payload", err)
		}
	})

	t.Run("transaction uuid but missing remote_id rejected", func(t *testing.T) {
		_, err := p.ParseWebhook(WebhookRequest{Payload: map[string]any{
			"transaction_uuid": "tx-123",
			"amount":           "10.00",
		}})
		if err == nil {
			t.Fatal("expected error for missing remote_id")
		}
		if !strings.Contains(err.Error(), "remote_id") {
			t.Errorf("error = %v, want missing remote_id", err)
		}
	})

	t.Run("misspelled transation_uuid key still recognized", func(t *testing.T) {
		// The provider tolerates QvaPay's historical "transation_uuid" typo;
		// with no remote_id it must still fail at the remote_id guard (not the
		// unrecognized-payload guard), proving the key was picked up.
		_, err := p.ParseWebhook(WebhookRequest{Payload: map[string]any{
			"transation_uuid": "tx-typo",
		}})
		if err == nil || !strings.Contains(err.Error(), "remote_id") {
			t.Fatalf("error = %v, want missing remote_id", err)
		}
	})
}

// ── Shared helper functions ──────────────────────────────────────────

func TestGetAnyStringKey(t *testing.T) {
	m := map[string]any{
		"a":   "first",
		"num": 42,
		"b":   "second",
	}
	tests := []struct {
		name string
		keys []string
		want string
	}{
		{"first match wins", []string{"a", "b"}, "first"},
		{"fallback to second key", []string{"missing", "b"}, "second"},
		{"numeric value stringified", []string{"num"}, "42"},
		{"no match returns empty", []string{"x", "y"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAnyStringKey(m, tt.keys...); got != tt.want {
				t.Errorf("getAnyStringKey(%v) = %q, want %q", tt.keys, got, tt.want)
			}
		})
	}
}

func TestGetStringKey(t *testing.T) {
	m := map[string]any{
		"empty": "",
		"nilv":  nil,
		"msg":   "boom",
	}
	tests := []struct {
		name string
		keys []string
		want string
	}{
		{"returns non-empty value", []string{"msg"}, "boom"},
		{"skips empty string", []string{"empty", "msg"}, "boom"},
		{"skips nil value", []string{"nilv", "msg"}, "boom"},
		{"all missing returns default", []string{"x"}, "Unknown error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStringKey(m, tt.keys...); got != tt.want {
				t.Errorf("getStringKey(%v) = %q, want %q", tt.keys, got, tt.want)
			}
		})
	}
}
