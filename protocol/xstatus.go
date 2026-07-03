package protocol

import (
	"fmt"
	"strings"
)

const (
	xCommandGetStatus      byte = 0x24
	xHeaderStatusReply     byte = 0x62
	xStatusDB0             byte = 0x22
	xStatusEmergencyStop   byte = 0x01
	xStatusTrackVoltageOff byte = 0x02
	xStatusShortCircuit    byte = 0x04
	xStatusProgrammingMode byte = 0x20
)

// XStatus is parsed from a LAN_X_GET_STATUS reply (spec §2.4 / §2.12).
type XStatus struct {
	CentralState byte
}

// GetXStatus returns a request for the command station central state (spec §2.4).
func GetXStatus() Message {
	return Message{
		Header: HeaderLANX,
		Data:   []byte{xHeaderGetVersion, xCommandGetStatus, xHeaderGetVersion ^ xCommandGetStatus},
	}
}

// XStatusFromMessages extracts central state from a Call reply.
func XStatusFromMessages(msgs []Message) (XStatus, error) {
	for _, msg := range msgs {
		if msg.Header != HeaderLANX {
			continue
		}
		status, err := ParseXStatus(msg.Data)
		if err != nil {
			continue
		}
		return status, nil
	}
	return XStatus{}, fmt.Errorf("z21: no LAN_X_GET_STATUS reply")
}

// ParseXStatus decodes a LAN_X_STATUS_CHANGED payload.
func ParseXStatus(data []byte) (XStatus, error) {
	if len(data) < 4 {
		return XStatus{}, fmt.Errorf("z21: X status reply too short (%d bytes)", len(data))
	}
	if data[0] != xHeaderStatusReply || data[1] != xStatusDB0 {
		return XStatus{}, fmt.Errorf("z21: not a LAN_X_STATUS_CHANGED reply")
	}
	return XStatus{CentralState: data[2]}, nil
}

// IsXStatusChanged reports whether data is LAN_X_STATUS_CHANGED (spec §2.12).
func IsXStatusChanged(data []byte) bool {
	_, err := ParseXStatus(data)
	return err == nil
}

// FormatXStatusFlags renders active central-state flags (spec §2.12).
func FormatXStatusFlags(state byte) string {
	var flags []string
	if state&xStatusEmergencyStop != 0 {
		flags = append(flags, "emergency stop")
	}
	if state&xStatusTrackVoltageOff != 0 {
		flags = append(flags, "track voltage off")
	}
	if state&xStatusShortCircuit != 0 {
		flags = append(flags, "short circuit")
	}
	if state&xStatusProgrammingMode != 0 {
		flags = append(flags, "programming mode")
	}
	if len(flags) == 0 {
		return "ok"
	}
	return strings.Join(flags, ", ")
}
