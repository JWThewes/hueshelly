package logging

import (
	"fmt"
	"io"
	"log"
	"os"
)

var Logger *log.Logger
var logFile *os.File

const defaultLogFile = "hueshelly.log"

func init() {
	Logger = log.New(os.Stdout, "", log.LstdFlags)
}

// Init configures logging to write to stdout and a local file.
func Init(path string) error {
	if path == "" {
		path = defaultLogFile
	}

	if logFile != nil {
		if err := logFile.Close(); err != nil {
			return fmt.Errorf("close existing log file: %w", err)
		}
		logFile = nil
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		Logger = log.New(os.Stdout, "", log.LstdFlags)
		return fmt.Errorf("open log file %q: %w", path, err)
	}

	logFile = file
	Logger = log.New(io.MultiWriter(os.Stdout, file), "", log.LstdFlags)
	Logger.Printf("Logging to %s", path)
	return nil
}

// Close closes the active logfile, if present.
func Close() error {
	if logFile == nil {
		return nil
	}
	if err := logFile.Close(); err != nil {
		return fmt.Errorf("close log file: %w", err)
	}
	logFile = nil
	return nil
}
