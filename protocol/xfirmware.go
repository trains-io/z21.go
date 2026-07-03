package protocol

import (
	"encoding/binary"
	"fmt"
)

const (
	xCommandGetFirmware      byte = 0xF1
	xFirmwareDB0             byte = 0x0A
	xHeaderFirmwareReply     byte = 0xF3
)

// XFirmware is parsed from a LAN_X_GET_FIRMWARE_VERSION reply (spec §2.15).
type XFirmware struct {
	VersionMSB byte
	VersionLSB byte
}

// GetXFirmware returns a request for the X-Bus firmware version (spec §2.15).
func GetXFirmware() Message {
	return Message{
		Header: HeaderLANX,
		Data:   []byte{xCommandGetFirmware, xFirmwareDB0, xCommandGetFirmware ^ xFirmwareDB0},
	}
}

// XFirmwareFromMessages extracts firmware version from a Call reply.
func XFirmwareFromMessages(msgs []Message) (XFirmware, error) {
	for _, msg := range msgs {
		if msg.Header != HeaderLANX {
			continue
		}
		fw, err := ParseXFirmware(msg.Data)
		if err != nil {
			continue
		}
		return fw, nil
	}
	return XFirmware{}, fmt.Errorf("z21: no LAN_X_GET_FIRMWARE_VERSION reply")
}

// ParseXFirmware decodes an X-Bus firmware version reply payload.
func ParseXFirmware(data []byte) (XFirmware, error) {
	if len(data) < 5 {
		return XFirmware{}, fmt.Errorf("z21: X firmware reply too short (%d bytes)", len(data))
	}
	if data[0] != xHeaderFirmwareReply || data[1] != xFirmwareDB0 {
		return XFirmware{}, fmt.Errorf("z21: not a LAN_X_GET_FIRMWARE_VERSION reply")
	}
	return XFirmware{
		VersionMSB: data[2],
		VersionLSB: data[3],
	}, nil
}

// FormatXFirmwareVersion renders the BCD-encoded X-Bus firmware version (spec §2.15).
func FormatXFirmwareVersion(fw XFirmware) string {
	var buf [4]byte
	buf[0] = fw.VersionLSB
	buf[1] = fw.VersionMSB
	return FormatFirmwareVersion(binary.LittleEndian.Uint32(buf[:]))
}
