//go:build !integration

package client_test

import (
	"bytes"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/trains-io/z21.go/client"
	"github.com/trains-io/z21.go/protocol"
)

func TestTracerHexdumpOnCall(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	reply, err := protocol.Message{
		Header: protocol.HeaderLANGetHWInfo,
		Data:   []byte{0x01, 0x00, 0x00, 0x00},
	}.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		buf := make([]byte, 1472)
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		_, _ = conn.WriteToUDP(reply, addr)
		_ = n
	}()

	var trace bytes.Buffer
	c, err := client.Dial(conn.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.SetTracer(client.NewTracer(&trace))

	if _, err := c.Call(t.Context(), protocol.GetHWInfo()); err != nil {
		t.Fatal(err)
	}

	out := trace.String()
	for _, want := range []string{
		">> ",
		"<< ",
		"LAN_GET_HWINFO (8 bytes)",
		"00000000",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("trace missing %q:\n%s", want, out)
		}
	}
}

func TestTracerCombinedPayload(t *testing.T) {
	payload, err := protocol.MarshalAll(
		protocol.SetBroadcastFlags(0x101),
		protocol.GetXStatus(),
	)
	if err != nil {
		t.Fatal(err)
	}

	var trace bytes.Buffer
	tr := client.NewTracer(&trace)
	tr.OnPacket(t.Context(), client.DirectionTX, "127.0.0.1:21107", payload, nil)

	out := trace.String()
	for _, want := range []string{
		">> 127.0.0.1:21107 (15 bytes, 2 datasets)",
		"[1/2] LAN_SET_BROADCASTFLAGS (8 bytes)",
		"[2/2] LAN_X_GET_STATUS (7 bytes)",
		"00000000",
		"00000008",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("trace missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "LAN_SET_BROADCASTFLAGS data=") {
		t.Fatalf("unexpected legacy combined label format:\n%s", out)
	}
}

func TestTracerNamesLANXGetVersion(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	reply, err := protocol.Message{
		Header: protocol.HeaderLANX,
		Data:   []byte{0x63, 0x21, 0x30, 0x12, 0x60},
	}.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		buf := make([]byte, 1472)
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		_, _ = conn.WriteToUDP(reply, addr)
		_ = n
	}()

	var trace bytes.Buffer
	c, err := client.Dial(conn.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.SetTracer(client.NewTracer(&trace))

	if _, err := c.Call(t.Context(), protocol.GetXVersion()); err != nil {
		t.Fatal(err)
	}

	out := trace.String()
	if strings.Count(out, "LAN_X_GET_VERSION") < 2 {
		t.Fatalf("expected request and reply named LAN_X_GET_VERSION:\n%s", out)
	}
	if strings.Contains(out, "LAN_X data=") || strings.Contains(out, "\n   LAN_X\n") {
		t.Fatalf("unexpected generic LAN_X label:\n%s", out)
	}
}
