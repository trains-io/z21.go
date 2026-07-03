package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetLocoModeWireFormat(t *testing.T) {
	msg := GetLocoMode(3)
	require.Equal(t, HeaderLANGetLocoMode, msg.Header)
	require.Equal(t, []byte{0x00, 0x03}, msg.Data)

	wire, err := msg.Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0x06, 0x00, 0x60, 0x00, 0x00, 0x03}, wire)
}

func TestSetLocoModeWireFormat(t *testing.T) {
	msg := SetLocoMode(3, OutputModeMM)
	require.Equal(t, HeaderLANSetLocoMode, msg.Header)
	require.Equal(t, []byte{0x00, 0x03, OutputModeMM}, msg.Data)

	wire, err := msg.Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0x07, 0x00, 0x61, 0x00, 0x00, 0x03, OutputModeMM}, wire)
}

func TestParseLocoMode(t *testing.T) {
	mode, err := ParseLocoMode([]byte{0x00, 0x03, OutputModeDCC})
	require.NoError(t, err)
	require.Equal(t, AddressMode{Address: 3, Mode: OutputModeDCC}, mode)
	require.Equal(t, "DCC", FormatOutputMode(mode.Mode))
}

func TestLocoModeFromMessages(t *testing.T) {
	msgs := []Message{{
		Header: HeaderLANGetLocoMode,
		Data:   []byte{0x01, 0x00, OutputModeMM},
	}}
	mode, err := LocoModeFromMessages(msgs)
	require.NoError(t, err)
	require.Equal(t, uint16(256), mode.Address)
	require.Equal(t, OutputModeMM, mode.Mode)
}

func TestGetTurnoutModeWireFormat(t *testing.T) {
	msg := GetTurnoutMode(2)
	require.Equal(t, HeaderLANGetTurnoutMode, msg.Header)
	require.Equal(t, []byte{0x00, 0x02}, msg.Data)

	wire, err := msg.Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0x06, 0x00, 0x70, 0x00, 0x00, 0x02}, wire)
}

func TestSetTurnoutModeWireFormat(t *testing.T) {
	msg := SetTurnoutMode(2, OutputModeDCC)
	require.Equal(t, HeaderLANSetTurnoutMode, msg.Header)
	require.Equal(t, []byte{0x00, 0x02, OutputModeDCC}, msg.Data)

	wire, err := msg.Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0x07, 0x00, 0x71, 0x00, 0x00, 0x02, OutputModeDCC}, wire)
}

func TestParseTurnoutMode(t *testing.T) {
	mode, err := ParseTurnoutMode([]byte{0x00, 0x02, OutputModeMM})
	require.NoError(t, err)
	require.Equal(t, AddressMode{Address: 2, Mode: OutputModeMM}, mode)
}

func TestTurnoutModeFromMessages(t *testing.T) {
	msgs := []Message{{
		Header: HeaderLANGetTurnoutMode,
		Data:   []byte{0x00, 0x05, OutputModeDCC},
	}}
	mode, err := TurnoutModeFromMessages(msgs)
	require.NoError(t, err)
	require.Equal(t, uint16(5), mode.Address)
	require.Equal(t, OutputModeDCC, mode.Mode)
}

func TestMessageNameSettings(t *testing.T) {
	tests := []struct {
		msg  Message
		want string
	}{
		{GetLocoMode(1), "LAN_GET_LOCOMODE"},
		{SetLocoMode(1, OutputModeDCC), "LAN_SET_LOCOMODE"},
		{GetTurnoutMode(1), "LAN_GET_TURNOUTMODE"},
		{SetTurnoutMode(1, OutputModeMM), "LAN_SET_TURNOUTMODE"},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, MessageName(tt.msg))
	}
}
