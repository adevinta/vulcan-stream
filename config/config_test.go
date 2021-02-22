package config

import (
	"testing"
)

const (
	// Logger
	logFile  = ""
	logLevel = "DEBUG"
	// Sender
	// httpPort     = 8080
	pingInterval = 5
)

func TestMustReadConfig(t *testing.T) {
	cfg := MustReadConfig("../_resources/config/test.toml")

	// Logger
	if cfg.Logger.LogFile != logFile {
		t.Errorf("Test failed, expected: '%s', got:  '%s'", logFile, cfg.Logger.LogFile)
	}
	if cfg.Logger.LogLevel != logLevel {
		t.Errorf("Test failed, expected: '%s', got:  '%s'", logLevel, cfg.Logger.LogLevel)
	}

	// TODO:

	// Sender
	// if cfg.Sender.HTTPPort != httpPort {
	// 	t.Errorf("Test failed, expected: '%d', got:  '%d'", httpPort, cfg.Sender.HTTPPort)
	// }
	if cfg.Sender.PingInterval != pingInterval {
		t.Errorf("Test failed, expected: '%d', got:  '%d'", pingInterval, cfg.Sender.PingInterval)
	}
}
