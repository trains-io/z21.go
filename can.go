package z21

import (
	"encoding/binary"
)

const (
	CAN_BROADCAST_NID uint16 = 0xd000
)

const (
	CANMessageTypeOccupancy uint8 = 0x00
	CANMessageTypeStatus    uint8 = 0x01
)

// LAN_CAN_DETECTOR
type CanDetector struct {
	NetworkID uint16 `json:"network_id"`
	Address   uint16 `json:"address"`
	Port      uint8  `json:"port"`
	Type      uint8  `json:"type"`
	Value1    uint16 `json:"value1"`
	Value2    uint16 `json:"value2"`
}

// ---------- Message interface ----------

func (m *CanDetector) String() string {
	return "can"
}

func (m *CanDetector) Pack() ([]byte, error) {
	bytes := make([]byte, 3)
	bytes[0] = CANMessageTypeOccupancy
	binary.LittleEndian.PutUint16(bytes[1:], m.NetworkID)
	return bytes, nil
}

func (m *CanDetector) Unpack(data []byte) error {
	if err := UnpackFields(
		data,
		&m.NetworkID,
		&m.Address,
		&m.Port,
		&m.Type,
		&m.Value1,
		&m.Value2,
	); err != nil {
		return err
	}
	return nil
}

func (m *CanDetector) EncapType() uint16 {
	return LAN_CAN_DETECTOR
}

// ---------- Correlatable interface ----------

func (m *CanDetector) Key() (string, bool) {
	return "", false
}
