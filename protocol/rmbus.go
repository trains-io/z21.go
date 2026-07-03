package protocol

import "fmt"

const (
	// RMBusInputsPerGroup is the number of feedback modules per group index (spec §7).
	RMBusInputsPerGroup = 10

	// RMBusProgramEndAddress ends the R-BUS programming sequence (spec §7.3).
	RMBusProgramEndAddress byte = 0x00
)

// RMBusStatus is a LAN_RMBUS_DATACHANGED payload (spec §7.1).
type RMBusStatus struct {
	GroupIndex uint8
	Status     [RMBusInputsPerGroup]byte
}

// GetRMBusData returns LAN_RMBUS_GETDATA for one group (spec §7.2).
// Group index 0 covers modules 1–10; index 1 covers modules 11–20.
func GetRMBusData(groupIndex uint8) Message {
	return Message{
		Header: HeaderLANRMBusGetData,
		Data:   []byte{groupIndex},
	}
}

// ProgramRMBusModule returns LAN_RMBUS_PROGRAMMODULE (spec §7.3).
// address is the new module address (1–20), or RMBusProgramEndAddress to finish programming.
func ProgramRMBusModule(address byte) (Message, error) {
	if address > 20 && address != RMBusProgramEndAddress {
		return Message{}, fmt.Errorf("z21: R-BUS module address must be 0 or 1..20")
	}
	return Message{
		Header: HeaderLANRMBusProgramModule,
		Data:   []byte{address},
	}, nil
}

// RMBusModuleAddress maps a group index and byte offset to a module address (1–20).
func RMBusModuleAddress(groupIndex, offset uint8) (uint8, error) {
	if offset >= RMBusInputsPerGroup {
		return 0, fmt.Errorf("z21: R-BUS module offset out of range")
	}
	return groupIndex*RMBusInputsPerGroup + offset + 1, nil
}

// RMBusActiveInputs returns 1-based input numbers with an active contact in status.
func RMBusActiveInputs(status byte) []uint8 {
	var inputs []uint8
	for bit := uint8(0); bit < 8; bit++ {
		if status&(1<<bit) != 0 {
			inputs = append(inputs, bit+1)
		}
	}
	return inputs
}

// ParseRMBusStatus decodes LAN_RMBUS_DATACHANGED / LAN_RMBUS_GETDATA reply (spec §7.1).
func ParseRMBusStatus(data []byte) (RMBusStatus, error) {
	if len(data) < 1+RMBusInputsPerGroup {
		return RMBusStatus{}, fmt.Errorf("z21: R-BUS status too short (%d bytes)", len(data))
	}
	var out RMBusStatus
	out.GroupIndex = data[0]
	copy(out.Status[:], data[1:1+RMBusInputsPerGroup])
	return out, nil
}

// RMBusStatusFromMessages extracts LAN_RMBUS_DATACHANGED replies.
func RMBusStatusFromMessages(msgs []Message) ([]RMBusStatus, error) {
	var out []RMBusStatus
	for _, msg := range msgs {
		if msg.Header != HeaderLANRMBusDataChanged && msg.Header != HeaderLANRMBusGetData {
			continue
		}
		status, err := ParseRMBusStatus(msg.Data)
		if err != nil {
			return nil, err
		}
		out = append(out, status)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("z21: no LAN_RMBUS_DATACHANGED reply")
	}
	return out, nil
}

// IsRMBusDataChanged reports whether msg is an R-BUS feedback broadcast.
func IsRMBusDataChanged(msg Message) bool {
	return msg.Header == HeaderLANRMBusDataChanged
}
