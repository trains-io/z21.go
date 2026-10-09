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

func TestEncodeTrackPowerBC(t *testing.T) {
	tests := []struct {
		on   bool
		want []byte
	}{
		{on: false, want: []byte{0x61, 0x00, 0x61}},
		{on: true, want: []byte{0x61, 0x01, 0x60}},
	}
	for _, tt := range tests {
		msg := protocol.EncodeTrackPowerBC(tt.on)
		if msg.Header != protocol.HeaderLANX {
			t.Fatalf("EncodeTrackPowerBC(%t).Header = %#x, want %#x", tt.on, msg.Header, protocol.HeaderLANX)
		}
		if string(msg.Data) != string(tt.want) {
			t.Fatalf("EncodeTrackPowerBC(%t).Data = % x, want % x", tt.on, msg.Data, tt.want)
		}
		got, err := protocol.TrackPowerFromMessages([]protocol.Message{msg})
		if err != nil {
			t.Fatalf("TrackPowerFromMessages(EncodeTrackPowerBC(%t)) error = %v", tt.on, err)
		}
		if got != tt.on {
			t.Fatalf("TrackPowerFromMessages(EncodeTrackPowerBC(%t)) = %t", tt.on, got)
		}
	}
}
