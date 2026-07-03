package protocol

import "fmt"

const (
	CodeNoLock           byte = 0x00
	CodeZ21StartLocked   byte = 0x01
	CodeZ21StartUnlocked byte = 0x02
)

// GetCode returns a request for the Z21 feature lock state (spec §2.21).
func GetCode() Message {
	return Message{Header: HeaderLANGetCode}
}

// CodeFromMessages extracts the lock code from a Call reply.
func CodeFromMessages(msgs []Message) (byte, error) {
	for _, msg := range msgs {
		if msg.Header != HeaderLANGetCode {
			continue
		}
		code, err := ParseCode(msg.Data)
		if err != nil {
			continue
		}
		return code, nil
	}
	return 0, fmt.Errorf("z21: no LAN_GET_CODE reply")
}

// ParseCode decodes the 8-bit lock code from reply data.
func ParseCode(data []byte) (byte, error) {
	if len(data) < 1 {
		return 0, fmt.Errorf("z21: code reply too short (%d bytes)", len(data))
	}
	return data[0], nil
}

// FormatLockCode renders the lock code for display (spec §2.21).
func FormatLockCode(code byte) string {
	switch code {
	case CodeNoLock:
		return "all features permitted"
	case CodeZ21StartLocked:
		return "z21 start locked (driving and switching blocked)"
	case CodeZ21StartUnlocked:
		return "z21 start unlocked (driving and switching permitted)"
	default:
		return fmt.Sprintf("unknown (0x%02x)", code)
	}
}
