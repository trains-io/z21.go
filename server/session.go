package server

import (
	"net/netip"
	"time"

	"github.com/trains-io/z21.go/protocol"
)

// DefaultSessionTimeout is how long a client may stay silent before the server
// forgets it (spec §1.1: clients must communicate at least once per minute).
const DefaultSessionTimeout = time.Minute

// session is the server-side state of one client, keyed by its UDP address.
type session struct {
	broadcastFlags uint32
	lastSeen       time.Time
}

// clientFor returns the client for addr, creating its session on first contact,
// and marks it as seen now.
func (s *Server) clientFor(addr netip.AddrPort) Client {
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[addr]
	if !ok || s.expired(sess, now) {
		s.pruneLocked(now)
		sess = &session{}
		s.sessions[addr] = sess
	}
	sess.lastSeen = now
	return Client{Addr: addr, BroadcastFlags: sess.broadcastFlags}
}

func (s *Server) expired(sess *session, now time.Time) bool {
	return s.sessionTimeout > 0 && now.Sub(sess.lastSeen) > s.sessionTimeout
}

// pruneLocked removes sessions that have been silent for longer than the
// session timeout. s.mu must be held.
func (s *Server) pruneLocked(now time.Time) {
	for addr, sess := range s.sessions {
		if s.expired(sess, now) {
			delete(s.sessions, addr)
		}
	}
}

// handleSession handles the requests that only change session state and reports
// whether msg was one of them (spec §2.2, §2.16, §2.17).
func (s *Server) handleSession(w *responseWriter, c Client, msg protocol.Message) bool {
	switch msg.Header {
	case protocol.HeaderLANSetBroadcastFlags:
		flags, err := protocol.ParseSetBroadcastFlags(msg.Data)
		if err != nil {
			s.log.Debug("z21 ignore malformed request", "from", c.Addr, "err", err)
			return true
		}
		s.mu.Lock()
		if sess, ok := s.sessions[c.Addr]; ok {
			sess.broadcastFlags = flags
		}
		s.mu.Unlock()
	case protocol.HeaderLANGetBroadcastFlags:
		w.Reply(protocol.EncodeBroadcastFlags(c.BroadcastFlags))
	case protocol.HeaderLANLogoff:
		s.mu.Lock()
		delete(s.sessions, c.Addr)
		s.mu.Unlock()
	default:
		return false
	}
	return true
}

// subscribers returns the addresses of active clients subscribed to any bit of
// flag, excluding except.
func (s *Server) subscribers(flag uint32, except netip.AddrPort) []netip.AddrPort {
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked(now)
	var out []netip.AddrPort
	for addr, sess := range s.sessions {
		if addr != except && sess.broadcastFlags&flag != 0 {
			out = append(out, addr)
		}
	}
	return out
}

// Broadcast sends msgs in one packet to every client subscribed to any bit of
// flag (protocol.BroadcastFlag*) and returns the number of clients it was sent to.
// It is safe to call from any goroutine, for example to push state changes that
// do not originate from a client request.
func (s *Server) Broadcast(flag uint32, msgs ...protocol.Message) int {
	return s.broadcast(flag, netip.AddrPort{}, msgs)
}

func (s *Server) broadcast(flag uint32, except netip.AddrPort, msgs []protocol.Message) int {
	if len(msgs) == 0 {
		return 0
	}
	out, err := protocol.MarshalAll(msgs...)
	if err != nil {
		s.log.Error("z21 marshal broadcast", "err", err)
		return 0
	}
	sent := 0
	for _, addr := range s.subscribers(flag, except) {
		if s.write(addr, out) {
			sent++
		}
	}
	return sent
}
