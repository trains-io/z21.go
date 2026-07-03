package protocol

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

// Z21 hardware type IDs from LAN_GET_HWINFO (spec §2.20).
const (
	HwTypeZ21Old           uint32 = 0x00000200
	HwTypeZ21New           uint32 = 0x00000201
	HwTypeSmartRail        uint32 = 0x00000202
	HwTypeZ21Small         uint32 = 0x00000203
	HwTypeZ21Start         uint32 = 0x00000204
	HwTypeSingleBooster    uint32 = 0x00000205
	HwTypeDualBooster      uint32 = 0x00000206
	HwTypeZ21XL            uint32 = 0x00000211
	HwTypeXLBooster        uint32 = 0x00000212
	HwTypeZ21SwitchDecoder uint32 = 0x00000301
	HwTypeZ21SignalDecoder uint32 = 0x00000302
)

// HWInfo is parsed from a LAN_GET_HWINFO reply.
type HWInfo struct {
	HwType          uint32
	FirmwareVersion uint32
}

// GetSerialNumber returns a request for the Z21 serial number (spec §2.1).
func GetSerialNumber() Message {
	return Message{Header: HeaderLANGetSerialNumber}
}

// HWInfoFromMessages extracts hardware info from a Call reply.
func HWInfoFromMessages(msgs []Message) (HWInfo, error) {
	for _, msg := range msgs {
		if msg.Header != HeaderLANGetHWInfo {
			continue
		}
		return ParseHWInfo(msg.Data)
	}
	return HWInfo{}, fmt.Errorf("z21: no LAN_GET_HWINFO reply")
}

// SerialFromMessages extracts the serial number from a Call reply.
func SerialFromMessages(msgs []Message) (string, bool) {
	for _, msg := range msgs {
		if msg.Header != HeaderLANGetSerialNumber {
			continue
		}
		serial, err := ParseSerialNumber(msg.Data)
		if err != nil {
			return "", false
		}
		return FormatSerialNumber(serial), true
	}
	return "", false
}

// ParseHWInfo decodes HwType and firmware version from reply data.
func ParseHWInfo(data []byte) (HWInfo, error) {
	if len(data) < 8 {
		return HWInfo{}, fmt.Errorf("z21: HW info too short (%d bytes)", len(data))
	}
	return HWInfo{
		HwType:          binary.LittleEndian.Uint32(data[0:4]),
		FirmwareVersion: binary.LittleEndian.Uint32(data[4:8]),
	}, nil
}

// ParseSerialNumber decodes the 32-bit little-endian serial number.
func ParseSerialNumber(data []byte) (uint32, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("z21: serial number too short (%d bytes)", len(data))
	}
	return binary.LittleEndian.Uint32(data[0:4]), nil
}

// FormatSerialNumber renders the serial for display.
func FormatSerialNumber(serial uint32) string {
	return strconv.FormatUint(uint64(serial), 10)
}

// FormatFirmwareVersion renders the BCD-encoded firmware version (spec §2.20).
func FormatFirmwareVersion(fw uint32) string {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], fw)
	major := bcdByte(b[1])
	minor := bcdByte(b[0])
	return fmt.Sprintf("%d.%02d", major, minor)
}

func bcdByte(b byte) int {
	return int((b>>4)*10 + (b & 0x0f))
}

// FormatHwType renders a human-readable hardware type name.
func FormatHwType(hwType uint32) string {
	switch hwType {
	case HwTypeZ21Old:
		return "black Z21 (2012)"
	case HwTypeZ21New:
		return "black Z21 (2013)"
	case HwTypeSmartRail:
		return "SmartRail (2012)"
	case HwTypeZ21Small:
		return "white z21 starter set (2013)"
	case HwTypeZ21Start:
		return "z21 start (2016)"
	case HwTypeSingleBooster:
		return "Z21 Single Booster (zLink)"
	case HwTypeDualBooster:
		return "Z21 Dual Booster (zLink)"
	case HwTypeZ21XL:
		return "Z21 XL Series (2020)"
	case HwTypeXLBooster:
		return "Z21 XL Booster (2021, zLink)"
	case HwTypeZ21SwitchDecoder:
		return "Z21 SwitchDecoder (zLink)"
	case HwTypeZ21SignalDecoder:
		return "Z21 SignalDecoder (zLink)"
	default:
		return fmt.Sprintf("unknown (0x%08x)", hwType)
	}
}
