package stream

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// ReceiverConfig defines required Vulcan Stream configuration
type ReceiverConfig struct {
	DBName        string
	DBUser        string
	DBPass        string
	DBHost        string
	DBPort        int
	DBSSLMode     string
	StreamChannel string
}

// Receiver defines Vulcan Steam message receiver
type Receiver struct {
	dbConnection       *sql.DB
	dbConnectionString string
	sender             *Sender
	logger             logrus.FieldLogger
	config             ReceiverConfig
	listener           *pq.Listener
}

// NewReceiver creates a Vulcan Stream receiver instance,
// initializing PostgreSQL listen instance with provided configuration.
func NewReceiver(l logrus.FieldLogger, c ReceiverConfig, s *Sender) (*Receiver, error) {
	receiver := Receiver{
		sender: s,
		logger: l,
		config: c,
	}

	connectionString := fmt.Sprintf(
		"dbname='%v' user='%v' password='%v' host=%v port=%v sslmode=%v",
		c.DBName, c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBSSLMode)

	dbConnection, err := sql.Open("postgres", connectionString)
	if err != nil {
		l.WithError(err).Debug()
		return nil, errors.New("Problem connecting to the database")
	}

	if err = dbConnection.Ping(); err != nil {
		l.WithError(err).Debug()
		return nil, errors.New("Problem pinging the database")
	}

	receiver.dbConnectionString = connectionString
	receiver.dbConnection = dbConnection

	dbEventCallback := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			l.WithError(err).Panic("Database connection state changed")
		}
	}

	receiver.listener = pq.NewListener(connectionString, 10*time.Second, time.Minute, dbEventCallback)
	if err := receiver.listener.Listen(c.StreamChannel); err != nil {
		l.WithError(err).Debug()
		return nil, errors.New("Problem when creating database listener")
	}

	return &receiver, nil
}

// Start publishes messages to VulcanStream through the Sender
func (r *Receiver) Start() {
	r.logger.Info("Stream Receiver started")

	for {
		select {
		case n := <-r.listener.Notify:
			r.logger.WithFields(logrus.Fields{
				"message": string([]byte(n.Extra)),
			}).Info("Message notification received")
			msg := Message{}
			if err := json.Unmarshal([]byte(n.Extra), &msg); err != nil {
				r.logger.WithError(err).Error("Notification unmarshall error")
			} else {
				r.sender.Broadcast(msg)
			}
		case <-time.After(90 * time.Second):
			go func() {
				err := r.listener.Ping()
				if err != nil {
					r.logger.WithError(err).Error("Listener ping error")
				}
			}()
		}
	}
}
