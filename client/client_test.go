package client_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/trains-io/z21.go/client"
	"github.com/trains-io/z21.go/protocol"
)

// minimalZ21Server responds to LAN_GET_HWINFO with a fixed payload.
func TestCallAgainstLocalStub(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("ListenUDP: %v", err)
	}
	defer conn.Close()

	reply, err := protocol.Message{
		Header: protocol.HeaderLANGetHWInfo,
		Data:   []byte{0x01, 0x00, 0x00, 0x00, 0x43, 0x01, 0x00, 0x00},
	}.Marshal()
	if err != nil {
		t.Fatalf("Marshal reply: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 1472)
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		msgs, err := protocol.ParseAll(buf[:n])
		if err != nil || len(msgs) != 1 || msgs[0].Header != protocol.HeaderLANGetHWInfo {
			return
		}
		_, _ = conn.WriteToUDP(reply, addr)
	}()

	c, err := client.Dial(conn.LocalAddr().String())
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer c.Close()

	msgs, err := c.Call(t.Context(), protocol.GetHWInfo())
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Header != protocol.HeaderLANGetHWInfo {
		t.Fatalf("unexpected reply: %#v", msgs)
	}

	<-done
}

func TestReadPacketContextCancel(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("ListenUDP: %v", err)
	}
	defer conn.Close()

	c, err := client.Dial(conn.LocalAddr().String())
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := c.ReadPacket(ctx)
		done <- err
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Fatalf("ReadPacket() error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("ReadPacket did not return after context cancel")
	}
}

func TestDialDefaultPort(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: client.DefaultPort})
	if err != nil {
		t.Fatalf("ListenUDP: %v", err)
	}
	defer conn.Close()

	c, err := client.Dial("127.0.0.1")
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer c.Close()

	if got, want := c.Addr(), "127.0.0.1:21105"; got != want {
		t.Fatalf("Addr() = %q, want %q", got, want)
	}
}

func TestDialInvalidAddress(t *testing.T) {
	_, err := client.Dial("")
	if err == nil {
		t.Fatal("expected error for empty address")
	}
}

func TestWithLogger(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("ListenUDP: %v", err)
	}
	defer conn.Close()

	c, err := client.Dial(conn.LocalAddr().String(), client.WithLogger(nil))
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer c.Close()

	if c.LocalAddr() == nil {
		t.Fatal("LocalAddr() = nil")
	}
}
