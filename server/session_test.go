package server_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/trains-io/z21.go/client"
	"github.com/trains-io/z21.go/protocol"
	"github.com/trains-io/z21.go/server"
)

// subscribe sets the client's broadcast flags and waits until the server has
// applied them.
func subscribe(t *testing.T, c *client.Client, flags uint32) {
	t.Helper()
	ctx := testContext(t)
	if err := c.Send(ctx, protocol.SetBroadcastFlags(flags)); err != nil {
		t.Fatalf("Send SetBroadcastFlags: %v", err)
	}
	if got := getBroadcastFlags(t, c); got != flags {
		t.Fatalf("broadcast flags = %#08x, want %#08x", got, flags)
	}
}

func getBroadcastFlags(t *testing.T, c *client.Client) uint32 {
	t.Helper()
	replies, err := c.Call(testContext(t), protocol.GetBroadcastFlags())
	if err != nil {
		t.Fatalf("Call GetBroadcastFlags: %v", err)
	}
	flags, err := protocol.ParseBroadcastFlags(replies[0].Data)
	if err != nil {
		t.Fatalf("ParseBroadcastFlags: %v", err)
	}
	return flags
}

func expectPacket(t *testing.T, c *client.Client, header uint16) {
	t.Helper()
	msgs, err := c.ReadPacket(testContext(t))
	if err != nil {
		t.Fatalf("ReadPacket: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Header != header {
		t.Fatalf("got %+v, want one dataset with header %#04x", msgs, header)
	}
}

func expectNoPacket(t *testing.T, c *client.Client) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if msgs, err := c.ReadPacket(ctx); err == nil {
		t.Fatalf("got unexpected packet %+v", msgs)
	}
}

func TestSessionRequestsAreHandledByServer(t *testing.T) {
	var (
		mu      sync.Mutex
		handled []uint16
	)
	srv := startServer(t, server.WithHandler(server.HandlerFunc(func(w server.ResponseWriter, r *server.Request) {
		mu.Lock()
		handled = append(handled, r.Message.Header)
		mu.Unlock()
		hwInfoHandler(w, r)
	})))
	c := dial(t, srv)

	subscribe(t, c, protocol.DefaultBroadcastFlags)
	if err := c.Send(testContext(t), protocol.Logoff()); err != nil {
		t.Fatalf("Send Logoff: %v", err)
	}
	if _, err := c.Call(testContext(t), protocol.GetHWInfo()); err != nil {
		t.Fatalf("Call GetHWInfo: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(handled) != 1 || handled[0] != protocol.HeaderLANGetHWInfo {
		t.Fatalf("handler saw headers %#04x, want only LAN_GET_HWINFO", handled)
	}
}

func TestRequestCarriesBroadcastFlags(t *testing.T) {
	got := make(chan uint32, 1)
	srv := startServer(t, server.WithHandler(server.HandlerFunc(func(w server.ResponseWriter, r *server.Request) {
		got <- r.Client.BroadcastFlags
		hwInfoHandler(w, r)
	})))
	c := dial(t, srv)

	subscribe(t, c, protocol.BroadcastFlagRBus)
	if _, err := c.Call(testContext(t), protocol.GetHWInfo()); err != nil {
		t.Fatalf("Call: %v", err)
	}
	if flags := <-got; flags != protocol.BroadcastFlagRBus {
		t.Fatalf("Client.BroadcastFlags = %#08x, want %#08x", flags, protocol.BroadcastFlagRBus)
	}
}

func TestLogoffRemovesSession(t *testing.T) {
	srv := startServer(t)
	c := dial(t, srv)

	subscribe(t, c, protocol.BroadcastFlagXpressNet)
	if err := c.Send(testContext(t), protocol.Logoff()); err != nil {
		t.Fatalf("Send Logoff: %v", err)
	}
	// The next request starts a fresh session with no flags.
	if got := getBroadcastFlags(t, c); got != 0 {
		t.Fatalf("broadcast flags after logoff = %#08x, want 0", got)
	}
	if n := srv.Broadcast(protocol.BroadcastFlagXpressNet, protocol.EncodeTrackPowerBC(true)); n != 0 {
		t.Fatalf("Broadcast reached %d clients after logoff, want 0", n)
	}
}

func TestResponseWriterBroadcastExcludesSender(t *testing.T) {
	srv := startServer(t, server.WithHandler(server.HandlerFunc(func(w server.ResponseWriter, r *server.Request) {
		if r.Message.Header == protocol.HeaderLANX {
			w.Broadcast(protocol.BroadcastFlagXpressNet, protocol.EncodeTrackPowerBC(true))
			w.Reply(protocol.EncodeTrackPowerBC(true))
		}
	})))
	sender, xbus, sysState := dial(t, srv), dial(t, srv), dial(t, srv)
	subscribe(t, sender, protocol.BroadcastFlagXpressNet)
	subscribe(t, xbus, protocol.BroadcastFlagXpressNet)
	subscribe(t, sysState, protocol.BroadcastFlagSystemState)

	if err := sender.Send(testContext(t), protocol.SetTrackPower(true)); err != nil {
		t.Fatalf("Send: %v", err)
	}

	expectPacket(t, sender, protocol.HeaderLANX) // the reply
	expectPacket(t, xbus, protocol.HeaderLANX)   // the broadcast
	expectNoPacket(t, sender)                    // no broadcast on top of the reply
	expectNoPacket(t, sysState)
}

func TestServerBroadcast(t *testing.T) {
	srv := startServer(t)
	a, b, other := dial(t, srv), dial(t, srv), dial(t, srv)
	subscribe(t, a, protocol.BroadcastFlagSystemState)
	subscribe(t, b, protocol.DefaultBroadcastFlags)
	subscribe(t, other, protocol.BroadcastFlagRBus)

	n := srv.Broadcast(protocol.BroadcastFlagSystemState, protocol.EncodeSystemState(protocol.SystemState{}))
	if n != 2 {
		t.Fatalf("Broadcast reached %d clients, want 2", n)
	}
	expectPacket(t, a, protocol.HeaderLANSystemStateDataChanged)
	expectPacket(t, b, protocol.HeaderLANSystemStateDataChanged)
	expectNoPacket(t, other)
}

func TestServerBroadcastWithoutMessages(t *testing.T) {
	srv := startServer(t)
	c := dial(t, srv)
	subscribe(t, c, protocol.BroadcastFlagSystemState)

	if n := srv.Broadcast(protocol.BroadcastFlagSystemState); n != 0 {
		t.Fatalf("Broadcast() reached %d clients, want 0", n)
	}
}
