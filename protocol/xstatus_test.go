package protocol

import "testing"

func TestGetXStatusWireFormat(t *testing.T) {
	msg := GetXStatus()
	if msg.Header != HeaderLANX {
		t.Fatalf("Header = %#x, want %#x", msg.Header, HeaderLANX)
	}
	want := []byte{0x21, 0x24, 0x05}
	if string(msg.Data) != string(want) {
		t.Fatalf("Data = % x, want % x", msg.Data, want)
	}
}

func TestParseXStatus(t *testing.T) {
	status, err := ParseXStatus([]byte{0x62, 0x22, 0x00, 0x40})
	if err != nil {
		t.Fatalf("ParseXStatus() error = %v", err)
	}
	if status.CentralState != 0x00 {
		t.Fatalf("CentralState = %#x, want 0x00", status.CentralState)
	}
	if got := FormatXStatusFlags(status.CentralState); got != "ok" {
		t.Fatalf("FormatXStatusFlags() = %q, want ok", got)
	}

	status, err = ParseXStatus([]byte{0x62, 0x22, 0x23, 0x41})
	if err != nil {
		t.Fatalf("ParseXStatus() error = %v", err)
	}
	if got := FormatXStatusFlags(status.CentralState); got != "emergency stop, track voltage off, programming mode" {
		t.Fatalf("FormatXStatusFlags() = %q", got)
	}
}

func TestXStatusFromMessages(t *testing.T) {
	msgs := []Message{{
		Header: HeaderLANX,
		Data:   []byte{0x62, 0x22, 0x02, 0x40},
	}}

	status, err := XStatusFromMessages(msgs)
	if err != nil {
		t.Fatalf("XStatusFromMessages() error = %v", err)
	}
	if got := FormatXStatusFlags(status.CentralState); got != "track voltage off" {
		t.Fatalf("FormatXStatusFlags() = %q, want track voltage off", got)
	}
}
