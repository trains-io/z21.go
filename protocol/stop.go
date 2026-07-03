package protocol

import "fmt"

const (
	xHeaderSetStop     byte = 0x80
	xChecksumSetStop   byte = 0x80
	xHeaderBCStopped   byte = 0x81
	xBCStoppedDB0      byte = 0x00
	xChecksumBCStopped byte = 0x81
)

// SetStop returns LAN_X_SET_STOP (spec §2.13).
// Locomotives are halted; track voltage remains on.
func SetStop() Message {
	return Message{
		Header: HeaderLANX,
		Data:   []byte{xHeaderSetStop, xChecksumSetStop},
	}
}

// BCStoppedFromMessages confirms a LAN_X_BC_STOPPED reply (spec §2.14).
func BCStoppedFromMessages(msgs []Message) error {
	for _, msg := range msgs {
		if msg.Header != HeaderLANX {
			continue
		}
		if err := ParseBCStopped(msg.Data); err != nil {
			continue
		}
		return nil
	}
	return fmt.Errorf("z21: no LAN_X_BC_STOPPED reply")
}

// ParseBCStopped decodes LAN_X_BC_STOPPED (spec §2.14).
func ParseBCStopped(data []byte) error {
	if len(data) < 3 {
		return fmt.Errorf("z21: emergency stop reply too short (%d bytes)", len(data))
	}
	if data[0] != xHeaderBCStopped || data[1] != xBCStoppedDB0 || data[2] != xChecksumBCStopped {
		return fmt.Errorf("z21: not a LAN_X_BC_STOPPED reply")
	}
	return nil
}

// IsBCStopped reports whether data is a LAN_X_BC_STOPPED broadcast.
func IsBCStopped(data []byte) bool {
	return ParseBCStopped(data) == nil
}
