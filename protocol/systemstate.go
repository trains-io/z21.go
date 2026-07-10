package protocol

import (
	"encoding/binary"
	"fmt"
	"strings"
)

const (
	centralStateExHighTemperature       byte = 0x01
	centralStateExPowerLost             byte = 0x02
	centralStateExShortCircuitExternal  byte = 0x04
	centralStateExShortCircuitInternal  byte = 0x08
	centralStateExRCN213                byte = 0x20
	capabilityDCC                       byte = 0x01
	capabilityMM                        byte = 0x02
	capabilityRailCom                   byte = 0x08
	capabilityLocoCmds                  byte = 0x10
	capabilityAccessoryCmds             byte = 0x20
	capabilityDetectorCmds              byte = 0x40
	capabilityNeedsUnlockCode           byte = 0x80
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
	if state&centralStateExHighTemperature != 0 {
		flags = append(flags, "high temperature")
	}
	if state&centralStateExPowerLost != 0 {
		flags = append(flags, "power lost")
	}
	if state&centralStateExShortCircuitExternal != 0 {
		flags = append(flags, "external short circuit")
	}
	if state&centralStateExShortCircuitInternal != 0 {
		flags = append(flags, "internal short circuit")
	}
	if state&centralStateExRCN213 != 0 {
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
	if caps&capabilityDCC != 0 {
		flags = append(flags, "DCC")
	}
	if caps&capabilityMM != 0 {
		flags = append(flags, "MM")
	}
	if caps&capabilityRailCom != 0 {
		flags = append(flags, "RailCom")
	}
	if caps&capabilityLocoCmds != 0 {
		flags = append(flags, "loco commands")
	}
	if caps&capabilityAccessoryCmds != 0 {
		flags = append(flags, "accessory commands")
	}
	if caps&capabilityDetectorCmds != 0 {
		flags = append(flags, "detector commands")
	}
	if caps&capabilityNeedsUnlockCode != 0 {
		flags = append(flags, "needs unlock code")
	}
	if len(flags) == 0 {
		return fmt.Sprintf("0x%02x", caps)
	}
	return strings.Join(flags, ", ")
}
