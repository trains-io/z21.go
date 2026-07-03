package protocol

import (
	"encoding/binary"
	"fmt"
	"strings"
)

const (
	// CANBoosterNameLen is the fixed name field size (spec §10.2.1).
	CANBoosterNameLen = 16

	canBoosterSystemStateLen = 10
)

// CAN booster state bitmasks (spec §10.2.3).
const (
	CANBoosterStateBrakeGenActive  uint16 = 0x0001
	CANBoosterStateShortCircuit    uint16 = 0x0020
	CANBoosterStateTrackVoltageOff uint16 = 0x0080
	CANBoosterStateOutputDisabled  uint16 = 0x0100
	CANBoosterStateRailComActive   uint16 = 0x0800
)

// CANBoosterTrackPower selects a LAN_CAN_BOOSTER_SET_TRACKPOWER action (spec §10.2.4).
type CANBoosterTrackPower byte

const (
	CANBoosterTrackPowerDeactivateAll  CANBoosterTrackPower = 0x00
	CANBoosterTrackPowerActivateAll    CANBoosterTrackPower = 0xFF
	CANBoosterTrackPowerDeactivateOut1 CANBoosterTrackPower = 0x10
	CANBoosterTrackPowerActivateOut1   CANBoosterTrackPower = 0x11
	CANBoosterTrackPowerDeactivateOut2 CANBoosterTrackPower = 0x20
	CANBoosterTrackPowerActivateOut2   CANBoosterTrackPower = 0x22
)

// CANBoosterSystemState is a LAN_CAN_BOOSTER_SYSTEMSTATE_CHGD payload (spec §10.2.3).
type CANBoosterSystemState struct {
	NetID      uint16
	OutputPort uint16
	State      uint16
	VCCVoltage uint16 // mV
	Current    uint16 // mA
}

// GetCANDeviceDescription returns LAN_CAN_DEVICE_GET_DESCRIPTION (spec §10.2.1).
func GetCANDeviceDescription(netID uint16) Message {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, netID)
	return Message{Header: HeaderLANCANDeviceGetDescription, Data: data}
}

// SetCANDeviceDescription returns LAN_CAN_DEVICE_SET_DESCRIPTION (spec §10.2.2).
func SetCANDeviceDescription(netID uint16, name string) (Message, error) {
	if strings.ContainsAny(name, "\"\\") {
		return Message{}, fmt.Errorf("z21: CAN device name must not contain \" or \\")
	}
	if len(name) >= CANBoosterNameLen {
		return Message{}, fmt.Errorf("z21: CAN device name too long (%d bytes, max %d)", len(name), CANBoosterNameLen-1)
	}

	data := make([]byte, 2+CANBoosterNameLen)
	binary.LittleEndian.PutUint16(data, netID)
	copy(data[2:], name)
	return Message{Header: HeaderLANCANDeviceSetDescription, Data: data}, nil
}

// SetCANBoosterTrackPower returns LAN_CAN_BOOSTER_SET_TRACKPOWER (spec §10.2.4).
func SetCANBoosterTrackPower(netID uint16, power CANBoosterTrackPower) Message {
	data := make([]byte, 3)
	binary.LittleEndian.PutUint16(data, netID)
	data[2] = byte(power)
	return Message{Header: HeaderLANCANBoosterSetTrackPower, Data: data}
}

// ParseCANDeviceDescription decodes a LAN_CAN_DEVICE_GET_DESCRIPTION reply (spec §10.2.1).
func ParseCANDeviceDescription(data []byte) (netID uint16, name string, err error) {
	if len(data) < 2+CANBoosterNameLen {
		return 0, "", fmt.Errorf("z21: CAN device description too short (%d bytes)", len(data))
	}
	netID = binary.LittleEndian.Uint16(data[0:2])
	name = parseCANDeviceName(data[2 : 2+CANBoosterNameLen])
	return netID, name, nil
}

// ParseCANBoosterSystemState decodes LAN_CAN_BOOSTER_SYSTEMSTATE_CHGD (spec §10.2.3).
func ParseCANBoosterSystemState(data []byte) (CANBoosterSystemState, error) {
	if len(data) < canBoosterSystemStateLen {
		return CANBoosterSystemState{}, fmt.Errorf("z21: CAN booster system state too short (%d bytes)", len(data))
	}
	return CANBoosterSystemState{
		NetID:      binary.LittleEndian.Uint16(data[0:2]),
		OutputPort: binary.LittleEndian.Uint16(data[2:4]),
		State:      binary.LittleEndian.Uint16(data[4:6]),
		VCCVoltage: binary.LittleEndian.Uint16(data[6:8]),
		Current:    binary.LittleEndian.Uint16(data[8:10]),
	}, nil
}

// CANBoosterSystemStatesFromMessages extracts LAN_CAN_BOOSTER_SYSTEMSTATE_CHGD replies.
func CANBoosterSystemStatesFromMessages(msgs []Message) ([]CANBoosterSystemState, error) {
	var out []CANBoosterSystemState
	for _, msg := range msgs {
		if msg.Header != HeaderLANCANBoosterSystemState {
			continue
		}
		state, err := ParseCANBoosterSystemState(msg.Data)
		if err != nil {
			return nil, err
		}
		out = append(out, state)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("z21: no LAN_CAN_BOOSTER_SYSTEMSTATE_CHGD reply")
	}
	return out, nil
}

// HasCANBoosterState reports whether state includes flag.
func HasCANBoosterState(state, flag uint16) bool {
	return state&flag != 0
}

// FormatCANBoosterState renders active CAN booster state flags.
func FormatCANBoosterState(state uint16) string {
	var flags []string
	if HasCANBoosterState(state, CANBoosterStateBrakeGenActive) {
		flags = append(flags, "BrakeGen")
	}
	if HasCANBoosterState(state, CANBoosterStateShortCircuit) {
		flags = append(flags, "ShortCircuit")
	}
	if HasCANBoosterState(state, CANBoosterStateTrackVoltageOff) {
		flags = append(flags, "TrackOff")
	}
	if HasCANBoosterState(state, CANBoosterStateOutputDisabled) {
		flags = append(flags, "OutputDisabled")
	}
	if HasCANBoosterState(state, CANBoosterStateRailComActive) {
		flags = append(flags, "RailCom")
	}
	if len(flags) == 0 {
		return "none"
	}
	out := flags[0]
	for _, f := range flags[1:] {
		out += "|" + f
	}
	return out
}

// IsCANBoosterSystemStateChanged reports whether msg is a CAN booster status broadcast.
func IsCANBoosterSystemStateChanged(msg Message) bool {
	return msg.Header == HeaderLANCANBoosterSystemState
}

func parseCANDeviceName(raw []byte) string {
	if i := strings.IndexByte(string(raw), 0); i >= 0 {
		return string(raw[:i])
	}
	return strings.TrimRight(string(raw), "\x00")
}
