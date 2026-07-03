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
