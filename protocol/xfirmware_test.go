package protocol

import "testing"

func TestGetXFirmwareWireFormat(t *testing.T) {
	msg := GetXFirmware()
	if msg.Header != HeaderLANX {
		t.Fatalf("Header = %#x, want %#x", msg.Header, HeaderLANX)
	}
	want := []byte{0xF1, 0x0A, 0xFB}
	if string(msg.Data) != string(want) {
		t.Fatalf("Data = % x, want % x", msg.Data, want)
	}
}

func TestParseXFirmware(t *testing.T) {
	fw, err := ParseXFirmware([]byte{0xF3, 0x0A, 0x01, 0x23, 0xDB})
	if err != nil {
		t.Fatalf("ParseXFirmware() error = %v", err)
	}
	if got := FormatXFirmwareVersion(fw); got != "1.23" {
		t.Fatalf("FormatXFirmwareVersion() = %q, want 1.23", got)
	}
}

func TestXFirmwareFromMessages(t *testing.T) {
	msgs := []Message{{
		Header: HeaderLANX,
		Data:   []byte{0xF3, 0x0A, 0x01, 0x20, 0xD8},
	}}

	fw, err := XFirmwareFromMessages(msgs)
	if err != nil {
		t.Fatalf("XFirmwareFromMessages() error = %v", err)
	}
	if got := FormatXFirmwareVersion(fw); got != "1.20" {
		t.Fatalf("FormatXFirmwareVersion() = %q, want 1.20", got)
	}
}
