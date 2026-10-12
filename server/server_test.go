package server_test

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/trains-io/z21.go/client"
	"github.com/trains-io/z21.go/protocol"
	"github.com/trains-io/z21.go/server"
)

var testHWInfo = protocol.HWInfo{HwType: protocol.HwTypeZ21New, FirmwareVersion: 0x0143}

// hwInfoHandler answers LAN_GET_HWINFO and LAN_GET_SERIAL_NUMBER.
func hwInfoHandler(w server.ResponseWriter, r *server.Request) {
	switch r.Message.Header {
	case protocol.HeaderLANGetHWInfo:
		w.Reply(protocol.EncodeHWInfo(testHWInfo))
	case protocol.HeaderLANGetSerialNumber:
		w.Reply(protocol.EncodeSerialNumber(123456))
	}
}

// startServer runs a server on 127.0.0.1:0 until the test ends.
func startServer(t *testing.T, opts ...server.Option) *server.Server {
	t.Helper()
	return startServerOn(t, "127.0.0.1:0", opts...)
}

func startServerOn(t *testing.T, addr string, opts ...server.Option) *server.Server {
	t.Helper()
	srv, err := server.New(addr, opts...)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- srv.Serve(ctx) }()
	t.Cleanup(func() {
		cancel()
		if err := <-done; err != nil {
			t.Errorf("Serve: %v", err)
		}
		srv.Close()
	})
	return srv
}

func dial(t *testing.T, srv *server.Server) *client.Client {
	t.Helper()
	c, err := client.Dial(srv.Addr().String())
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

func testContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestServeReplies(t *testing.T) {
	srv := startServer(t, server.WithHandler(server.HandlerFunc(hwInfoHandler)))
	c := dial(t, srv)

	replies, err := c.Call(testContext(t), protocol.GetHWInfo())
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if len(replies) != 1 {
		t.Fatalf("got %d replies, want 1", len(replies))
	}
	info, err := protocol.ParseHWInfo(replies[0].Data)
	if err != nil {
		t.Fatalf("ParseHWInfo: %v", err)
	}
	if info != testHWInfo {
		t.Fatalf("HWInfo = %+v, want %+v", info, testHWInfo)
	}
}

func TestServeRepliesToAllDatasetsInOnePacket(t *testing.T) {
	srv := startServer(t, server.WithHandler(server.HandlerFunc(hwInfoHandler)))
	c := dial(t, srv)
	ctx := testContext(t)

	if err := c.SendMany(ctx, protocol.GetHWInfo(), protocol.GetSerialNumber()); err != nil {
		t.Fatalf("SendMany: %v", err)
	}
	msgs, err := c.ReadPacket(ctx)
	if err != nil {
		t.Fatalf("ReadPacket: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("got %d datasets in reply packet, want 2", len(msgs))
	}
	if msgs[0].Header != protocol.HeaderLANGetHWInfo || msgs[1].Header != protocol.HeaderLANGetSerialNumber {
		t.Fatalf("reply headers = %#04x, %#04x; want HW info then serial number", msgs[0].Header, msgs[1].Header)
	}
}

func TestRequestClientAddr(t *testing.T) {
	tests := []struct {
		name   string
		listen string
	}{
		{name: "IPv4 socket", listen: "127.0.0.1:0"},
		{name: "dual-stack socket reports unmapped IPv4", listen: ":0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make(chan netip.AddrPort, 1)
			srv := startServerOn(t, tt.listen, server.WithHandler(server.HandlerFunc(func(w server.ResponseWriter, r *server.Request) {
				got <- r.Client.Addr
				w.Reply(protocol.EncodeHWInfo(testHWInfo))
			})))
			port := srv.Addr().(*net.UDPAddr).Port
			c, err := client.Dial(netip.AddrPortFrom(netip.MustParseAddr("127.0.0.1"), uint16(port)).String())
			if err != nil {
				t.Fatalf("Dial: %v", err)
			}
			defer c.Close()

			if _, err := c.Call(testContext(t), protocol.GetHWInfo()); err != nil {
				t.Fatalf("Call: %v", err)
			}
			want := c.LocalAddr().(*net.UDPAddr).AddrPort()
			if addr := <-got; addr != want {
				t.Fatalf("Client.Addr = %v, want %v", addr, want)
			}
		})
	}
}

func TestServeIgnoresMalformedPacket(t *testing.T) {
	srv := startServer(t, server.WithHandler(server.HandlerFunc(hwInfoHandler)))
	c := dial(t, srv)

	raw, err := net.Dial("udp", srv.Addr().String())
	if err != nil {
		t.Fatalf("Dial raw: %v", err)
	}
	defer raw.Close()
	if _, err := raw.Write([]byte{0x02, 0x00}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if _, err := c.Call(testContext(t), protocol.GetHWInfo()); err != nil {
		t.Fatalf("Call after malformed packet: %v", err)
	}
}

func TestServeWithoutHandlerDropsRequests(t *testing.T) {
	srv := startServer(t)
	c := dial(t, srv)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if _, err := c.Call(ctx, protocol.GetHWInfo()); err == nil {
		t.Fatal("Call succeeded, want timeout without a handler")
	}
}

func TestServeHandlesDatasetsInOrder(t *testing.T) {
	var (
		mu      sync.Mutex
		headers []uint16
	)
	srv := startServer(t, server.WithHandler(server.HandlerFunc(func(w server.ResponseWriter, r *server.Request) {
		mu.Lock()
		headers = append(headers, r.Message.Header)
		mu.Unlock()
		hwInfoHandler(w, r)
	})))
	c := dial(t, srv)

	if _, err := c.Call(testContext(t), protocol.GetSerialNumber(), protocol.GetHWInfo()); err != nil {
		t.Fatalf("Call: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	want := []uint16{protocol.HeaderLANGetSerialNumber, protocol.HeaderLANGetHWInfo}
	if len(headers) != 2 || headers[0] != want[0] || headers[1] != want[1] {
		t.Fatalf("handled headers = %#04x, want %#04x", headers, want)
	}
}

func TestServeReturnsNilOnCancel(t *testing.T) {
	srv, err := server.New("127.0.0.1:0")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- srv.Serve(ctx) }()
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Serve = %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not return after cancel")
	}
}

func TestServeReturnsErrServerClosedAfterClose(t *testing.T) {
	tests := []struct {
		name string
		opts []server.Option
	}{
		{name: "one socket"},
		{name: "two sockets", opts: []server.Option{server.WithAdditionalAddr("127.0.0.1:0")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, err := server.New("127.0.0.1:0", tt.opts...)
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			done := make(chan error, 1)
			go func() { done <- srv.Serve(context.Background()) }()
			if err := srv.Close(); err != nil {
				t.Fatalf("Close: %v", err)
			}

			select {
			case err := <-done:
				if !errors.Is(err, server.ErrServerClosed) {
					t.Fatalf("Serve = %v, want ErrServerClosed", err)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("Serve did not return after Close")
			}
		})
	}
}

func TestNewInvalidAddr(t *testing.T) {
	if _, err := server.New("not-an-addr"); err == nil {
		t.Fatal("New succeeded, want error")
	}
	if _, err := server.New("127.0.0.1:0", server.WithAdditionalAddr("not-an-addr")); err == nil {
		t.Fatal("New with invalid additional addr succeeded, want error")
	}
}

func TestNewReleasesSocketsOnError(t *testing.T) {
	busy, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("ListenUDP: %v", err)
	}
	defer busy.Close()
	free, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("ListenUDP: %v", err)
	}
	addr := free.LocalAddr().String()
	free.Close()

	if _, err := server.New(addr, server.WithAdditionalAddr(busy.LocalAddr().String())); err == nil {
		t.Fatal("New succeeded on a busy additional addr, want error")
	}
	srv, err := server.New(addr)
	if err != nil {
		t.Fatalf("New after failed New: %v (socket on %s not released)", err, addr)
	}
	srv.Close()
}

func TestAdditionalAddrSharesSessions(t *testing.T) {
	srv := startServer(t,
		server.WithAdditionalAddr("127.0.0.1:0"),
		server.WithHandler(server.HandlerFunc(func(w server.ResponseWriter, r *server.Request) {
			hwInfoHandler(w, r)
			if r.Message.Header == protocol.HeaderLANGetSerialNumber {
				w.Broadcast(protocol.BroadcastFlagSystemState, protocol.EncodeSystemState(protocol.SystemState{}))
			}
		})),
	)
	addrs := srv.Addrs()
	if len(addrs) != 2 || addrs[0].String() != srv.Addr().String() {
		t.Fatalf("Addrs = %v, want Addr and one additional address", addrs)
	}
	primary := dial(t, srv)
	alternate, err := client.Dial(addrs[1].String())
	if err != nil {
		t.Fatalf("Dial alternate: %v", err)
	}
	t.Cleanup(func() { alternate.Close() })

	// The client socket is connected, so replies only arrive from the socket it dialed.
	subscribe(t, alternate, protocol.BroadcastFlagSystemState)
	if _, err := alternate.Call(testContext(t), protocol.GetHWInfo()); err != nil {
		t.Fatalf("Call on alternate socket: %v", err)
	}

	if _, err := primary.Call(testContext(t), protocol.GetSerialNumber()); err != nil {
		t.Fatalf("Call on primary socket: %v", err)
	}
	expectPacket(t, alternate, protocol.HeaderLANSystemStateDataChanged)

	if n := srv.Broadcast(protocol.BroadcastFlagSystemState, protocol.EncodeSystemState(protocol.SystemState{})); n != 1 {
		t.Fatalf("Broadcast sent to %d clients, want 1", n)
	}
	expectPacket(t, alternate, protocol.HeaderLANSystemStateDataChanged)

	local := map[netip.AddrPort]string{}
	for _, c := range srv.Clients() {
		local[c.Addr] = c.LocalAddr.String()
	}
	for c, want := range map[*client.Client]string{primary: addrs[0].String(), alternate: addrs[1].String()} {
		addr := c.LocalAddr().(*net.UDPAddr).AddrPort()
		if got := local[addr]; got != want {
			t.Fatalf("LocalAddr of %v = %q, want %q", addr, got, want)
		}
	}
}
