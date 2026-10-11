package servertest_test

import (
	"context"
	"testing"
	"time"

	"github.com/trains-io/z21.go/protocol"
	"github.com/trains-io/z21.go/server"
	"github.com/trains-io/z21.go/server/servertest"
)

func testContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestServer(t *testing.T) {
	want := protocol.HWInfo{HwType: protocol.HwTypeZ21New, FirmwareVersion: 0x0143}
	ts := servertest.NewServer(server.HandlerFunc(func(w server.ResponseWriter, r *server.Request) {
		if r.Message.Header == protocol.HeaderLANGetHWInfo {
			w.Reply(protocol.EncodeHWInfo(want))
		}
	}))
	defer ts.Close()

	c, err := ts.Dial()
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer c.Close()

	replies, err := c.Call(testContext(t), protocol.GetHWInfo())
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	got, err := protocol.HWInfoFromMessages(replies)
	if err != nil {
		t.Fatalf("HWInfoFromMessages: %v", err)
	}
	if got != want {
		t.Fatalf("HWInfo = %+v, want %+v", got, want)
	}
}

func TestServerBroadcast(t *testing.T) {
	ts := servertest.NewServer(server.NewMux())
	defer ts.Close()

	c, err := ts.Dial()
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer c.Close()

	ctx := testContext(t)
	if err := c.Send(ctx, protocol.SetBroadcastFlags(protocol.BroadcastFlagSystemState)); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if _, err := c.Call(ctx, protocol.GetBroadcastFlags()); err != nil {
		t.Fatalf("Call GetBroadcastFlags: %v", err)
	}

	if n := ts.Server.Broadcast(protocol.BroadcastFlagSystemState, protocol.EncodeSystemState(protocol.SystemState{})); n != 1 {
		t.Fatalf("Broadcast reached %d clients, want 1", n)
	}
	msgs, err := c.ReadPacket(ctx)
	if err != nil {
		t.Fatalf("ReadPacket: %v", err)
	}
	if _, err := protocol.SystemStateFromMessages(msgs); err != nil {
		t.Fatalf("SystemStateFromMessages: %v", err)
	}
}

func TestCloseTwice(t *testing.T) {
	ts := servertest.NewServer(server.NewMux())
	ts.Close()
	ts.Close()
}

func TestServerOptions(t *testing.T) {
	var seen bool
	obs := observerFunc(func(server.Packet) { seen = true })
	ts := servertest.NewServer(server.NewMux(), server.WithObserver(obs))

	c, err := ts.Dial()
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer c.Close()
	if _, err := c.Call(testContext(t), protocol.GetBroadcastFlags()); err != nil {
		t.Fatalf("Call: %v", err)
	}
	ts.Close()
	if !seen {
		t.Fatal("observer passed as option was not used")
	}
}

type observerFunc func(server.Packet)

func (f observerFunc) OnPacket(p server.Packet) { f(p) }
