package stream

import (
	"fmt"
	"net/http"
	"testing"

	metrics "github.com/adevinta/vulcan-metrics-client"
	"github.com/gorilla/websocket"
)

const (
	// Sender
	httpPort     = 8080
	httpStream   = "test"
	pingInterval = 1
	// Receiver
	dbName        = "stream"
	dbUser        = "postgres"
	dbPass        = ""
	dbHost        = "localhost"
	dbPort        = 5432
	dbSSLMode     = "disable"
	streamChannel = "events"
	// Logger
	receiverLogFile  = ""
	receiverLogLevel = "debug"
)

type mockMetricsClient struct {
	metrics.Client
}

func TestReceiverMainActions(t *testing.T) {
	// Initialize loggerConfig
	lc := LoggerConfig{
		LogFile:  receiverLogFile,
		LogLevel: receiverLogLevel,
	}

	logger, logFile, err := NewLogger(lc)
	if err != nil {
		t.Fatalf("error creating the logger: %v", err)
	}
	defer func() {
		_ = logFile.Close()
	}()

	mc := &mockMetricsClient{}

	// Initialize senderConfig
	sc := SenderConfig{
		HTTPPort:     httpPort,
		HTTPStream:   httpStream,
		PingInterval: pingInterval,
	}

	// Initialize receiverConfig
	rc := ReceiverConfig{
		DBName:        dbName,
		DBUser:        dbUser,
		DBPass:        dbPass,
		DBHost:        dbHost,
		DBPort:        dbPort,
		DBSSLMode:     dbSSLMode,
		StreamChannel: streamChannel,
	}

	sender := NewSender(logger, sc)
	go sender.Start()
	receiver, err := NewReceiver(logger, rc, sender, mc)
	if err != nil {
		t.Fatalf("error creating the receiver: %v", err)
	}
	go receiver.Start()

	streamEndpoint := fmt.Sprintf("ws://localhost:%v/stream", sc.HTTPPort)
	conn, _, err := websocket.DefaultDialer.Dial(streamEndpoint, http.Header{})
	if err != nil {
		t.Fatalf("error connecting to stream: %v", err)
	}
	defer conn.Close()

	msg := Message{}
	for {
		err = websocket.ReadJSON(conn, &msg)
		if err != nil {
			fmt.Println("error receiving message: ", err.Error())
			break
		}
		if msg.Action == "ping" {
			sender.Broadcast(Message{Action: "test", AgentID: "00-00-00-00"})
		}
		if msg.Action == "test" && msg.AgentID == "00-00-00-00" {
			break
		}
	}

	statusEndpoint := fmt.Sprintf("http://localhost:%v/status", sc.HTTPPort)
	resp, err := http.Get(statusEndpoint)
	if err != nil {
		t.Error("fail making status http request")
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			resp.StatusCode, http.StatusOK)
	}
}
