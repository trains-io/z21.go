package protocol

import "testing"

func TestGetBroadcastFlagsWireFormat(t *testing.T) {
	msg := GetBroadcastFlags()
	if msg.Header != HeaderLANGetBroadcastFlags {
		t.Fatalf("Header = %#x, want %#x", msg.Header, HeaderLANGetBroadcastFlags)
	}
	if len(msg.Data) != 0 {
		t.Fatalf("Data = % x, want empty", msg.Data)
	}
}

func TestParseBroadcastFlags(t *testing.T) {
	flags, err := ParseBroadcastFlags([]byte{0x01, 0x01, 0x00, 0x00})
	if err != nil {
		t.Fatalf("ParseBroadcastFlags() error = %v", err)
	}
	if flags != DefaultBroadcastFlags {
		t.Fatalf("flags = %#x, want %#x", flags, DefaultBroadcastFlags)
	}
}

func TestBroadcastFlagsFromMessages(t *testing.T) {
	msgs := []Message{{
		Header: HeaderLANGetBroadcastFlags,
		Data:   []byte{0x03, 0x01, 0x00, 0x00},
	}}

	flags, err := BroadcastFlagsFromMessages(msgs)
	if err != nil {
		t.Fatalf("BroadcastFlagsFromMessages() error = %v", err)
	}
	if flags != 0x103 {
		t.Fatalf("flags = %#x, want 0x103", flags)
	}
}

func TestDefaultBroadcastFlags(t *testing.T) {
	if !HasBroadcastFlag(DefaultBroadcastFlags, BroadcastFlagXpressNet) ||
		!HasBroadcastFlag(DefaultBroadcastFlags, BroadcastFlagSystemState) {
		t.Fatalf("default flags = %#x", DefaultBroadcastFlags)
	}
}
