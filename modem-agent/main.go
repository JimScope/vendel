package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/xlab/at"
	"github.com/xlab/at/sms"
)

// ModemConfig holds the configuration for a single modem.
type ModemConfig struct {
	APIKey      string
	DeviceID    string // derived from API key (set after startup recovery)
	CommandPort string
	NotifyPort  string
}

// Config holds the global configuration.
type Config struct {
	EnderURL string
	Modems   []ModemConfig
}

func main() {
	cfg := loadConfig()

	log.Printf("ender-modem-agent starting with %d modem(s)", len(cfg.Modems))

	var wg sync.WaitGroup
	for _, modemCfg := range cfg.Modems {
		wg.Add(1)
		go func(m ModemConfig) {
			defer wg.Done()
			runModem(m, cfg.EnderURL)
		}(modemCfg)
	}

	// Wait for signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Printf("received signal %s, shutting down", sig)
}

func loadConfig() Config {
	enderURL := os.Getenv("ENDER_URL")
	if enderURL == "" {
		enderURL = "http://localhost:8090"
	}

	modemStr := os.Getenv("MODEMS")
	if modemStr == "" {
		log.Fatal("MODEMS env var is required (format: api_key:command_port[:notify_port],...")
	}

	var modems []ModemConfig
	for _, entry := range strings.Split(modemStr, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.SplitN(entry, ":", 3)
		if len(parts) < 2 {
			log.Fatalf("invalid MODEMS entry %q: expected api_key:command_port[:notify_port]", entry)
		}

		m := ModemConfig{
			APIKey:      parts[0],
			CommandPort: parts[1],
		}
		if len(parts) == 3 && parts[2] != "" {
			m.NotifyPort = parts[2]
		} else {
			m.NotifyPort = m.CommandPort
		}
		modems = append(modems, m)
	}

	if len(modems) == 0 {
		log.Fatal("no modems configured")
	}

	return Config{
		EnderURL: enderURL,
		Modems:   modems,
	}
}

func runModem(cfg ModemConfig, enderURL string) {
	logPrefix := fmt.Sprintf("[%s]", cfg.CommandPort)
	log.Printf("%s starting modem on command=%s notify=%s", logPrefix, cfg.CommandPort, cfg.NotifyPort)

	// Open modem via xlab/at
	dev := &at.Device{
		CommandPort: cfg.CommandPort,
		NotifyPort:  cfg.NotifyPort,
	}
	if err := dev.Open(); err != nil {
		log.Printf("%s failed to open modem: %v", logPrefix, err)
		return
	}
	defer dev.Close()

	if err := dev.Init(at.DeviceE173()); err != nil {
		log.Printf("%s failed to init modem: %v", logPrefix, err)
		return
	}
	log.Printf("%s modem initialized", logPrefix)

	// Use the command port path as a temporary device ID for logging.
	// The real device ID comes from the backend via the API key,
	// but we use the API key directly for SSE topic subscription.
	// The backend maps dk_ keys to device records.
	client := NewEnderClient(enderURL, cfg.APIKey, cfg.CommandPort)

	// Recover any pending messages from before agent started
	pending, err := client.FetchPending()
	if err != nil {
		log.Printf("%s failed to fetch pending messages: %v", logPrefix, err)
	} else if len(pending) > 0 {
		log.Printf("%s processing %d pending message(s)", logPrefix, len(pending))
		for _, msg := range pending {
			sendAndReport(dev, client, msg, logPrefix)
		}
	}

	// Start incoming SMS monitoring
	go dev.Watch()
	go func() {
		for msg := range dev.IncomingSms() {
			log.Printf("%s incoming SMS from %s", logPrefix, msg.Address)
			if reportErr := client.ReportIncoming(
				string(msg.Address),
				msg.Text,
				time.Now().UTC().Format(time.RFC3339),
			); reportErr != nil {
				log.Printf("%s failed to report incoming SMS: %v", logPrefix, reportErr)
			}
		}
	}()

	// Subscribe to SSE for real-time message assignment.
	// ConnectSSE reconnects automatically on disconnect.
	log.Printf("%s connecting to SSE for real-time dispatch", logPrefix)
	client.ConnectSSE(func(msg PendingMessage) {
		log.Printf("%s received message %s -> %s", logPrefix, msg.MessageID, msg.Recipient)
		sendAndReport(dev, client, msg, logPrefix)
	})
}

func sendAndReport(dev *at.Device, client *EnderClient, msg PendingMessage, logPrefix string) {
	if err := dev.SendSMS(msg.Body, sms.PhoneNumber(msg.Recipient)); err != nil {
		log.Printf("%s send failed for %s: %v", logPrefix, msg.MessageID, err)
		if reportErr := client.ReportStatus(msg.MessageID, "failed", err.Error()); reportErr != nil {
			log.Printf("%s failed to report status: %v", logPrefix, reportErr)
		}
	} else {
		log.Printf("%s sent %s to %s", logPrefix, msg.MessageID, msg.Recipient)
		if reportErr := client.ReportStatus(msg.MessageID, "sent", ""); reportErr != nil {
			log.Printf("%s failed to report status: %v", logPrefix, reportErr)
		}
	}
}
