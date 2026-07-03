package client

import (
	"context"
	"fmt"
	"io"
	"strings"

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
	t.log(direction, remoteAddr, raw, msgs)
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

func (t *Tracer) log(direction, remoteAddr string, data []byte, msgs []protocol.Message) {
	fmt.Fprintf(t.Out, "%s %s (%d bytes)\n", direction, remoteAddr, len(data))
	writeHexdump(t.Out, data)
	t.writeHeaders(data, msgs)
}

func (t *Tracer) writeHeaders(data []byte, msgs []protocol.Message) {
	if len(msgs) == 0 {
		parsed, err := protocol.ParseAll(data)
		if err != nil || len(parsed) == 0 {
			return
		}
		msgs = parsed
	}
	var parts []string
	for _, msg := range msgs {
		name := protocol.MessageName(msg)
		if len(msg.Data) > 0 {
			parts = append(parts, fmt.Sprintf("%s data=%d", name, len(msg.Data)))
		} else {
			parts = append(parts, name)
		}
	}
	fmt.Fprintf(t.Out, "   %s\n", strings.Join(parts, ", "))
}

func writeHexdump(w io.Writer, data []byte) {
	const width = 16
	for offset := 0; offset < len(data); offset += width {
		end := offset + width
		if end > len(data) {
			end = len(data)
		}
		chunk := data[offset:end]

		fmt.Fprintf(w, "%08x  ", offset)
		for i := 0; i < width; i++ {
			if i < len(chunk) {
				fmt.Fprintf(w, "%02x ", chunk[i])
			} else {
				fmt.Fprint(w, "   ")
			}
			if i == 7 {
				fmt.Fprint(w, " ")
			}
		}
		fmt.Fprint(w, " |")
		for _, b := range chunk {
			if b >= 32 && b < 127 {
				fmt.Fprintf(w, "%c", b)
			} else {
				fmt.Fprint(w, ".")
			}
		}
		fmt.Fprintln(w, "|")
	}
}
