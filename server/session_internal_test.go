package server

import (
	"net/netip"
	"slices"
	"testing"
	"time"

	"github.com/trains-io/z21.go/protocol"
)

type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time          { return c.t }
func (c *fakeClock) advance(d time.Duration) { c.t = c.t.Add(d) }

func newTestServer(t *testing.T, opts ...Option) (*Server, *fakeClock) {
	t.Helper()
	s, err := New("127.0.0.1:0", opts...)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	clock := &fakeClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	s.now = clock.now
	return s, clock
}

func send(t *testing.T, s *Server, from netip.AddrPort, msgs ...protocol.Message) {
	t.Helper()
	b, err := protocol.MarshalAll(msgs...)
	if err != nil {
		t.Fatalf("MarshalAll: %v", err)
	}
	s.handlePacket(peer{conn: s.conns[0], addr: from}, b)
}

var (
	clientA = netip.MustParseAddrPort("127.0.0.1:40001")
	clientB = netip.MustParseAddrPort("127.0.0.1:40002")
)

func TestSessionTimeout(t *testing.T) {
	tests := []struct {
		name      string
		opts      []Option
		heartbeat time.Duration // when clientA sends another request; 0 = never
		elapsed   time.Duration // when the broadcast is sent
		wantSent  int
	}{
		{name: "active within timeout", elapsed: 30 * time.Second, wantSent: 1},
		{name: "exactly at timeout", elapsed: DefaultSessionTimeout, wantSent: 1},
		{name: "expired after timeout", elapsed: DefaultSessionTimeout + time.Second, wantSent: 0},
		{name: "request keeps session alive", heartbeat: 50 * time.Second, elapsed: 100 * time.Second, wantSent: 1},
		{name: "custom timeout", opts: []Option{WithSessionTimeout(5 * time.Second)}, elapsed: 6 * time.Second, wantSent: 0},
		{name: "timeout disabled", opts: []Option{WithSessionTimeout(0)}, elapsed: time.Hour, wantSent: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, clock := newTestServer(t, tt.opts...)
			send(t, s, clientA, protocol.SetBroadcastFlags(protocol.BroadcastFlagSystemState))

			if tt.heartbeat > 0 {
				clock.advance(tt.heartbeat)
				send(t, s, clientA, protocol.GetHWInfo())
				clock.advance(tt.elapsed - tt.heartbeat)
			} else {
				clock.advance(tt.elapsed)
			}

			got := s.Broadcast(protocol.BroadcastFlagSystemState, protocol.EncodeSystemState(protocol.SystemState{}))
			if got != tt.wantSent {
				t.Fatalf("Broadcast reached %d clients, want %d", got, tt.wantSent)
			}
		})
	}
}

func TestExpiredSessionStartsFresh(t *testing.T) {
	s, clock := newTestServer(t)
	send(t, s, clientA, protocol.SetBroadcastFlags(protocol.BroadcastFlagSystemState))
	clock.advance(DefaultSessionTimeout + time.Second)

	if c := s.clientFor(peer{conn: s.conns[0], addr: clientA}); c.BroadcastFlags != 0 {
		t.Fatalf("BroadcastFlags after expiry = %#08x, want 0", c.BroadcastFlags)
	}
}

func locos(t *testing.T, s *Server, addr netip.AddrPort) []uint16 {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[addr]
	if !ok {
		t.Fatalf("no session for %v", addr)
	}
	return slices.Clone(sess.locos)
}

func TestGetLocoInfoSubscribes(t *testing.T) {
	tests := []struct {
		name  string
		flags uint32
		req   []uint16 // loco addresses requested in order
		want  []uint16
	}{
		{name: "requires XpressNet flag", flags: protocol.BroadcastFlagSystemState, req: []uint16{3}, want: nil},
		{name: "subscribes", flags: protocol.BroadcastFlagXpressNet, req: []uint16{3, 1234}, want: []uint16{3, 1234}},
		{name: "no duplicates", flags: protocol.BroadcastFlagXpressNet, req: []uint16{3, 4, 3}, want: []uint16{3, 4}},
		{
			name:  "drops oldest beyond 16",
			flags: protocol.BroadcastFlagXpressNet,
			req:   []uint16{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17},
			want:  []uint16{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := newTestServer(t)
			send(t, s, clientA, protocol.SetBroadcastFlags(tt.flags))
			for _, addr := range tt.req {
				send(t, s, clientA, protocol.GetLocoInfo(addr))
			}
			if got := locos(t, s, clientA); !slices.Equal(got, tt.want) {
				t.Fatalf("subscribed locos = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetBroadcastFlagsInSamePacketEnablesSubscription(t *testing.T) {
	s, _ := newTestServer(t)
	send(t, s, clientA, protocol.SetBroadcastFlags(protocol.BroadcastFlagXpressNet), protocol.GetLocoInfo(3))

	if got := locos(t, s, clientA); !slices.Equal(got, []uint16{3}) {
		t.Fatalf("subscribed locos = %v, want [3]", got)
	}
}

func TestLocoInfoMatch(t *testing.T) {
	tests := []struct {
		name  string
		flags uint32
		locos []uint16
		want  bool
	}{
		{name: "subscribed with XpressNet", flags: protocol.BroadcastFlagXpressNet, locos: []uint16{3}, want: true},
		{name: "subscribed to other loco", flags: protocol.BroadcastFlagXpressNet, locos: []uint16{4}, want: false},
		{name: "subscribed without XpressNet", flags: protocol.BroadcastFlagSystemState, locos: []uint16{3}, want: false},
		{name: "all loco info", flags: protocol.BroadcastFlagAllLocoInfo, want: true},
		{name: "no flags", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sess := &session{broadcastFlags: tt.flags, locos: tt.locos}
			if got := locoInfoMatch(3)(sess); got != tt.want {
				t.Fatalf("locoInfoMatch(3) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSessionPrunesExpiredSessions(t *testing.T) {
	s, clock := newTestServer(t)
	send(t, s, clientA, protocol.GetHWInfo())
	clock.advance(DefaultSessionTimeout + time.Second)
	send(t, s, clientB, protocol.GetHWInfo())

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[clientA]; ok {
		t.Fatal("expired session of client A still present")
	}
	if _, ok := s.sessions[clientB]; !ok {
		t.Fatal("session of client B missing")
	}
}
