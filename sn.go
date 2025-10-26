package z21

import (
	"encoding/binary"
)

type SerialNumber struct {
	SerialNumber uint32
}

// LAN_GET_SERIAL_NUMBER

// ---------- Message interface ----------

func (m *SerialNumber) Pack() ([]byte, error) {
	return PackFields()
}

func (m *SerialNumber) Unpack(data []byte) error {
	var sn uint32
	err := UnpackFields(data, &sn)
	if err != nil {
		return err
	}
	m.decodeSerialNumber(sn)
	return nil
}

func (m *SerialNumber) EncapType() uint16 {
	return LAN_GET_SERIAL_NUMBER
}

// ---------- Correlatable interface ----------

func (m *SerialNumber) Key() (string, bool) {
	d := []byte{byte(LAN_GET_SERIAL_NUMBER)}
	f, err := fingerprint(d)
	if err != nil {
		return "", false
	}
	return f, true
}

func (m *SerialNumber) decodeSerialNumber(sn uint32) {
	b := []byte{
		byte(sn),
		byte(sn >> 8),
		byte(sn >> 16),
		byte(sn >> 24),
	}
	m.SerialNumber = binary.LittleEndian.Uint32(b)
}
