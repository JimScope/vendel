package services

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/routine"
	"github.com/pocketbase/pocketbase/tools/types"
)

// SendSMS orchestrates the entire SMS sending process.
// Recipients are distributed among devices via round-robin.
func SendSMS(app core.App, userId string, recipients []string, body string, deviceId string) ([]*core.Record, error) {
	if len(recipients) == 0 {
		return nil, fmt.Errorf("no recipients provided")
	}

	if err := CheckSMSQuota(app, userId, len(recipients)); err != nil {
		return nil, err
	}

	devices, err := resolveDevices(app, userId, deviceId)
	if err != nil {
		return nil, err
	}

	messages, err := createMessageRecords(app, userId, recipients, body, devices)
	if err != nil {
		return nil, err
	}

	if err := IncrementSMSCount(app, userId, len(recipients)); err != nil {
		app.Logger().Warn("failed to increment SMS count", slog.Any("error", err))
	}

	if len(devices) > 0 {
		routine.FireAndForget(func() { DispatchMessages(app, messages) })
	}

	return messages, nil
}

// resolveDevices returns the target device(s) for sending.
// If deviceId is specified, validates ownership. Otherwise, returns all user devices.
func resolveDevices(app core.App, userId, deviceId string) ([]*core.Record, error) {
	if deviceId != "" {
		device, err := app.FindRecordById("sms_devices", deviceId)
		if err != nil {
			return nil, fmt.Errorf("device not found: %w", err)
		}
		if device.GetString("user") != userId {
			return nil, fmt.Errorf("device does not belong to user")
		}
		return []*core.Record{device}, nil
	}

	records, err := app.FindRecordsByFilter(
		"sms_devices",
		"user = {:userId} && (fcm_token != '' || device_type = 'modem')",
		"-created",
		0, 0,
		dbx.Params{"userId": userId},
	)
	if err != nil || len(records) == 0 {
		return nil, nil
	}
	return records, nil
}

// createMessageRecords creates sms_messages records, assigning devices via round-robin.
func createMessageRecords(app core.App, userId string, recipients []string, body string, devices []*core.Record) ([]*core.Record, error) {
	collection, err := app.FindCollectionByNameOrId("sms_messages")
	if err != nil {
		return nil, fmt.Errorf("sms_messages collection not found: %w", err)
	}

	batchId := ""
	if len(recipients) > 1 {
		batchId = core.GenerateDefaultRandomId()
	}

	messages := make([]*core.Record, 0, len(recipients))
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

	return messages, nil
}

// ProcessSMSAck handles device acknowledgment for a sent SMS.
// The deviceId must match the message's assigned device.
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

	switch status {
	case "sent":
		record.Set("sent_at", types.NowDateTime())
	case "delivered":
		record.Set("delivered_at", types.NowDateTime())
	}

	if err := app.Save(record); err != nil {
		return err
	}

	eventMap := map[string]string{"sent": "sms_sent", "delivered": "sms_delivered", "failed": "sms_failed"}
	if event, ok := eventMap[status]; ok {
		routine.FireAndForget(func() { TriggerWebhooks(app, record.GetString("user"), record, event) })
	}

	return nil
}

// HandleIncomingSMS processes an incoming SMS from a device and triggers webhooks.
func HandleIncomingSMS(app core.App, userId string, deviceId string, fromNumber string, body string, timestamp string) (*core.Record, error) {
	// Deduplicate: check for identical incoming SMS within the last 5 minutes
	cutoff := time.Now().UTC().Add(-5 * time.Minute).Format(time.RFC3339)
	bodyHash, _ := ComputeBodyHash(body)
	existing, err := app.FindFirstRecordByFilter(
		"sms_messages",
		"message_type = 'incoming' && device = {:deviceId} && from_number = {:from} && body_hash = {:hash} && created >= {:cutoff}",
		dbx.Params{"deviceId": deviceId, "from": fromNumber, "hash": bodyHash, "cutoff": cutoff},
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

	routine.FireAndForget(func() { TriggerWebhooks(app, userId, record, "sms_received") })

	return record, nil
}
