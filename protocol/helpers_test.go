package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsHelpers(t *testing.T) {
	locoData := appendLANXXOR([]byte{0xEF, 0x00, 0x01, 0x00, 0x00, 0x00})
	require.True(t, IsLocoInfo(locoData))
	require.False(t, IsLocoInfo([]byte{0xEF}))

	turnoutData := appendLANXXOR([]byte{0x43, 0x00, 0x01, 0x01})
	require.True(t, IsTurnoutInfo(turnoutData))
	require.False(t, IsTurnoutInfo([]byte{0x43}))

	extData := appendLANXXOR([]byte{0x44, 0x00, 0x01, 0x00, 0x00})
	require.True(t, IsExtAccessoryInfo(extData))

	cvResult := appendLANXXOR([]byte{0x64, 0x14, 0x00, 0x00, 0x05})
	require.True(t, IsCVResult(cvResult))
	require.False(t, IsCVResult([]byte{0x64, 0x14}))

	require.True(t, IsCVNack([]byte{0x61, 0x13, 0x72}))
	require.True(t, IsCVNackSC([]byte{0x61, 0x12, 0x73}))
	require.True(t, IsBCStopped([]byte{0x81, 0x00, 0x81}))
	require.True(t, IsRMBusDataChanged(Message{Header: HeaderLANRMBusDataChanged}))
	require.True(t, IsRailComDataChanged(Message{Header: HeaderLANRailComDataChanged}))
	require.True(t, IsCANBoosterSystemStateChanged(Message{Header: HeaderLANCANBoosterSystemState}))
	require.True(t, IsCANDetectorReport(Message{Header: HeaderLANCANDetector, Data: make([]byte, 10)}))
}

func TestFromMessagesExtractors(t *testing.T) {
	hwData := []byte{0x01, 0x02, 0x00, 0x00, 0x43, 0x01, 0x00, 0x00}
	hw, err := HWInfoFromMessages([]Message{{Header: HeaderLANGetHWInfo, Data: hwData}})
	require.NoError(t, err)
	require.Equal(t, uint32(HwTypeZ21New), hw.HwType)

	turnoutData := appendLANXXOR([]byte{0x43, 0x00, 0x05, 0x02})
	info, err := TurnoutInfoFromMessages([]Message{{Header: HeaderLANX, Data: turnoutData}})
	require.NoError(t, err)
	require.Equal(t, uint16(5), info.Address)
	require.Equal(t, TurnoutPositionOutput2, info.Position)

	extData := appendLANXXOR([]byte{0x44, 0x00, 0x03, 0x05, ExtAccessoryStatusValid})
	ext, err := ExtAccessoryInfoFromMessages([]Message{{Header: HeaderLANX, Data: extData}})
	require.NoError(t, err)
	require.Equal(t, uint16(3), ext.Address)
	require.Equal(t, byte(5), ext.Value)
}

func TestFormatBroadcastFlags(t *testing.T) {
	require.Equal(t, "0x00000101", FormatBroadcastFlags(DefaultBroadcastFlags))
}

func TestLogoffWireFormat(t *testing.T) {
	msg := Logoff()
	require.Equal(t, HeaderLANLogoff, msg.Header)
	wire, err := msg.Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0x04, 0x00, 0x30, 0x00}, wire)
}

func TestFormatCentralStateExFlags(t *testing.T) {
	flags := FormatCentralStateExFlags(0x01)
	require.NotEmpty(t, flags)
}
