package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "config.json")
	content := `{
		"hueBridgeIp": "192.168.1.2",
		"hueUser": "test-user",
		"serverPort": 8090,
		"restorePreviousLightState": true
	}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write test config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HueBridgeIP != "192.168.1.2" {
		t.Fatalf("HueBridgeIP = %q, want %q", cfg.HueBridgeIP, "192.168.1.2")
	}
	if cfg.HueUser != "test-user" {
		t.Fatalf("HueUser = %q, want %q", cfg.HueUser, "test-user")
	}
	if cfg.ServerPort != 8090 {
		t.Fatalf("ServerPort = %d, want %d", cfg.ServerPort, 8090)
	}
	if !cfg.RestorePreviousLightState {
		t.Fatalf("RestorePreviousLightState = false, want true")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "config.json")
	if err := os.WriteFile(path, []byte("{not-json}"), 0o644); err != nil {
		t.Fatalf("write test config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("Load() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Fatalf("Load() error = %q, want message to contain %q", err.Error(), "decode")
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "missing hue user",
			cfg: Config{
				ServerPort: 8090,
			},
			wantErr: "hueUser is required",
		},
		{
			name: "invalid port",
			cfg: Config{
				HueUser:    "abc",
				ServerPort: 70000,
			},
			wantErr: "serverPort must be between 1 and 65535",
		},
		{
			name: "valid",
			cfg: Config{
				HueUser:    "abc",
				ServerPort: 8090,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.cfg.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("Validate() error = nil, want %q", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("Validate() error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}
