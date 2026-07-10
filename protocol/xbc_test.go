package protocol

import "testing"

func TestParseXBusBCMessages(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		parse   func([]byte) error
		isMatch func([]byte) bool
	}{
		{
			name:    "programming mode",
			data:    []byte{0x61, 0x02, 0x63},
			parse:   ParseBCProgrammingMode,
			isMatch: IsBCProgrammingMode,
		},
		{
			name:    "short circuit",
			data:    []byte{0x61, 0x08, 0x69},
			parse:   ParseBCShortCircuit,
			isMatch: IsBCShortCircuit,
		},
		{
			name:    "unknown command",
			data:    []byte{0x61, 0x82, 0xE3},
			parse:   ParseUnknownCommand,
			isMatch: IsUnknownCommand,
		},
		{
			name:    "track power on",
			data:    []byte{0x61, 0x01, 0x60},
			parse:   func(data []byte) error { _, err := ParseTrackPowerBC(data); return err },
			isMatch: IsTrackPowerBC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.parse(tt.data); err != nil {
				t.Fatalf("parse() error = %v", err)
			}
			if !tt.isMatch(tt.data) {
				t.Fatal("isMatch() = false, want true")
			}
			if err := tt.parse([]byte{0x61, 0xFF}); err == nil {
				t.Fatal("expected parse error for invalid DB0")
			}
		})
	}
}

func TestParseXBusBCChecksum(t *testing.T) {
	if err := ParseBCProgrammingMode([]byte{0x61, 0x02, 0x00}); err == nil {
		t.Fatal("expected checksum error")
	}
}

func TestXBusBCMessageNames(t *testing.T) {
	tests := []struct {
		data []byte
		want string
	}{
		{[]byte{0x61, 0x02, 0x63}, "LAN_X_BC_PROGRAMMING_MODE"},
		{[]byte{0x61, 0x08, 0x69}, "LAN_X_BC_TRACK_SHORT_CIRCUIT"},
		{[]byte{0x61, 0x82, 0xE3}, "LAN_X_UNKNOWN_COMMAND"},
	}

	for _, tt := range tests {
		msg := Message{Header: HeaderLANX, Data: tt.data}
		if got := MessageName(msg); got != tt.want {
			t.Fatalf("MessageName(% x) = %q, want %q", tt.data, got, tt.want)
		}
	}
}

func TestIsXStatusChanged(t *testing.T) {
	data := []byte{0x62, 0x22, 0x04, 0x40}
	if !IsXStatusChanged(data) {
		t.Fatal("IsXStatusChanged() = false, want true")
	}
}

func TestIsSystemStateDataChanged(t *testing.T) {
	msg := Message{Header: HeaderLANSystemStateDataChanged, Data: make([]byte, 16)}
	if !IsSystemStateDataChanged(msg) {
		t.Fatal("IsSystemStateDataChanged() = false, want true")
	}
	if IsSystemStateDataChanged(Message{Header: HeaderLANGetHWInfo}) {
		t.Fatal("IsSystemStateDataChanged() = true for wrong header")
	}
}

func TestFormatCentralState(t *testing.T) {
	if got := FormatCentralState(0x04); got != "short circuit" {
		t.Fatalf("FormatCentralState() = %q, want short circuit", got)
	}
}
