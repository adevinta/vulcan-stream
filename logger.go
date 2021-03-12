/*
Copyright 2021 Adevinta
*/

package stream

import (
	"os"

	"github.com/sirupsen/logrus"
)

// LoggerConfig defines required Vulcan Logger configuration.
type LoggerConfig struct {
	LogFile  string
	LogLevel string
}

// NewLogger provides a logrus FieldLogger.
func NewLogger(lc LoggerConfig) (logrus.FieldLogger, *os.File, error) {
	logger := logrus.New().WithFields(logrus.Fields{
		"app": "VULCAN-STREAM",
	})
	logger.Logger.Formatter = &logrus.TextFormatter{
		FullTimestamp: true,
		DisableColors: true,
	}

	var file *os.File
	var err error

	if lc.LogFile != "" {
		file, err = os.OpenFile(lc.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			return nil, nil, err
		}
		logger.Logger.Out = file
	}

	switch lc.LogLevel {
	case "DEBUG":
		logger.Logger.Level = logrus.DebugLevel
	case "INFO":
		logger.Logger.Level = logrus.InfoLevel
	case "WARN":
		logger.Logger.Level = logrus.WarnLevel
	case "ERROR":
		logger.Logger.Level = logrus.ErrorLevel
	case "FATAL":
		logger.Logger.Level = logrus.FatalLevel
	case "PANIC":
		logger.Logger.Level = logrus.PanicLevel
	default:
		logger.Logger.Level = logrus.DebugLevel
	}

	return logger, file, nil
}
