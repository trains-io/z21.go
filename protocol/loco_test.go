package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetLocoInfoWireFormat(t *testing.T) {
	msg := GetLocoInfo(3)
	require.Equal(t, []byte{0xE3, 0xF0, 0x00, 0x03, 0x10}, msg.Data)

	wire, err := msg.Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0x09, 0x00, 0x40, 0x00, 0xE3, 0xF0, 0x00, 0x03, 0x10}, wire)
}

func TestSetLocoDriveWireFormat(t *testing.T) {
	msg := SetLocoDrive(3, LocoSpeedSteps128, true, 0x05)
	require.Equal(t, []byte{0xE4, 0x13, 0x00, 0x03, 0x85, 0x71}, msg.Data)
}

func TestSetLocoFunctionWireFormat(t *testing.T) {
	msg := SetLocoFunction(3, 1, LocoFunctionOn)
	require.Equal(t, xHeaderSetLocoDrive, msg.Data[0])
	require.Equal(t, xCommandSetLocoFunction, msg.Data[1])
	require.Equal(t, byte(0x41), msg.Data[4]) // TT=01, N=1
}

func TestSetLocoFunctionGroupWireFormat(t *testing.T) {
	msg := SetLocoFunctionGroup(3, LocoFunctionGroupF0F4, 0x1F)
	require.Equal(t, byte(LocoFunctionGroupF0F4), msg.Data[1])
}

func TestSetLocoBinaryStateWireFormat(t *testing.T) {
	msg := SetLocoBinaryState(3, 29, true)
	require.Equal(t, []byte{0xE5, 0x5F, 0x00, 0x03}, msg.Data[:4])
}

func TestSetLocoEStopWireFormat(t *testing.T) {
	msg := SetLocoEStop(128)
	require.Equal(t, byte(0xC0), msg.Data[1])
	require.Equal(t, byte(0x80), msg.Data[2])
}

func TestPurgeLocoWireFormat(t *testing.T) {
	msg := PurgeLoco(3)
	require.Equal(t, []byte{0xE3, 0x44, 0x00, 0x03, 0xA4}, msg.Data)
}

func TestParseLocoInfo(t *testing.T) {
	data := appendLANXXOR([]byte{
		0xEF, 0x00, 0x03, 0x04, 0x85, 0x0A, 0xFF,
	})
	info, err := ParseLocoInfo(data)
	require.NoError(t, err)
	require.Equal(t, uint16(3), info.Address)
	require.Equal(t, byte(4), info.SpeedSteps)
	require.True(t, info.Forward)
	require.Equal(t, byte(5), info.Speed)
	require.False(t, info.Headlight)
	require.False(t, info.DoubleTraction)
	require.False(t, info.SmartSearch)
	require.Equal(t, byte(0x0A), info.FunctionsF1F4)
	require.Equal(t, byte(0xFF), info.FunctionsF5F12)
}

func TestParseLocoInfoDB4(t *testing.T) {
	tests := []struct {
		name           string
		db4            byte
		doubleTraction bool
		smartSearch    bool
		headlight      bool
		f1f4           byte
	}{
		{name: "none", db4: 0x00},
		{name: "F1-F4 only", db4: 0x0F, f1f4: 0x0F},
		{name: "F0 only", db4: 0x10, headlight: true},
		{name: "smart search only", db4: 0x20, smartSearch: true},
		{name: "double traction only", db4: 0x40, doubleTraction: true},
		{name: "all", db4: 0x7F, doubleTraction: true, smartSearch: true, headlight: true, f1f4: 0x0F},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := appendLANXXOR([]byte{0xEF, 0x00, 0x03, 0x04, 0x00, tt.db4})
			info, err := ParseLocoInfo(data)
			require.NoError(t, err)
			require.Equal(t, tt.doubleTraction, info.DoubleTraction, "DoubleTraction")
			require.Equal(t, tt.smartSearch, info.SmartSearch, "SmartSearch")
			require.Equal(t, tt.headlight, info.Headlight, "Headlight")
			require.Equal(t, tt.f1f4, info.FunctionsF1F4, "FunctionsF1F4")
		})
	}
}

func TestLocoInfoFromMessages(t *testing.T) {
	data := appendLANXXOR([]byte{0xEF, 0x00, 0x01, 0x00, 0x00, 0x00})
	msgs := []Message{{Header: HeaderLANX, Data: data}}
	info, err := LocoInfoFromMessages(msgs)
	require.NoError(t, err)
	require.Equal(t, uint16(1), info.Address)
}

func TestEncodeLocoAddressBytes(t *testing.T) {
	msb, lsb := encodeLocoAddressBytes(128)
	require.Equal(t, byte(0xC0), msb)
	require.Equal(t, byte(0x80), lsb)
	require.Equal(t, uint16(128), parseLocoAddressBytes(msb, lsb))
}
