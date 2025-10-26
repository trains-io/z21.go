package z21

type Stop struct{}

// LAN_X_SET_STOP

// ---------- Message interface ----------

func (m *Stop) Pack() ([]byte, error) {
	return []byte{0x80, 0x80}, nil
}

func (m *Stop) Unpack(data []byte) error {
	return nil
}

func (m *Stop) EncapType() uint16 {
	return LAN_X
}

// ---------- Correlatable interface ----------

func (m *Stop) Key() (string, bool) {
	d := []byte{byte(LAN_X), byte(LAN_X_BC_STOPPED)}
	f, err := fingerprint(d)
	if err != nil {
		return "", false
	}
	return f, true
}
