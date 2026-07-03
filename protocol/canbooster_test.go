package protocol

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetCANDeviceDescriptionWireFormat(t *testing.T) {
	msg := GetCANDeviceDescription(0xC101)
	require.Equal(t, HeaderLANCANDeviceGetDescription, msg.Header)
	require.Equal(t, []byte{0x01, 0xC1}, msg.Data)

	wire, err := msg.Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0x06, 0x00, 0xC8, 0x00, 0x01, 0xC1}, wire)
}

func TestSetCANDeviceDescription(t *testing.T) {
	msg, err := SetCANDeviceDescription(0xC101, "Booster 1")
	require.NoError(t, err)
	require.Equal(t, HeaderLANCANDeviceSetDescription, msg.Header)
	require.Equal(t, "Booster 1\x00", string(msg.Data[2:2+len("Booster 1\x00")]))

	_, err = SetCANDeviceDescription(0xC101, `bad"name`)
	require.Error(t, err)
}

func TestParseCANDeviceDescription(t *testing.T) {
	data := make([]byte, 2+CANBoosterNameLen)
	binary.LittleEndian.PutUint16(data, 0xC102)
	copy(data[2:], "Track A\x00")

	netID, name, err := ParseCANDeviceDescription(data)
	require.NoError(t, err)
	require.Equal(t, uint16(0xC102), netID)
	require.Equal(t, "Track A", name)
}

func TestSetCANBoosterTrackPowerWireFormat(t *testing.T) {
	msg := SetCANBoosterTrackPower(0xC101, CANBoosterTrackPowerActivateAll)
	require.Equal(t, HeaderLANCANBoosterSetTrackPower, msg.Header)
	require.Equal(t, []byte{0x01, 0xC1, 0xFF}, msg.Data)
}

func TestParseCANBoosterSystemState(t *testing.T) {
	data := []byte{
		0x01, 0xC1,
		0x01, 0x00,
		0x81, 0x08,
		0x10, 0x27,
		0x64, 0x00,
	}
	state, err := ParseCANBoosterSystemState(data)
	require.NoError(t, err)
	require.Equal(t, uint16(0xC101), state.NetID)
	require.Equal(t, uint16(1), state.OutputPort)
	require.Equal(t, uint16(0x0881), state.State)
	require.Equal(t, uint16(10000), state.VCCVoltage)
	require.Equal(t, uint16(100), state.Current)
	require.Equal(t, "BrakeGen|TrackOff|RailCom", FormatCANBoosterState(state.State))
}

func TestCANBoosterSystemStatesFromMessages(t *testing.T) {
	msgs := []Message{{Header: HeaderLANCANBoosterSystemState, Data: make([]byte, canBoosterSystemStateLen)}}
	out, err := CANBoosterSystemStatesFromMessages(msgs)
	require.NoError(t, err)
	require.Len(t, out, 1)
}

func TestParseCANDetectorLocoAddress(t *testing.T) {
	forward := ParseCANDetectorLocoAddress(0x4003)
	require.Equal(t, uint16(3), forward.Address)
	require.Equal(t, CANDetectorLocoDirectionForward, forward.Direction)

	backward := ParseCANDetectorLocoAddress(0x800A)
	require.Equal(t, uint16(10), backward.Address)
	require.Equal(t, CANDetectorLocoDirectionBackward, backward.Direction)

	require.True(t, IsCANDetectorLocoAddressType(0x11))
	require.False(t, IsCANDetectorLocoAddressType(0x01))
}

func TestCANDetectorReportLocoAddresses(t *testing.T) {
	report := CANDetectorReport{
		Type:   0x11,
		Value1: 0x4005,
		Value2: 0,
	}
	addrs := report.LocoAddresses()
	require.Equal(t, uint16(5), addrs[0].Address)
	require.Equal(t, CANDetectorLocoDirectionForward, addrs[0].Direction)
	require.Equal(t, uint16(0), addrs[1].Address)
}
