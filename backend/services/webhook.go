package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"ender/services/payment"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

const webhookMaxRetries = 3

// webhookRetryBackoffs defines the delay before each retry attempt.
var webhookRetryBackoffs = []time.Duration{
	1 * time.Minute,  // after 1st failure
	5 * time.Minute,  // after 2nd failure
	15 * time.Minute, // after 3rd failure
}

// WebhookDeliveryResult holds the outcome of a webhook delivery attempt.
type WebhookDeliveryResult struct {
	LogRecord      *core.Record
	DeliveryStatus string
	ResponseStatus int
	DurationMs     int
	ErrorMessage   string
}

// SendWebhookForMessage delivers a webhook HTTP POST for an SMS message.
func SendWebhookForMessage(app core.App, webhook *core.Record, message *core.Record, event string) error {
	if !webhook.GetBool("active") {
		return fmt.Errorf("webhook inactive")
	}

	payload := map[string]any{
		"event":      event,
		"body":       message.GetString("body"),
		"message_id": message.Id,
		"timestamp":  message.GetString("created"),
	}

	switch event {
	case "sms_received":
		payload["from"] = message.GetString("from_number")
	case "sms_sent", "sms_delivered", "sms_failed":
		payload["to"] = message.GetString("to")
		payload["status"] = message.GetString("status")
		if v := message.GetString("error_message"); v != "" {
			payload["error_message"] = v
		}
		if v := message.GetString("sent_at"); v != "" {
			payload["sent_at"] = v
		}
		if v := message.GetString("delivered_at"); v != "" {
			payload["delivered_at"] = v
		}
	}

	result := deliverWebhook(app, webhook, payload, event)

	// Mark message as webhook_sent on success
	if result.DeliveryStatus == "success" {
		message.Set("webhook_sent", true)
		if err := app.Save(message); err != nil {
			app.Logger().Warn("failed to update webhook_sent", slog.Any("error", err))
		}
	}

	if result.DeliveryStatus == "failed" {
		return fmt.Errorf("webhook delivery failed: %s", result.ErrorMessage)
	}

	return nil
}

// SendTestWebhook sends a synthetic test payload to the webhook and returns the delivery result.
func SendTestWebhook(app core.App, webhook *core.Record) *WebhookDeliveryResult {
	payload := map[string]any{
		"event":      "test",
		"message_id": "test_" + GenerateSecureKey("", 12),
		"body":       "Test webhook from Ender",
		"from":       "+1234567890",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	return deliverWebhook(app, webhook, payload, "test")
}

// deliverWebhook performs the HTTP request, measures timing, and logs the delivery.
func deliverWebhook(app core.App, webhook *core.Record, payload map[string]any, event string) *WebhookDeliveryResult {
	url := webhook.GetString("url")

	payloadJSON, err := marshalSorted(payload)
	if err != nil {
		return logDelivery(app, webhook, event, url, payload, 0, "", "failed", fmt.Sprintf("marshal payload: %v", err), 0)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// HMAC-SHA256 signature if secret is configured
	secretKey := webhook.GetString("secret_key")
	if secretKey != "" {
		decrypted, err := DecryptSecret(secretKey)
		if err != nil {
			app.Logger().Warn("failed to decrypt webhook secret", slog.Any("error", err))
			decrypted = secretKey // fallback to raw value
		}
		sig := generateHMAC(decrypted, string(payloadJSON))
		headers["X-Webhook-Signature"] = sig
	}

	// Get timeout from system config (capped at 30s)
	timeout := 10
	config, err := app.FindFirstRecordByFilter("system_config", "key = 'webhook_timeout'")
	if err == nil && config != nil {
		if t := config.GetInt("value"); t > 0 {
			timeout = t
		}
	}
	if timeout > 30 {
		timeout = 30
	}

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	req, err := http.NewRequest("POST", url, bytes.NewReader(payloadJSON))
	if err != nil {
		return logDelivery(app, webhook, event, url, payload, 0, "", "failed", fmt.Sprintf("create request: %v", err), 0)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := client.Do(req)
	durationMs := int(time.Since(start).Milliseconds())

	if err != nil {
		return logDelivery(app, webhook, event, url, payload, 0, "", "failed", fmt.Sprintf("request failed: %v", err), durationMs)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	respBodyStr := string(respBody)
	if len(respBodyStr) > 2000 {
		respBodyStr = respBodyStr[:2000]
	}

	status := "success"
	errMsg := ""
	if resp.StatusCode >= 400 {
		status = "failed"
		errMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return logDelivery(app, webhook, event, url, payload, resp.StatusCode, respBodyStr, status, errMsg, durationMs)
}

// logDelivery creates a webhook_delivery_logs record and returns the result.
func logDelivery(app core.App, webhook *core.Record, event, url string, payload map[string]any, responseStatus int, responseBody, deliveryStatus, errMsg string, durationMs int) *WebhookDeliveryResult {
	result := &WebhookDeliveryResult{
		DeliveryStatus: deliveryStatus,
		ResponseStatus: responseStatus,
		DurationMs:     durationMs,
		ErrorMessage:   errMsg,
	}

	col, err := app.FindCollectionByNameOrId("webhook_delivery_logs")
	if err != nil {
		app.Logger().Warn("webhook_delivery_logs collection not found", slog.Any("error", err))
		return result
	}

	record := core.NewRecord(col)
	record.Set("webhook", webhook.Id)
	record.Set("event", event)
	record.Set("url", url)
	record.Set("request_body", payload)
	record.Set("response_status", responseStatus)
	record.Set("response_body", responseBody)
	record.Set("delivery_status", deliveryStatus)
	record.Set("error_message", errMsg)
	record.Set("duration_ms", durationMs)

	// Schedule first retry for initial failed deliveries
	if deliveryStatus == "failed" {
		record.Set("retry_count", 0)
		nextRetry := time.Now().UTC().Add(webhookRetryBackoffs[0])
		record.Set("next_retry_at", nextRetry.Format(time.RFC3339))
	}

	if err := app.Save(record); err != nil {
		app.Logger().Warn("failed to save webhook delivery log", slog.Any("error", err))
	} else {
		result.LogRecord = record
	}

	return result
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

// RetryFailedWebhooks retries failed webhook deliveries with exponential backoff.
func RetryFailedWebhooks(app core.App) error {
	if !webhookBreaker.Allow() {
		app.Logger().Warn("Webhook circuit breaker open, skipping webhook retries")
		return nil
	}

	now := time.Now().UTC().Format(time.RFC3339)

	records, err := app.FindRecordsByFilter(
		"webhook_delivery_logs",
		"delivery_status = 'failed' && next_retry_at != '' && next_retry_at <= {:now} && retry_count < {:maxRetries}",
		"-created", 50, 0,
		dbx.Params{"now": now, "maxRetries": webhookMaxRetries},
	)
	if err != nil {
		return err
	}

	retried := 0
	for _, record := range records {
		webhookId := record.GetString("webhook")
		webhook, err := app.FindRecordById("webhook_configs", webhookId)
		if err != nil {
			app.Logger().Warn("webhook config not found for retry", slog.String("webhook", webhookId))
			// Clear next_retry_at so we don't keep trying
			record.Set("next_retry_at", "")
			_ = app.Save(record)
			continue
		}

		if !webhook.GetBool("active") {
			record.Set("next_retry_at", "")
			_ = app.Save(record)
			continue
		}

		// Reconstruct payload from stored request_body
		var payload map[string]any
		raw := record.Get("request_body")
		switch v := raw.(type) {
		case map[string]any:
			payload = v
		case string:
			if err := json.Unmarshal([]byte(v), &payload); err != nil {
				app.Logger().Warn("failed to parse stored request_body", slog.Any("error", err))
				record.Set("next_retry_at", "")
				_ = app.Save(record)
				continue
			}
		default:
			rawJSON, _ := json.Marshal(raw)
			if err := json.Unmarshal(rawJSON, &payload); err != nil {
				record.Set("next_retry_at", "")
				_ = app.Save(record)
				continue
			}
		}

		event := record.GetString("event")
		result := deliverWebhook(app, webhook, payload, event)

		retryCount := record.GetInt("retry_count") + 1
		record.Set("retry_count", retryCount)

		if result.DeliveryStatus == "success" {
			record.Set("delivery_status", "success")
			record.Set("response_status", result.ResponseStatus)
			record.Set("next_retry_at", "")
			record.Set("error_message", "")
			record.Set("duration_ms", result.DurationMs)
			webhookBreaker.RecordSuccess()
		} else {
			record.Set("error_message", result.ErrorMessage)
			record.Set("duration_ms", result.DurationMs)
			if result.ResponseStatus > 0 {
				record.Set("response_status", result.ResponseStatus)
			}
			if retryCount < webhookMaxRetries {
				nextRetry := time.Now().UTC().Add(webhookRetryBackoffs[retryCount])
				record.Set("next_retry_at", nextRetry.Format(time.RFC3339))
			} else {
				record.Set("next_retry_at", "")
			}
			webhookBreaker.RecordFailure()
		}

		if err := app.Save(record); err != nil {
			app.Logger().Warn("failed to update webhook retry log", slog.Any("error", err))
		}
		retried++
	}

	if retried > 0 {
		app.Logger().Info("Retried failed webhooks", slog.Int("retried", retried))
	}
	return nil
}

// RetryWebhookDelivery manually retries a single failed webhook delivery.
func RetryWebhookDelivery(app core.App, logId string) (*WebhookDeliveryResult, error) {
	record, err := app.FindRecordById("webhook_delivery_logs", logId)
	if err != nil {
		return nil, fmt.Errorf("delivery log not found")
	}

	if record.GetString("delivery_status") != "failed" {
		return nil, fmt.Errorf("can only retry failed deliveries")
	}

	webhookId := record.GetString("webhook")
	webhook, err := app.FindRecordById("webhook_configs", webhookId)
	if err != nil {
		return nil, fmt.Errorf("webhook config not found")
	}

	// Parse stored payload
	var payload map[string]any
	raw := record.Get("request_body")
	switch v := raw.(type) {
	case map[string]any:
		payload = v
	case string:
		if err := json.Unmarshal([]byte(v), &payload); err != nil {
			return nil, fmt.Errorf("failed to parse stored request_body")
		}
	default:
		rawJSON, _ := json.Marshal(raw)
		if err := json.Unmarshal(rawJSON, &payload); err != nil {
			return nil, fmt.Errorf("failed to parse stored request_body")
		}
	}

	event := record.GetString("event")
	result := deliverWebhook(app, webhook, payload, event)

	// Link the new log entry to the original
	if result.LogRecord != nil {
		result.LogRecord.Set("original_log", logId)
		_ = app.Save(result.LogRecord)
	}

	// Clear auto-retry on the original to prevent duplicate retries
	record.Set("next_retry_at", "")
	_ = app.Save(record)

	return result, nil
}

// GetSystemConfigValue reads a value from the system_config collection.
func GetSystemConfigValue(app core.App, key string) string {
	record, err := app.FindFirstRecordByFilter(
		"system_config",
		"key = {:key}",
		dbx.Params{"key": key},
	)
	if err != nil || record == nil {
		return ""
	}
	return record.GetString("value")
}

// GetAppSettings returns public app settings.
func GetAppSettings(app core.App) map[string]any {
	keys := []string{"app_name", "support_email"}
	result := make(map[string]any)
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
	// Add configured payment providers
	var providers []map[string]string
	for _, p := range payment.GetProviders() {
		providers = append(providers, map[string]string{
			"name":         p.Name(),
			"display_name": p.DisplayName(),
		})
	}
	result["payment_providers"] = providers
	return result
}
