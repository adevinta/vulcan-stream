package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	stream "github.com/adevinta/vulcan-stream"
	"github.com/adevinta/vulcan-stream/config"
	"github.com/gorilla/websocket"
)

const (
	streamReadTimeout = 20
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
func wsClient(l *log.Logger, c *config.Config, t string, ch chan bool) {
	l.Print("Building vulcan-stream URL Endpoint string")
	streamEndpoint := fmt.Sprintf("ws://localhost:%v/stream",
		c.API.Port)

	l.Printf("Client connecting to vulcan-stream URL: %v", streamEndpoint)
	conn, _, err := websocket.DefaultDialer.Dial(streamEndpoint, http.Header{})
	if err != nil {
		log.Fatalf("Error while connecting to topic: %v", err)
	}
	defer conn.Close()

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
					ch <- true
				} else {
					l.Printf("Incorrect stream message received: %+v", message)
					ch <- false
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

// notifyEvent generates a PostgreSQL notification
// Requires:
// - Logger
// - Config
// - Token string (the key which will generate an event that can be identified)
func notifyEvent(l *log.Logger, c *config.Config, t string) {
	// TODO:
	// l.Print("Waiting for stream to be ready")
	// time.Sleep(3000 * time.Millisecond)
	// l.Print("Creating database connection string")
	// connectionString := fmt.Sprintf(
	// 	"dbname='%v' user='%v' password='%v' host=%v port=%v sslmode=%v",
	// 	c.Receiver.DBName, c.Receiver.DBUser, c.Receiver.DBPass, c.Receiver.DBHost, c.Receiver.DBPort, c.Receiver.DBSSLMode)
	// l.Print("Creating database connection")
	// dbConnection, err := sql.Open("postgres", connectionString)
	// if err != nil {
	// 	l.Fatal("Failed connecting to the database")
	// }
	// l.Print("Database connection created")
	// l.Print("Creating event message")
	// notification := fmt.Sprintf("NOTIFY %v, '{\"action\":\"test\",\"check_id\":\"%v\"}'", c.Receiver.StreamChannel, t)
	// l.Printf("Event message created: %v", notification)
	// l.Print("Notifying message to the database")
	// _, err = dbConnection.Exec(notification)
	// if err != nil {
	// 	l.Printf("Error notifying the database: %v", err)
	// }
	// l.Print("Message notified to the database")
}

func main() {
	os.Exit(0)

	// TODO: Update test to v2

	logger := log.New(os.Stderr, "vulcan-stream-test-client: ", log.LstdFlags|log.Lshortfile)
	logger.Print("Starting vulcan-stream-test-client")
	// Read config file
	if len(os.Args) != 2 {
		log.Fatal("Usage: vulcan-stream-test-client config-file")
	}
	configFile := os.Args[1]

	logger.Print("Reading vulcan-stream config file")
	config := config.MustReadConfig(configFile)
	logger.Print("Config file read successfully")

	ch := make(chan bool)
	token := uuid()
	logger.Printf("Magic token to test message streaming: %v", token)
	logger.Print("Starting ServerSideEvents client")
	go wsClient(logger, &config, token, ch)
	go notifyEvent(logger, &config, token)
	go timeout(logger, ch)

	if <-ch {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
