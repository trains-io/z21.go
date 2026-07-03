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
	require.True(t, info.Headlight)
	require.Equal(t, byte(0x0A), info.FunctionsF1F4)
	require.Equal(t, byte(0xFF), info.FunctionsF5F12)
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
