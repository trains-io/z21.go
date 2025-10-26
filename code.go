package z21

const (
	Z21_NO_LOCK        uint8 = 0x00
	Z21_START_LOCKED   uint8 = 0x01
	Z21_START_UNLOCKED uint8 = 0x02
)

// LAN_GET_CODE
type Code struct {
	Code uint8
}

// ---------- Message interface ----------

func (m *Code) Pack() ([]byte, error) {
	return []byte{}, nil
}

func (m *Code) Unpack(data []byte) error {
	return UnpackFields(data, &m.Code)
}

func (m *Code) EncapType() uint16 {
	return LAN_GET_CODE
}

// ---------- Correlatable interface ----------

func (m *Code) Key() (string, bool) {
	d := []byte{byte(LAN_GET_CODE)}
	f, err := fingerprint(d)
	if err != nil {
		return "", false
	}
	return f, true
}
