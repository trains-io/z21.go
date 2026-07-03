package protocol

import "testing"

func TestGetXVersionWireFormat(t *testing.T) {
	msg := GetXVersion()
	if msg.Header != HeaderLANX {
		t.Fatalf("Header = %#x, want %#x", msg.Header, HeaderLANX)
	}
	want := []byte{0x21, 0x21, 0x00}
	if string(msg.Data) != string(want) {
		t.Fatalf("Data = % x, want % x", msg.Data, want)
	}
}

func TestParseXVersion(t *testing.T) {
	info, err := ParseXVersion([]byte{0x63, 0x21, 0x30, 0x12, 0x60})
	if err != nil {
		t.Fatalf("ParseXVersion() error = %v", err)
	}
	if got := FormatXBusVersion(info.XBusVersion); got != "3.0" {
		t.Fatalf("FormatXBusVersion() = %q, want 3.0", got)
	}
	if got := FormatCommandStationID(info.CommandStationID); got != "Z21" {
		t.Fatalf("FormatCommandStationID() = %q, want Z21", got)
	}
}

func TestXVersionFromMessages(t *testing.T) {
	msgs := []Message{{
		Header: HeaderLANX,
		Data:   []byte{0x63, 0x21, 0x36, 0x12, 0x60},
	}}

	info, err := XVersionFromMessages(msgs)
	if err != nil {
		t.Fatalf("XVersionFromMessages() error = %v", err)
	}
	if got := FormatXBusVersion(info.XBusVersion); got != "3.6" {
		t.Fatalf("FormatXBusVersion() = %q, want 3.6", got)
	}
}
