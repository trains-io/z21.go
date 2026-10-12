package protocol

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// CentralState bits, reported in SystemState.CentralState and XStatus.CentralState
// (spec §2.12, §2.18).
const (
	CentralStateEmergencyStop   byte = 0x01
	CentralStateTrackVoltageOff byte = 0x02
	CentralStateShortCircuit    byte = 0x04
	CentralStateProgrammingMode byte = 0x20
)

// CentralStateEx bits, reported in SystemState.CentralStateEx (spec §2.18).
const (
	CentralStateExHighTemperature      byte = 0x01
	CentralStateExPowerLost            byte = 0x02
	CentralStateExShortCircuitExternal byte = 0x04
	CentralStateExShortCircuitInternal byte = 0x08
	CentralStateExRCN213               byte = 0x20
)

// Capability bits, reported in SystemState.Capabilities (spec §2.18).
const (
	CapabilityDCC             byte = 0x01
	CapabilityMM              byte = 0x02
	CapabilityRailCom         byte = 0x08
	CapabilityLocoCmds        byte = 0x10
	CapabilityAccessoryCmds   byte = 0x20
	CapabilityDetectorCmds    byte = 0x40
	CapabilityNeedsUnlockCode byte = 0x80
)

// SystemState is parsed from a LAN_SYSTEMSTATE_DATACHANGED reply (spec §2.18).
type SystemState struct {
	MainCurrent         int16
	ProgCurrent         int16
	FilteredMainCurrent int16
	Temperature         int16
	SupplyVoltage       uint16
	VCCVoltage          uint16
	CentralState        byte
	CentralStateEx      byte
	Capabilities        byte
}

// SystemStateGetData returns a request for the current system state (spec §2.19).
func SystemStateGetData() Message {
	return Message{Header: HeaderLANSystemStateGetData}
}

// EncodeSystemState builds LAN_SYSTEMSTATE_DATACHANGED (spec §2.18).
// The reserved byte 14 is always zero.
func EncodeSystemState(s SystemState) Message {
	data := make([]byte, 16)
	binary.LittleEndian.PutUint16(data[0:2], uint16(s.MainCurrent))
	binary.LittleEndian.PutUint16(data[2:4], uint16(s.ProgCurrent))
	binary.LittleEndian.PutUint16(data[4:6], uint16(s.FilteredMainCurrent))
	binary.LittleEndian.PutUint16(data[6:8], uint16(s.Temperature))
	binary.LittleEndian.PutUint16(data[8:10], s.SupplyVoltage)
	binary.LittleEndian.PutUint16(data[10:12], s.VCCVoltage)
	data[12] = s.CentralState
	data[13] = s.CentralStateEx
	data[15] = s.Capabilities
	return Message{Header: HeaderLANSystemStateDataChanged, Data: data}
}

// SystemStateFromMessages extracts system state from a Call reply.
func SystemStateFromMessages(msgs []Message) (SystemState, error) {
	for _, msg := range msgs {
		if msg.Header != HeaderLANSystemStateDataChanged {
			continue
		}
		state, err := ParseSystemState(msg.Data)
		if err != nil {
			continue
		}
		return state, nil
	}
	return SystemState{}, fmt.Errorf("z21: no LAN_SYSTEMSTATE_DATACHANGED reply")
}

// ParseSystemState decodes the 16-byte system state payload.
func ParseSystemState(data []byte) (SystemState, error) {
	if len(data) < 16 {
		return SystemState{}, fmt.Errorf("z21: system state too short (%d bytes)", len(data))
	}
	return SystemState{
		MainCurrent:         int16(binary.LittleEndian.Uint16(data[0:2])),
		ProgCurrent:         int16(binary.LittleEndian.Uint16(data[2:4])),
		FilteredMainCurrent: int16(binary.LittleEndian.Uint16(data[4:6])),
		Temperature:         int16(binary.LittleEndian.Uint16(data[6:8])),
		SupplyVoltage:       binary.LittleEndian.Uint16(data[8:10]),
		VCCVoltage:          binary.LittleEndian.Uint16(data[10:12]),
		CentralState:        data[12],
		CentralStateEx:      data[13],
		Capabilities:        data[15],
	}, nil
}

// IsSystemStateDataChanged reports whether msg is LAN_SYSTEMSTATE_DATACHANGED (spec §2.18).
func IsSystemStateDataChanged(msg Message) bool {
	if msg.Header != HeaderLANSystemStateDataChanged {
		return false
	}
	_, err := ParseSystemState(msg.Data)
	return err == nil
}

// FormatCentralState renders CentralState flags (spec §2.18; same bitmask as §2.12).
func FormatCentralState(state byte) string {
	return FormatXStatusFlags(state)
}

// FormatCentralStateExFlags renders active CentralStateEx flags (spec §2.18).
func FormatCentralStateExFlags(state byte) string {
	var flags []string
	if state&CentralStateExHighTemperature != 0 {
		flags = append(flags, "high temperature")
	}
	if state&CentralStateExPowerLost != 0 {
		flags = append(flags, "power lost")
	}
	if state&CentralStateExShortCircuitExternal != 0 {
		flags = append(flags, "external short circuit")
	}
	if state&CentralStateExShortCircuitInternal != 0 {
		flags = append(flags, "internal short circuit")
	}
	if state&CentralStateExRCN213 != 0 {
		flags = append(flags, "RCN-213 turnout addresses")
	}
	if len(flags) == 0 {
		return "ok"
	}
	return strings.Join(flags, ", ")
}

// FormatCapabilities renders active capability flags (spec §2.18).
func FormatCapabilities(caps byte) string {
	if caps == 0 {
		return "(not reported)"
	}
	var flags []string
	if caps&CapabilityDCC != 0 {
		flags = append(flags, "DCC")
	}
	if caps&CapabilityMM != 0 {
		flags = append(flags, "MM")
	}
	if caps&CapabilityRailCom != 0 {
		flags = append(flags, "RailCom")
	}
	if caps&CapabilityLocoCmds != 0 {
		flags = append(flags, "loco commands")
	}
	if caps&CapabilityAccessoryCmds != 0 {
		flags = append(flags, "accessory commands")
	}
	if caps&CapabilityDetectorCmds != 0 {
		flags = append(flags, "detector commands")
	}
	if caps&CapabilityNeedsUnlockCode != 0 {
		flags = append(flags, "needs unlock code")
	}
	if len(flags) == 0 {
		return fmt.Sprintf("0x%02x", caps)
	}
	return strings.Join(flags, ", ")
}
