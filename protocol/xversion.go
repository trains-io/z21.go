package protocol

import "fmt"

// HeaderLANX is the dataset header for X-Bus encapsulated commands (spec §2.3+).
const HeaderLANX uint16 = 0x0040

const (
	xHeaderGetVersion      byte = 0x21
	xHeaderGetVersionReply byte = 0x63
	// LAN_X GET_VERSION uses the same byte twice; XOR checksum is always 0 (spec §2.3).
	xChecksumGetVersion byte = 0x00
	commandStationIDZ21 byte = 0x12
)

// XVersion is parsed from a LAN_X_GET_VERSION reply (spec §2.3).
type XVersion struct {
	XBusVersion      byte
	CommandStationID byte
}

// GetXVersion returns a request for the X-Bus protocol version (spec §2.3).
func GetXVersion() Message {
	return Message{
		Header: HeaderLANX,
		Data:   []byte{xHeaderGetVersion, xHeaderGetVersion, xChecksumGetVersion},
	}
}

// XVersionFromMessages extracts X-Bus version info from a Call reply.
func XVersionFromMessages(msgs []Message) (XVersion, error) {
	for _, msg := range msgs {
		if msg.Header != HeaderLANX {
			continue
		}
		info, err := ParseXVersion(msg.Data)
		if err != nil {
			continue
		}
		return info, nil
	}
	return XVersion{}, fmt.Errorf("z21: no LAN_X_GET_VERSION reply")
}

// ParseXVersion decodes an X-Bus version reply payload.
func ParseXVersion(data []byte) (XVersion, error) {
	if len(data) < 5 {
		return XVersion{}, fmt.Errorf("z21: X version reply too short (%d bytes)", len(data))
	}
	if data[0] != xHeaderGetVersionReply || data[1] != xHeaderGetVersion {
		return XVersion{}, fmt.Errorf("z21: not a LAN_X_GET_VERSION reply")
	}
	return XVersion{
		XBusVersion:      data[2],
		CommandStationID: data[3],
	}, nil
}

// FormatXBusVersion renders the BCD-encoded X-Bus version byte (spec §2.3).
func FormatXBusVersion(version byte) string {
	return fmt.Sprintf("%d.%d", (version>>4)&0x0f, version&0x0f)
}

// FormatCommandStationID renders the command station family ID.
func FormatCommandStationID(id byte) string {
	if id == commandStationIDZ21 {
		return "Z21"
	}
	return fmt.Sprintf("0x%02x", id)
}
