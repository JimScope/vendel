package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// SendWebhookForMessage delivers a webhook HTTP POST for an SMS message.
func SendWebhookForMessage(app core.App, webhook *core.Record, message *core.Record) error {
	if !webhook.GetBool("active") {
		return fmt.Errorf("webhook inactive")
	}

	payload := map[string]any{
		"event":      "sms_received",
		"from":       message.GetString("from_number"),
		"body":       message.GetString("body"),
		"timestamp":  message.GetString("created"),
		"message_id": message.Id,
	}

	payloadJSON, err := marshalSorted(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// HMAC-SHA256 signature if secret is configured
	secretKey := webhook.GetString("secret_key")
	if secretKey != "" {
		sig := generateHMAC(secretKey, string(payloadJSON))
		headers["X-Webhook-Signature"] = sig
	}

	// Get timeout from system config
	timeout := 10
	configs, err := app.FindRecordsByFilter("system_config", "key = 'webhook_timeout'", "", 1, 0)
	if err == nil && len(configs) > 0 {
		if t := configs[0].GetInt("value"); t > 0 {
			timeout = t
		}
	}

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	req, err := http.NewRequest("POST", webhook.GetString("url"), bytes.NewReader(payloadJSON))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("webhook returned %d: %s", resp.StatusCode, string(body))
	}

	// Mark message as webhook_sent
	message.Set("webhook_sent", true)
	if err := app.Save(message); err != nil {
		log.Printf("WARNING: failed to update webhook_sent: %v", err)
	}

	return nil
}

// VerifyWebhookSignature verifies an HMAC-SHA256 signature.
func VerifyWebhookSignature(secretKey, payload, signature string) bool {
	expected := generateHMAC(secretKey, payload)
	return hmac.Equal([]byte(expected), []byte(signature))
}

func generateHMAC(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// marshalSorted produces deterministic JSON with sorted keys.
func marshalSorted(m map[string]any) ([]byte, error) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		keyJSON, _ := json.Marshal(k)
		buf.Write(keyJSON)
		buf.WriteByte(':')
		valJSON, _ := json.Marshal(m[k])
		// Remove trailing whitespace for compact output
		buf.Write(bytes.TrimRight(valJSON, " "))
	}
	buf.WriteByte('}')

	// Verify it's valid JSON
	var check map[string]any
	if err := json.Unmarshal(buf.Bytes(), &check); err != nil {
		// Fall back to standard marshal
		return json.Marshal(m)
	}

	return buf.Bytes(), nil
}

// GetSystemConfigValue reads a value from the system_config collection.
func GetSystemConfigValue(app core.App, key string) string {
	records, err := app.FindRecordsByFilter(
		"system_config",
		"key = {:key}",
		"", 1, 0,
		map[string]any{"key": key},
	)
	if err != nil || len(records) == 0 {
		return ""
	}
	return records[0].GetString("value")
}

// GetAppSettings returns public app settings.
func GetAppSettings(app core.App) map[string]string {
	keys := []string{"app_name", "support_email"}
	result := make(map[string]string)
	for _, k := range keys {
		result[k] = GetSystemConfigValue(app, k)
	}
	// Fill defaults
	if result["app_name"] == "" {
		result["app_name"] = "Ender"
	}
	// Add maintenance status
	if strings.ToLower(GetSystemConfigValue(app, "maintenance_mode")) == "true" {
		result["maintenance_mode"] = "true"
	} else {
		result["maintenance_mode"] = "false"
	}
	return result
}
