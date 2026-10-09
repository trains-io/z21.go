package protocol

import (
	"bytes"
	"testing"
)

func TestGetCodeWireFormat(t *testing.T) {
	msg := GetCode()
	if msg.Header != HeaderLANGetCode {
		t.Fatalf("Header = %#x, want %#x", msg.Header, HeaderLANGetCode)
	}
	if len(msg.Data) != 0 {
		t.Fatalf("Data = % x, want empty", msg.Data)
	}
}

func TestParseCode(t *testing.T) {
	code, err := ParseCode([]byte{CodeNoLock})
	if err != nil {
		t.Fatalf("ParseCode() error = %v", err)
	}
	if got := FormatLockCode(code); got != "all features permitted" {
		t.Fatalf("FormatLockCode() = %q", got)
	}
}

func TestCodeFromMessages(t *testing.T) {
	msgs := []Message{{
		Header: HeaderLANGetCode,
		Data:   []byte{CodeZ21StartUnlocked},
	}}

	code, err := CodeFromMessages(msgs)
	if err != nil {
		t.Fatalf("CodeFromMessages() error = %v", err)
	}
	if got := FormatLockCode(code); got != "z21 start unlocked (driving and switching permitted)" {
		t.Fatalf("FormatLockCode() = %q", got)
	}
}

func TestEncodeCodeWire(t *testing.T) {
	got, err := EncodeCode(CodeZ21StartLocked).Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	want := []byte{0x05, 0x00, 0x18, 0x00, 0x01}
	if !bytes.Equal(got, want) {
		t.Fatalf("EncodeCode(CodeZ21StartLocked) = % x, want % x", got, want)
	}
}

func TestEncodeCodeRoundTrip(t *testing.T) {
	for _, in := range []byte{CodeNoLock, CodeZ21StartLocked, CodeZ21StartUnlocked, 0xff} {
		got, err := CodeFromMessages([]Message{EncodeCode(in)})
		if err != nil {
			t.Fatalf("CodeFromMessages(EncodeCode(%#02x)) error = %v", in, err)
		}
		if got != in {
			t.Fatalf("CodeFromMessages(EncodeCode(%#02x)) = %#02x", in, got)
		}
	}
}
