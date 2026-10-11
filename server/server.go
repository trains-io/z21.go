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

// AlternatePort is the second UDP port a Z21 answers on (spec §1.1). Clients
// should prefer DefaultPort; serve both with WithAdditionalAddr.
const AlternatePort = 21106

// maxPacketSize is the largest UDP payload the server reads.
const maxPacketSize = 65535

// ErrServerClosed is returned by Serve after Close.
var ErrServerClosed = errors.New("z21: server closed")

// Server is a Z21 LAN server listening on one or more UDP sockets that share
// the same client sessions.
type Server struct {
	conns           []*net.UDPConn // conns[0] is bound to the address passed to New
	additionalAddrs []string
	handler         Handler
	log             *slog.Logger
	observer        Observer

	sessionTimeout time.Duration
	now            func() time.Time

	// handleMu serializes packet handling across sockets, so the handler is
	// never called concurrently.
	handleMu sync.Mutex

	mu       sync.Mutex
	sessions map[netip.AddrPort]*session
}

// peer is a client address and the server socket the client talks to.
type peer struct {
	conn *net.UDPConn
	addr netip.AddrPort
}

// New binds a UDP socket on addr (for example ":21105" or "127.0.0.1:0") and
// on every address given with WithAdditionalAddr. Call Serve to start handling
// requests.
func New(addr string, opts ...Option) (*Server, error) {
	s := &Server{
		sessionTimeout: DefaultSessionTimeout,
		now:            time.Now,
		sessions:       make(map[netip.AddrPort]*session),
	}
	applyOptions(s, opts)
	if s.log == nil {
		s.log = slog.New(slog.DiscardHandler)
	}

	for _, a := range append([]string{addr}, s.additionalAddrs...) {
		conn, err := listen(a)
		if err != nil {
			_ = s.Close()
			return nil, err
		}
		s.conns = append(s.conns, conn)
	}
	return s, nil
}

func listen(addr string) (*net.UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("z21: resolve %q: %w", addr, err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, fmt.Errorf("z21: listen %q: %w", addr, err)
	}
	return conn, nil
}

// Addr returns the local address of the socket bound to the address passed to New.
func (s *Server) Addr() net.Addr {
	return s.conns[0].LocalAddr()
}

// Addrs returns the local addresses of all sockets, starting with Addr.
func (s *Server) Addrs() []net.Addr {
	addrs := make([]net.Addr, len(s.conns))
	for i, conn := range s.conns {
		addrs[i] = conn.LocalAddr()
	}
	return addrs
}

// Close closes all sockets. A running Serve returns ErrServerClosed.
func (s *Server) Close() error {
	var errs []error
	for _, conn := range s.conns {
		errs = append(errs, conn.Close())
	}
	return errors.Join(errs...)
}

// Serve reads and handles packets on all sockets until ctx is cancelled
// (returns nil) or the server is closed (returns ErrServerClosed). If reading
// from one socket fails, Serve stops the others and returns that error.
func (s *Server) Serve(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errs := make(chan error, len(s.conns))
	for _, conn := range s.conns {
		go func() { errs <- s.serveConn(ctx, conn) }()
	}
	err := <-errs
	cancel()
	for range len(s.conns) - 1 {
		<-errs
	}
	return err
}

func (s *Server) serveConn(ctx context.Context, conn *net.UDPConn) error {
	stop := context.AfterFunc(ctx, func() {
		_ = conn.SetReadDeadline(time.Now())
	})
	defer stop()

	buf := make([]byte, maxPacketSize)
	for {
		n, remote, err := conn.ReadFromUDPAddrPort(buf)
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
		s.handlePacket(peer{conn: conn, addr: remote}, buf[:n])
	}
}

// handlePacket handles the datasets of one packet in order. Session requests
// (broadcast flags, logoff) are handled by the server; all others go to the
// handler.
func (s *Server) handlePacket(from peer, payload []byte) {
	s.handleMu.Lock()
	defer s.handleMu.Unlock()

	msgs, err := protocol.ParseAll(payload)
	s.observe(Packet{Direction: DirectionRX, Remote: from.addr, Raw: payload, Messages: msgs})
	if err != nil {
		s.log.Debug("z21 ignore malformed packet", "from", from.addr, "err", err)
		return
	}

	w := &responseWriter{srv: s, sender: from.addr}
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

	s.sendReply(from, w.replies)
	for _, b := range w.broadcasts {
		s.send(b.flag, b.to, b.msgs)
	}
}

func (s *Server) sendReply(to peer, replies []protocol.Message) {
	if len(replies) == 0 {
		return
	}
	out, err := protocol.MarshalAll(replies...)
	if err != nil {
		s.log.Error("z21 marshal reply", "to", to.addr, "err", err)
		return
	}
	s.observe(Packet{Direction: DirectionTX, Remote: to.addr, Raw: out, Messages: replies})
	s.write(to, out)
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

func (s *Server) write(to peer, out []byte) bool {
	if _, err := to.conn.WriteToUDPAddrPort(out, to.addr); err != nil {
		s.log.Error("z21 write", "to", to.addr, "err", err)
		return false
	}
	return true
}
