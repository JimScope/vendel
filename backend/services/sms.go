package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/routine"
	"github.com/pocketbase/pocketbase/tools/types"
)

// SendSMS orchestrates the entire SMS sending process.
// Recipients are distributed among devices via round-robin.
func SendSMS(app core.App, userId string, recipients []string, body string, deviceId string) ([]*core.Record, error) {
	count := len(recipients)
	if count == 0 {
		return nil, fmt.Errorf("no recipients provided")
	}

	// Check quota
	if err := CheckSMSQuota(app, userId, count); err != nil {
		return nil, err
	}

	// Determine devices
	var devices []*core.Record
	if deviceId != "" {
		device, err := app.FindRecordById("sms_devices", deviceId)
		if err != nil {
			return nil, fmt.Errorf("device not found: %w", err)
		}
		if device.GetString("user") != userId {
			return nil, fmt.Errorf("device does not belong to user")
		}
		devices = []*core.Record{device}
	} else {
		records, err := app.FindRecordsByFilter(
			"sms_devices",
			"user = {:userId} && (fcm_token != '' || device_type = 'modem')",
			"-created",
			0, 0,
			dbx.Params{"userId": userId},
		)
		if err == nil && len(records) > 0 {
			devices = records
		}
	}

	// Generate batch ID for bulk sends
	batchId := ""
	if count > 1 {
		batchId = core.GenerateDefaultRandomId()
	}

	// Create message records
	collection, err := app.FindCollectionByNameOrId("sms_messages")
	if err != nil {
		return nil, fmt.Errorf("sms_messages collection not found: %w", err)
	}

	var messages []*core.Record
	for i, recipient := range recipients {
		record := core.NewRecord(collection)
		record.Set("to", recipient)
		record.Set("body", body)
		record.Set("user", userId)
		record.Set("message_type", "outgoing")
		record.Set("webhook_sent", false)

		if batchId != "" {
			record.Set("batch_id", batchId)
		}

		if len(devices) > 0 {
			device := devices[i%len(devices)]
			record.Set("device", device.Id)
			record.Set("status", "assigned")
			record.Set("from_number", device.GetString("phone_number"))
		} else {
			record.Set("status", "pending")
		}

		if err := app.Save(record); err != nil {
			return nil, fmt.Errorf("failed to create message: %w", err)
		}
		messages = append(messages, record)
	}

	// Increment SMS count
	if err := IncrementSMSCount(app, userId, count); err != nil {
		app.Logger().Warn("failed to increment SMS count", slog.Any("error", err))
	}

	// Dispatch FCM notifications in background (replaces QStash)
	if len(devices) > 0 {
		routine.FireAndForget(func() { DispatchMessages(app, messages) })
	}

	return messages, nil
}

// ProcessSMSAck handles device acknowledgment for a sent SMS.
// The deviceId must match the message's assigned device (prevents cross-device manipulation).
func ProcessSMSAck(app core.App, deviceId string, messageId string, status string, errorMessage string) error {
	record, err := app.FindRecordById("sms_messages", messageId)
	if err != nil {
		return fmt.Errorf("message not found: %w", err)
	}

	if record.GetString("device") != deviceId {
		return fmt.Errorf("message does not belong to this device")
	}

	record.Set("status", status)
	if errorMessage != "" {
		record.Set("error_message", errorMessage)
	}
	if status == "sent" {
		record.Set("sent_at", types.NowDateTime())
	} else if status == "delivered" {
		record.Set("delivered_at", types.NowDateTime())
	}

	if err := app.Save(record); err != nil {
		return err
	}

	// Trigger webhooks for status transitions
	eventMap := map[string]string{"sent": "sms_sent", "delivered": "sms_delivered", "failed": "sms_failed"}
	if event, ok := eventMap[status]; ok {
		routine.FireAndForget(func() { TriggerWebhooks(app, record.GetString("user"), record, event) })
	}

	return nil
}

// HandleIncomingSMS processes an incoming SMS from a device and triggers webhooks.
// deviceId is used for deduplication and traceability.
func HandleIncomingSMS(app core.App, userId string, deviceId string, fromNumber string, body string, timestamp string) (*core.Record, error) {
	// Deduplicate: check for identical incoming SMS within the last 5 minutes
	cutoff := time.Now().UTC().Add(-5 * time.Minute).Format(time.RFC3339)
	existing, err := app.FindFirstRecordByFilter(
		"sms_messages",
		"message_type = 'incoming' && device = {:deviceId} && from_number = {:from} && body = {:body} && created >= {:cutoff}",
		dbx.Params{"deviceId": deviceId, "from": fromNumber, "body": body, "cutoff": cutoff},
	)
	if err == nil && existing != nil {
		return existing, nil
	}

	collection, err := app.FindCollectionByNameOrId("sms_messages")
	if err != nil {
		return nil, err
	}

	record := core.NewRecord(collection)
	record.Set("user", userId)
	record.Set("device", deviceId)
	record.Set("to", "")
	record.Set("from_number", fromNumber)
	record.Set("body", body)
	record.Set("status", "received")
	record.Set("message_type", "incoming")
	record.Set("webhook_sent", false)

	if err := app.Save(record); err != nil {
		return nil, err
	}

	// Trigger webhooks in background
	routine.FireAndForget(func() { TriggerWebhooks(app, userId, record, "sms_received") })

	return record, nil
}

var webhookBreaker = NewCircuitBreaker("webhook", 5, 60*time.Second)

// TriggerWebhooks finds active webhook configs for a user and fires matching webhooks.
func TriggerWebhooks(app core.App, userId string, message *core.Record, event string) {
	if !webhookBreaker.Allow() {
		app.Logger().Warn("Webhook circuit breaker open, skipping webhooks",
			slog.String("event", event), slog.String("user", userId))
		return
	}
	webhooks, err := app.FindRecordsByFilter(
		"webhook_configs",
		"user = {:userId} && active = true",
		"", 0, 0,
		dbx.Params{"userId": userId},
	)
	if err != nil || len(webhooks) == 0 {
		return
	}

	for _, wh := range webhooks {
		events := wh.GetString("events")
		if containsEvent(events, event) {
			wh := wh // capture loop variable for closure
			routine.FireAndForget(func() {
				if err := SendWebhookForMessage(app, wh, message, event); err != nil {
					webhookBreaker.RecordFailure()
					app.Logger().Warn("webhook delivery failed", slog.Any("error", err))
				} else {
					webhookBreaker.RecordSuccess()
				}
			})
		}
	}
}

const maxRetries = 3

// retryBackoffs defines the minimum wait time before each retry attempt.
var retryBackoffs = []time.Duration{
	15 * time.Minute, // after 1st failure
	1 * time.Hour,    // after 2nd failure
	6 * time.Hour,    // after 3rd failure
}

// isPermanentFailure returns true for errors that should not be retried.
func isPermanentFailure(errMsg string) bool {
	permanent := []string{
		"invalid number",
		"blocked",
		"unsubscribed",
		"blacklisted",
		"not a valid phone",
	}
	lower := strings.ToLower(errMsg)
	for _, p := range permanent {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

// RetryFailedMessages retries failed outgoing messages with exponential backoff
// and a maximum of 3 retry attempts. Permanent failures are skipped.
func RetryFailedMessages(app core.App) error {
	cutoff := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)

	records, err := app.FindRecordsByFilter(
		"sms_messages",
		"status = 'failed' && message_type = 'outgoing' && retry_count < {:maxRetries} && created >= {:cutoff}",
		"", 50, 0,
		dbx.Params{"maxRetries": maxRetries, "cutoff": cutoff},
	)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	retried := 0
	skipped := 0
	for _, record := range records {
		// Skip permanent failures
		if isPermanentFailure(record.GetString("error_message")) {
			skipped++
			continue
		}

		// Enforce exponential backoff based on retry_count
		retryCount := record.GetInt("retry_count")
		if retryCount > 0 && retryCount <= len(retryBackoffs) {
			lastRetry := record.GetDateTime("last_retry_at").Time()
			if !lastRetry.IsZero() {
				requiredWait := retryBackoffs[retryCount-1]
				if now.Sub(lastRetry) < requiredWait {
					continue // not enough time has passed
				}
			}
		}

		record.Set("status", "pending")
		record.Set("retry_count", retryCount+1)
		record.Set("last_retry_at", types.NowDateTime())
		record.Set("error_message", "")
		if err := app.Save(record); err == nil {
			retried++
		}
	}

	app.Logger().Info("Retried failed SMS messages",
		slog.Int("retried", retried), slog.Int("skipped_permanent", skipped))
	return nil
}

// containsEvent checks if a JSON array string contains a specific event.
func containsEvent(eventsJSON string, event string) bool {
	if eventsJSON == "" {
		return false
	}
	var events []string
	if err := json.Unmarshal([]byte(eventsJSON), &events); err != nil {
		return strings.Contains(eventsJSON, event)
	}
	for _, e := range events {
		if e == event {
			return true
		}
	}
	return false
}
