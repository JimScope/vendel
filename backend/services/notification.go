package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/pocketbase/pocketbase/core"
	"google.golang.org/api/option"
)

const (
	FCMMaxPayloadBytes     = 4096
	FCMPayloadOverhead     = 256
	FCMChunkDelaySeconds   = 5
)

var fcmClient *messaging.Client

// InitFCM initializes the Firebase Admin SDK from environment.
func InitFCM() {
	credJSON := os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON")
	if credJSON == "" {
		log.Println("WARNING: FIREBASE_SERVICE_ACCOUNT_JSON not set. FCM will not work.")
		return
	}

	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsJSON([]byte(credJSON)))
	if err != nil {
		log.Printf("ERROR: Failed to initialize Firebase app: %v", err)
		return
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to get FCM client: %v", err)
		return
	}

	fcmClient = client
	log.Println("Firebase Admin SDK initialized successfully.")
}

// MessageRef holds the ID and recipient for an FCM payload.
type MessageRef struct {
	MessageID string `json:"message_id"`
	Recipient string `json:"recipient"`
}

// DispatchMessages groups messages by device and sends FCM notifications via goroutines.
// This replaces QStash entirely.
func DispatchMessages(app core.App, messages []*core.Record, body string) {
	if len(messages) == 0 {
		return
	}

	// Group by device
	byDevice := make(map[string][]*core.Record)
	for _, msg := range messages {
		deviceId := msg.GetString("device")
		if deviceId != "" {
			byDevice[deviceId] = append(byDevice[deviceId], msg)
		}
	}

	for deviceId, deviceMessages := range byDevice {
		device, err := app.FindRecordById("sms_devices", deviceId)
		if err != nil {
			log.Printf("WARNING: Device %s not found: %v", deviceId, err)
			continue
		}

		fcmToken := device.GetString("fcm_token")
		if fcmToken == "" {
			log.Printf("WARNING: Device %s has no FCM token", deviceId)
			continue
		}

		// Build message refs
		refs := make([]MessageRef, 0, len(deviceMessages))
		for _, msg := range deviceMessages {
			refs = append(refs, MessageRef{
				MessageID: msg.Id,
				Recipient: msg.GetString("to"),
			})
		}

		// Chunk for FCM 4KB limit
		chunks := chunkMessagesForFCM(refs, body)

		// Dispatch each chunk in a goroutine with staggered delay and timeout
		for i, chunk := range chunks {
			delay := time.Duration(i*FCMChunkDelaySeconds) * time.Second
			go func(token string, chunk []MessageRef, body string, delay time.Duration) {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second+delay)
				defer cancel()

				if delay > 0 {
					select {
					case <-time.After(delay):
					case <-ctx.Done():
						log.Printf("ERROR: FCM dispatch timed out during delay for token %s", token[:20])
						return
					}
				}
				if err := sendFCMNotification(token, chunk, body); err != nil {
					log.Printf("ERROR: FCM send failed for token %s: %v", token[:20], err)
					for _, ref := range chunk {
						markMessageFailed(app, ref.MessageID, err.Error())
					}
				}
			}(fcmToken, chunk, body, delay)
		}
	}
}

func sendFCMNotification(token string, refs []MessageRef, body string) error {
	if fcmClient == nil {
		return fmt.Errorf("FCM not initialized")
	}

	refsJSON, err := json.Marshal(refs)
	if err != nil {
		return fmt.Errorf("marshal refs: %w", err)
	}

	msg := &messaging.Message{
		Token: token,
		Data: map[string]string{
			"messages": string(refsJSON),
			"body":     body,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := fcmClient.Send(ctx, msg)
	if err != nil {
		return err
	}

	log.Printf("FCM message sent: %s", resp)
	return nil
}

func markMessageFailed(app core.App, messageId, errMsg string) {
	record, err := app.FindRecordById("sms_messages", messageId)
	if err != nil {
		return
	}
	record.Set("status", "failed")
	record.Set("error_message", errMsg)
	_ = app.Save(record)
}

// chunkMessagesForFCM splits messages into chunks fitting within FCM's 4KB payload limit.
func chunkMessagesForFCM(refs []MessageRef, body string) [][]MessageRef {
	if len(refs) == 0 {
		return nil
	}

	// Check if everything fits
	if estimatePayloadSize(refs, body) <= FCMMaxPayloadBytes {
		return [][]MessageRef{refs}
	}

	var chunks [][]MessageRef
	var current []MessageRef

	for _, ref := range refs {
		candidate := append(current, ref)
		if estimatePayloadSize(candidate, body) > FCMMaxPayloadBytes {
			if len(current) > 0 {
				chunks = append(chunks, current)
			}
			current = []MessageRef{ref}
		} else {
			current = candidate
		}
	}

	if len(current) > 0 {
		chunks = append(chunks, current)
	}

	return chunks
}

func estimatePayloadSize(refs []MessageRef, body string) int {
	data, _ := json.Marshal(map[string]string{
		"messages": mustMarshalString(refs),
		"body":     body,
	})
	return len(data)
}

func mustMarshalString(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
