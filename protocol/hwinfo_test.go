package protocol

import (
	"testing"
)

func TestParseHWInfo(t *testing.T) {
	t.Parallel()

	info, err := ParseHWInfo([]byte{0x00, 0x02, 0x00, 0x00, 0x20, 0x01, 0x00, 0x00})
	if err != nil {
		t.Fatalf("ParseHWInfo() error = %v", err)
	}
	if info.HwType != 0x00000200 {
		t.Fatalf("HwType = %#x, want 0x00000200", info.HwType)
	}
	if got := FormatHwType(info.HwType); got != "black Z21 (2012)" {
		t.Fatalf("FormatHwType() = %q, want black Z21 (2012)", got)
	}
	if got := FormatFirmwareVersion(info.FirmwareVersion); got != "1.20" {
		t.Fatalf("FormatFirmwareVersion() = %q, want 1.20", got)
	}
}

func TestFormatHwTypeKnownValues(t *testing.T) {
	t.Parallel()

	if got := FormatHwType(HwTypeZ21New); got != "black Z21 (2013)" {
		t.Fatalf("FormatHwType() = %q", got)
	}
	if got := FormatHwType(HwTypeZ21Start); got != "z21 start (2016)" {
		t.Fatalf("FormatHwType() = %q", got)
	}
}

func TestFormatFirmwareVersion143(t *testing.T) {
	t.Parallel()

	if got := FormatFirmwareVersion(0x00000143); got != "1.43" {
		t.Fatalf("FormatFirmwareVersion() = %q, want 1.43", got)
	}
}

func TestSerialFromMessages(t *testing.T) {
	t.Parallel()

	msgs := []Message{{
		Header: HeaderLANGetSerialNumber,
		Data:   []byte{0xa3, 0xcf, 0x01, 0x00},
	}}

	got, ok := SerialFromMessages(msgs)
	if !ok {
		t.Fatal("SerialFromMessages() ok = false")
	}
	if got != "118691" {
		t.Fatalf("SerialFromMessages() = %q, want 118691", got)
	}
}
