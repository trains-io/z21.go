package server_test

import (
	"bytes"
	"log/slog"
	"net"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/trains-io/z21.go/protocol"
	"github.com/trains-io/z21.go/server"
)

// recordingObserver keeps a copy of every packet.
type recordingObserver struct {
	mu      sync.Mutex
	packets []server.Packet
}

func (o *recordingObserver) OnPacket(p server.Packet) {
	p.Raw = slices.Clone(p.Raw)
	o.mu.Lock()
	defer o.mu.Unlock()
	o.packets = append(o.packets, p)
}

func (o *recordingObserver) find(match func(server.Packet) bool) (server.Packet, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for _, p := range o.packets {
		if match(p) {
			return p, true
		}
	}
	return server.Packet{}, false
}

// syncBuffer is a bytes.Buffer safe for concurrent use.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func TestObserverSeesRequestAndReply(t *testing.T) {
	obs := &recordingObserver{}
	srv := startServer(t, server.WithHandler(server.HandlerFunc(hwInfoHandler)), server.WithObserver(obs))
	c := dial(t, srv)
	from := c.LocalAddr().(*net.UDPAddr).AddrPort()

	if _, err := c.Call(testContext(t), protocol.GetHWInfo()); err != nil {
		t.Fatalf("Call: %v", err)
	}

	rx, ok := obs.find(func(p server.Packet) bool { return p.Direction == server.DirectionRX })
	if !ok {
		t.Fatal("no RX packet observed")
	}
	if rx.Remote != from || len(rx.Messages) != 1 || rx.Messages[0].Header != protocol.HeaderLANGetHWInfo {
		t.Fatalf("RX packet = %+v, want LAN_GET_HWINFO from %v", rx, from)
	}
	tx, ok := obs.find(func(p server.Packet) bool { return p.Direction == server.DirectionTX })
	if !ok {
		t.Fatal("no TX packet observed")
	}
	if tx.Remote != from || tx.Broadcast || len(tx.Messages) != 1 || tx.Messages[0].Header != protocol.HeaderLANGetHWInfo {
		t.Fatalf("TX packet = %+v, want HW info reply to %v", tx, from)
	}
}

func TestObserverSeesMalformedPacket(t *testing.T) {
	obs := &recordingObserver{}
	srv := startServer(t, server.WithHandler(server.HandlerFunc(hwInfoHandler)), server.WithObserver(obs))
	c := dial(t, srv)

	raw, err := net.Dial("udp", srv.Addr().String())
	if err != nil {
		t.Fatalf("Dial raw: %v", err)
	}
	defer raw.Close()
	if _, err := raw.Write([]byte{0x02, 0x00}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	// The server handles packets in order, so this reply means the malformed one was seen.
	if _, err := c.Call(testContext(t), protocol.GetHWInfo()); err != nil {
		t.Fatalf("Call: %v", err)
	}

	from := raw.LocalAddr().(*net.UDPAddr).AddrPort()
	p, ok := obs.find(func(p server.Packet) bool { return p.Remote == from })
	if !ok {
		t.Fatal("malformed packet not observed")
	}
	if !bytes.Equal(p.Raw, []byte{0x02, 0x00}) || p.Messages != nil {
		t.Fatalf("observed %+v, want raw 02 00 without messages", p)
	}
}

func TestObserverSeesBroadcast(t *testing.T) {
	obs := &recordingObserver{}
	srv := startServer(t, server.WithObserver(obs))
	a, b := dial(t, srv), dial(t, srv)
	subscribe(t, a, protocol.BroadcastFlagSystemState)
	subscribe(t, b, protocol.BroadcastFlagSystemState)

	srv.Broadcast(protocol.BroadcastFlagSystemState, protocol.EncodeSystemState(protocol.SystemState{}))
	srv.Broadcast(protocol.BroadcastFlagRBus, protocol.EncodeRMBusStatus(protocol.RMBusStatus{}))

	p, ok := obs.find(func(p server.Packet) bool { return p.Broadcast && p.Flag == protocol.BroadcastFlagSystemState })
	if !ok {
		t.Fatal("system state broadcast not observed")
	}
	want := []string{a.LocalAddr().String(), b.LocalAddr().String()}
	got := make([]string, len(p.Recipients))
	for i, addr := range p.Recipients {
		got[i] = addr.String()
	}
	slices.Sort(want)
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Fatalf("recipients = %v, want %v", got, want)
	}

	p, ok = obs.find(func(p server.Packet) bool { return p.Broadcast && p.Flag == protocol.BroadcastFlagRBus })
	if !ok {
		t.Fatal("broadcast without recipients not observed")
	}
	if len(p.Recipients) != 0 {
		t.Fatalf("recipients = %v, want none", p.Recipients)
	}
}

func TestTracer(t *testing.T) {
	var out syncBuffer
	srv := startServer(t, server.WithHandler(server.HandlerFunc(hwInfoHandler)), server.WithObserver(server.NewTracer(&out)))
	c := dial(t, srv)
	subscribe(t, c, protocol.BroadcastFlagSystemState)

	if _, err := c.Call(testContext(t), protocol.GetHWInfo()); err != nil {
		t.Fatalf("Call: %v", err)
	}
	srv.Broadcast(protocol.BroadcastFlagSystemState, protocol.EncodeSystemState(protocol.SystemState{}))
	srv.Broadcast(protocol.BroadcastFlagRBus, protocol.EncodeRMBusStatus(protocol.RMBusStatus{}))

	trace := out.String()
	for _, want := range []string{
		"<< " + c.LocalAddr().String() + " (4 bytes)",
		">> " + c.LocalAddr().String() + " (12 bytes)",
		"LAN_GET_HWINFO",
		">> broadcast System state -> " + c.LocalAddr().String(),
		">> broadcast R-Bus feedback (0 recipients)",
		"00000000",
	} {
		if !strings.Contains(trace, want) {
			t.Errorf("trace missing %q:\n%s", want, trace)
		}
	}
}

func TestNewTracerNil(t *testing.T) {
	if tr := server.NewTracer(nil); tr != nil {
		t.Fatalf("NewTracer(nil) = %v, want nil", tr)
	}
}

func TestHandlerPanicIsRecovered(t *testing.T) {
	var logs syncBuffer
	srv := startServer(t,
		server.WithLogger(slog.New(slog.NewTextHandler(&logs, nil))),
		server.WithHandler(server.HandlerFunc(func(w server.ResponseWriter, r *server.Request) {
			if r.Message.Header == protocol.HeaderLANGetSerialNumber {
				panic("boom")
			}
			hwInfoHandler(w, r)
		})),
	)
	c := dial(t, srv)

	// The panicking dataset loses its reply; the other dataset in the packet is still answered.
	if err := c.SendMany(testContext(t), protocol.GetSerialNumber(), protocol.GetHWInfo()); err != nil {
		t.Fatalf("SendMany: %v", err)
	}
	expectPacket(t, c, protocol.HeaderLANGetHWInfo)
	if _, err := c.Call(testContext(t), protocol.GetHWInfo()); err != nil {
		t.Fatalf("server stopped answering after panic: %v", err)
	}
	if !strings.Contains(logs.String(), "z21 handler panic") || !strings.Contains(logs.String(), "boom") {
		t.Fatalf("panic not logged:\n%s", logs.String())
	}
}
