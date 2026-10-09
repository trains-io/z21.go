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

// Client describes the sender of a request as seen when the request arrived.
type Client struct {
	Addr           netip.AddrPort
	BroadcastFlags uint32 // set by LAN_SET_BROADCASTFLAGS (spec §2.16)
}

// ResponseWriter sends replies and broadcasts for a request.
type ResponseWriter interface {
	// Reply queues msgs for the sender. Replies to all datasets of one incoming
	// UDP packet go back together in a single packet after the last dataset
	// has been handled.
	Reply(msgs ...protocol.Message)

	// Broadcast immediately sends msgs to every other client subscribed to any
	// bit of flag (protocol.BroadcastFlag*). The sender is excluded; reply to it
	// with Reply.
	Broadcast(flag uint32, msgs ...protocol.Message)

	// BroadcastLocoInfo immediately sends LAN_X_LOCO_INFO for info.Address to
	// every client subscribed to that loco, including the sender, as a real
	// Z21 does after a drive or function command (spec §4.4). Clients subscribe
	// with LAN_X_GET_LOCO_INFO; the server tracks this before calling the
	// handler.
	BroadcastLocoInfo(info protocol.LocoInfo)
}

type responseWriter struct {
	srv     *Server
	sender  netip.AddrPort
	replies []protocol.Message
}

func (w *responseWriter) Reply(msgs ...protocol.Message) {
	w.replies = append(w.replies, msgs...)
}

func (w *responseWriter) Broadcast(flag uint32, msgs ...protocol.Message) {
	w.srv.send(w.srv.recipients(w.sender, flagMatch(flag)), msgs)
}

func (w *responseWriter) BroadcastLocoInfo(info protocol.LocoInfo) {
	w.srv.BroadcastLocoInfo(info)
}
