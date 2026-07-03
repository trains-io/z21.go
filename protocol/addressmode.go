package protocol

import (
	"encoding/binary"
	"fmt"
)

// Output format modes for locomotive and accessory decoder addresses (spec §3).
const (
	OutputModeDCC byte = 0x00
	OutputModeMM  byte = 0x01
)

// AddressMode is a persisted DCC/MM output format for one address (spec §3).
type AddressMode struct {
	Address uint16
	Mode    byte
}

func encodeAddressBE(addr uint16) []byte {
	return []byte{byte(addr >> 8), byte(addr)}
}

func parseAddressModeReply(data []byte) (AddressMode, error) {
	if len(data) < 3 {
		return AddressMode{}, fmt.Errorf("z21: address mode reply too short (%d bytes)", len(data))
	}
	return AddressMode{
		Address: binary.BigEndian.Uint16(data[0:2]),
		Mode:    data[2],
	}, nil
}

// FormatOutputMode renders a DCC/MM mode byte (spec §3).
func FormatOutputMode(mode byte) string {
	switch mode {
	case OutputModeDCC:
		return "DCC"
	case OutputModeMM:
		return "MM"
	default:
		return fmt.Sprintf("unknown (0x%02x)", mode)
	}
}
