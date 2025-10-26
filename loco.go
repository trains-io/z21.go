package z21

// LAN_X_GET_LOCO_INFO
type LocoInfo struct {
	Address uint16
}

// ---------- Message interface ----------

func (m *LocoInfo) Pack() ([]byte, error) {
	return []byte{0xe3, 0xf0, 0x00, 0x01, 0x12}, nil
}

func (m *LocoInfo) Unpack(data []byte) error {
	return nil
}

func (m *LocoInfo) EncapType() uint16 {
	return LAN_X
}

// ---------- Correlatable interface ----------

func (m *LocoInfo) Key() (string, bool) {
	d := []byte{byte(LAN_X), byte(LAN_X_LOCO_INFO)}
	f, err := fingerprint(d)
	if err != nil {
		return "", false
	}
	return f, true
}
