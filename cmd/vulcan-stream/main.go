package main

import (
	"log"
	"os"

	stream "github.com/adevinta/vulcan-stream"
	"github.com/adevinta/vulcan-stream/config"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: vulcan-stream config-file")
	}
	configFile := os.Args[1]

	// Read config file.
	config := config.MustReadConfig(configFile)

	// Initialize logger.
	logger, logFile, err := stream.NewLogger(config.Logger)
	if err != nil {
		log.Fatalf("unable to send logs to file %s", config.Logger.LogFile)
	}
	defer func() {
		// We don't care about the error because a log file may not even exist.
		_ = logFile.Close()
	}()

	logger.Info("Starting Vulcan Stream")

	sender := stream.NewSender(logger, config.Sender)
	go sender.Start()

	receiver, err := stream.NewReceiver(logger, config.Receiver, sender)
	if err != nil {
		logger.WithError(err).Panic()
	}

	go receiver.Start()

	select {}
}
