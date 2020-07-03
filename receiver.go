package stream

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	metrics "github.com/adevinta/vulcan-metrics-client"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

const (
	// Metrics
	notifiedMetric    = "vulcan.stream.mssgs.notified"
	broadcastedMetric = "vulcan.stream.mssgs.broadcasted"

	componentTag = "component:stream"
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
	metricsClient      metrics.Client
}

// NewReceiver creates a Vulcan Stream receiver instance,
// initializing PostgreSQL listen instance with provided configuration.
func NewReceiver(l logrus.FieldLogger, c ReceiverConfig, s *Sender, mc metrics.Client) (*Receiver, error) {
	receiver := Receiver{
		sender:        s,
		logger:        l,
		config:        c,
		metricsClient: mc,
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

			r.incrNotifiedMssgs()

			msg := Message{}
			if err := json.Unmarshal([]byte(n.Extra), &msg); err != nil {
				r.logger.WithError(err).Error("Notification unmarshall error")
			} else {
				r.incrBroadcastedMssgs(msg)
				r.sender.Broadcast(msg)
			}
		case <-time.After(90 * time.Second):
			go func() {
				err := r.listener.Ping()
				if err != nil {
					r.logger.WithError(err).Error("Listener ping error")
				} else {
					r.incrBroadcastedMssgs(Message{Action: "ping"})
				}
			}()
		}
	}
}

// incrNotifiedMssgs increments the metric for notified mssgs.
func (r *Receiver) incrNotifiedMssgs() {
	r.pushMetrics(notifiedMetric, []string{componentTag})
}

// incrBroadcastedMssgs increments the metric for broadcasted mssgs
// including a tag for the requested action.
func (r *Receiver) incrBroadcastedMssgs(msg Message) {
	tags := []string{
		componentTag,
		fmt.Sprint("action:", msg.Action),
	}
	r.pushMetrics(broadcastedMetric, tags)
}

func (r *Receiver) pushMetrics(metric string, tags []string) {
	r.metricsClient.Push(metrics.Metric{
		Name:  metric,
		Typ:   metrics.Count,
		Value: 1,
		Tags:  tags,
	})
}
