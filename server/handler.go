package server

import (
	"net/netip"

	"github.com/trains-io/z21.go/protocol"
)

// Handler responds to one Z21 request dataset.
//
// ServeZ21 is called once per dataset, in the order the datasets arrive. Datasets
// from all clients are handled sequentially, so a handler must not block.
type Handler interface {
	ServeZ21(w ResponseWriter, r *Request)
}

// HandlerFunc adapts an ordinary function to a Handler.
type HandlerFunc func(w ResponseWriter, r *Request)

// ServeZ21 calls f(w, r).
func (f HandlerFunc) ServeZ21(w ResponseWriter, r *Request) { f(w, r) }

// Request is a single dataset received from a client.
type Request struct {
	Message protocol.Message
	Client  Client
}

// Client identifies the sender of a request.
type Client struct {
	Addr netip.AddrPort
}

// ResponseWriter collects the replies to a request.
//
// Replies to all datasets of one incoming UDP packet are sent back to the client
// together in a single packet after the last dataset has been handled.
type ResponseWriter interface {
	Reply(msgs ...protocol.Message)
}

type responseWriter struct {
	replies []protocol.Message
}

func (w *responseWriter) Reply(msgs ...protocol.Message) {
	w.replies = append(w.replies, msgs...)
}
