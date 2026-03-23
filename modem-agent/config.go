package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// ModemConfig holds the configuration for a single modem.
type ModemConfig struct {
	APIKey      string `yaml:"api_key"`
	CommandPort string `yaml:"command_port"`
	NotifyPort  string `yaml:"notify_port"`
	SimPIN      string `yaml:"sim_pin"`
	Profile     string `yaml:"profile"`
}

// Config holds the global configuration.
type Config struct {
	VendelURL string        `yaml:"vendel_url"`
	Modems    []ModemConfig `yaml:"modems"`
}

func loadConfig() Config {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "modems.yaml"
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("failed to read config file %s: %v", configFile, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("failed to parse config file %s: %v", configFile, err)
	}

	// VENDEL_URL env var overrides the config file value
	if envURL := os.Getenv("VENDEL_URL"); envURL != "" {
		cfg.VendelURL = envURL
	}
	if cfg.VendelURL == "" {
		cfg.VendelURL = "http://localhost:8090"
	}

	if err := validateConfig(&cfg); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	return cfg
}

func validateConfig(cfg *Config) error {
	if len(cfg.Modems) == 0 {
		return fmt.Errorf("no modems configured")
	}
	for i := range cfg.Modems {
		m := &cfg.Modems[i]
		if m.APIKey == "" {
			return fmt.Errorf("modem[%d]: api_key is required", i)
		}
		if m.CommandPort == "" {
			return fmt.Errorf("modem[%d]: command_port is required", i)
		}
		// Default notify_port to command_port (single-port modem)
		if m.NotifyPort == "" {
			m.NotifyPort = m.CommandPort
		}
	}
	return nil
}
