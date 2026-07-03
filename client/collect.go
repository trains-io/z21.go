package client

import (
	"context"
	"errors"
	"time"

	"github.com/trains-io/z21.go/protocol"
)

// Collect reads UDP datagrams until ctx is cancelled or times out.
func (c *Client) Collect(ctx context.Context) ([]protocol.Message, error) {
	var all []protocol.Message
	for {
		msgs, err := c.ReadPacket(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return all, nil
			}
			return all, err
		}
		all = append(all, msgs...)
	}
}

// RequestCollect sends req then collects replies until collectWindow elapses.
func (c *Client) RequestCollect(ctx context.Context, req protocol.Message, collectWindow time.Duration) ([]protocol.Message, error) {
	if err := c.Send(ctx, req); err != nil {
		return nil, err
	}
	if collectWindow <= 0 {
		collectWindow = 2 * time.Second
	}
	collectCtx, cancel := context.WithTimeout(ctx, collectWindow)
	defer cancel()
	return c.Collect(collectCtx)
}
