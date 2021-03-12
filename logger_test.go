/*
Copyright 2021 Adevinta
*/

package stream

import (
	"testing"
)

func TestNewLogger(t *testing.T) {
	lc := LoggerConfig{LogFile: "logfile", LogLevel: "INFO"}
	logger, logFile, err := NewLogger(lc)

	if logger == nil {
		t.Errorf("logger is nil")
	}

	if logFile == nil {
		t.Errorf("logFile is nil")
	}

	if err != nil {
		t.Errorf("error while creating new logger: %s", err)
	}
}
