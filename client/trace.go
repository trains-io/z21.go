package client

import (
	"context"
	"fmt"
	"io"

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

type wireDataset struct {
	msg    protocol.Message
	offset int
	wire   []byte
}

func (t *Tracer) log(direction, remoteAddr string, data []byte, _ []protocol.Message) {
	datasets, err := wireDatasets(data)
	if err != nil || len(datasets) == 0 {
		fmt.Fprintf(t.Out, "%s %s (%d bytes)\n", direction, remoteAddr, len(data))
		writeHexdump(t.Out, data)
		return
	}

	switch len(datasets) {
	case 1:
		ds := datasets[0]
		fmt.Fprintf(t.Out, "%s %s (%d bytes)\n", direction, remoteAddr, len(data))
		fmt.Fprintf(t.Out, "   %s\n", datasetSummary(ds.msg, len(ds.wire)))
		writeHexdump(t.Out, ds.wire)
	default:
		fmt.Fprintf(t.Out, "%s %s (%d bytes, %d datasets)\n", direction, remoteAddr, len(data), len(datasets))
		for i, ds := range datasets {
			fmt.Fprintf(t.Out, "   [%d/%d] %s\n", i+1, len(datasets), datasetSummary(ds.msg, len(ds.wire)))
			writeIndentedHexdump(t.Out, ds.offset, ds.wire)
		}
	}
}

func wireDatasets(raw []byte) ([]wireDataset, error) {
	var out []wireDataset
	rest := raw
	offset := 0
	for len(rest) > 0 {
		msg, tail, err := protocol.Unmarshal(rest)
		if err != nil {
			return nil, err
		}
		consumed := len(rest) - len(tail)
		out = append(out, wireDataset{
			msg:    msg,
			offset: offset,
			wire:   rest[:consumed],
		})
		offset += consumed
		rest = tail
	}
	return out, nil
}

func datasetSummary(msg protocol.Message, wireLen int) string {
	name := protocol.MessageName(msg)
	if wireLen > 0 {
		return fmt.Sprintf("%s (%d bytes)", name, wireLen)
	}
	return name
}

func writeHexdump(w io.Writer, data []byte) {
	writeHexdumpAt(w, 0, data, "")
}

func writeIndentedHexdump(w io.Writer, baseOffset int, data []byte) {
	writeHexdumpAt(w, baseOffset, data, "       ")
}

func writeHexdumpAt(w io.Writer, baseOffset int, data []byte, indent string) {
	const width = 16
	for offset := 0; offset < len(data); offset += width {
		end := offset + width
		if end > len(data) {
			end = len(data)
		}
		chunk := data[offset:end]

		fmt.Fprintf(w, "%s%08x  ", indent, baseOffset+offset)
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
