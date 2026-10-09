package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeLANX(t *testing.T) {
	tests := []struct {
		name    string
		xHeader byte
		data    []byte
		want    []byte
	}{
		{name: "LAN_X_GET_VERSION", xHeader: 0x21, data: []byte{0x21}, want: []byte{0x21, 0x21, 0x00}},
		{name: "LAN_X_BC_TRACK_POWER_ON", xHeader: 0x61, data: []byte{0x01}, want: []byte{0x61, 0x01, 0x60}},
		{name: "LAN_X_GET_LOCO_INFO address 3", xHeader: 0xE3, data: []byte{0xF0, 0x00, 0x03}, want: []byte{0xE3, 0xF0, 0x00, 0x03, 0x10}},
		{name: "no data bytes", xHeader: 0x81, want: []byte{0x81, 0x81}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := EncodeLANX(tt.xHeader, tt.data...)
			require.Equal(t, HeaderLANX, got.Header)
			require.Equal(t, tt.want, got.Data)
		})
	}
}

func TestEncodeLANXMatchesBuilders(t *testing.T) {
	tests := []struct {
		name string
		got  Message
		want Message
	}{
		{name: "GetXVersion", got: EncodeLANX(xHeaderGetVersion, xHeaderGetVersion), want: GetXVersion()},
		{name: "SetTrackPower on", got: EncodeLANX(xHeaderXCommand, xCommandSetTrackPowerOn), want: SetTrackPower(true)},
		{name: "SetTrackPower off", got: EncodeLANX(xHeaderXCommand, xCommandSetTrackPowerOff), want: SetTrackPower(false)},
		{name: "GetLocoInfo", got: EncodeLANX(xHeaderGetLocoInfo, xCommandGetLocoInfo, 0xC0, 0x80), want: GetLocoInfo(128)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, tt.got)
		})
	}
}

func TestParseLANXRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		xHeader byte
		data    []byte
	}{
		{name: "no data bytes", xHeader: 0x81, data: []byte{}},
		{name: "one data byte", xHeader: 0x61, data: []byte{0x01}},
		{name: "loco info", xHeader: 0xEF, data: []byte{0x00, 0x03, 0x04, 0x80, 0x00, 0x00, 0x00, 0x00}},
		{name: "all bits set", xHeader: 0xFF, data: []byte{0xFF, 0xFF}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			xHeader, data, err := ParseLANX(EncodeLANX(tt.xHeader, tt.data...))
			require.NoError(t, err)
			require.Equal(t, tt.xHeader, xHeader)
			require.Equal(t, tt.data, data)
		})
	}
}

func TestParseLANXErrors(t *testing.T) {
	tests := []struct {
		name string
		msg  Message
	}{
		{name: "wrong header", msg: Message{Header: HeaderLANGetHWInfo, Data: []byte{0x21, 0x21, 0x00}}},
		{name: "empty", msg: Message{Header: HeaderLANX}},
		{name: "only X-header", msg: Message{Header: HeaderLANX, Data: []byte{0x21}}},
		{name: "bad checksum", msg: Message{Header: HeaderLANX, Data: []byte{0x61, 0x01, 0x61}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, _, err := ParseLANX(tt.msg)
			require.Error(t, err)
		})
	}
}

func TestParseLANXDoesNotAlias(t *testing.T) {
	msg := EncodeLANX(0x61, 0x01)
	_, data, err := ParseLANX(msg)
	require.NoError(t, err)

	data[0] = 0xAA
	require.Equal(t, []byte{0x61, 0x01, 0x60}, msg.Data)
}
