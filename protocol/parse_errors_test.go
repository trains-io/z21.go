package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseErrors(t *testing.T) {
	t.Run("ParseHWInfo", func(t *testing.T) {
		_, err := ParseHWInfo([]byte{0x01})
		require.ErrorContains(t, err, "too short")
	})

	t.Run("ParseSerialNumber", func(t *testing.T) {
		_, err := ParseSerialNumber(nil)
		require.ErrorContains(t, err, "too short")
	})

	t.Run("ParseSystemState", func(t *testing.T) {
		_, err := ParseSystemState(make([]byte, 8))
		require.ErrorContains(t, err, "too short")
	})

	t.Run("ParseCVResult", func(t *testing.T) {
		_, err := ParseCVResult([]byte{0x64})
		require.ErrorContains(t, err, "too short")
		bad := appendLANXXOR([]byte{0x64, 0x14, 0x00, 0x00, 0x05})
		bad[len(bad)-1] ^= 0xFF
		_, err = ParseCVResult(bad)
		require.ErrorContains(t, err, "checksum")
	})

	t.Run("ParseTurnoutInfo", func(t *testing.T) {
		_, err := ParseTurnoutInfo([]byte{0x43})
		require.ErrorContains(t, err, "too short")
	})

	t.Run("ParseLocoInfo", func(t *testing.T) {
		_, err := ParseLocoInfo([]byte{0xEF})
		require.ErrorContains(t, err, "too short")
	})

	t.Run("ParseRMBusStatus", func(t *testing.T) {
		_, err := ParseRMBusStatus([]byte{0x00})
		require.ErrorContains(t, err, "too short")
	})

	t.Run("ParseRailComData", func(t *testing.T) {
		_, err := ParseRailComData(make([]byte, 4))
		require.ErrorContains(t, err, "too short")
	})

	t.Run("ParseCANDetector", func(t *testing.T) {
		_, err := ParseCANDetector(make([]byte, 4))
		require.ErrorContains(t, err, "too short")
	})

	t.Run("ProgramRMBusModule", func(t *testing.T) {
		_, err := ProgramRMBusModule(21)
		require.ErrorContains(t, err, "1..20")
	})

	t.Run("SetCANDeviceDescription", func(t *testing.T) {
		_, err := SetCANDeviceDescription(0xC101, `bad"name`)
		require.ErrorContains(t, err, "must not contain")
		_, err = SetCANDeviceDescription(0xC101, string(make([]byte, 16)))
		require.ErrorContains(t, err, "too long")
	})

	t.Run("Unmarshal", func(t *testing.T) {
		_, _, err := Unmarshal([]byte{0x01})
		require.Error(t, err)
		_, _, err = Unmarshal([]byte{0x02, 0x00, 0x00, 0x00})
		require.ErrorContains(t, err, "invalid data length")
	})
}
