//go:build !integration

package client_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/trains-io/z21.go/client"
	"github.com/trains-io/z21.go/protocol"
)

func TestSend(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	require.NoError(t, err)
	defer conn.Close()

	received := make(chan []byte, 1)
	go func() {
		buf := make([]byte, 1472)
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, _, err := conn.ReadFromUDP(buf)
		if err == nil {
			received <- append([]byte(nil), buf[:n]...)
		}
		close(received)
	}()

	c, err := client.Dial(conn.LocalAddr().String())
	require.NoError(t, err)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	require.NoError(t, c.Send(ctx, protocol.GetHWInfo()))

	raw := <-received
	require.NotEmpty(t, raw)
	msgs, err := protocol.ParseAll(raw)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.Equal(t, protocol.HeaderLANGetHWInfo, msgs[0].Header)
}

func TestCollect(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	require.NoError(t, err)
	defer conn.Close()

	replyA, err := protocol.GetHWInfo().Marshal()
	require.NoError(t, err)
	replyB, err := protocol.GetBroadcastFlags().Marshal()
	require.NoError(t, err)

	go func() {
		buf := make([]byte, 1472)
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		_, _ = conn.WriteToUDP(replyA, addr)
		time.Sleep(20 * time.Millisecond)
		_, _ = conn.WriteToUDP(replyB, addr)
	}()

	c, err := client.Dial(conn.LocalAddr().String())
	require.NoError(t, err)
	defer c.Close()

	// Trigger server replies.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = c.Send(ctx, protocol.GetHWInfo())

	collectCtx, collectCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer collectCancel()

	msgs, err := c.Collect(collectCtx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(msgs), 2)

	var sawHW, sawFlags bool
	for _, msg := range msgs {
		switch msg.Header {
		case protocol.HeaderLANGetHWInfo:
			sawHW = true
		case protocol.HeaderLANGetBroadcastFlags:
			sawFlags = true
		}
	}
	require.True(t, sawHW)
	require.True(t, sawFlags)
}

func TestRequestCollect(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	require.NoError(t, err)
	defer conn.Close()

	reply, err := protocol.Message{
		Header: protocol.HeaderLANGetBroadcastFlags,
		Data:   []byte{0x01, 0x00, 0x00, 0x00},
	}.Marshal()
	require.NoError(t, err)

	go func() {
		buf := make([]byte, 1472)
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		msgs, err := protocol.ParseAll(buf[:n])
		if err != nil || len(msgs) != 1 || msgs[0].Header != protocol.HeaderLANGetBroadcastFlags {
			return
		}
		_, _ = conn.WriteToUDP(reply, addr)
	}()

	c, err := client.Dial(conn.LocalAddr().String())
	require.NoError(t, err)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	msgs, err := c.RequestCollect(ctx, protocol.GetBroadcastFlags(), 500*time.Millisecond)
	require.NoError(t, err)
	require.NotEmpty(t, msgs)

	flags, err := protocol.BroadcastFlagsFromMessages(msgs)
	require.NoError(t, err)
	require.Equal(t, uint32(1), flags)
}

func TestCollectReturnsOnCancel(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	require.NoError(t, err)
	defer conn.Close()

	c, err := client.Dial(conn.LocalAddr().String())
	require.NoError(t, err)
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan []protocol.Message, 1)
	go func() {
		msgs, _ := c.Collect(ctx)
		done <- msgs
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case msgs := <-done:
		require.Empty(t, msgs)
	case <-time.After(time.Second):
		t.Fatal("Collect did not return after cancel")
	}
}
