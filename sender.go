package stream

import (
	"fmt"
	"net/http"
	"time"

	"github.com/danfaizer/gowse"
	"github.com/sirupsen/logrus"
)

// SenderConfig defines required Vulcan websocket event server configuration
type SenderConfig struct {
	HTTPPort     int
	HTTPStream   string
	PingInterval time.Duration
}

// Sender defines a websocket event server
type Sender struct {
	topic  *gowse.Topic
	logger logrus.FieldLogger
	config SenderConfig
}

// statusHandler simply returns 200 OK responses
func statusHandler(w http.ResponseWriter, r *http.Request) {
}

// NewSender creates a Vulcan Stream sender instance
func NewSender(l logrus.FieldLogger, c SenderConfig) *Sender {
	server := gowse.NewServer(l.WithFields(logrus.Fields{}))
	topic := server.CreateTopic(c.HTTPStream)
	return &Sender{topic: topic, logger: l, config: c}
}

// Start initializes a websocket event server instance with provided configuration
func (s *Sender) Start() {
	port := fmt.Sprintf(":%v", s.config.HTTPPort)
	mux := http.NewServeMux()
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		if err := s.topic.SubscriberHandler(w, r); err != nil {
			s.logger.Printf("error handling subscriber request: %+v", err)
		}
	})
	mux.HandleFunc("/status", statusHandler)
	go s.ping()
	s.logger.WithFields(logrus.Fields{
		"details": port,
	}).Info("Vulcan Stream Sender started")
	s.logger.Panic(http.ListenAndServe(port, mux))
}

// Broadcast emits msg to the specified Stream channel
func (s *Sender) Broadcast(msg Message) {
	s.topic.Broadcast(msg)
	s.logger.WithFields(logrus.Fields{
		"msg": msg,
	}).Info("Message pushed to the stream successfully")
}

// ping starts a scheduler which will broadcast pings at configured interval
func (s *Sender) ping() {
	pingMsg := Message{Action: "ping"}

	ticker := time.NewTicker(s.config.PingInterval * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.topic.Broadcast(pingMsg)
	}
}
