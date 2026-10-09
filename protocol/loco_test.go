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

func TestEncodeLocoInfoWire(t *testing.T) {
	msg := EncodeLocoInfo(LocoInfo{
		Address:        3,
		Busy:           true,
		SpeedSteps:     4,
		Forward:        true,
		Speed:          5,
		Headlight:      true,
		FunctionsF1F4:  0x01,
		FunctionsF5F12: 0x01,
	})
	got, err := msg.Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{
		0x0F, 0x00, 0x40, 0x00, // DataLen 15, LAN_X
		0xEF,       // X-header LAN_X_LOCO_INFO
		0x00, 0x03, // address 3
		0x0C,             // busy, 128 steps
		0x85,             // forward, speed 5
		0x11,             // F0, F1
		0x01,             // F5
		0x00, 0x00, 0x00, // F13-F20, F21-F28, F29-F31
		0x75, // XOR
	}, got)
}

func TestEncodeLocoInfoLongAddress(t *testing.T) {
	msg := EncodeLocoInfo(LocoInfo{Address: 1234})
	require.Equal(t, []byte{0xC4, 0xD2}, msg.Data[1:3])
}

func TestEncodeLocoInfoRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		in   LocoInfo
	}{
		{name: "zero", in: LocoInfo{}},
		{name: "short address, 14 steps, reverse", in: LocoInfo{Address: 127, SpeedSteps: 0, Speed: 0x0F}},
		{name: "long address, 28 steps", in: LocoInfo{Address: 128, SpeedSteps: 2, Forward: true, Speed: 0x1F}},
		{name: "max address, MM, busy", in: LocoInfo{Address: 10239, MMFormat: true, Busy: true, SpeedSteps: 4, Speed: 0x7F}},
		{name: "DB4 flags only", in: LocoInfo{Address: 3, DoubleTraction: true, SmartSearch: true, Headlight: true}},
		{
			name: "all functions",
			in: LocoInfo{
				Address:         3,
				SpeedSteps:      4,
				Headlight:       true,
				FunctionsF1F4:   0x0F,
				FunctionsF5F12:  0xFF,
				FunctionsF13F20: 0xFF,
				FunctionsF21F28: 0xFF,
				FunctionsF29F31: 0x07,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LocoInfoFromMessages([]Message{EncodeLocoInfo(tt.in)})
			require.NoError(t, err)
			require.Equal(t, tt.in, got)
		})
	}
}

func TestEncodeLocoInfoMasksFields(t *testing.T) {
	got, err := ParseLocoInfo(EncodeLocoInfo(LocoInfo{
		SpeedSteps:      0xFF,
		Speed:           0xFF,
		FunctionsF1F4:   0xFF,
		FunctionsF29F31: 0xFF,
	}).Data)
	require.NoError(t, err)
	require.Equal(t, byte(0x07), got.SpeedSteps)
	require.False(t, got.Forward, "speed overflow must not set the direction bit")
	require.Equal(t, byte(0x7F), got.Speed)
	require.Equal(t, byte(0x0F), got.FunctionsF1F4)
	require.False(t, got.Headlight, "F1-F4 overflow must not set F0")
	require.Equal(t, byte(0x07), got.FunctionsF29F31)
}

var testLocoAddresses = []uint16{0, 3, 127, 128, 1234, 10239}

func TestParseGetLocoInfo(t *testing.T) {
	for _, addr := range testLocoAddresses {
		got, err := ParseGetLocoInfo(GetLocoInfo(addr).Data)
		require.NoError(t, err)
		require.Equal(t, addr, got)
	}
}

func TestParseSetLocoDrive(t *testing.T) {
	for _, addr := range testLocoAddresses {
		for _, steps := range []byte{LocoSpeedSteps14, LocoSpeedSteps28, LocoSpeedSteps128} {
			for _, in := range []LocoDrive{
				{Address: addr, SpeedSteps: steps},
				{Address: addr, SpeedSteps: steps, Forward: true, Speed: 0x7F},
			} {
				got, err := ParseSetLocoDrive(SetLocoDrive(in.Address, in.SpeedSteps, in.Forward, in.Speed).Data)
				require.NoError(t, err)
				require.Equal(t, in, got)
			}
		}
	}
}

func TestParseSetLocoFunction(t *testing.T) {
	for _, addr := range testLocoAddresses {
		for _, action := range []LocoFunctionAction{LocoFunctionOff, LocoFunctionOn, LocoFunctionToggle} {
			for _, fn := range []uint8{0, 31, 63} {
				gotAddr, gotFn, gotAction, err := ParseSetLocoFunction(SetLocoFunction(addr, fn, action).Data)
				require.NoError(t, err)
				require.Equal(t, addr, gotAddr)
				require.Equal(t, fn, gotFn)
				require.Equal(t, action, gotAction)
			}
		}
	}
}

func TestParseSetLocoFunctionGroup(t *testing.T) {
	groups := []LocoFunctionGroup{
		LocoFunctionGroupF0F4, LocoFunctionGroupF5F8, LocoFunctionGroupF9F12, LocoFunctionGroupF13F20,
		LocoFunctionGroupF21F28, LocoFunctionGroupF29F36, LocoFunctionGroupF37F44, LocoFunctionGroupF45F52,
		LocoFunctionGroupF53F60, LocoFunctionGroupF61F68,
	}
	for _, group := range groups {
		addr, gotGroup, functions, err := ParseSetLocoFunctionGroup(SetLocoFunctionGroup(1234, group, 0xA5).Data)
		require.NoError(t, err)
		require.Equal(t, uint16(1234), addr)
		require.Equal(t, group, gotGroup)
		require.Equal(t, byte(0xA5), functions)
	}
}

func TestParseSetLocoBinaryState(t *testing.T) {
	for _, binAddr := range []uint16{0, 1, 127, 128, 32767} {
		for _, on := range []bool{false, true} {
			addr, gotBin, gotOn, err := ParseSetLocoBinaryState(SetLocoBinaryState(3, binAddr, on).Data)
			require.NoError(t, err)
			require.Equal(t, uint16(3), addr)
			require.Equal(t, binAddr, gotBin)
			require.Equal(t, on, gotOn)
		}
	}
}

func TestParseSetLocoEStopAndPurgeLoco(t *testing.T) {
	for _, addr := range testLocoAddresses {
		got, err := ParseSetLocoEStop(SetLocoEStop(addr).Data)
		require.NoError(t, err)
		require.Equal(t, addr, got)

		got, err = ParsePurgeLoco(PurgeLoco(addr).Data)
		require.NoError(t, err)
		require.Equal(t, addr, got)
	}
}

func TestParseLocoRequestsRejectOtherCommands(t *testing.T) {
	drive := SetLocoDrive(3, LocoSpeedSteps128, true, 10).Data
	function := SetLocoFunction(3, 0, LocoFunctionOn).Data
	group := SetLocoFunctionGroup(3, LocoFunctionGroupF0F4, 0x01).Data
	getInfo := GetLocoInfo(3).Data
	purge := PurgeLoco(3).Data

	tests := []struct {
		name  string
		parse func([]byte) error
		data  []byte
	}{
		{name: "drive parser on function", parse: func(d []byte) error { _, err := ParseSetLocoDrive(d); return err }, data: function},
		{name: "drive parser on group", parse: func(d []byte) error { _, err := ParseSetLocoDrive(d); return err }, data: group},
		{name: "function parser on drive", parse: func(d []byte) error { _, _, _, err := ParseSetLocoFunction(d); return err }, data: drive},
		{name: "group parser on function", parse: func(d []byte) error { _, _, _, err := ParseSetLocoFunctionGroup(d); return err }, data: function},
		{name: "group parser on drive", parse: func(d []byte) error { _, _, _, err := ParseSetLocoFunctionGroup(d); return err }, data: drive},
		{name: "get loco info parser on purge", parse: func(d []byte) error { _, err := ParseGetLocoInfo(d); return err }, data: purge},
		{name: "purge parser on get loco info", parse: func(d []byte) error { _, err := ParsePurgeLoco(d); return err }, data: getInfo},
		{name: "e-stop parser on get loco info", parse: func(d []byte) error { _, err := ParseSetLocoEStop(d); return err }, data: getInfo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, tt.parse(tt.data))
		})
	}
}

func TestParseLocoRequestsErrors(t *testing.T) {
	badChecksum := GetLocoInfo(3).Data
	badChecksum[len(badChecksum)-1] ^= 0xFF

	_, err := ParseGetLocoInfo(badChecksum)
	require.ErrorContains(t, err, "checksum")

	_, err = ParseGetLocoInfo([]byte{0xE3, 0xF0, 0x00})
	require.ErrorContains(t, err, "too short")

	_, _, _, err = ParseSetLocoFunction(appendLANXXOR([]byte{0xE4, 0xF8, 0x00, 0x03, 0xC0}))
	require.ErrorContains(t, err, "action")

	_, err = ParseSetLocoDrive(appendLANXXOR([]byte{0xE4, 0x11, 0x00, 0x03, 0x00}))
	require.Error(t, err, "speed steps nibble 1 is undefined")
}
