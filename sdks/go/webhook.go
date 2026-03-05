package vendel

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// VerifyWebhookSignature verifies a Vendel webhook X-Webhook-Signature header.
//
// The signature is an HMAC-SHA256 hex digest computed over the raw JSON
// payload string using the webhook secret as the key.
//
// Pass the raw request body as payload — do not re-marshal it.
func VerifyWebhookSignature(payload, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
