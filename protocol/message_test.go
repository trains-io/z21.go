package protocol_test

import (
	"bytes"
	"testing"

	"github.com/trains-io/z21.go/protocol"
)

func TestMarshalKnownRequests(t *testing.T) {
	tests := []struct {
		name string
		msg  protocol.Message
		want []byte
	}{
		{
			name: "LAN_SYSTEMSTATE_GETDATA",
			msg:  protocol.SystemStateGetData(),
			want: []byte{0x04, 0x00, 0x85, 0x00},
		},
		{
			name: "LAN_GET_HWINFO",
			msg:  protocol.GetHWInfo(),
			want: []byte{0x04, 0x00, 0x1A, 0x00},
		},
		{
			name: "LAN_GET_BROADCASTFLAGS",
			msg:  protocol.GetBroadcastFlags(),
			want: []byte{0x04, 0x00, 0x51, 0x00},
		},
		{
			name: "LAN_SET_BROADCASTFLAGS",
			msg:  protocol.SetBroadcastFlags(0x101),
			want: []byte{0x08, 0x00, 0x50, 0x00, 0x01, 0x01, 0x00, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.msg.Marshal()
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("Marshal() = % X, want % X", got, tt.want)
			}
		})
	}
}

func TestUnmarshalRoundTrip(t *testing.T) {
	original := protocol.SystemStateGetData()
	encoded, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	msg, rest, err := protocol.Unmarshal(encoded)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(rest) != 0 {
		t.Fatalf("Unmarshal() rest = % X, want empty", rest)
	}
	if msg.Header != original.Header {
		t.Fatalf("Header = %04X, want %04X", msg.Header, original.Header)
	}
	if len(msg.Data) != 0 {
		t.Fatalf("Data = % X, want empty", msg.Data)
	}
}

func TestParseCombinedUDP(t *testing.T) {
	// Spec §1.3: multiple independent datasets in one UDP packet.
	a, err := protocol.GetHWInfo().Marshal()
	if err != nil {
		t.Fatalf("Marshal HWInfo: %v", err)
	}
	b, err := protocol.SystemStateGetData().Marshal()
	if err != nil {
		t.Fatalf("Marshal SystemState: %v", err)
	}
	payload := append(a, b...)

	msgs, err := protocol.ParseAll(payload)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("ParseAll() len = %d, want 2", len(msgs))
	}
	if msgs[0].Header != protocol.HeaderLANGetHWInfo {
		t.Fatalf("msgs[0].Header = %04X", msgs[0].Header)
	}
	if msgs[1].Header != protocol.HeaderLANSystemStateGetData {
		t.Fatalf("msgs[1].Header = %04X", msgs[1].Header)
	}
}

func TestMarshalAllRoundTrip(t *testing.T) {
	reqs := []protocol.Message{
		protocol.GetSerialNumber(),
		protocol.GetHWInfo(),
		protocol.GetXVersion(),
		protocol.GetXFirmware(),
		protocol.GetCode(),
	}

	payload, err := protocol.MarshalAll(reqs...)
	if err != nil {
		t.Fatalf("MarshalAll() error = %v", err)
	}

	msgs, err := protocol.ParseAll(payload)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(msgs) != len(reqs) {
		t.Fatalf("ParseAll() len = %d, want %d", len(msgs), len(reqs))
	}
	for i, req := range reqs {
		if msgs[i].Header != req.Header {
			t.Fatalf("msgs[%d].Header = %04X, want %04X", i, msgs[i].Header, req.Header)
		}
	}
}

func TestUnmarshalTruncatedDataLen(t *testing.T) {
	// Z21Posix replies to LAN_GET_CODE with DataLen=8 but only 5 bytes on the wire.
	payload := []byte{0x08, 0x00, 0x18, 0x00, 0x00}

	msg, rest, err := protocol.Unmarshal(payload)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(rest) != 0 {
		t.Fatalf("Unmarshal() rest = % X, want empty", rest)
	}
	if msg.Header != protocol.HeaderLANGetCode {
		t.Fatalf("Header = %04X, want %04X", msg.Header, protocol.HeaderLANGetCode)
	}
	if len(msg.Data) != 1 || msg.Data[0] != 0x00 {
		t.Fatalf("Data = % x, want [00]", msg.Data)
	}
}
