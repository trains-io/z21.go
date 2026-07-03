package protocol

import "fmt"

const (
	xHeaderXBusBC        byte = 0x61
	xBCDB0TrackPowerOff  byte = 0x00
	xBCDB0TrackPowerOn   byte = 0x01
	xBCDB0ProgrammingMode byte = 0x02
	xBCDB0ShortCircuit   byte = 0x08
	xBCDB0CVNackSC       byte = 0x12
	xBCDB0CVNack         byte = 0x13
	xBCDB0UnknownCommand byte = 0x82
)

func parseXBusBC(data []byte, db0 byte) error {
	if len(data) < 2 {
		return fmt.Errorf("z21: X-Bus broadcast too short (%d bytes)", len(data))
	}
	if data[0] != xHeaderXBusBC || data[1] != db0 {
		return fmt.Errorf("z21: not a LAN_X X-Bus broadcast with DB0=0x%02x", db0)
	}
	if len(data) >= 3 && data[2] != xHeaderXBusBC^db0 {
		return fmt.Errorf("z21: invalid X-Bus broadcast checksum")
	}
	return nil
}

// ParseBCProgrammingMode decodes LAN_X_BC_PROGRAMMING_MODE (spec §2.9).
func ParseBCProgrammingMode(data []byte) error {
	return parseXBusBC(data, xBCDB0ProgrammingMode)
}

// IsBCProgrammingMode reports whether data is LAN_X_BC_PROGRAMMING_MODE.
func IsBCProgrammingMode(data []byte) bool {
	return ParseBCProgrammingMode(data) == nil
}

// ParseBCShortCircuit decodes LAN_X_BC_TRACK_SHORT_CIRCUIT (spec §2.10).
func ParseBCShortCircuit(data []byte) error {
	return parseXBusBC(data, xBCDB0ShortCircuit)
}

// IsBCShortCircuit reports whether data is LAN_X_BC_TRACK_SHORT_CIRCUIT.
func IsBCShortCircuit(data []byte) bool {
	return ParseBCShortCircuit(data) == nil
}

// ParseUnknownCommand decodes LAN_X_UNKNOWN_COMMAND (spec §2.11).
func ParseUnknownCommand(data []byte) error {
	return parseXBusBC(data, xBCDB0UnknownCommand)
}

// IsUnknownCommand reports whether data is LAN_X_UNKNOWN_COMMAND.
func IsUnknownCommand(data []byte) bool {
	return ParseUnknownCommand(data) == nil
}

func xBusBCMessageName(data []byte) string {
	if len(data) < 2 || data[0] != xHeaderXBusBC {
		return "LAN_X_BC"
	}
	switch data[1] {
	case xBCDB0TrackPowerOff:
		return "LAN_X_BC_TRACK_POWER_OFF"
	case xBCDB0TrackPowerOn:
		return "LAN_X_BC_TRACK_POWER_ON"
	case xBCDB0ProgrammingMode:
		return "LAN_X_BC_PROGRAMMING_MODE"
	case xBCDB0ShortCircuit:
		return "LAN_X_BC_TRACK_SHORT_CIRCUIT"
	case xBCDB0CVNackSC:
		return "LAN_X_CV_NACK_SC"
	case xBCDB0CVNack:
		return "LAN_X_CV_NACK"
	case xBCDB0UnknownCommand:
		return "LAN_X_UNKNOWN_COMMAND"
	default:
		return fmt.Sprintf("LAN_X_BC db0=0x%02x", data[1])
	}
}
