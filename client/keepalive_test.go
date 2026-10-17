//go:build !integration

package client_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/trains-io/z21.go/protocol"
	"github.com/trains-io/z21.go/server"
	"github.com/trains-io/z21.go/server/servertest"
)

func TestKeepAliveKeepsSession(t *testing.T) {
	const (
		sessionTimeout = 200 * time.Millisecond
		interval       = 50 * time.Millisecond
		flags          = protocol.BroadcastFlagSystemState
	)

	keepAlives := make(chan struct{}, 100)
	mux := stationMux()
	counting := server.HandlerFunc(func(w server.ResponseWriter, r *server.Request) {
		if r.Message.Header == protocol.HeaderLANGetSerialNumber {
			keepAlives <- struct{}{}
		}
		mux.ServeZ21(w, r)
	})
	srv := servertest.NewServer(counting, server.WithSessionTimeout(sessionTimeout))
	t.Cleanup(srv.Close)
	c := dialStation(t, srv)

	if err := c.Send(t.Context(), protocol.SetBroadcastFlags(flags)); err != nil {
		t.Fatalf("SetBroadcastFlags: %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- c.KeepAlive(ctx, interval) }()

	// Ten keep-alives span more than twice the session timeout.
	timeout := time.After(5 * time.Second)
	for range 10 {
		select {
		case <-keepAlives:
		case <-timeout:
			t.Fatal("keep-alives did not arrive")
		}
	}

	clients := srv.Server.Clients()
	if len(clients) != 1 || clients[0].BroadcastFlags != flags {
		t.Fatalf("Clients() = %+v, want one client with broadcast flags %#x (session kept)", clients, flags)
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("KeepAlive after cancel = %v, want nil", err)
	}
}

func TestKeepAliveSendsPrependedMessages(t *testing.T) {
	const flags = protocol.BroadcastFlagXpressNet | protocol.BroadcastFlagSystemState

	keepAlives := make(chan struct{}, 100)
	mux := stationMux()
	counting := server.HandlerFunc(func(w server.ResponseWriter, r *server.Request) {
		mux.ServeZ21(w, r)
		if r.Message.Header == protocol.HeaderLANGetSerialNumber {
			keepAlives <- struct{}{}
		}
	})
	srv := servertest.NewServer(counting)
	t.Cleanup(srv.Close)
	c := dialStation(t, srv)
	c.SetPrependMessages(protocol.SetBroadcastFlags(flags))

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- c.KeepAlive(ctx, 10*time.Millisecond) }()

	select {
	case <-keepAlives:
	case <-time.After(5 * time.Second):
		t.Fatal("keep-alive did not arrive")
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("KeepAlive after cancel = %v, want nil", err)
	}

	clients := srv.Server.Clients()
	if len(clients) != 1 || clients[0].BroadcastFlags != flags {
		t.Fatalf("Clients() = %+v, want one client with broadcast flags %#x (set by the keep-alive)", clients, flags)
	}
}

func TestKeepAliveClosedClient(t *testing.T) {
	c := dialStation(t, startStation(t, nil))
	c.Close()

	err := c.KeepAlive(t.Context(), time.Millisecond)
	if !errors.Is(err, net.ErrClosed) {
		t.Fatalf("KeepAlive on closed client = %v, want net.ErrClosed", err)
	}
}

func TestKeepAliveInvalidInterval(t *testing.T) {
	c := dialStation(t, startStation(t, nil))

	for _, interval := range []time.Duration{0, -time.Second} {
		if err := c.KeepAlive(t.Context(), interval); err == nil {
			t.Errorf("KeepAlive(%v) = nil, want error", interval)
		}
	}
}
