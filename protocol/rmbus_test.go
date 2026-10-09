package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRMBusDataWireFormat(t *testing.T) {
	msg := GetRMBusData(1)
	require.Equal(t, HeaderLANRMBusGetData, msg.Header)
	require.Equal(t, []byte{0x01}, msg.Data)

	wire, err := msg.Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0x05, 0x00, 0x81, 0x00, 0x01}, wire)
}

func TestProgramRMBusModule(t *testing.T) {
	msg, err := ProgramRMBusModule(12)
	require.NoError(t, err)
	require.Equal(t, HeaderLANRMBusProgramModule, msg.Header)
	require.Equal(t, []byte{0x0C}, msg.Data)

	end, err := ProgramRMBusModule(RMBusProgramEndAddress)
	require.NoError(t, err)
	require.Equal(t, byte(0x00), end.Data[0])

	_, err = ProgramRMBusModule(21)
	require.Error(t, err)
}

func TestParseRMBusStatusSpecExample(t *testing.T) {
	data := []byte{
		0x01,
		0x01, 0x00, 0xC5, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	status, err := ParseRMBusStatus(data)
	require.NoError(t, err)
	require.Equal(t, uint8(1), status.GroupIndex)

	addr11, err := RMBusModuleAddress(status.GroupIndex, 0)
	require.NoError(t, err)
	require.Equal(t, uint8(11), addr11)
	require.Equal(t, []uint8{1}, RMBusActiveInputs(status.Status[0]))

	addr13, err := RMBusModuleAddress(status.GroupIndex, 2)
	require.NoError(t, err)
	require.Equal(t, uint8(13), addr13)
	require.Equal(t, []uint8{1, 3, 7, 8}, RMBusActiveInputs(status.Status[2]))
}

func TestRMBusStatusFromMessages(t *testing.T) {
	msgs := []Message{{
		Header: HeaderLANRMBusDataChanged,
		Data:   append([]byte{0x00}, make([]byte, RMBusInputsPerGroup)...),
	}}
	statuses, err := RMBusStatusFromMessages(msgs)
	require.NoError(t, err)
	require.Len(t, statuses, 1)
}

func TestEncodeRMBusStatusWire(t *testing.T) {
	got, err := EncodeRMBusStatus(RMBusStatus{
		GroupIndex: 1,
		Status:     [RMBusInputsPerGroup]byte{0x01, 0x00, 0xC5},
	}).Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{
		0x0F, 0x00, 0x80, 0x00, // DataLen 15, LAN_RMBUS_DATACHANGED
		0x01, // group index
		0x01, 0x00, 0xC5, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}, got)
}

func TestEncodeRMBusStatusRoundTrip(t *testing.T) {
	full := RMBusStatus{GroupIndex: 1}
	for i := range full.Status {
		full.Status[i] = 0xFF
	}
	for _, in := range []RMBusStatus{{}, full} {
		msg := EncodeRMBusStatus(in)
		require.True(t, IsRMBusDataChanged(msg))
		got, err := RMBusStatusFromMessages([]Message{msg})
		require.NoError(t, err)
		require.Equal(t, []RMBusStatus{in}, got)
	}
}

func TestParseGetRMBusData(t *testing.T) {
	for _, group := range []uint8{0, 1} {
		got, err := ParseGetRMBusData(GetRMBusData(group).Data)
		require.NoError(t, err)
		require.Equal(t, group, got)
	}
	_, err := ParseGetRMBusData(nil)
	require.Error(t, err)
}

func TestParseProgramRMBusModule(t *testing.T) {
	for _, addr := range []byte{RMBusProgramEndAddress, 1, 20} {
		req, err := ProgramRMBusModule(addr)
		require.NoError(t, err)
		got, err := ParseProgramRMBusModule(req.Data)
		require.NoError(t, err)
		require.Equal(t, addr, got)
	}
	_, err := ParseProgramRMBusModule([]byte{21})
	require.Error(t, err)
	_, err = ParseProgramRMBusModule(nil)
	require.Error(t, err)
}
