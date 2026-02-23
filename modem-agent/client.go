package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// PendingMessage represents an SMS message to be sent.
type PendingMessage struct {
	MessageID string `json:"message_id"`
	Recipient string `json:"recipient"`
	Body      string `json:"body"`
}

// EnderClient communicates with the Ender backend.
type EnderClient struct {
	baseURL  string
	apiKey   string
	deviceID string // resolved from backend via FetchPending
	http     *http.Client
}

// NewEnderClient creates a new Ender API client.
// The deviceID is resolved from the backend on the first call to FetchPending.
func NewEnderClient(baseURL, apiKey string) *EnderClient {
	return &EnderClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchPending recovers messages assigned while the agent was offline.
// It also resolves the device record ID from the backend (stored in c.deviceID).
func (c *EnderClient) FetchPending() ([]PendingMessage, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/sms/pending", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch pending: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetch pending: status %d: %s", resp.StatusCode, body)
	}

	var result struct {
		DeviceID string           `json:"device_id"`
		Messages []PendingMessage `json:"messages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode pending response: %w", err)
	}

	if result.DeviceID != "" {
		c.deviceID = result.DeviceID
	}

	return result.Messages, nil
}

// ReportStatus reports message delivery status back to the server.
func (c *EnderClient) ReportStatus(messageID, status, errorMessage string) error {
	payload, _ := json.Marshal(map[string]string{
		"message_id":    messageID,
		"status":        status,
		"error_message": errorMessage,
	})

	req, err := http.NewRequest("POST", c.baseURL+"/api/sms/report", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("report status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("report status: status %d: %s", resp.StatusCode, body)
	}
	return nil
}

// ReportIncoming reports an incoming SMS received on the modem.
func (c *EnderClient) ReportIncoming(fromNumber, body, timestamp string) error {
	payload, _ := json.Marshal(map[string]string{
		"from_number": fromNumber,
		"body":        body,
		"timestamp":   timestamp,
	})

	req, err := http.NewRequest("POST", c.baseURL+"/api/sms/incoming", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("report incoming: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("report incoming: status %d: %s", resp.StatusCode, body)
	}
	return nil
}

// ConnectSSE establishes an SSE connection to PocketBase and subscribes to the modem topic.
// It reconnects automatically with exponential backoff on disconnect.
// The onMessage callback is called for each incoming message assignment.
func (c *EnderClient) ConnectSSE(onMessage func(PendingMessage)) {
	backoff := time.Second
	maxBackoff := 60 * time.Second

	for {
		err := c.runSSE(onMessage)
		if err != nil {
			log.Printf("[%s] SSE disconnected: %v, reconnecting in %s", c.deviceID, err, backoff)
		}
		time.Sleep(backoff)
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

func (c *EnderClient) runSSE(onMessage func(PendingMessage)) error {
	// Step 1: Connect to SSE endpoint
	sseClient := &http.Client{
		Timeout: 0, // no timeout for SSE
	}

	req, err := http.NewRequest("GET", c.baseURL+"/api/realtime", nil)
	if err != nil {
		return fmt.Errorf("create SSE request: %w", err)
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := sseClient.Do(req)
	if err != nil {
		return fmt.Errorf("SSE connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SSE connect: status %d: %s", resp.StatusCode, body)
	}

	// Parse SSE stream to get clientId from the PB_CONNECT event
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var clientID string
	var eventName string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "event:") {
			eventName = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}

		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

			if eventName == "PB_CONNECT" {
				var connect struct {
					ClientID string `json:"clientId"`
				}
				if err := json.Unmarshal([]byte(data), &connect); err != nil {
					return fmt.Errorf("parse PB_CONNECT: %w", err)
				}
				clientID = connect.ClientID
				log.Printf("[%s] SSE connected, clientId=%s", c.deviceID, clientID)

				// Step 2: Subscribe to our modem topic
				if err := c.subscribe(clientID); err != nil {
					return fmt.Errorf("subscribe: %w", err)
				}
				log.Printf("[%s] subscribed to modem/%s", c.deviceID, c.deviceID)
				continue
			}

			// Handle modem message events
			topic := "modem/" + c.deviceID
			if eventName == topic {
				var msg PendingMessage
				if err := json.Unmarshal([]byte(data), &msg); err != nil {
					log.Printf("[%s] failed to parse SSE message: %v", c.deviceID, err)
					continue
				}
				onMessage(msg)
			}

			eventName = ""
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("SSE read: %w", err)
	}
	return fmt.Errorf("SSE stream ended")
}

// subscribe sends a POST to /api/realtime to register subscriptions.
func (c *EnderClient) subscribe(clientID string) error {
	topic := "modem/" + c.deviceID
	payload, _ := json.Marshal(map[string]any{
		"clientId":      clientID,
		"subscriptions": []string{topic},
	})

	req, err := http.NewRequest("POST", c.baseURL+"/api/realtime", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("subscribe failed: status %d: %s", resp.StatusCode, body)
	}
	return nil
}
