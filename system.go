package z21

import (
	"encoding/binary"
)

// LAN_SYSTEMSTATE_GETDATA
type SysData struct {
	MainCurrent         uint16 `json:"main_current"`
	ProgCurrent         uint16 `json:"program_current"`
	FilteredMainCurrent uint16 `json:"filtered_main_current"`
	Temperature         uint16 `json:"temperature"`
	SupplyVoltage       uint16 `json:"supply_voltage"`
	VccVoltage          uint16 `json:"vcc_voltage"`
	CentralState        Mask8  `json:"central_state"`
	CentralStateEx      Mask8  `json:"central_state_extended"`
	Reserved            uint8  `json:"reserved"`
	Capabilities        Mask8  `json:"capabilities"`
}

// ---------- Message interface ----------

func (m *SysData) String() string {
	return "system"
}

func (m *SysData) Pack() ([]byte, error) {
	return []byte{}, nil
}

func (m *SysData) Unpack(data []byte) error {
	var state = make([]byte, 16)
	if err := UnpackFields(data, &state); err != nil {
		return err
	}
	m.decodeState(state)

	return nil
}

func (m *SysData) EncapType() uint16 {
	return LAN_SYSTEMSTATE_GETDATA
}

// ---------- helpers ----------

func (m *SysData) decodeState(state []byte) {
	m.MainCurrent = binary.LittleEndian.Uint16(state[0:2])
	m.ProgCurrent = binary.LittleEndian.Uint16(state[2:4])
	m.FilteredMainCurrent = binary.LittleEndian.Uint16(state[4:6])
	m.Temperature = binary.LittleEndian.Uint16(state[6:8])
	m.SupplyVoltage = binary.LittleEndian.Uint16(state[8:10])
	m.VccVoltage = binary.LittleEndian.Uint16(state[10:12])
	m.CentralState = Mask8(state[12])
	m.CentralStateEx = Mask8(state[13])
	m.Reserved = state[14]
	m.Capabilities = Mask8(state[15])
}

// ---------- Correlatable interface ----------

func (m *SysData) Key() (string, bool) {
	d := []byte{byte(LAN_SYSTEMSTATE_DATACHANGED)}
	f, err := fingerprint(d)
	if err != nil {
		return "", false
	}
	return f, true
}
