package server

import (
	"net"
	"net/netip"
	"slices"
	"time"

	"github.com/trains-io/z21.go/protocol"
)

// DefaultSessionTimeout is how long a client may stay silent before the server
// forgets it (spec §1.1: clients must communicate at least once per minute).
const DefaultSessionTimeout = time.Minute

// maxLocoSubscriptions is the number of loco addresses a client can subscribe
// to; the oldest is dropped first (spec §4).
const maxLocoSubscriptions = 16

// session is the server-side state of one client, keyed by its UDP address.
type session struct {
	conn           *net.UDPConn // socket the client last sent to; replies and broadcasts use it
	broadcastFlags uint32
	lastSeen       time.Time
	locos          []uint16 // subscribed loco addresses, oldest first
}

func (sess *session) subscribedTo(addr uint16) bool {
	return slices.Contains(sess.locos, addr)
}

// subscribeLoco handles the subscription side effect of LAN_X_GET_LOCO_INFO,
// which only applies to clients with BroadcastFlagXpressNet set (spec §4.1).
func (s *Server) subscribeLoco(c Client, msg protocol.Message) {
	if msg.Header != protocol.HeaderLANX || c.BroadcastFlags&protocol.BroadcastFlagXpressNet == 0 {
		return
	}
	addr, err := protocol.ParseGetLocoInfo(msg.Data)
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[c.Addr]
	if !ok || sess.subscribedTo(addr) {
		return
	}
	if len(sess.locos) == maxLocoSubscriptions {
		sess.locos = slices.Delete(sess.locos, 0, 1)
	}
	sess.locos = append(sess.locos, addr)
}

// clientFor returns the client for p, creating its session on first contact,
// and marks it as seen now on the socket it used.
func (s *Server) clientFor(p peer) Client {
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[p.addr]
	if !ok || s.expired(sess, now) {
		s.pruneLocked(now)
		sess = &session{}
		s.sessions[p.addr] = sess
	}
	sess.conn = p.conn
	sess.lastSeen = now
	return Client{Addr: p.addr, BroadcastFlags: sess.broadcastFlags}
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

// recipients returns the active clients whose session matches, excluding except.
func (s *Server) recipients(except netip.AddrPort, match func(*session) bool) []peer {
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked(now)
	var out []peer
	for addr, sess := range s.sessions {
		if addr != except && match(sess) {
			out = append(out, peer{conn: sess.conn, addr: addr})
		}
	}
	return out
}

func flagMatch(flag uint32) func(*session) bool {
	return func(sess *session) bool { return sess.broadcastFlags&flag != 0 }
}

// locoInfoMatch selects clients that receive LAN_X_LOCO_INFO for addr: those
// subscribed to addr with BroadcastFlagXpressNet set, and those with
// BroadcastFlagAllLocoInfo (spec §2.16, §4.4).
func locoInfoMatch(addr uint16) func(*session) bool {
	return func(sess *session) bool {
		if sess.broadcastFlags&protocol.BroadcastFlagAllLocoInfo != 0 {
			return true
		}
		return sess.broadcastFlags&protocol.BroadcastFlagXpressNet != 0 && sess.subscribedTo(addr)
	}
}

// Broadcast sends msgs in one packet to every client subscribed to any bit of
// flag (protocol.BroadcastFlag*) and returns the number of clients it was sent to.
// It is safe to call from any goroutine, for example to push state changes that
// do not originate from a client request.
func (s *Server) Broadcast(flag uint32, msgs ...protocol.Message) int {
	return s.send(flag, s.recipients(netip.AddrPort{}, flagMatch(flag)), msgs)
}

// BroadcastLocoInfo sends LAN_X_LOCO_INFO for info.Address to the clients that
// subscribed to that loco and to clients with BroadcastFlagAllLocoInfo, and
// returns the number of clients it was sent to. It is safe to call from any
// goroutine.
func (s *Server) BroadcastLocoInfo(info protocol.LocoInfo) int {
	return s.send(
		locoInfoFlags,
		s.recipients(netip.AddrPort{}, locoInfoMatch(info.Address)),
		[]protocol.Message{protocol.EncodeLocoInfo(info)},
	)
}

// locoInfoFlags are the broadcast flags reported to observers for loco info.
const locoInfoFlags = protocol.BroadcastFlagXpressNet | protocol.BroadcastFlagAllLocoInfo

// send writes msgs as one packet to each address in to and returns the number
// of successful writes. flag is only reported to the observer.
func (s *Server) send(flag uint32, to []peer, msgs []protocol.Message) int {
	if len(msgs) == 0 {
		return 0
	}
	out, err := protocol.MarshalAll(msgs...)
	if err != nil {
		s.log.Error("z21 marshal broadcast", "err", err)
		return 0
	}
	addrs := make([]netip.AddrPort, len(to))
	for i, p := range to {
		addrs[i] = p.addr
	}
	s.observe(Packet{Direction: DirectionTX, Broadcast: true, Flag: flag, Recipients: addrs, Raw: out, Messages: msgs})
	sent := 0
	for _, p := range to {
		if s.write(p, out) {
			sent++
		}
	}
	return sent
}
