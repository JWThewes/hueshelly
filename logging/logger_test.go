package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCreatesLogFile(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "hueshelly.log")

	if err := Init(logPath); err != nil {
		t.Fatalf("Init() error = %v, want nil", err)
	}

	Logger.Println("test log message")

	if err := Close(); err != nil {
		t.Fatalf("Close() error = %v, want nil", err)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v, want nil", err)
	}

	logOutput := string(content)
	if !strings.Contains(logOutput, "test log message") {
		t.Fatalf("log file content = %q, want message", logOutput)
	}
}
