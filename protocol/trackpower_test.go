package protocol_test

import (
	"testing"

	"github.com/trains-io/z21.go/protocol"
)

func TestSetTrackPowerOn(t *testing.T) {
	msg := protocol.SetTrackPower(true)
	if msg.Header != protocol.HeaderLANX {
		t.Fatalf("Header = %#x, want %#x", msg.Header, protocol.HeaderLANX)
	}
	want := []byte{0x21, 0x81, 0xa0}
	if string(msg.Data) != string(want) {
		t.Fatalf("Data = %v, want %v", msg.Data, want)
	}
}

func TestSetTrackPowerOff(t *testing.T) {
	msg := protocol.SetTrackPower(false)
	want := []byte{0x21, 0x80, 0xa1}
	if string(msg.Data) != string(want) {
		t.Fatalf("Data = %v, want %v", msg.Data, want)
	}
}

func TestTrackPowerFromMessages(t *testing.T) {
	on, err := protocol.TrackPowerFromMessages([]protocol.Message{{
		Header: protocol.HeaderLANX,
		Data:   []byte{0x61, 0x01},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !on {
		t.Fatal("expected track power on")
	}

	off, err := protocol.TrackPowerFromMessages([]protocol.Message{{
		Header: protocol.HeaderLANX,
		Data:   []byte{0x61, 0x00},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if off {
		t.Fatal("expected track power off")
	}
}
