/*
Copyright 2021 Adevinta
*/

package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	stream "github.com/adevinta/vulcan-stream"
	"github.com/adevinta/vulcan-stream/config"
	"github.com/gorilla/websocket"
)

const (
	streamReadTimeout = 30
)

// generates random UUID
func uuid() string {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return ""
	}
	uuid[8] = uuid[8]&^0xc0 | 0x80
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

// timeout sends false to test exit channel if test hadn't finished before streamReadTimeout
func timeout(l *log.Logger, ch chan bool) {
	time.Sleep(time.Second * streamReadTimeout)
	l.Print("Message reception timeout")
	ch <- false
}

// wsClient starts a websocket client
// The client requires:
// - Logger
// - Config
// - Token string (the key which will verify that stream event has propagated properly)
// - Channel which will confirm that the message received is well formed
func wsClient(l *log.Logger, c config.Config, t string, resCh chan bool, conCh chan struct{}) {
	l.Print("Building vulcan-stream URL Endpoint string")
	streamEndpoint := fmt.Sprintf("ws://localhost:%v/stream",
		c.API.Port)

	l.Printf("Client connecting to vulcan-stream URL: %v", streamEndpoint)
	conn, _, err := websocket.DefaultDialer.Dial(streamEndpoint, http.Header{})
	if err != nil {
		log.Fatalf("Error while connecting to topic: %v", err)
	}
	defer conn.Close()

	conCh <- struct{}{}

	message := stream.Message{}
	done := make(chan error)
	go func() {
		var err error
		for {
			err = websocket.ReadJSON(conn, &message)
			if err != nil {
				fmt.Println("Error receiving message: ", err.Error())
				break
			}
			if message.Action != "ping" {
				if message.CheckID == t {
					l.Printf("Stream message read successfully: %+v", message)
					resCh <- true
				} else {
					l.Printf("Incorrect stream message received: %+v", message)
					resCh <- false
				}
			}
		}
		done <- err
	}()
	<-done

	if err != nil {
		log.Fatalf("Error while reading from topic: %v", err)
	}
}

// abortCheck performs an HTTP request to stream's abort endpoint
// in order to abort a check which matches the input token.
// Requires:
// - Config
// - Token string (the key which was specified as identifiable check ID)
func abortCheck(c config.Config, t string) error {
	abortEndpoint := fmt.Sprintf("http://localhost:%d/abort", c.API.Port)
	abortPayload := bytes.NewBuffer([]byte(fmt.Sprintf(`{"checks": ["%v"]}`, t)))
	_, err := http.Post(abortEndpoint, "application/json", abortPayload)
	if err != nil {
		return err
	}
	return nil
}

// verifyChecks verifies that the input token t is included in the aborted
// checks list returned by checks stream endpoint.
// Requires:
// - Config
// - Token string (the key which was specified as identifiable check ID)
func verifyChecks(c config.Config, t string) error {
	checksEndpoint := fmt.Sprintf("http://localhost:%d/checks", c.API.Port)
	resp, err := http.Get(checksEndpoint)
	if err != nil {
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var checks []string
	err = json.Unmarshal(respBody, &checks)
	if err != nil {
		return err
	}
	if len(checks) != 1 || checks[0] != t {
		return fmt.Errorf("checks do not contain t\ngot: %v", checks)
	}
	return nil
}

func main() {
	logger := log.New(os.Stderr, "vulcan-stream-test-client: ", log.LstdFlags|log.Lshortfile)
	logger.Print("Starting vulcan-stream-test-client")

	logger.Print("Waiting for stream to be ready")
	time.Sleep(3000 * time.Millisecond)

	// Read config file
	if len(os.Args) != 2 {
		log.Fatal("Usage: vulcan-stream-test-client config-file")
	}
	configFile := os.Args[1]

	logger.Print("Reading vulcan-stream config file")
	config := config.MustReadConfig(configFile)
	logger.Print("Config file read successfully")

	// Test WS communication
	resCh := make(chan bool)
	go timeout(logger, resCh)

	token := uuid()

	logger.Print("Starting stream WS client")
	conCh := make(chan struct{})
	go wsClient(logger, config, token, resCh, conCh)
	<-conCh // wait for wsClient to be connected

	logger.Print("Sending abort request to stream API")
	if err := abortCheck(config, token); err != nil {
		logger.Printf("Error sending abort request to stream API: %v", err)
		os.Exit(1)
	}
	logger.Print("Abort request successfully sent to stream API")

	if !<-resCh {
		os.Exit(1)
	}

	// Test checks endpoint
	if err := verifyChecks(config, token); err != nil {
		logger.Printf("Error verifying checks: %v", err)
		os.Exit(1)
	}
	logger.Print("Checks endpoint response verified successfully")
}
