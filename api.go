/*
Copyright 2021 Adevinta
*/

package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	metrics "github.com/adevinta/vulcan-metrics-client"
	"github.com/sirupsen/logrus"
)

const (
	// abort check action
	actionAbort = "abort"

	// metrics
	metricNotified    = "vulcan.stream.mssgs.notified"
	metricBroadcasted = "vulcan.stream.mssgs.broadcasted"

	componentTag = "component:stream"
)

// APIConfig represents the config
// necessary for stream API.
type APIConfig struct {
	Port int
}

// API represents the stream REST API.
type API struct {
	sender  *Sender
	storage Storage
	logger  logrus.FieldLogger
	metrics metrics.Client
	mux     *http.ServeMux
	port    int
}

// AbortRequest represents the body
// for an abort cheks request.
type AbortRequest struct {
	Checks []string `json:"checks"`
}

// NewAPI builds a new stream  API.
func NewAPI(port int, sender *Sender, storage Storage, logger logrus.FieldLogger,
	metrics metrics.Client) *API {

	a := &API{
		sender:  sender,
		storage: storage,
		logger:  logger,
		metrics: metrics,
		mux:     http.NewServeMux(),
		port:    port,
	}

	a.mux.HandleFunc("/stream", a.connHandler)
	a.mux.HandleFunc("/checks", a.checksHandler)
	a.mux.HandleFunc("/abort", a.abortHandler)
	a.mux.HandleFunc("/status", a.statusHandler)

	return a
}

// Start starts the stream API.
func (a *API) Start() {
	a.logger.WithFields(logrus.Fields{
		"details": a.port,
	}).Info("Vulcan Stream API started")

	go a.sender.Start()

	a.logger.Panic(http.ListenAndServe(fmt.Sprintf(":%v", a.port), a.mux))
}

// connHandler handles a new connection to the stream.
func (a *API) connHandler(w http.ResponseWriter, r *http.Request) {
	a.sender.HandleConn(w, r)
}

// checksHandler returns the list of currently aborted checks.
func (a *API) checksHandler(w http.ResponseWriter, r *http.Request) {
	checks, err := a.storage.GetAbortedChecks(context.Background())
	if err != nil {
		writeErr(w, err)
		return
	}

	checksArray, err := json.Marshal(checks)
	if err != nil {
		writeErr(w, err)
		return
	}

	w.Write(checksArray)
}

// abortHandler handles an abort checks request.
func (a *API) abortHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, err)
		return
	}

	var req AbortRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		writeErr(w, err)
		return
	}

	err = a.storage.AddAbortedChecks(context.Background(), req.Checks)
	if err != nil {
		writeErr(w, err)
		return
	}

	a.incrNotifiedMssgs(len(req.Checks))

	// TODO: should we broadcast
	// checks async once they are
	// stored so we don't block
	// sync HTTP request?
	// E.g: have a backlog in sender
	// and a goroutine which consumes
	// from that.

	for _, c := range req.Checks {
		m := Message{
			CheckID: c,
			Action:  actionAbort,
		}
		a.sender.Broadcast(m)
		a.incrBroadcastedMssgs(m)
	}
}

func (a *API) statusHandler(w http.ResponseWriter, r *http.Request) { /* 200 OK */ }

func writeErr(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf("err: %v", err)))
}

func (a *API) incrNotifiedMssgs(count int) {
	a.metrics.Push(metrics.Metric{
		Name:  metricNotified,
		Typ:   metrics.Count,
		Value: float64(count),
		Tags:  []string{componentTag},
	})
}

func (a *API) incrBroadcastedMssgs(m Message) {
	a.metrics.Push(metrics.Metric{
		Name:  metricBroadcasted,
		Typ:   metrics.Count,
		Value: 1,
		Tags:  []string{componentTag, fmt.Sprint("action:", m.Action)},
	})
}
