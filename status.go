package z21

const (
	EMERGENCY_STOP          uint8 = 0x01 // bit 0
	TRACK_VOLTAGE_OFF       uint8 = 0x02 // bit 1
	SHORT_CIRCUIT           uint8 = 0x04 // bit 2
	PROGRAMMING_MODE_ACTIVE uint8 = 0x20 // bit 5
)

// LAN_X_GET_STATUS
type Status struct {
	Mask Mask8
}

// ---------- Message interface ----------

func (m *Status) Pack() ([]byte, error) {
	return []byte{0x21, 0x24, 0x05}, nil
}

func (m *Status) Unpack(data []byte) error {
	return UnpackFields(data[2:3], &m.Mask)
}

func (m *Status) EncapType() uint16 {
	return LAN_X
}

// ---------- Correlatable interface ----------

func (m *Status) Key() (string, bool) {
	d := []byte{byte(LAN_X), byte(LAN_X_STATUS_CHANGED)}
	f, err := fingerprint(d)
	if err != nil {
		return "", false
	}
	return f, true
}
