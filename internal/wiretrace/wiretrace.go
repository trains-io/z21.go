// Package wiretrace formats Z21 UDP payloads as human-readable hexdumps for the
// client and server tracers.
package wiretrace

import (
	"fmt"
	"io"

	"github.com/trains-io/z21.go/protocol"
)

// Dump writes a header line "label (N bytes)" followed by a hexdump of raw,
// split per dataset when raw parses as Z21 datasets.
func Dump(w io.Writer, label string, raw []byte) {
	datasets, err := wireDatasets(raw)
	if err != nil || len(datasets) == 0 {
		fmt.Fprintf(w, "%s (%d bytes)\n", label, len(raw))
		writeHexdumpAt(w, 0, raw, "")
		return
	}

	if len(datasets) == 1 {
		ds := datasets[0]
		fmt.Fprintf(w, "%s (%d bytes)\n", label, len(raw))
		fmt.Fprintf(w, "   %s\n", datasetSummary(ds.msg, len(ds.wire)))
		writeHexdumpAt(w, 0, ds.wire, "")
		return
	}

	fmt.Fprintf(w, "%s (%d bytes, %d datasets)\n", label, len(raw), len(datasets))
	for i, ds := range datasets {
		fmt.Fprintf(w, "   [%d/%d] %s\n", i+1, len(datasets), datasetSummary(ds.msg, len(ds.wire)))
		writeHexdumpAt(w, ds.offset, ds.wire, "       ")
	}
}

type wireDataset struct {
	msg    protocol.Message
	offset int
	wire   []byte
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

func writeHexdumpAt(w io.Writer, baseOffset int, data []byte, indent string) {
	const width = 16
	for offset := 0; offset < len(data); offset += width {
		end := min(offset+width, len(data))
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
