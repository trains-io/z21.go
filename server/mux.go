package server

import (
	"fmt"
	"sync"

	"github.com/trains-io/z21.go/protocol"
)

// Mux routes requests to handlers by LAN header and, for LAN_X datasets, by
// X-header and DB0. A Mux is itself a Handler. It is safe for concurrent use.
//
// LAN_X requests are matched in this order:
//
//  1. a handler registered with HandleLANXDB0 for the X-header and DB0;
//  2. a handler registered with HandleLANX for the X-header;
//  3. a handler registered with Handle for protocol.HeaderLANX;
//  4. otherwise the Mux replies LAN_X_UNKNOWN_COMMAND (spec §2.11), as it does
//     for malformed LAN_X requests (too short or wrong checksum).
//
// Requests with any other unregistered header are ignored. Session requests
// (broadcast flags, logoff) are handled by the Server and never reach a Mux.
type Mux struct {
	mu   sync.RWMutex
	lan  map[uint16]Handler
	x    map[byte]Handler
	xDB0 map[[2]byte]Handler
}

// NewMux returns an empty Mux.
func NewMux() *Mux {
	return &Mux{
		lan:  make(map[uint16]Handler),
		x:    make(map[byte]Handler),
		xDB0: make(map[[2]byte]Handler),
	}
}

// Handle registers h for requests with the LAN header (protocol.HeaderLAN*).
// It panics if h is nil or a handler is already registered for header.
func (m *Mux) Handle(header uint16, h Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	register(m.lan, header, h, fmt.Sprintf("LAN header %#04x", header))
}

// HandleLANX registers h for LAN_X requests with xHeader, whatever their DB0.
// It panics if h is nil or a handler is already registered for xHeader.
func (m *Mux) HandleLANX(xHeader byte, h Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	register(m.x, xHeader, h, fmt.Sprintf("X-header %#02x", xHeader))
}

// HandleLANXDB0 registers h for LAN_X requests with xHeader and first data byte
// db0. It panics if h is nil or a handler is already registered for the pair.
func (m *Mux) HandleLANXDB0(xHeader, db0 byte, h Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	register(m.xDB0, [2]byte{xHeader, db0}, h, fmt.Sprintf("X-header %#02x DB0 %#02x", xHeader, db0))
}

func register[K comparable](routes map[K]Handler, key K, h Handler, name string) {
	if h == nil {
		panic("z21: nil handler for " + name)
	}
	if _, ok := routes[key]; ok {
		panic("z21: multiple registrations for " + name)
	}
	routes[key] = h
}

// ServeZ21 dispatches r to the matching handler.
func (m *Mux) ServeZ21(w ResponseWriter, r *Request) {
	if h := m.handler(r.Message); h != nil {
		h.ServeZ21(w, r)
	}
}

func (m *Mux) handler(msg protocol.Message) Handler {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if msg.Header != protocol.HeaderLANX {
		return m.lan[msg.Header]
	}
	xHeader, data, err := protocol.ParseLANX(msg)
	if err != nil {
		return unknownCommand
	}
	if len(data) > 0 {
		if h, ok := m.xDB0[[2]byte{xHeader, data[0]}]; ok {
			return h
		}
	}
	if h, ok := m.x[xHeader]; ok {
		return h
	}
	if h, ok := m.lan[protocol.HeaderLANX]; ok {
		return h
	}
	return unknownCommand
}

var unknownCommand = HandlerFunc(func(w ResponseWriter, _ *Request) {
	w.Reply(protocol.EncodeUnknownCommand())
})
