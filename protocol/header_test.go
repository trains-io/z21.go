package protocol

import "testing"

func TestHeaderName(t *testing.T) {
	if got := HeaderName(HeaderLANGetHWInfo); got != "LAN_GET_HWINFO" {
		t.Fatalf("HeaderName() = %q, want LAN_GET_HWINFO", got)
	}
	if got := HeaderName(0x00ff); got != "0x00ff" {
		t.Fatalf("HeaderName() = %q, want 0x00ff", got)
	}
}

func TestMessageNameLANX(t *testing.T) {
	tests := []struct {
		msg  Message
		want string
	}{
		{GetXVersion(), "LAN_X_GET_VERSION"},
		{GetXStatus(), "LAN_X_GET_STATUS"},
		{GetXFirmware(), "LAN_X_GET_FIRMWARE_VERSION"},
		{SetTrackPower(true), "LAN_X_SET_TRACK_POWER_ON"},
		{SetTrackPower(false), "LAN_X_SET_TRACK_POWER_OFF"},
		{SetStop(), "LAN_X_SET_STOP"},
		{Message{Header: HeaderLANX, Data: []byte{0x81, 0x00, 0x81}}, "LAN_X_BC_STOPPED"},
		{Message{Header: HeaderLANX, Data: []byte{0x63, 0x21, 0x30, 0x12, 0x60}}, "LAN_X_GET_VERSION"},
		{Message{Header: HeaderLANX, Data: []byte{0x62, 0x22, 0x00, 0x40}}, "LAN_X_STATUS_CHANGED"},
		{Message{Header: HeaderLANX, Data: []byte{0x61, 0x01}}, "LAN_X_BC_TRACK_POWER_ON"},
		{Message{Header: HeaderLANX, Data: []byte{0x61, 0x02, 0x63}}, "LAN_X_BC_PROGRAMMING_MODE"},
		{Message{Header: HeaderLANX, Data: []byte{0x61, 0x08, 0x69}}, "LAN_X_BC_TRACK_SHORT_CIRCUIT"},
		{Message{Header: HeaderLANX, Data: []byte{0x61, 0x82, 0xE3}}, "LAN_X_UNKNOWN_COMMAND"},
		{GetHWInfo(), "LAN_GET_HWINFO"},
		{GetAllCANDetectors(), "LAN_CAN_DETECTOR"},
		{GetRMBusData(0), "LAN_RMBUS_GETDATA"},
		{GetRailComData(1), "LAN_RAILCOM_GETDATA"},
		{GetCANDeviceDescription(0xC101), "LAN_CAN_DEVICE_GET_DESCRIPTION"},
		{SetCANBoosterTrackPower(0xC101, CANBoosterTrackPowerActivateAll), "LAN_CAN_BOOSTER_SET_TRACKPOWER"},
	}

	for _, tt := range tests {
		if got := MessageName(tt.msg); got != tt.want {
			t.Fatalf("MessageName(%#v) = %q, want %q", tt.msg, got, tt.want)
		}
	}
}
