package stream

import (
	"net/http"
	"time"

	"github.com/danfaizer/gowse"
	"github.com/sirupsen/logrus"
)

// SenderConfig defines required Vulcan websocket event server configuration
type SenderConfig struct {
	HTTPStream   string
	PingInterval time.Duration
}

// Sender defines a websocket event server
type Sender struct {
	topic  *gowse.Topic
	logger logrus.FieldLogger
	config SenderConfig
}

// NewSender creates a Vulcan Stream sender instance
func NewSender(l logrus.FieldLogger, c SenderConfig) *Sender {
	server := gowse.NewServer(l.WithFields(logrus.Fields{}))
	topic := server.CreateTopic(c.HTTPStream)
	return &Sender{topic: topic, logger: l, config: c}
}

// Start initializes a websocket event server instance with provided configuration
func (s *Sender) Start() {
	go s.ping()
	s.logger.Info("Vulcan Stream Sender started")
}

// HandleConn handles a connection to sender web socket topic.
func (s *Sender) HandleConn(w http.ResponseWriter, r *http.Request) {
	if err := s.topic.SubscriberHandler(w, r); err != nil {
		s.logger.Error("error handling subscriber request: %+v", err)
	}
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
