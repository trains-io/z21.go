package client_test

import (
	"bytes"
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/trains-io/z21.go/client"
	"github.com/trains-io/z21.go/protocol"
)

type recordingObserver struct {
	mu    sync.Mutex
	events []recordedEvent
}

type recordedEvent struct {
	dir    client.Direction
	remote string
	rawLen int
	msgLen int
}

func (o *recordingObserver) OnPacket(_ context.Context, dir client.Direction, remote string, raw []byte, msgs []protocol.Message) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.events = append(o.events, recordedEvent{
		dir:    dir,
		remote: remote,
		rawLen: len(raw),
		msgLen: len(msgs),
	})
}

func TestWithObserverOnDial(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	go func() {
		buf := make([]byte, 1472)
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		reply, _ := protocol.Message{Header: protocol.HeaderLANGetHWInfo, Data: []byte{0x01}}.Marshal()
		_, _ = conn.WriteToUDP(reply, addr)
		_ = n
	}()

	var rec recordingObserver
	c, err := client.Dial(conn.LocalAddr().String(), client.WithObserver(&rec))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	if _, err := c.Call(t.Context(), protocol.GetHWInfo()); err != nil {
		t.Fatal(err)
	}

	rec.mu.Lock()
	defer rec.mu.Unlock()
	if len(rec.events) != 2 {
		t.Fatalf("len(events) = %d, want tx+rx", len(rec.events))
	}
	if rec.events[0].dir != client.DirectionTX || rec.events[1].dir != client.DirectionRX {
		t.Fatalf("events = %+v", rec.events)
	}
}

func TestMultiObserver(t *testing.T) {
	var a, b recordingObserver
	multi := client.MultiObserver{Observers: []client.Observer{&a, &b}}

	var buf bytes.Buffer
	tr := client.NewTracer(&buf)
	multi.Observers = append(multi.Observers, tr)

	raw := []byte{0x04, 0x00, 0x1a, 0x00, 0x21, 0x00, 0x00, 0x00}
	msgs := []protocol.Message{{Header: protocol.HeaderLANGetHWInfo}}
	multi.OnPacket(context.Background(), client.DirectionTX, "127.0.0.1:21105", raw, msgs)

	a.mu.Lock()
	b.mu.Lock()
	defer a.mu.Unlock()
	defer b.mu.Unlock()

	if len(a.events) != 1 || len(b.events) != 1 {
		t.Fatalf("events a=%d b=%d, want 1 each", len(a.events), len(b.events))
	}
	if buf.Len() == 0 {
		t.Fatal("expected hexdump tracer output")
	}
}

func TestNopObserver(t *testing.T) {
	client.NopObserver{}.OnPacket(context.Background(), client.DirectionTX, "x", []byte{1}, nil)
}

func TestSetObserverNil(t *testing.T) {
	c, err := client.Dial("127.0.0.1:1")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.SetObserver(nil)
}
