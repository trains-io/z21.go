package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCVAddressFromNumber(t *testing.T) {
	require.Equal(t, CVAddress(0), CVAddressFromNumber(1))
	require.Equal(t, uint16(29), CVAddress(28).Number())
}

func TestReadCVWireFormat(t *testing.T) {
	msg := ReadCV(CVAddressFromNumber(1))
	require.Equal(t, []byte{0x23, 0x11, 0x00, 0x00, 0x32}, msg.Data)
}

func TestWriteCVWireFormat(t *testing.T) {
	msg := WriteCV(CVAddressFromNumber(2), 0x2A)
	require.Equal(t, []byte{0x24, 0x12, 0x00, 0x01, 0x2A, 0x1D}, msg.Data)
}

func TestWriteMMByteWireFormat(t *testing.T) {
	msg := WriteMMByte(0x00, 0x05)
	require.Equal(t, []byte{0x24, 0xFF, 0x00, 0x00, 0x05, 0xDE}, msg.Data)
}

func TestReadDCCRegisterWireFormat(t *testing.T) {
	msg := ReadDCCRegister(0x01)
	require.Equal(t, []byte{0x22, 0x11, 0x01, 0x32}, msg.Data)
}

func TestWriteDCCRegisterWireFormat(t *testing.T) {
	msg := WriteDCCRegister(0x01, 0x05)
	require.Equal(t, []byte{0x23, 0x12, 0x01, 0x05, 0x35}, msg.Data)
}

func TestParseCVResult(t *testing.T) {
	data := appendLANXXOR([]byte{0x64, 0x14, 0x00, 0x00, 0x05})
	result, err := ParseCVResult(data)
	require.NoError(t, err)
	require.Equal(t, CVAddress(0), result.Address)
	require.Equal(t, byte(0x05), result.Value)
}

func TestCVResultFromMessages(t *testing.T) {
	data := appendLANXXOR([]byte{0x64, 0x14, 0x00, 0x01, 0x07})
	msgs := []Message{{Header: HeaderLANX, Data: data}}
	result, err := CVResultFromMessages(msgs)
	require.NoError(t, err)
	require.Equal(t, CVAddress(1), result.Address)
	require.Equal(t, byte(0x07), result.Value)
}

func TestCVResultFromMessagesNack(t *testing.T) {
	msgs := []Message{{Header: HeaderLANX, Data: []byte{0x61, 0x13, 0x72}}}
	_, err := CVResultFromMessages(msgs)
	require.ErrorContains(t, err, "NACK")
}

func TestCVResultFromMessagesNackSC(t *testing.T) {
	msgs := []Message{{Header: HeaderLANX, Data: []byte{0x61, 0x12, 0x73}}}
	_, err := CVResultFromMessages(msgs)
	require.ErrorContains(t, err, "short circuit")
}

func TestParseCVNack(t *testing.T) {
	require.NoError(t, ParseCVNack([]byte{0x61, 0x13, 0x72}))
	require.NoError(t, ParseCVNackSC([]byte{0x61, 0x12, 0x73}))
}

func TestPOMLocoWriteByteWireFormat(t *testing.T) {
	msg := POMLocoWriteByte(3, CVAddressFromNumber(1), 0x05)
	require.Equal(t, []byte{0xE6, 0x30, 0x00, 0x03, 0xEC, 0x00, 0x05, 0x3C}, msg.Data)
}

func TestPOMLocoWriteBitWireFormat(t *testing.T) {
	msg := POMLocoWriteBit(3, CVAddressFromNumber(29), 4, true)
	require.Equal(t, byte(0xE8), msg.Data[4]&0xFC)
	require.Equal(t, byte(0x24), msg.Data[6]) // bit 4 set
}

func TestPOMLocoReadByteWireFormat(t *testing.T) {
	msg := POMLocoReadByte(128, CVAddressFromNumber(28))
	require.Equal(t, byte(0xC0), msg.Data[2])
	require.Equal(t, byte(0x80), msg.Data[3])
	require.Equal(t, byte(0xE4), msg.Data[4]&0xFC)
}

func TestPOMAccessoryWriteByteWireFormat(t *testing.T) {
	output := uint8(2)
	msg := POMAccessoryWriteByte(42, CVAddressFromNumber(1), 0x0A, &output)
	packed := uint16(msg.Data[2])<<8 | uint16(msg.Data[3])
	require.Equal(t, uint16(42<<4|0x0A), packed)
	require.Equal(t, byte(0xEC), msg.Data[4]&0xFC)
}

func TestPOMAccessoryWriteBitWireFormat(t *testing.T) {
	output := uint8(1)
	msg := POMAccessoryWriteBit(10, CVAddressFromNumber(29), 3, true, &output)
	require.Equal(t, byte(0xE8), msg.Data[4]&0xFC)
	require.Equal(t, byte(0x23), msg.Data[6]) // bit 3 + value
}

func TestPOMAccessoryReadByteWireFormat(t *testing.T) {
	msg := POMAccessoryReadByte(5, CVAddressFromNumber(28), nil)
	require.Equal(t, byte(0x31), msg.Data[1])
	require.Equal(t, byte(0xE4), msg.Data[4]&0xFC)
	require.Equal(t, byte(0x00), msg.Data[6])
}

func TestCVMessageNames(t *testing.T) {
	require.Equal(t, "LAN_X_CV_READ", MessageName(ReadCV(CVAddress(0))))
	require.Equal(t, "LAN_X_CV_WRITE", MessageName(WriteCV(CVAddress(0), 1)))
	require.Equal(t, "LAN_X_MM_WRITE_BYTE", MessageName(WriteMMByte(0, 5)))
	require.Equal(t, "LAN_X_DCC_READ_REGISTER", MessageName(ReadDCCRegister(1)))
	require.Equal(t, "LAN_X_DCC_WRITE_REGISTER", MessageName(WriteDCCRegister(1, 5)))
	require.Equal(t, "LAN_X_CV_RESULT", MessageName(Message{Header: HeaderLANX, Data: appendLANXXOR([]byte{0x64, 0x14, 0x00, 0x00, 0x05})}))
	require.Equal(t, "LAN_X_CV_NACK", MessageName(Message{Header: HeaderLANX, Data: []byte{0x61, 0x13, 0x72}}))
	require.Equal(t, "LAN_X_CV_POM_WRITE_BYTE", MessageName(POMLocoWriteByte(3, CVAddress(0), 5)))
	require.Equal(t, "LAN_X_CV_POM_READ_BYTE", MessageName(POMLocoReadByte(3, CVAddress(0))))
	require.Equal(t, "LAN_X_CV_POM_ACCESSORY_WRITE_BYTE", MessageName(POMAccessoryWriteByte(1, CVAddress(0), 1, nil)))
}
