/*
Copyright 2019 Adevinta
*/

package config

import (
	"testing"
)

const (
	// Logger
	logFile  = ""
	logLevel = "DEBUG"
	// API
	httpPort = 8080
	// Storage
	storageHost = "127.0.0.1"
	storagePort = 6379
	// Sender
	httpStream   = "stream"
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
	// API
	if cfg.API.Port != httpPort {
		t.Errorf("Test failed, expected: '%d', got:  '%d'", httpPort, cfg.API.Port)
	}
	// Storage
	if cfg.Storage.Host != storageHost {
		t.Errorf("Test failed, expected: '%s', got:  '%s'", storageHost, cfg.Storage.Host)
	}
	if cfg.Storage.Port != storagePort {
		t.Errorf("Test failed, expected: '%d', got:  '%d'", storagePort, cfg.Storage.Port)
	}
	// Sender
	if cfg.Sender.HTTPStream != httpStream {
		t.Errorf("Test failed, expected: '%s', got:  '%s'", httpStream, cfg.Sender.HTTPStream)
	}
	if cfg.Sender.PingInterval != pingInterval {
		t.Errorf("Test failed, expected: '%d', got:  '%d'", pingInterval, cfg.Sender.PingInterval)
	}
}
