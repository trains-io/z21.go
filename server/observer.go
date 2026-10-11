package server

import (
	"fmt"
	"io"
	"net/netip"
	"strings"
	"sync"

	"github.com/trains-io/z21.go/internal/wiretrace"
	"github.com/trains-io/z21.go/protocol"
)

// Direction is the UDP flow relative to the server.
type Direction int

const (
	DirectionRX Direction = iota // received from a client
	DirectionTX                  // sent to one or more clients
)

// String returns a short label for trace output.
func (d Direction) String() string {
	switch d {
	case DirectionRX:
		return "rx"
	case DirectionTX:
		return "tx"
	default:
		return "unknown"
	}
}

// Packet is one UDP packet received or sent by the server.
type Packet struct {
	Direction Direction
	Remote    netip.AddrPort // sender (RX) or recipient of a reply (TX); zero for broadcasts

	Broadcast  bool
	Flag       uint32           // broadcast flag(s) selecting the recipients
	Recipients []netip.AddrPort // clients the broadcast was sent to; may be empty

	Raw      []byte             // UDP payload; only valid during OnPacket
	Messages []protocol.Message // parsed datasets; nil if Raw is malformed
}

// Observer receives every packet the server receives and sends.
//
// OnPacket is called synchronously, from the Serve goroutine and from any
// goroutine calling Broadcast, so implementations must be fast and safe for
// concurrent use.
type Observer interface {
	OnPacket(p Packet)
}

// Tracer writes hexdumps of the server's UDP traffic. It implements Observer.
type Tracer struct {
	mu  sync.Mutex
	out io.Writer
}

// NewTracer returns a tracer writing to w, or nil when w is nil.
func NewTracer(w io.Writer) *Tracer {
	if w == nil {
		return nil
	}
	return &Tracer{out: w}
}

// OnPacket implements Observer.
func (t *Tracer) OnPacket(p Packet) {
	if t == nil || len(p.Raw) == 0 {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	wiretrace.Dump(t.out, packetLabel(p), p.Raw)
}

func packetLabel(p Packet) string {
	if !p.Broadcast {
		if p.Direction == DirectionRX {
			return "<< " + p.Remote.String()
		}
		return ">> " + p.Remote.String()
	}
	label := ">> broadcast " + broadcastFlagName(p.Flag)
	if len(p.Recipients) == 0 {
		return label + " (0 recipients)"
	}
	names := make([]string, len(p.Recipients))
	for i, addr := range p.Recipients {
		names[i] = addr.String()
	}
	return label + " -> " + strings.Join(names, ", ")
}

func broadcastFlagName(flag uint32) string {
	var names []string
	for _, f := range protocol.KnownBroadcastFlags {
		if flag&f.Value != 0 {
			names = append(names, f.Name)
		}
	}
	if len(names) == 0 {
		return fmt.Sprintf("%#08x", flag)
	}
	return strings.Join(names, " | ")
}
