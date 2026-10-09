package protocol

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRailComDataWireFormat(t *testing.T) {
	msg := GetRailComData(3)
	require.Equal(t, HeaderLANRailComGetData, msg.Header)
	require.Equal(t, []byte{0x01, 0x03, 0x00}, msg.Data)

	wire, err := msg.Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0x07, 0x00, 0x89, 0x00, 0x01, 0x03, 0x00}, wire)
}

func TestParseRailComData(t *testing.T) {
	data := make([]byte, railComDataMinLen)
	binary.LittleEndian.PutUint16(data[0:2], 128)
	binary.LittleEndian.PutUint32(data[2:6], 42)
	binary.LittleEndian.PutUint16(data[6:8], 1)
	data[9] = RailComOptionSpeed1 | RailComOptionQoS
	data[10] = 55
	data[11] = 7

	rc, err := ParseRailComData(data)
	require.NoError(t, err)
	require.Equal(t, uint16(128), rc.LocoAddress)
	require.Equal(t, uint32(42), rc.ReceiveCounter)
	require.Equal(t, uint16(1), rc.ErrorCounter)
	require.Equal(t, byte(55), rc.Speed)
	require.Equal(t, byte(7), rc.QoS)
	require.Equal(t, "Speed1|QoS", FormatRailComOptions(rc.Options))
}

func TestRailComDataFromMessages(t *testing.T) {
	data := make([]byte, railComDataMinLen)
	binary.LittleEndian.PutUint16(data[0:2], 1)
	msgs := []Message{{Header: HeaderLANRailComDataChanged, Data: data}}
	out, err := RailComDataFromMessages(msgs)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, uint16(1), out[0].LocoAddress)
}

func TestEncodeRailComDataWire(t *testing.T) {
	got, err := EncodeRailComData(RailComData{
		LocoAddress:    3,
		ReceiveCounter: 0x01020304,
		ErrorCounter:   0x0506,
		Options:        RailComOptionSpeed1 | RailComOptionQoS,
		Speed:          42,
		QoS:            99,
	}).Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{
		0x11, 0x00, 0x88, 0x00, // DataLen 17, LAN_RAILCOM_DATACHANGED
		0x03, 0x00, // loco address
		0x04, 0x03, 0x02, 0x01, // receive counter
		0x06, 0x05, // error counter
		0x00,   // reserved
		0x05,   // options
		42, 99, // speed, QoS
		0x00, // reserved
	}, got)
}

func TestEncodeRailComDataRoundTrip(t *testing.T) {
	for _, in := range []RailComData{
		{},
		{LocoAddress: 10239, ReceiveCounter: 0xFFFFFFFF, ErrorCounter: 0xFFFF, Options: 0x07, Speed: 0xFF, QoS: 0xFF},
	} {
		got, err := RailComDataFromMessages([]Message{EncodeRailComData(in)})
		require.NoError(t, err)
		require.Equal(t, []RailComData{in}, got)
		require.True(t, IsRailComDataChanged(EncodeRailComData(in)))
	}
}

func TestParseGetRailComData(t *testing.T) {
	for _, addr := range []uint16{RailComPollNextAddress, 3, 128, 10239} {
		got, err := ParseGetRailComData(GetRailComData(addr).Data)
		require.NoError(t, err)
		require.Equal(t, addr, got)
	}

	got, err := ParseGetRailComData(nil)
	require.NoError(t, err, "deprecated request without data")
	require.Equal(t, RailComPollNextAddress, got)

	got, err = ParseGetRailComData([]byte{0x00, 0x03, 0x00})
	require.NoError(t, err)
	require.Equal(t, RailComPollNextAddress, got, "type 0 polls the next loco")

	_, err = ParseGetRailComData([]byte{0x01, 0x03})
	require.ErrorContains(t, err, "too short")
}
