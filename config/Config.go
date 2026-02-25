package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config stores all runtime settings loaded from config.json.
type Config struct {
	HueBridgeIP               string `json:"hueBridgeIp"`
	HueUser                   string `json:"hueUser"`
	ServerPort                int    `json:"serverPort"`
	RestorePreviousLightState bool   `json:"restorePreviousLightState"`
}

// Load reads and validates configuration from disk.
func Load(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file %q: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode config file %q: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("validate config file %q: %w", path, err)
	}

	return cfg, nil
}

// Validate checks the required configuration fields.
func (cfg Config) Validate() error {
	if cfg.HueUser == "" {
		return fmt.Errorf("hueUser is required")
	}
	if cfg.ServerPort <= 0 || cfg.ServerPort > 65535 {
		return fmt.Errorf("serverPort must be between 1 and 65535")
	}
	return nil
}
