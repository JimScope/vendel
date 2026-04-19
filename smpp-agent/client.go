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

// RemoteSMPPConfig is the bind configuration fetched from the backend.
type RemoteSMPPConfig struct {
	DeviceID           string `json:"device_id"`
	Host               string `json:"host"`
	Port               int    `json:"port"`
	SystemID           string `json:"system_id"`
	Password           string `json:"password"`
	SystemType         string `json:"system_type"`
	BindMode           string `json:"bind_mode"`
	SourceTON          int    `json:"source_ton"`
	SourceNPI          int    `json:"source_npi"`
	DestTON            int    `json:"dest_ton"`
	DestNPI            int    `json:"dest_npi"`
	UseTLS             bool   `json:"use_tls"`
	EnquireLinkSeconds int    `json:"enquire_link_seconds"`
	DefaultDataCoding  int    `json:"default_data_coding"`
	SubmitThrottleTPS  int    `json:"submit_throttle_tps"`
	SourceAddr         string `json:"source_addr"`
}

// VendelClient communicates with the Vendel backend.
type VendelClient struct {
	baseURL  string
	apiKey   string
	deviceID string
	http     *http.Client
}

func NewVendelClient(baseURL, apiKey string) *VendelClient {
	return &VendelClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchSMPPConfig fetches the SMPP bind configuration for this device from
// the backend. This also implicitly resolves the device record ID.
func (c *VendelClient) FetchSMPPConfig() (*RemoteSMPPConfig, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/smpp/config", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch smpp config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetch smpp config: status %d: %s", resp.StatusCode, body)
	}

	var cfg RemoteSMPPConfig
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode smpp config: %w", err)
	}
	c.deviceID = cfg.DeviceID
	return &cfg, nil
}

// FetchPending claims pending messages assigned while the agent was offline.
func (c *VendelClient) FetchPending() ([]PendingMessage, error) {
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
func (c *VendelClient) ReportStatus(messageID, status, errorMessage string) error {
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

// ReportIncoming reports an incoming SMS (MO) received via the SMPP bind.
func (c *VendelClient) ReportIncoming(fromNumber, body, timestamp string) error {
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

// ConnectSSE subscribes to the backend SSE stream for the "smpp/<deviceId>"
// topic and invokes onMessage for every dispatched message.
// It reconnects automatically with exponential backoff on disconnect.
func (c *VendelClient) ConnectSSE(onMessage func(PendingMessage)) {
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

func (c *VendelClient) runSSE(onMessage func(PendingMessage)) error {
	sseClient := &http.Client{Timeout: 0}

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

				if err := c.subscribe(clientID); err != nil {
					return fmt.Errorf("subscribe: %w", err)
				}
				log.Printf("[%s] subscribed to smpp/%s", c.deviceID, c.deviceID)
				continue
			}

			topic := "smpp/" + c.deviceID
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

func (c *VendelClient) subscribe(clientID string) error {
	topic := "smpp/" + c.deviceID
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
