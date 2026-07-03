package client

import (
	"context"

	"github.com/trains-io/z21.go/protocol"
)

// Direction is UDP flow relative to this client.
type Direction int

const (
	DirectionTX Direction = iota
	DirectionRX
)

// String returns a short label for trace output.
func (d Direction) String() string {
	switch d {
	case DirectionTX:
		return "tx"
	case DirectionRX:
		return "rx"
	default:
		return "unknown"
	}
}

// Observer receives wire-level Z21 UDP events.
//
// Implementations must be fast and non-blocking, or offload work themselves;
// the client calls observers synchronously on the hot path.
type Observer interface {
	OnPacket(ctx context.Context, dir Direction, remote string, raw []byte, msgs []protocol.Message)
}

// NopObserver discards all events. It is the default when no observer is set.
type NopObserver struct{}

func (NopObserver) OnPacket(context.Context, Direction, string, []byte, []protocol.Message) {}

// MultiObserver fans out events to multiple observers.
type MultiObserver struct {
	Observers []Observer
}

func (m MultiObserver) OnPacket(ctx context.Context, dir Direction, remote string, raw []byte, msgs []protocol.Message) {
	for _, o := range m.Observers {
		if o != nil {
			o.OnPacket(ctx, dir, remote, raw, msgs)
		}
	}
}
