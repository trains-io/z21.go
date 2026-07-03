package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTurnoutInfoWireFormat(t *testing.T) {
	msg := GetTurnoutInfo(4)
	require.Equal(t, []byte{0x43, 0x00, 0x04, 0x47}, msg.Data)

	wire, err := msg.Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0x08, 0x00, 0x40, 0x00, 0x43, 0x00, 0x04, 0x47}, wire)
}

func TestSetTurnoutWireFormat(t *testing.T) {
	msg := SetTurnout(4, TurnoutSwitch{Activate: true, Output2: true})
	require.Equal(t, []byte{0x53, 0x00, 0x04, 0x89, 0xDE}, msg.Data)

	deactivate := SetTurnout(4, TurnoutSwitch{Activate: false, Output2: true})
	require.Equal(t, byte(0x81), deactivate.Data[3])

	queued := SetTurnout(24, TurnoutSwitch{Activate: true, Output2: true, Queue: true})
	require.Equal(t, byte(0xA9), queued.Data[3])
}

func TestParseTurnoutInfo(t *testing.T) {
	data := appendLANXXOR([]byte{0x43, 0x00, 0x04, 0x02})
	info, err := ParseTurnoutInfo(data)
	require.NoError(t, err)
	require.Equal(t, uint16(4), info.Address)
	require.Equal(t, TurnoutPositionOutput2, info.Position)
	require.Equal(t, "output 2", FormatTurnoutPosition(info.Position))
}

func TestSetExtAccessoryWireFormat(t *testing.T) {
	msg := SetExtAccessory(4, 5)
	require.Equal(t, []byte{0x54, 0x00, 0x04, 0x05, 0x00, 0x55}, msg.Data)

	wire, err := msg.Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0x0A, 0x00, 0x40, 0x00, 0x54, 0x00, 0x04, 0x05, 0x00, 0x55}, wire)

	stop := SetExtAccessory(2047, 0)
	require.Equal(t, []byte{0x54, 0x07, 0xFF, 0x00, 0x00, 0xAC}, stop.Data)
}

func TestGetExtAccessoryInfoWireFormat(t *testing.T) {
	msg := GetExtAccessoryInfo(4)
	require.Equal(t, []byte{0x44, 0x00, 0x04, 0x00, 0x40}, msg.Data)
}

func TestParseExtAccessoryInfo(t *testing.T) {
	data := appendLANXXOR([]byte{0x44, 0x00, 0x04, 0x05, ExtAccessoryStatusValid})
	info, err := ParseExtAccessoryInfo(data)
	require.NoError(t, err)
	require.Equal(t, uint16(4), info.Address)
	require.Equal(t, byte(5), info.Value)
	require.Equal(t, ExtAccessoryStatusValid, info.Status)
}

func TestEncodeTurnoutSwitchCommand(t *testing.T) {
	require.Equal(t, byte(0x89), EncodeTurnoutSwitchCommand(TurnoutSwitch{Activate: true, Output2: true}))
	require.Equal(t, byte(0x81), EncodeTurnoutSwitchCommand(TurnoutSwitch{Output2: true}))
	require.Equal(t, byte(0xA9), EncodeTurnoutSwitchCommand(TurnoutSwitch{Activate: true, Output2: true, Queue: true}))
}

func TestDrivingSwitchingMessageNames(t *testing.T) {
	tests := []struct {
		msg  Message
		want string
	}{
		{GetLocoInfo(1), "LAN_X_GET_LOCO_INFO"},
		{SetLocoDrive(1, LocoSpeedSteps128, true, 1), "LAN_X_SET_LOCO_DRIVE"},
		{SetLocoFunction(1, 0, LocoFunctionOn), "LAN_X_SET_LOCO_FUNCTION"},
		{SetLocoFunctionGroup(1, LocoFunctionGroupF0F4, 0), "LAN_X_SET_LOCO_FUNCTION_GROUP"},
		{SetLocoBinaryState(1, 29, true), "LAN_X_SET_LOCO_BINARY_STATE"},
		{Message{Header: HeaderLANX, Data: appendLANXXOR([]byte{0xEF, 0, 1, 0, 0, 0})}, "LAN_X_LOCO_INFO"},
		{SetLocoEStop(1), "LAN_X_SET_LOCO_E_STOP"},
		{PurgeLoco(1), "LAN_X_PURGE_LOCO"},
		{GetTurnoutInfo(1), "LAN_X_GET_TURNOUT_INFO"},
		{SetTurnout(1, TurnoutSwitch{Activate: true}), "LAN_X_SET_TURNOUT"},
		{Message{Header: HeaderLANX, Data: appendLANXXOR([]byte{0x43, 0, 1, 0})}, "LAN_X_TURNOUT_INFO"},
		{GetExtAccessoryInfo(4), "LAN_X_GET_EXT_ACCESSORY_INFO"},
		{SetExtAccessory(4, 5), "LAN_X_SET_EXT_ACCESSORY"},
		{Message{Header: HeaderLANX, Data: appendLANXXOR([]byte{0x44, 0, 4, 5, 0})}, "LAN_X_EXT_ACCESSORY_INFO"},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, MessageName(tt.msg), "%#v", tt.msg)
	}
}
