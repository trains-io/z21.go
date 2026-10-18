//go:build !integration

package client_test

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/trains-io/z21.go/client"
	"github.com/trains-io/z21.go/protocol"
	"github.com/trains-io/z21.go/server"
)

func TestCall(t *testing.T) {
	c := dialStation(t, startStation(t, nil))

	msgs, err := c.Call(t.Context(), protocol.GetHWInfo())
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	info, err := protocol.HWInfoFromMessages(msgs)
	if err != nil {
		t.Fatalf("HWInfoFromMessages() error = %v", err)
	}
	if info != testHWInfo {
		t.Fatalf("HWInfo = %+v, want %+v", info, testHWInfo)
	}
}

func TestCallLocoCommandWithoutReply(t *testing.T) {
	c := dialStation(t, startStation(t, nil), client.WithCallTimeout(time.Second))

	msgs, err := c.Call(t.Context(), protocol.SetLocoDrive(3, protocol.LocoSpeedSteps128, true, 40))
	if err != nil || msgs != nil {
		t.Fatalf("Call(SetLocoDrive) = %v, %v; want no replies and no error without waiting", msgs, err)
	}
}

func TestCallCombinedDeviceInfo(t *testing.T) {
	c := dialStation(t, startStation(t, nil))

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
	if serial, ok := protocol.SerialFromMessages(msgs); !ok || serial != protocol.FormatSerialNumber(testSerial) {
		t.Fatalf("SerialFromMessages() = %q, ok=%v", serial, ok)
	}
	if info, err := protocol.HWInfoFromMessages(msgs); err != nil || info != testHWInfo {
		t.Fatalf("HWInfoFromMessages() = %+v, err=%v", info, err)
	}
	if v, err := protocol.XVersionFromMessages(msgs); err != nil || v != testXVersion {
		t.Fatalf("XVersionFromMessages() = %+v, err=%v", v, err)
	}
	if fw, err := protocol.XFirmwareFromMessages(msgs); err != nil || fw != testXFirmware {
		t.Fatalf("XFirmwareFromMessages() = %+v, err=%v", fw, err)
	}
	if code, err := protocol.CodeFromMessages(msgs); err != nil || code != protocol.CodeNoLock {
		t.Fatalf("CodeFromMessages() = %#x, err=%v", code, err)
	}
}

// replyStub serves the first request it receives on a raw socket with
// replies, one datagram each, for server behaviour the server package cannot
// produce.
func replyStub(t *testing.T, wantDatasets int, replies []protocol.Message) string {
	t.Helper()
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("ListenUDP: %v", err)
	}
	done := make(chan struct{})
	t.Cleanup(func() {
		conn.Close()
		<-done
	})

	go func() {
		defer close(done)
		buf := make([]byte, 1472)
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		msgs, err := protocol.ParseAll(buf[:n])
		if err != nil || len(msgs) != wantDatasets {
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
	return conn.LocalAddr().String()
}

var deviceInfoReplies = []protocol.Message{
	protocol.EncodeSerialNumber(testSerial),
	protocol.EncodeHWInfo(testHWInfo),
	protocol.EncodeXVersion(testXVersion),
	protocol.EncodeXFirmware(testXFirmware),
	protocol.EncodeCode(protocol.CodeNoLock),
}

var deviceInfoRequests = []protocol.Message{
	protocol.GetSerialNumber(),
	protocol.GetHWInfo(),
	protocol.GetXVersion(),
	protocol.GetXFirmware(),
	protocol.GetCode(),
}

func TestCallCombinedDeviceInfoSeparateReplies(t *testing.T) {
	c, err := client.Dial(replyStub(t, 5, deviceInfoReplies))
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer c.Close()

	msgs, err := c.Call(t.Context(), deviceInfoRequests...)
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
}

func TestCallCombinedDeviceInfoWithUnsolicitedBroadcast(t *testing.T) {
	broadcast := protocol.Message{Header: protocol.HeaderLANSystemStateDataChanged, Data: make([]byte, 16)}
	replies := append([]protocol.Message{deviceInfoReplies[0], broadcast}, deviceInfoReplies[1:]...)
	c, err := client.Dial(replyStub(t, 5, replies))
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer c.Close()

	msgs, err := c.Call(t.Context(), deviceInfoRequests...)
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if !protocol.AllRequestsAnswered(deviceInfoRequests, msgs) {
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
}

func TestCallWithPrependBroadcastFlags(t *testing.T) {
	var (
		mu    sync.Mutex
		flags []uint32
	)
	mux := stationMux()
	srv := startStation(t, server.HandlerFunc(func(w server.ResponseWriter, r *server.Request) {
		mu.Lock()
		flags = append(flags, r.Client.BroadcastFlags)
		mu.Unlock()
		mux.ServeZ21(w, r)
	}))
	c := dialStation(t, srv)
	c.SetPrependMessages(protocol.SetBroadcastFlags(0x101))

	msgs, err := c.Call(t.Context(), protocol.GetHWInfo())
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Header != protocol.HeaderLANGetHWInfo {
		t.Fatalf("unexpected reply: %#v", msgs)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(flags) != 1 || flags[0] != 0x101 {
		t.Fatalf("broadcast flags seen by handler = %#x, want [0x101] (prepended in the same packet)", flags)
	}
}

func TestCallWithPrependGetXStatus(t *testing.T) {
	c := dialStation(t, startStation(t, nil))
	c.SetPrependMessages(protocol.SetBroadcastFlags(0x101))

	msgs, err := c.Call(t.Context(), protocol.GetXStatus())
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if status, err := protocol.XStatusFromMessages(msgs); err != nil || status != testXStatus {
		t.Fatalf("XStatusFromMessages() = %+v, err=%v, msgs=%#v", status, err, msgs)
	}
}

func TestReadPacketContextCancel(t *testing.T) {
	c := dialStation(t, startStation(t, nil))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	time.AfterFunc(50*time.Millisecond, cancel)

	start := time.Now()
	_, err := c.ReadPacket(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ReadPacket() error = %v, want context.Canceled", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("ReadPacket returned %v after cancel, want promptly", elapsed)
	}
}

// silentPeer returns the address of a UDP socket that never replies.
func silentPeer(t *testing.T) string {
	t.Helper()
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("ListenUDP: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn.LocalAddr().String()
}

func dialAddr(t *testing.T, addr string, opts ...client.Option) *client.Client {
	t.Helper()
	c, err := client.Dial(addr, opts...)
	if err != nil {
		t.Fatalf("Dial(%q): %v", addr, err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

func TestCallWithoutDeadlineUsesCallTimeout(t *testing.T) {
	tests := []struct {
		name string
		reqs []protocol.Message
	}{
		{name: "single request", reqs: []protocol.Message{protocol.GetHWInfo()}},
		{name: "several requests", reqs: deviceInfoRequests},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := dialAddr(t, silentPeer(t), client.WithCallTimeout(200*time.Millisecond))

			start := time.Now()
			_, err := c.Call(context.Background(), tt.reqs...)
			if !errors.Is(err, client.ErrIncomplete) || !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("Call() error = %v, want ErrIncomplete and context.DeadlineExceeded", err)
			}
			if elapsed := time.Since(start); elapsed > 2*time.Second {
				t.Fatalf("Call returned after %v, want about 200ms", elapsed)
			}
		})
	}
}

func TestCallContextDeadline(t *testing.T) {
	c := dialAddr(t, silentPeer(t))

	ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
	defer cancel()
	msgs, err := c.Call(ctx, protocol.GetHWInfo())
	if !errors.Is(err, client.ErrIncomplete) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Call() error = %v, want ErrIncomplete and context.DeadlineExceeded", err)
	}
	if len(msgs) != 0 {
		t.Fatalf("Call() replies = %#v, want none", msgs)
	}
}

func TestCallContextCancel(t *testing.T) {
	c := dialAddr(t, silentPeer(t))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	time.AfterFunc(50*time.Millisecond, cancel)

	start := time.Now()
	_, err := c.Call(ctx, protocol.GetHWInfo())
	if !errors.Is(err, client.ErrIncomplete) || !errors.Is(err, context.Canceled) {
		t.Fatalf("Call() error = %v, want ErrIncomplete and context.Canceled", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("Call returned %v after cancel, want promptly", elapsed)
	}
}

func TestCallSkipsBroadcastBeforeReply(t *testing.T) {
	broadcast := protocol.Message{Header: protocol.HeaderLANSystemStateDataChanged, Data: make([]byte, 16)}
	c := dialAddr(t, replyStub(t, 1, []protocol.Message{broadcast, protocol.EncodeHWInfo(testHWInfo)}))

	msgs, err := c.Call(t.Context(), protocol.GetHWInfo())
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	info, err := protocol.HWInfoFromMessages(msgs)
	if err != nil {
		t.Fatalf("HWInfoFromMessages(%#v) error = %v", msgs, err)
	}
	if info != testHWInfo {
		t.Fatalf("HWInfo = %+v, want %+v", info, testHWInfo)
	}
}

func TestCallPartialReplies(t *testing.T) {
	c := dialAddr(t, replyStub(t, 5, deviceInfoReplies[:2]))

	ctx, cancel := context.WithTimeout(t.Context(), 300*time.Millisecond)
	defer cancel()
	msgs, err := c.Call(ctx, deviceInfoRequests...)
	if !errors.Is(err, client.ErrIncomplete) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Call() error = %v, want ErrIncomplete and context.DeadlineExceeded", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("Call() replies = %d, want the 2 received", len(msgs))
	}
}

func TestDialDefaultPort(t *testing.T) {
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
	c := dialStation(t, startStation(t, nil), client.WithLogger(nil))

	if c.LocalAddr() == nil {
		t.Fatal("LocalAddr() = nil")
	}
}
