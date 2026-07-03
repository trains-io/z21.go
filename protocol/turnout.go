package protocol

import "fmt"

const (
	xHeaderGetTurnoutInfo    byte = 0x43
	xHeaderSetTurnout        byte = 0x53
	xHeaderGetExtAccessory   byte = 0x44
	xHeaderSetExtAccessory   byte = 0x54

	ExtAccessoryStatusValid   byte = 0x00
	ExtAccessoryStatusUnknown byte = 0xFF
)

// TurnoutPosition is ZZ in LAN_X_TURNOUT_INFO (spec §5.3).
type TurnoutPosition byte

const (
	TurnoutPositionNotSwitched TurnoutPosition = 0x00
	TurnoutPositionOutput1     TurnoutPosition = 0x01
	TurnoutPositionOutput2     TurnoutPosition = 0x02
	TurnoutPositionInvalid     TurnoutPosition = 0x03
)

// TurnoutSwitch describes LAN_X_SET_TURNOUT activation (spec §5.2).
type TurnoutSwitch struct {
	Activate bool
	Output2  bool
	Queue    bool
}

// EncodeTurnoutSwitchCommand returns the 10Q0A00P command byte (spec §5.2).
func EncodeTurnoutSwitchCommand(sw TurnoutSwitch) byte {
	var cmd byte = 0x80
	if sw.Queue {
		cmd |= 0x20
	}
	if sw.Activate {
		cmd |= 0x08
	}
	if sw.Output2 {
		cmd |= 0x01
	}
	return cmd
}

// TurnoutInfo is parsed from LAN_X_TURNOUT_INFO (spec §5.3).
type TurnoutInfo struct {
	Address  uint16
	Position TurnoutPosition
}

// ExtAccessoryInfo is parsed from LAN_X_EXT_ACCESSORY_INFO (spec §5.6).
type ExtAccessoryInfo struct {
	Address uint16
	Value   byte
	Status  byte
}

// GetTurnoutInfo returns LAN_X_GET_TURNOUT_INFO (spec §5.1).
func GetTurnoutInfo(address uint16) Message {
	addr := encodeFunctionAddressBytes(address)
	return Message{
		Header: HeaderLANX,
		Data:   appendLANXXOR(append([]byte{xHeaderGetTurnoutInfo}, addr...)),
	}
}

// SetTurnout returns LAN_X_SET_TURNOUT (spec §5.2).
func SetTurnout(address uint16, sw TurnoutSwitch) Message {
	addr := encodeFunctionAddressBytes(address)
	return Message{
		Header: HeaderLANX,
		Data: appendLANXXOR([]byte{
			xHeaderSetTurnout,
			addr[0],
			addr[1],
			EncodeTurnoutSwitchCommand(sw),
		}),
	}
}

// GetExtAccessoryInfo returns LAN_X_GET_EXT_ACCESSORY_INFO (spec §5.5).
func GetExtAccessoryInfo(address uint16) Message {
	addr := encodeFunctionAddressBytes(address)
	return Message{
		Header: HeaderLANX,
		Data: appendLANXXOR([]byte{
			xHeaderGetExtAccessory,
			addr[0],
			addr[1],
			0x00,
		}),
	}
}

// SetExtAccessory returns LAN_X_SET_EXT_ACCESSORY (spec §5.4).
func SetExtAccessory(address uint16, value byte) Message {
	addr := encodeFunctionAddressBytes(address)
	return Message{
		Header: HeaderLANX,
		Data: appendLANXXOR([]byte{
			xHeaderSetExtAccessory,
			addr[0],
			addr[1],
			value,
			0x00,
		}),
	}
}

// TurnoutInfoFromMessages extracts LAN_X_TURNOUT_INFO.
func TurnoutInfoFromMessages(msgs []Message) (TurnoutInfo, error) {
	for _, msg := range msgs {
		if msg.Header != HeaderLANX {
			continue
		}
		info, err := ParseTurnoutInfo(msg.Data)
		if err != nil {
			continue
		}
		return info, nil
	}
	return TurnoutInfo{}, fmt.Errorf("z21: no LAN_X_TURNOUT_INFO reply")
}

// ParseTurnoutInfo decodes LAN_X_TURNOUT_INFO (spec §5.3).
func ParseTurnoutInfo(data []byte) (TurnoutInfo, error) {
	if len(data) < 5 {
		return TurnoutInfo{}, fmt.Errorf("z21: turnout info too short (%d bytes)", len(data))
	}
	if data[0] != xHeaderGetTurnoutInfo {
		return TurnoutInfo{}, fmt.Errorf("z21: not a LAN_X_TURNOUT_INFO reply")
	}
	if data[len(data)-1] != lanXXOR(data[:len(data)-1]) {
		return TurnoutInfo{}, fmt.Errorf("z21: invalid LAN_X_TURNOUT_INFO checksum")
	}
	addr, err := parseFunctionAddressBytes(data[1:3])
	if err != nil {
		return TurnoutInfo{}, err
	}
	return TurnoutInfo{
		Address:  addr,
		Position: TurnoutPosition(data[3] & 0x03),
	}, nil
}

// IsTurnoutInfo reports whether data is LAN_X_TURNOUT_INFO.
func IsTurnoutInfo(data []byte) bool {
	_, err := ParseTurnoutInfo(data)
	return err == nil
}

// ExtAccessoryInfoFromMessages extracts LAN_X_EXT_ACCESSORY_INFO.
func ExtAccessoryInfoFromMessages(msgs []Message) (ExtAccessoryInfo, error) {
	for _, msg := range msgs {
		if msg.Header != HeaderLANX {
			continue
		}
		info, err := ParseExtAccessoryInfo(msg.Data)
		if err != nil {
			continue
		}
		return info, nil
	}
	return ExtAccessoryInfo{}, fmt.Errorf("z21: no LAN_X_EXT_ACCESSORY_INFO reply")
}

// ParseExtAccessoryInfo decodes LAN_X_EXT_ACCESSORY_INFO (spec §5.6).
func ParseExtAccessoryInfo(data []byte) (ExtAccessoryInfo, error) {
	if len(data) < 6 {
		return ExtAccessoryInfo{}, fmt.Errorf("z21: ext accessory info too short (%d bytes)", len(data))
	}
	if data[0] != xHeaderGetExtAccessory {
		return ExtAccessoryInfo{}, fmt.Errorf("z21: not a LAN_X_EXT_ACCESSORY_INFO reply")
	}
	if data[len(data)-1] != lanXXOR(data[:len(data)-1]) {
		return ExtAccessoryInfo{}, fmt.Errorf("z21: invalid LAN_X_EXT_ACCESSORY_INFO checksum")
	}
	addr, err := parseFunctionAddressBytes(data[1:3])
	if err != nil {
		return ExtAccessoryInfo{}, err
	}
	return ExtAccessoryInfo{
		Address: addr,
		Value:   data[3],
		Status:  data[4],
	}, nil
}

// IsExtAccessoryInfo reports whether data is LAN_X_EXT_ACCESSORY_INFO.
func IsExtAccessoryInfo(data []byte) bool {
	_, err := ParseExtAccessoryInfo(data)
	return err == nil
}

// FormatTurnoutPosition renders a turnout position label (spec §5.3).
func FormatTurnoutPosition(pos TurnoutPosition) string {
	switch pos {
	case TurnoutPositionNotSwitched:
		return "not switched"
	case TurnoutPositionOutput1:
		return "output 1"
	case TurnoutPositionOutput2:
		return "output 2"
	case TurnoutPositionInvalid:
		return "invalid"
	default:
		return fmt.Sprintf("unknown (0x%02x)", pos)
	}
}
