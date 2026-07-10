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

func TestCallCombinedDeviceInfo(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("ListenUDP: %v", err)
	}
	defer conn.Close()

	replies := []protocol.Message{
		{
			Header: protocol.HeaderLANGetSerialNumber,
			Data:   []byte{0x0E, 0xB0, 0x04, 0x00},
		},
		{
			Header: protocol.HeaderLANGetHWInfo,
			Data:   []byte{0x01, 0x02, 0x00, 0x00, 0x42, 0x01, 0x00, 0x00},
		},
		{
			Header: protocol.HeaderLANX,
			Data:   []byte{0x63, 0x21, 0x30, 0x12, 0x60},
		},
		{
			Header: protocol.HeaderLANX,
			Data:   []byte{0xF3, 0x0A, 0x01, 0x42, 0xB8},
		},
		{
			Header: protocol.HeaderLANGetCode,
			Data:   []byte{0x00},
		},
	}
	reply, err := protocol.MarshalAll(replies...)
	if err != nil {
		t.Fatalf("MarshalAll reply: %v", err)
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
		if err != nil || len(msgs) != 5 {
			return
		}
		want := []uint16{
			protocol.HeaderLANGetSerialNumber,
			protocol.HeaderLANGetHWInfo,
			protocol.HeaderLANX,
			protocol.HeaderLANX,
			protocol.HeaderLANGetCode,
		}
		for i, header := range want {
			if msgs[i].Header != header {
				return
			}
		}
		_, _ = conn.WriteToUDP(reply, addr)
	}()

	c, err := client.Dial(conn.LocalAddr().String())
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer c.Close()

	msgs, err := c.Call(t.Context(),
		protocol.GetSerialNumber(),
		protocol.GetHWInfo(),
		protocol.GetXVersion(),
		protocol.GetXFirmware(),
		protocol.GetCode(),
	)
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if len(msgs) != 5 {
		t.Fatalf("reply len = %d, want 5", len(msgs))
	}
	if serial, ok := protocol.SerialFromMessages(msgs); !ok || serial == "" {
		t.Fatalf("SerialFromMessages() = %q, ok=%v", serial, ok)
	}
	if _, err := protocol.HWInfoFromMessages(msgs); err != nil {
		t.Fatalf("HWInfoFromMessages() error = %v", err)
	}
	if _, err := protocol.XVersionFromMessages(msgs); err != nil {
		t.Fatalf("XVersionFromMessages() error = %v", err)
	}
	if _, err := protocol.XFirmwareFromMessages(msgs); err != nil {
		t.Fatalf("XFirmwareFromMessages() error = %v", err)
	}
	if code, err := protocol.CodeFromMessages(msgs); err != nil || code != protocol.CodeNoLock {
		t.Fatalf("CodeFromMessages() = %#x, err=%v", code, err)
	}

	<-done
}

func TestCallCombinedDeviceInfoSeparateReplies(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("ListenUDP: %v", err)
	}
	defer conn.Close()

	replies := []protocol.Message{
		{Header: protocol.HeaderLANGetSerialNumber, Data: []byte{0x0E, 0xB0, 0x04, 0x00}},
		{Header: protocol.HeaderLANGetHWInfo, Data: []byte{0x01, 0x02, 0x00, 0x00, 0x42, 0x01, 0x00, 0x00}},
		{Header: protocol.HeaderLANX, Data: []byte{0x63, 0x21, 0x30, 0x12, 0x60}},
		{Header: protocol.HeaderLANX, Data: []byte{0xF3, 0x0A, 0x01, 0x42, 0xB8}},
		{Header: protocol.HeaderLANGetCode, Data: []byte{0x00}},
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
		if err != nil || len(msgs) != 5 {
			return
		}
		for _, reply := range replies {
			payload, err := reply.Marshal()
			if err != nil {
				return
			}
			if _, err := conn.WriteToUDP(payload, addr); err != nil {
				return
			}
		}
	}()

	c, err := client.Dial(conn.LocalAddr().String())
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer c.Close()

	msgs, err := c.Call(t.Context(),
		protocol.GetSerialNumber(),
		protocol.GetHWInfo(),
		protocol.GetXVersion(),
		protocol.GetXFirmware(),
		protocol.GetCode(),
	)
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if len(msgs) != 5 {
		t.Fatalf("reply len = %d, want 5", len(msgs))
	}
	if _, err := protocol.HWInfoFromMessages(msgs); err != nil {
		t.Fatalf("HWInfoFromMessages() error = %v", err)
	}
	if _, err := protocol.CodeFromMessages(msgs); err != nil {
		t.Fatalf("CodeFromMessages() error = %v", err)
	}

	<-done
}

func TestCallCombinedDeviceInfoWithUnsolicitedBroadcast(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("ListenUDP: %v", err)
	}
	defer conn.Close()

	replies := []protocol.Message{
		{Header: protocol.HeaderLANGetSerialNumber, Data: []byte{0x0E, 0xB0, 0x04, 0x00}},
		{Header: protocol.HeaderLANSystemStateDataChanged, Data: make([]byte, 16)},
		{Header: protocol.HeaderLANGetHWInfo, Data: []byte{0x01, 0x02, 0x00, 0x00, 0x42, 0x01, 0x00, 0x00}},
		{Header: protocol.HeaderLANX, Data: []byte{0x63, 0x21, 0x30, 0x12, 0x60}},
		{Header: protocol.HeaderLANX, Data: []byte{0xF3, 0x0A, 0x01, 0x42, 0xB8}},
		{Header: protocol.HeaderLANGetCode, Data: []byte{0x00}},
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
		if err != nil || len(msgs) != 5 {
			return
		}
		for _, reply := range replies {
			payload, err := reply.Marshal()
			if err != nil {
				return
			}
			if _, err := conn.WriteToUDP(payload, addr); err != nil {
				return
			}
		}
	}()

	c, err := client.Dial(conn.LocalAddr().String())
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer c.Close()

	msgs, err := c.Call(t.Context(),
		protocol.GetSerialNumber(),
		protocol.GetHWInfo(),
		protocol.GetXVersion(),
		protocol.GetXFirmware(),
		protocol.GetCode(),
	)
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if !protocol.AllRequestsAnswered([]protocol.Message{
		protocol.GetSerialNumber(),
		protocol.GetHWInfo(),
		protocol.GetXVersion(),
		protocol.GetXFirmware(),
		protocol.GetCode(),
	}, msgs) {
		t.Fatalf("expected all requests answered, got %#v", msgs)
	}
	for _, msg := range msgs {
		if msg.Header == protocol.HeaderLANSystemStateDataChanged {
			t.Fatalf("unexpected broadcast in filtered replies: %#v", msgs)
		}
	}
	if _, err := protocol.HWInfoFromMessages(msgs); err != nil {
		t.Fatalf("HWInfoFromMessages() error = %v", err)
	}
	if _, err := protocol.CodeFromMessages(msgs); err != nil {
		t.Fatalf("CodeFromMessages() error = %v", err)
	}

	<-done
}

func TestCallWithPrependBroadcastFlags(t *testing.T) {
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
		if err != nil || len(msgs) != 2 {
			return
		}
		if msgs[0].Header != protocol.HeaderLANSetBroadcastFlags {
			return
		}
		if msgs[1].Header != protocol.HeaderLANGetHWInfo {
			return
		}
		_, _ = conn.WriteToUDP(reply, addr)
	}()

	c, err := client.Dial(conn.LocalAddr().String())
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer c.Close()
	c.SetPrependMessages(protocol.SetBroadcastFlags(0x101))

	msgs, err := c.Call(t.Context(), protocol.GetHWInfo())
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Header != protocol.HeaderLANGetHWInfo {
		t.Fatalf("unexpected reply: %#v", msgs)
	}

	<-done
}

func TestCallWithPrependGetXStatus(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("ListenUDP: %v", err)
	}
	defer conn.Close()

	statusReply, err := protocol.Message{
		Header: protocol.HeaderLANX,
		Data:   []byte{0x62, 0x22, 0x00, 0x40},
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
		if err != nil || len(msgs) != 2 {
			return
		}
		if msgs[0].Header != protocol.HeaderLANSetBroadcastFlags {
			return
		}
		if msgs[1].Header != protocol.HeaderLANX || msgs[1].Data[1] != 0x24 {
			return
		}
		_, _ = conn.WriteToUDP(statusReply, addr)
	}()

	c, err := client.Dial(conn.LocalAddr().String())
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer c.Close()
	c.SetPrependMessages(protocol.SetBroadcastFlags(0x101))

	msgs, err := c.Call(t.Context(), protocol.GetXStatus())
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if _, err := protocol.XStatusFromMessages(msgs); err != nil {
		t.Fatalf("XStatusFromMessages() error = %v, msgs=%#v", err, msgs)
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
