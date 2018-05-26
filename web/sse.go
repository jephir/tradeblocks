package web

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

// event represents an SSE event
type event []byte

// https://robots.thoughtbot.com/writing-a-server-sent-events-server-in-go
type sse struct {
	broadcast      chan event
	newClient      chan chan event
	closeClient    chan chan event
	clients        map[chan event]struct{}
	connectHandler func() []event
}

func newSSE() *sse {
	s := &sse{
		broadcast:   make(chan event, 1),
		newClient:   make(chan chan event),
		closeClient: make(chan chan event),
		clients:     make(map[chan event]struct{}),
	}
	go s.listen()
	return s
}

func (s *sse) listen() {
	for {
		select {
		case e := <-s.newClient:
			s.clients[e] = struct{}{}
		case e := <-s.closeClient:
			delete(s.clients, e)
		case e := <-s.broadcast:
			for c := range s.clients {
				c <- e
			}
		}
	}
}

func (s *sse) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	f, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming not supported.", http.StatusInternalServerError)
		return
	}

	// Set the headers related to event streaming.
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	c := make(chan event)
	s.newClient <- c

	// Remove this client from the map of connected clients
	// when this handler exits.
	defer func() {
		s.closeClient <- c
	}()

	// Listen to connection close and un-register client
	if notifier, ok := rw.(http.CloseNotifier); ok {
		n := notifier.CloseNotify()
		go func() {
			<-n
			s.closeClient <- c
		}()
	}

	// Send initial data
	if s.connectHandler != nil {
		for _, e := range s.connectHandler() {
			writeSSE(rw, e)
		}
		f.Flush()
	}

	for {
		// Write to the ResponseWriter
		writeSSE(rw, <-c)

		// Flush the data immediatly instead of buffering it for later.
		f.Flush()
	}
}

func writeSSE(w io.Writer, e event) {
	// Server Sent Events compatible
	if _, err := fmt.Fprintf(w, "data: %s\n\n", e); err != nil {
		log.Println(err)
	}
}
