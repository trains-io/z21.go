package z21

// LAN_LOGOFF
type Logoff struct{}

// ---------- Message interface ----------

func (m *Logoff) Pack() ([]byte, error) {
	return []byte{}, nil
}

func (m *Logoff) Unpack(data []byte) error {
	return nil
}

func (m *Logoff) EncapType() uint16 {
	return LAN_LOGOFF
}

// ---------- Correlatable interface ----------

func (m *Logoff) Key() (string, bool) {
	return "", false
}
