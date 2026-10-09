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

func TestEncodeCANDeviceDescription(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		wantName string
	}{
		{name: "empty", in: "", wantName: ""},
		{name: "short", in: "Booster 1", wantName: "Booster 1"},
		{name: "max length", in: "abcdefghijklmno", wantName: "abcdefghijklmno"},
		{name: "truncated", in: "abcdefghijklmnopqrst", wantName: "abcdefghijklmno"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := EncodeCANDeviceDescription(0xC101, tt.in)
			require.Equal(t, HeaderLANCANDeviceGetDescription, msg.Header)
			require.Len(t, msg.Data, 2+CANBoosterNameLen)
			require.Equal(t, byte(0), msg.Data[len(msg.Data)-1], "name field must stay NUL-terminated")

			netID, name, err := ParseCANDeviceDescription(msg.Data)
			require.NoError(t, err)
			require.Equal(t, uint16(0xC101), netID)
			require.Equal(t, tt.wantName, name)
		})
	}
}

func TestEncodeCANBoosterSystemState(t *testing.T) {
	in := CANBoosterSystemState{
		NetID:      0xC101,
		OutputPort: 1,
		State:      CANBoosterStateTrackVoltageOff,
		VCCVoltage: 10000,
		Current:    100,
	}
	msg := EncodeCANBoosterSystemState(in)
	require.Equal(t, HeaderLANCANBoosterSystemState, msg.Header)
	require.Equal(t, []byte{0x01, 0xC1, 0x01, 0x00, 0x80, 0x00, 0x10, 0x27, 0x64, 0x00}, msg.Data)
	require.True(t, IsCANBoosterSystemStateChanged(msg))

	for _, in := range []CANBoosterSystemState{{}, in, {NetID: 0xffff, OutputPort: 2, State: 0xffff, VCCVoltage: 0xffff, Current: 0xffff}} {
		got, err := CANBoosterSystemStatesFromMessages([]Message{EncodeCANBoosterSystemState(in)})
		require.NoError(t, err)
		require.Equal(t, []CANBoosterSystemState{in}, got)
	}
}

func TestParseGetCANDeviceDescription(t *testing.T) {
	for _, in := range []uint16{0, 0xC101, 0xffff} {
		got, err := ParseGetCANDeviceDescription(GetCANDeviceDescription(in).Data)
		require.NoError(t, err)
		require.Equal(t, in, got)
	}

	_, err := ParseGetCANDeviceDescription([]byte{0x01})
	require.Error(t, err)
}

func TestParseSetCANDeviceDescription(t *testing.T) {
	for _, name := range []string{"", "Booster 1", "abcdefghijklmno"} {
		req, err := SetCANDeviceDescription(0xC101, name)
		require.NoError(t, err)

		netID, got, err := ParseSetCANDeviceDescription(req.Data)
		require.NoError(t, err)
		require.Equal(t, uint16(0xC101), netID)
		require.Equal(t, name, got)
	}

	_, _, err := ParseSetCANDeviceDescription(make([]byte, 2+CANBoosterNameLen-1))
	require.Error(t, err)
}

func TestParseSetCANBoosterTrackPower(t *testing.T) {
	powers := []CANBoosterTrackPower{
		CANBoosterTrackPowerDeactivateAll,
		CANBoosterTrackPowerActivateAll,
		CANBoosterTrackPowerDeactivateOut1,
		CANBoosterTrackPowerActivateOut1,
		CANBoosterTrackPowerDeactivateOut2,
		CANBoosterTrackPowerActivateOut2,
	}
	for _, power := range powers {
		netID, got, err := ParseSetCANBoosterTrackPower(SetCANBoosterTrackPower(0xC101, power).Data)
		require.NoError(t, err)
		require.Equal(t, uint16(0xC101), netID)
		require.Equal(t, power, got)
	}

	_, _, err := ParseSetCANBoosterTrackPower([]byte{0x01, 0xC1})
	require.Error(t, err)
}
