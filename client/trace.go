package client

import (
	"context"
	"io"

	"github.com/trains-io/z21.go/internal/wiretrace"
	"github.com/trains-io/z21.go/protocol"
)

// Tracer writes hexdumps of Z21 UDP payloads. It implements Observer.
// Attach with Client.SetTracer or WithObserver(NewTracer(w)).
type Tracer struct {
	Out io.Writer
}

// NewTracer returns a tracer that writes to w.
func NewTracer(w io.Writer) *Tracer {
	if w == nil {
		return nil
	}
	return &Tracer{Out: w}
}

// OnPacket implements Observer.
func (t *Tracer) OnPacket(_ context.Context, dir Direction, remoteAddr string, raw []byte, msgs []protocol.Message) {
	if t == nil || t.Out == nil || len(raw) == 0 {
		return
	}
	direction := ">>"
	if dir == DirectionRX {
		direction = "<<"
	}
	wiretrace.Dump(t.Out, direction+" "+remoteAddr, raw)
}

// SetTracer enables or disables UDP payload tracing. Pass nil to disable.
// Deprecated: use SetObserver or WithObserver instead.
func (c *Client) SetTracer(t *Tracer) {
	if c == nil {
		return
	}
	if t == nil {
		c.observer = nil
		return
	}
	c.observer = t
}

// SetObserver enables or disables wire-level observation. Pass nil to disable.
func (c *Client) SetObserver(o Observer) {
	if c == nil {
		return
	}
	c.observer = o
}
