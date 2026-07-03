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
		"LAN_GET_HWINFO",
		"00000000",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("trace missing %q:\n%s", want, out)
		}
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
