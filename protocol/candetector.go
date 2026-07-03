package protocol

import (
	"encoding/binary"
	"fmt"
)

const (
	cANDetectorPollType byte = 0x00

	// CANDetectorPollAll requests status from every CAN occupancy detector (spec §10.1).
	CANDetectorPollAll uint16 = 0xD000

	// CANDetectorTypeOccupancy is the occupancy status report (spec §10.1).
	CANDetectorTypeOccupancy byte = 0x01

	// CANDetectorTypeLocoAddressBase is the first RailCom loco-address pair type (spec §10.1).
	CANDetectorTypeLocoAddressBase byte = 0x11
	// CANDetectorTypeLocoAddressMax is the last RailCom loco-address pair type (spec §10.1).
	CANDetectorTypeLocoAddressMax byte = 0x1F

	cANDetectorLocoAddressMask     uint16 = 0x3FFF
	cANDetectorLocoDirectionMask   uint16 = 0xC000
	cANDetectorLocoDirectionForward  uint16 = 0x4000
	cANDetectorLocoDirectionBackward uint16 = 0x8000
)

// CANDetectorLocoDirection is the RailCom direction encoded in Value1/Value2 (spec §10.1).
type CANDetectorLocoDirection uint8

const (
	CANDetectorLocoDirectionNone CANDetectorLocoDirection = iota
	CANDetectorLocoDirectionForward
	CANDetectorLocoDirectionBackward
)

// CANDetectorLocoAddress is a decoded locomotive address from a CAN detector report.
type CANDetectorLocoAddress struct {
	Address   uint16
	Direction CANDetectorLocoDirection
}

// CANDetectorReport is a LAN_CAN_DETECTOR reply (spec §10.1).
type CANDetectorReport struct {
	NetID  uint16
	Addr   uint16
	Port   uint8
	Type   uint8
	Value1 uint16
	Value2 uint16
}

// GetCANDetector returns a LAN_CAN_DETECTOR poll for one NetID (spec §10.1).
func GetCANDetector(netID uint16) Message {
	return Message{
		Header: HeaderLANCANDetector,
		Data:   []byte{cANDetectorPollType, byte(netID), byte(netID >> 8)},
	}
}

// GetAllCANDetectors polls every CAN occupancy detector (NetID 0xD000).
func GetAllCANDetectors() Message {
	return GetCANDetector(CANDetectorPollAll)
}

// ParseCANDetector decodes a LAN_CAN_DETECTOR reply payload.
func ParseCANDetector(data []byte) (CANDetectorReport, error) {
	if len(data) < 10 {
		return CANDetectorReport{}, fmt.Errorf("z21: CAN detector reply too short (%d bytes)", len(data))
	}
	return CANDetectorReport{
		NetID:  binary.LittleEndian.Uint16(data[0:2]),
		Addr:   binary.LittleEndian.Uint16(data[2:4]),
		Port:   data[4],
		Type:   data[5],
		Value1: binary.LittleEndian.Uint16(data[6:8]),
		Value2: binary.LittleEndian.Uint16(data[8:10]),
	}, nil
}

// CANDetectorReportsFromMessages extracts LAN_CAN_DETECTOR replies.
func CANDetectorReportsFromMessages(msgs []Message) ([]CANDetectorReport, error) {
	var out []CANDetectorReport
	for _, msg := range msgs {
		if msg.Header != HeaderLANCANDetector {
			continue
		}
		report, err := ParseCANDetector(msg.Data)
		if err != nil {
			return nil, err
		}
		out = append(out, report)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("z21: no LAN_CAN_DETECTOR reply")
	}
	return out, nil
}

// OccupancyStatusLabel returns a short label for a type-0x01 occupancy Value1.
func OccupancyStatusLabel(value uint16) string {
	switch value {
	case 0x0000:
		return "free"
	case 0x0100:
		return "free"
	case 0x1000, 0x1100:
		return "occupied"
	case 0x1201, 0x1202, 0x1203:
		return "overload"
	default:
		return fmt.Sprintf("0x%04x", value)
	}
}

// IsCANDetectorLocoAddressType reports whether t is a RailCom loco-address pair type.
func IsCANDetectorLocoAddressType(t byte) bool {
	return t >= CANDetectorTypeLocoAddressBase && t <= CANDetectorTypeLocoAddressMax
}

// ParseCANDetectorLocoAddress decodes a Value1/Value2 locomotive address field (spec §10.1).
// A value of 0 means no locomotive or end of list.
func ParseCANDetectorLocoAddress(value uint16) CANDetectorLocoAddress {
	if value == 0 {
		return CANDetectorLocoAddress{}
	}
	out := CANDetectorLocoAddress{Address: value & cANDetectorLocoAddressMask}
	switch value & cANDetectorLocoDirectionMask {
	case cANDetectorLocoDirectionForward:
		out.Direction = CANDetectorLocoDirectionForward
	case cANDetectorLocoDirectionBackward:
		out.Direction = CANDetectorLocoDirectionBackward
	}
	return out
}

// LocoAddresses returns decoded locomotive addresses for RailCom loco-address reports.
func (r CANDetectorReport) LocoAddresses() [2]CANDetectorLocoAddress {
	return [2]CANDetectorLocoAddress{
		ParseCANDetectorLocoAddress(r.Value1),
		ParseCANDetectorLocoAddress(r.Value2),
	}
}

// IsCANDetectorReport reports whether msg is a LAN_CAN_DETECTOR message.
func IsCANDetectorReport(msg Message) bool {
	return msg.Header == HeaderLANCANDetector && len(msg.Data) >= 10
}
