package protocol

import "testing"

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
