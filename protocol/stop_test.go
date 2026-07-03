package protocol

import "testing"

func TestSetStopWireFormat(t *testing.T) {
	msg := SetStop()
	if msg.Header != HeaderLANX {
		t.Fatalf("Header = %#x, want %#x", msg.Header, HeaderLANX)
	}
	want := []byte{0x80, 0x80}
	if string(msg.Data) != string(want) {
		t.Fatalf("Data = % x, want % x", msg.Data, want)
	}

	wire, err := msg.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	wantWire := []byte{0x06, 0x00, 0x40, 0x00, 0x80, 0x80}
	if string(wire) != string(wantWire) {
		t.Fatalf("wire = % x, want % x", wire, wantWire)
	}
}

func TestParseBCStopped(t *testing.T) {
	if err := ParseBCStopped([]byte{0x81, 0x00, 0x81}); err != nil {
		t.Fatalf("ParseBCStopped() error = %v", err)
	}
	if err := ParseBCStopped([]byte{0x81, 0x00, 0x00}); err == nil {
		t.Fatal("expected error for invalid checksum")
	}
}

func TestBCStoppedFromMessages(t *testing.T) {
	if err := BCStoppedFromMessages([]Message{{
		Header: HeaderLANX,
		Data:   []byte{0x81, 0x00, 0x81},
	}}); err != nil {
		t.Fatalf("BCStoppedFromMessages() error = %v", err)
	}

	if err := BCStoppedFromMessages([]Message{{
		Header: HeaderLANGetHWInfo,
		Data:   []byte{1, 2, 3, 4},
	}}); err == nil {
		t.Fatal("expected error without BC_STOPPED reply")
	}
}

func TestMessageNameSetStop(t *testing.T) {
	if got := MessageName(SetStop()); got != "LAN_X_SET_STOP" {
		t.Fatalf("MessageName(SetStop()) = %q, want LAN_X_SET_STOP", got)
	}
	if got := MessageName(Message{
		Header: HeaderLANX,
		Data:   []byte{0x81, 0x00, 0x81},
	}); got != "LAN_X_BC_STOPPED" {
		t.Fatalf("MessageName(BC_STOPPED) = %q, want LAN_X_BC_STOPPED", got)
	}
}
