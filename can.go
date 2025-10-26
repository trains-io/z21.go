package z21

const (
	CAN_BROADCAST_ADDR uint16 = 0xd000
)

// LAN_CAN_DETECTOR
type CanDetector struct {
	Address uint16
}

// ---------- Message interface ----------

func (m *CanDetector) Pack() ([]byte, error) {
	return []byte{0x00, byte(m.Address)}, nil
}

func (m *CanDetector) Unpack(data []byte) error {
	return nil
}

func (m *CanDetector) EncapType() uint16 {
	return LAN_CAN_DETECTOR
}

// ---------- Correlatable interface ----------

func (m *CanDetector) Key() (string, bool) {
	d := []byte{byte(LAN_CAN_DETECTOR)}
	f, err := fingerprint(d)
	if err != nil {
		return "", false
	}
	return f, true
}
