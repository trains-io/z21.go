package z21

// LAN_X_SET_TRACK_POWER_ON and LAN_X_SET_TRACK_POWER_OFF,
// LAN_X_BC_TRACK_POWER_ON and LAN_X_BC_TRACK_POWER_OFF
type TrackPower struct {
	On bool
}

// ---------- Message interface ----------

func (m *TrackPower) Pack() ([]byte, error) {
	if m.On {
		return []byte{0x21, 0x81, 0xa0}, nil
	}
	return []byte{0x21, 0x80, 0xa1}, nil

}

func (m *TrackPower) Unpack(data []byte) error {
	m.On = data[1] != 0x00
	return nil
}

func (m *TrackPower) EncapType() uint16 {
	return LAN_X
}

// ---------- Correlatable interface ----------

func (m *TrackPower) Key() (string, bool) {
	var d []byte
	if m.On {
		d = []byte{0x40, 0x61, 0x01}
	} else {
		d = []byte{0x40, 0x61, 0x00}
	}

	f, err := fingerprint(d)
	if err != nil {
		return "", false
	}
	return f, true

}

type TrackPowerStatus struct {
	On bool
}
