package protocol

import (
	"bytes"
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

func TestEncodeHWInfoWire(t *testing.T) {
	t.Parallel()

	got, err := EncodeHWInfo(HWInfo{HwType: HwTypeZ21Old, FirmwareVersion: 0x00000120}).Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	want := []byte{0x0c, 0x00, 0x1a, 0x00, 0x00, 0x02, 0x00, 0x00, 0x20, 0x01, 0x00, 0x00}
	if !bytes.Equal(got, want) {
		t.Fatalf("EncodeHWInfo(Z21 old, 1.20) = % x, want % x", got, want)
	}
}

func TestEncodeHWInfoRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []HWInfo{
		{},
		{HwType: HwTypeZ21New, FirmwareVersion: 0x00000143},
		{HwType: HwTypeZ21XL, FirmwareVersion: 0xffffffff},
	}
	for _, in := range tests {
		msgs := []Message{EncodeHWInfo(in)}
		got, err := HWInfoFromMessages(msgs)
		if err != nil {
			t.Fatalf("HWInfoFromMessages(EncodeHWInfo(%+v)) error = %v", in, err)
		}
		if got != in {
			t.Fatalf("HWInfoFromMessages(EncodeHWInfo(%+v)) = %+v", in, got)
		}
	}
}

func TestEncodeSerialNumber(t *testing.T) {
	t.Parallel()

	msg := EncodeSerialNumber(118691)
	if msg.Header != HeaderLANGetSerialNumber {
		t.Fatalf("EncodeSerialNumber().Header = %#04x, want %#04x", msg.Header, HeaderLANGetSerialNumber)
	}
	want := []byte{0xa3, 0xcf, 0x01, 0x00}
	if !bytes.Equal(msg.Data, want) {
		t.Fatalf("EncodeSerialNumber(118691).Data = % x, want % x", msg.Data, want)
	}

	for _, in := range []uint32{0, 118691, 0xffffffff} {
		got, err := ParseSerialNumber(EncodeSerialNumber(in).Data)
		if err != nil {
			t.Fatalf("ParseSerialNumber(EncodeSerialNumber(%d)) error = %v", in, err)
		}
		if got != in {
			t.Fatalf("ParseSerialNumber(EncodeSerialNumber(%d)) = %d", in, got)
		}
	}
}
