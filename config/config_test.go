package config

import (
	"testing"
)

const (
	// Logger
	logFile  = ""
	logLevel = "DEBUG"
	// Receiver
	dbName        = "stream"
	dbUser        = "postgres"
	dbPass        = ""
	dbHost        = "localhost"
	dbPort        = 5432
	dbSSLMode     = "disable"
	streamChannel = "events"
	// Sender
	httpPort     = 8080
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
	// Receiver
	if cfg.Receiver.DBUser != dbUser {
		t.Errorf("Test failed, expected: '%s', got:  '%s'", dbUser, cfg.Receiver.DBUser)
	}
	if cfg.Receiver.DBHost != dbHost {
		t.Errorf("Test failed, expected: '%s', got:  '%s'", dbHost, cfg.Receiver.DBHost)
	}
	if cfg.Receiver.DBName != dbName {
		t.Errorf("Test failed, expected: '%s', got:  '%s'", dbName, cfg.Receiver.DBName)
	}
	if cfg.Receiver.DBPass != dbPass {
		t.Errorf("Test failed, expected: '%s', got:  '%s'", dbPass, cfg.Receiver.DBPass)
	}
	if cfg.Receiver.DBPort != dbPort {
		t.Errorf("Test failed, expected: '%d', got:  '%d'", dbPort, cfg.Receiver.DBPort)
	}
	if cfg.Receiver.DBSSLMode != dbSSLMode {
		t.Errorf("Test failed, expected: '%s', got:  '%s'", dbSSLMode, cfg.Receiver.DBSSLMode)
	}
	if cfg.Receiver.StreamChannel != streamChannel {
		t.Errorf("Test failed, expected: '%s', got:  '%s'", streamChannel, cfg.Receiver.StreamChannel)
	}
	// Sender
	if cfg.Sender.HTTPPort != httpPort {
		t.Errorf("Test failed, expected: '%d', got:  '%d'", httpPort, cfg.Sender.HTTPPort)
	}
	if cfg.Sender.PingInterval != pingInterval {
		t.Errorf("Test failed, expected: '%d', got:  '%d'", pingInterval, cfg.Sender.PingInterval)
	}
}
