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

func TestEncodeBroadcastFlagsWire(t *testing.T) {
	got, err := EncodeBroadcastFlags(DefaultBroadcastFlags | BroadcastFlagAllLocoInfo).Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	want := []byte{0x08, 0x00, 0x51, 0x00, 0x01, 0x01, 0x01, 0x00}
	if string(got) != string(want) {
		t.Fatalf("EncodeBroadcastFlags(0x00010101) = % x, want % x", got, want)
	}
}

func TestEncodeBroadcastFlagsRoundTrip(t *testing.T) {
	for _, in := range []uint32{0, DefaultBroadcastFlags, BroadcastFlagAllLocoInfo, 0xffffffff} {
		got, err := BroadcastFlagsFromMessages([]Message{EncodeBroadcastFlags(in)})
		if err != nil {
			t.Fatalf("BroadcastFlagsFromMessages(EncodeBroadcastFlags(%#08x)) error = %v", in, err)
		}
		if got != in {
			t.Fatalf("BroadcastFlagsFromMessages(EncodeBroadcastFlags(%#08x)) = %#08x", in, got)
		}
	}
}

func TestParseSetBroadcastFlags(t *testing.T) {
	for _, in := range []uint32{0, DefaultBroadcastFlags, BroadcastFlagAllLocoInfo, 0xffffffff} {
		req := SetBroadcastFlags(in)
		if req.Header != HeaderLANSetBroadcastFlags {
			t.Fatalf("SetBroadcastFlags(%#08x).Header = %#x, want %#x", in, req.Header, HeaderLANSetBroadcastFlags)
		}
		got, err := ParseSetBroadcastFlags(req.Data)
		if err != nil {
			t.Fatalf("ParseSetBroadcastFlags(SetBroadcastFlags(%#08x)) error = %v", in, err)
		}
		if got != in {
			t.Fatalf("ParseSetBroadcastFlags(SetBroadcastFlags(%#08x)) = %#08x", in, got)
		}
	}
}

func TestParseSetBroadcastFlagsTooShort(t *testing.T) {
	for _, data := range [][]byte{nil, {0x01, 0x01, 0x00}} {
		if _, err := ParseSetBroadcastFlags(data); err == nil {
			t.Fatalf("ParseSetBroadcastFlags(% x) error = nil, want error", data)
		}
	}
}
