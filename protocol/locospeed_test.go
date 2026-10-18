package protocol

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeLocoSpeed(t *testing.T) {
	tests := []struct {
		name       string
		speedSteps byte
		step       uint8
		want       byte
	}{
		{"14 stop", LocoSpeedSteps14, 0, 0x00},
		{"14 step 1", LocoSpeedSteps14, 1, 0x02},
		{"14 step 14", LocoSpeedSteps14, 14, 0x0F},
		{"28 stop", LocoSpeedSteps28, 0, 0x00},
		{"28 step 1", LocoSpeedSteps28, 1, 0x02},
		{"28 step 2", LocoSpeedSteps28, 2, 0x12},
		{"28 step 3", LocoSpeedSteps28, 3, 0x03},
		{"28 step 27", LocoSpeedSteps28, 27, 0x0F},
		{"28 step 28", LocoSpeedSteps28, 28, 0x1F},
		{"128 stop", LocoSpeedSteps128, 0, 0x00},
		{"128 step 1", LocoSpeedSteps128, 1, 0x02},
		{"128 step 126", LocoSpeedSteps128, 126, 0x7F},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeLocoSpeed(tt.speedSteps, tt.step)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestEncodeLocoSpeedErrors(t *testing.T) {
	tests := []struct {
		name       string
		speedSteps byte
		step       uint8
	}{
		{"14 step 15", LocoSpeedSteps14, 15},
		{"28 step 29", LocoSpeedSteps28, 29},
		{"128 step 127", LocoSpeedSteps128, 127},
		{"unknown speed steps", 0x01, 0},
		{"LAN_X_LOCO_INFO 128 speed steps", LocoInfoSpeedSteps128, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EncodeLocoSpeed(tt.speedSteps, tt.step)
			require.Error(t, err)
		})
	}
}

func TestDecodeLocoSpeed(t *testing.T) {
	tests := []struct {
		name       string
		speedSteps byte
		speed      byte
		wantStep   uint8
		wantEStop  bool
	}{
		{"14 stop", LocoSpeedSteps14, 0x00, 0, false},
		{"14 e-stop", LocoSpeedSteps14, 0x01, 0, true},
		{"14 step 14", LocoSpeedSteps14, 0x0F, 14, false},
		{"14 ignores bit 4", LocoSpeedSteps14, 0x12, 1, false},
		{"28 stop", LocoSpeedSteps28, 0x00, 0, false},
		{"28 stop, other encoding", LocoSpeedSteps28, 0x10, 0, false},
		{"28 e-stop", LocoSpeedSteps28, 0x01, 0, true},
		{"28 e-stop, other encoding", LocoSpeedSteps28, 0x11, 0, true},
		{"28 step 2", LocoSpeedSteps28, 0x12, 2, false},
		{"28 step 28", LocoSpeedSteps28, 0x1F, 28, false},
		{"128 stop", LocoSpeedSteps128, 0x00, 0, false},
		{"128 e-stop", LocoSpeedSteps128, 0x01, 0, true},
		{"128 step 126", LocoSpeedSteps128, 0x7F, 126, false},
		{"128 ignores direction", LocoSpeedSteps128, 0x85, 4, false},
		{"unknown like 128", 0x01, 0x7F, 126, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step, eStop := DecodeLocoSpeed(tt.speedSteps, tt.speed)
			require.Equal(t, tt.wantStep, step)
			require.Equal(t, tt.wantEStop, eStop)
		})
	}
}

func TestLocoSpeedRoundTrip(t *testing.T) {
	for _, speedSteps := range []byte{LocoSpeedSteps14, LocoSpeedSteps28, LocoSpeedSteps128} {
		t.Run(fmt.Sprintf("speed steps %#02x", speedSteps), func(t *testing.T) {
			seen := map[byte]bool{LocoSpeedEStop: true}
			for step := range MaxLocoSpeedStep(speedSteps) + 1 {
				speed, err := EncodeLocoSpeed(speedSteps, step)
				require.NoError(t, err)
				require.False(t, seen[speed], "step %d: speed %#02x used twice", step, speed)
				seen[speed] = true

				got, eStop := DecodeLocoSpeed(speedSteps, speed)
				require.Equal(t, step, got)
				require.False(t, eStop)
			}
		})
	}
}

func TestLocoSpeedStepsConversion(t *testing.T) {
	for _, tt := range []struct{ drive, info byte }{
		{LocoSpeedSteps14, LocoInfoSpeedSteps14},
		{LocoSpeedSteps28, LocoInfoSpeedSteps28},
		{LocoSpeedSteps128, LocoInfoSpeedSteps128},
	} {
		require.Equal(t, tt.info, LocoInfoSpeedSteps(tt.drive))
		require.Equal(t, tt.drive, LocoDriveSpeedSteps(tt.info))
	}
}

func TestLocoInfoSpeedStep(t *testing.T) {
	info := LocoInfo{SpeedSteps: LocoInfoSpeedSteps128, Speed: 0x7F}
	step, eStop := info.SpeedStep()
	require.Equal(t, uint8(126), step)
	require.False(t, eStop)

	info = LocoInfo{SpeedSteps: LocoInfoSpeedSteps28, Speed: 0x12}
	step, eStop = info.SpeedStep()
	require.Equal(t, uint8(2), step)
	require.False(t, eStop)

	info.Speed = LocoSpeedEStop
	_, eStop = info.SpeedStep()
	require.True(t, eStop)
}
