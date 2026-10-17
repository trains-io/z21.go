package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/trains-io/z21.go/protocol"
)

// DefaultKeepAliveInterval is half the minute after which a Z21 forgets a
// silent client (spec §1.1).
const DefaultKeepAliveInterval = 30 * time.Second

// KeepAlive sends LAN_GET_SERIAL_NUMBER every interval until ctx is done, so
// that a client that only listens for broadcasts keeps its session and with it
// its broadcast flags. It does not wait for the replies: Call skips them as
// unsolicited datasets, ReadPacket returns them like any other.
//
// The datasets set with SetPrependMessages go in front, as with Send and Call.
// Prepending LAN_SET_BROADCASTFLAGS thus restores the subscriptions when the
// Z21 has forgotten the client anyway, for example after a restart.
//
// KeepAlive blocks; run it in its own goroutine. Unlike the other methods it
// may run concurrently with them, as long as the setters (SetObserver,
// SetPrependMessages, …) are not called meanwhile. The observer then sees
// packets from several goroutines and must be safe for concurrent use.
//
// KeepAlive returns nil when ctx is done and an error matching net.ErrClosed
// once the client is closed. Other write errors, for example while the Z21 is
// unreachable, are logged and the next keep-alive is sent as scheduled.
func (c *Client) KeepAlive(ctx context.Context, interval time.Duration) error {
	if c == nil || c.conn == nil {
		return errors.New("z21: not connected")
	}
	if interval <= 0 {
		return fmt.Errorf("z21: keep-alive interval %v is not positive", interval)
	}
	payload, err := c.marshalOutbound([]protocol.Message{protocol.GetSerialNumber()})
	if err != nil {
		return fmt.Errorf("z21: keep-alive: %w", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}

		if _, err := c.conn.Write(payload); err != nil {
			if errors.Is(err, net.ErrClosed) {
				return fmt.Errorf("z21: keep-alive: %w", err)
			}
			if c.log != nil {
				c.log.Warn("keep-alive not sent", "addr", c.Addr(), "err", err)
			}
			continue
		}
		c.observePacket(ctx, DirectionTX, payload, nil)
	}
}
