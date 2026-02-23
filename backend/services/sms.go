package services

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
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
			"user = {:userId} && fcm_token != ''",
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
		go DispatchMessages(app, messages, body)
	}

	return messages, nil
}

// ProcessSMSAck handles device acknowledgment for a sent SMS.
func ProcessSMSAck(app core.App, messageId string, status string, errorMessage string) error {
	record, err := app.FindRecordById("sms_messages", messageId)
	if err != nil {
		return fmt.Errorf("message not found: %w", err)
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
		go TriggerWebhooks(app, record.GetString("user"), record, event)
	}

	return nil
}

// HandleIncomingSMS processes an incoming SMS from a device and triggers webhooks.
func HandleIncomingSMS(app core.App, userId string, fromNumber string, body string, timestamp string) (*core.Record, error) {
	collection, err := app.FindCollectionByNameOrId("sms_messages")
	if err != nil {
		return nil, err
	}

	record := core.NewRecord(collection)
	record.Set("user", userId)
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
	go TriggerWebhooks(app, userId, record, "sms_received")

	return record, nil
}

// TriggerWebhooks finds active webhook configs for a user and fires matching webhooks.
func TriggerWebhooks(app core.App, userId string, message *core.Record, event string) {
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
			go func(webhook *core.Record) {
				if err := SendWebhookForMessage(app, webhook, message, event); err != nil {
					app.Logger().Warn("webhook delivery failed", slog.Any("error", err))
				}
			}(wh)
		}
	}
}

// RetryFailedMessages retries failed outgoing messages from the last 24 hours.
func RetryFailedMessages(app core.App) error {
	cutoff := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)

	records, err := app.FindRecordsByFilter(
		"sms_messages",
		"status = 'failed' && message_type = 'outgoing' && created >= {:cutoff}",
		"", 50, 0,
		dbx.Params{"cutoff": cutoff},
	)
	if err != nil {
		return err
	}

	retried := 0
	for _, record := range records {
		record.Set("status", "pending")
		if err := app.Save(record); err == nil {
			retried++
		}
	}

	app.Logger().Info("Retried failed SMS messages", slog.Int("count", retried))
	return nil
}
