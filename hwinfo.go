package z21

import (
	"fmt"
)

// LAN_GET_HWINFO
type HwInfo struct {
	Hardware        Hardware
	FirmwareVersion string
}

type Hardware struct {
	HardwareType uint32
	Name         string
}

// ---------- Message interface ----------

func (m *HwInfo) Pack() ([]byte, error) {
	return []byte{}, nil
}

func (m *HwInfo) Unpack(data []byte) error {
	var hwid uint32
	var version uint32
	if err := UnpackFields(data, &hwid, &version); err != nil {
		return err
	}
	m.decodeHardwareID(hwid)
	m.decodeFirmwareVersion(version)

	return nil
}

func (m *HwInfo) EncapType() uint16 {
	return LAN_GET_HWINFO
}

func (hw Hardware) String() string {
	name := hw.Name
	if name == "" {
		name = fmt.Sprintf("0x%02x", hw.HardwareType)
	}
	return name
}

// ---------- helpers ----------

func (m *HwInfo) decodeHardwareID(hwid uint32) {
	hw := Hardware{HardwareType: hwid}
	switch hwid {
	case 0x00000200:
		hw.Name = "black Z21 (2012)"
	case 0x00000201:
		hw.Name = "black Z21 (2013)"
	case 0x00000202:
		hw.Name = "SmartRail (2012)"
	case 0x00000203:
		hw.Name = "white Z21 Starter (2013)"
	case 0x00000204:
		hw.Name = "white Z21 Starter (2016)"
	case 0x00000205:
		hw.Name = "Z21 Single Booster"
	case 0x00000206:
		hw.Name = "Z21 Dual Booster"
	case 0x00000211:
		hw.Name = "Z21 XL Series (2020)"
	case 0x00000212:
		hw.Name = "Z21 XL Booster (2021)"
	case 0x00000301:
		hw.Name = "Z21 Switch Decoder"
	case 0x00000302:
		hw.Name = "Z21 Signal Decoder"
	default:
		hw.Name = ""
	}
	m.Hardware = hw
}

func (m *HwInfo) decodeFirmwareVersion(version uint32) {
	b3 := byte((version >> 8) & 0xFF)
	b4 := byte(version & 0xFF)

	decodeNibble := func(b byte) (int, error) {
		high, low, err := decodeBCD(b)
		if err != nil {
			return 0, err
		}
		return high*10 + low, nil
	}

	major, err := decodeNibble(b3)
	if err != nil {
		m.FirmwareVersion = ""
	}
	minor, err := decodeNibble(b4)
	if err != nil {
		m.FirmwareVersion = ""
	}
	m.FirmwareVersion = fmt.Sprintf("%d.%d", major, minor)
}

// ---------- Correlatable interface ----------

func (m *HwInfo) Key() (string, bool) {
	d := []byte{byte(LAN_GET_HWINFO)}
	f, err := fingerprint(d)
	if err != nil {
		return "", false
	}
	return f, true
}
