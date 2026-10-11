// Package server implements a reusable Z21 LAN protocol server: the UDP
// transport and the dispatch of request datasets to a Handler.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/netip"
	"runtime/debug"
	"sync"
	"time"

	"github.com/trains-io/z21.go/protocol"
)

// DefaultPort is the Z21 LAN UDP port (spec §1.1).
const DefaultPort = 21105

// maxPacketSize is the largest UDP payload the server reads.
const maxPacketSize = 65535

// ErrServerClosed is returned by Serve after Close.
var ErrServerClosed = errors.New("z21: server closed")

// Server is a Z21 LAN server listening on a UDP socket.
type Server struct {
	conn     *net.UDPConn
	handler  Handler
	log      *slog.Logger
	observer Observer

	sessionTimeout time.Duration
	now            func() time.Time

	mu       sync.Mutex
	sessions map[netip.AddrPort]*session
}

// New binds a UDP socket on addr (for example ":21105" or "127.0.0.1:0").
// Call Serve to start handling requests.
func New(addr string, opts ...Option) (*Server, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("z21: resolve %q: %w", addr, err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, fmt.Errorf("z21: listen %q: %w", addr, err)
	}

	s := &Server{
		conn:           conn,
		sessionTimeout: DefaultSessionTimeout,
		now:            time.Now,
		sessions:       make(map[netip.AddrPort]*session),
	}
	applyOptions(s, opts)
	if s.log == nil {
		s.log = slog.New(slog.DiscardHandler)
	}
	return s, nil
}

// Addr returns the local address the server is bound to.
func (s *Server) Addr() net.Addr {
	return s.conn.LocalAddr()
}

// Close closes the socket. A running Serve returns ErrServerClosed.
func (s *Server) Close() error {
	return s.conn.Close()
}

// Serve reads and handles packets until ctx is cancelled (returns nil) or the
// server is closed (returns ErrServerClosed).
func (s *Server) Serve(ctx context.Context) error {
	stop := context.AfterFunc(ctx, func() {
		_ = s.conn.SetReadDeadline(time.Now())
	})
	defer stop()

	buf := make([]byte, maxPacketSize)
	for {
		n, remote, err := s.conn.ReadFromUDPAddrPort(buf)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			if errors.Is(err, net.ErrClosed) {
				return ErrServerClosed
			}
			return fmt.Errorf("z21: read: %w", err)
		}
		remote = netip.AddrPortFrom(remote.Addr().Unmap(), remote.Port())
		s.handlePacket(remote, buf[:n])
	}
}

// handlePacket handles the datasets of one packet in order. Session requests
// (broadcast flags, logoff) are handled by the server; all others go to the
// handler.
func (s *Server) handlePacket(from netip.AddrPort, payload []byte) {
	msgs, err := protocol.ParseAll(payload)
	s.observe(Packet{Direction: DirectionRX, Remote: from, Raw: payload, Messages: msgs})
	if err != nil {
		s.log.Debug("z21 ignore malformed packet", "from", from, "err", err)
		return
	}

	w := &responseWriter{srv: s, sender: from}
	for _, msg := range msgs {
		c := s.clientFor(from)
		if s.handleSession(w, c, msg) {
			continue
		}
		s.subscribeLoco(c, msg)
		if s.handler == nil {
			continue
		}
		s.serveZ21(w, &Request{Message: msg, Client: c})
	}
	if len(w.replies) == 0 {
		return
	}

	out, err := protocol.MarshalAll(w.replies...)
	if err != nil {
		s.log.Error("z21 marshal reply", "to", from, "err", err)
		return
	}
	s.observe(Packet{Direction: DirectionTX, Remote: from, Raw: out, Messages: w.replies})
	s.write(from, out)
}

// serveZ21 calls the handler and recovers from a panic, so a faulty handler
// only loses the reply to one dataset instead of stopping the server.
func (s *Server) serveZ21(w ResponseWriter, r *Request) {
	defer func() {
		if p := recover(); p != nil {
			s.log.Error("z21 handler panic", "from", r.Client.Addr, "request", protocol.MessageName(r.Message),
				"panic", p, "stack", string(debug.Stack()))
		}
	}()
	s.handler.ServeZ21(w, r)
}

func (s *Server) observe(p Packet) {
	if s.observer != nil {
		s.observer.OnPacket(p)
	}
}

func (s *Server) write(to netip.AddrPort, out []byte) bool {
	if _, err := s.conn.WriteToUDPAddrPort(out, to); err != nil {
		s.log.Error("z21 write", "to", to, "err", err)
		return false
	}
	return true
}
