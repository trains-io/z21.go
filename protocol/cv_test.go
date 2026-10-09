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
	require.Equal(t, byte(0x0C), msg.Data[6]) // 0000VPPP: V=1, PPP=4
}

func TestEncodePOMBitParam(t *testing.T) {
	tests := []struct {
		bitPos uint8
		on     bool
		want   byte
	}{
		{bitPos: 0, on: false, want: 0x00},
		{bitPos: 0, on: true, want: 0x08},
		{bitPos: 7, on: false, want: 0x07},
		{bitPos: 7, on: true, want: 0x0F},
		{bitPos: 8, on: false, want: 0x00}, // positions above 7 are masked
	}
	for _, tt := range tests {
		if got := encodePOMBitParam(tt.bitPos, tt.on); got != tt.want {
			t.Fatalf("encodePOMBitParam(%d, %t) = %#02x, want %#02x", tt.bitPos, tt.on, got, tt.want)
		}
	}
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
	require.Equal(t, byte(0x0B), msg.Data[6]) // 0000VPPP: V=1, PPP=3
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

func TestEncodeCVResultWire(t *testing.T) {
	got, err := EncodeCVResult(CVResult{Address: CVAddressFromNumber(1), Value: 0x05}).Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0x0A, 0x00, 0x40, 0x00, 0x64, 0x14, 0x00, 0x00, 0x05, 0x75}, got)
}

func TestEncodeCVResultRoundTrip(t *testing.T) {
	for _, in := range []CVResult{
		{},
		{Address: CVAddressFromNumber(29), Value: 0x06},
		{Address: 1023, Value: 0xFF},
	} {
		got, err := CVResultFromMessages([]Message{EncodeCVResult(in)})
		require.NoError(t, err)
		require.Equal(t, in, got)
	}
}

func TestEncodeCVNacks(t *testing.T) {
	nack, err := EncodeCVNack().Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0x07, 0x00, 0x40, 0x00, 0x61, 0x13, 0x72}, nack)

	nackSC, err := EncodeCVNackSC().Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0x07, 0x00, 0x40, 0x00, 0x61, 0x12, 0x73}, nackSC)

	_, err = CVResultFromMessages([]Message{EncodeCVNack()})
	require.ErrorContains(t, err, "NACK")
	_, err = CVResultFromMessages([]Message{EncodeCVNackSC()})
	require.ErrorContains(t, err, "short circuit")
}

var testCVAddresses = []CVAddress{0, CVAddressFromNumber(29), 255, 1023}

func TestParseReadWriteCV(t *testing.T) {
	for _, cv := range testCVAddresses {
		got, err := ParseReadCV(ReadCV(cv).Data)
		require.NoError(t, err)
		require.Equal(t, cv, got)

		for _, value := range []byte{0x00, 0x05, 0xFF} {
			gotCV, gotValue, err := ParseWriteCV(WriteCV(cv, value).Data)
			require.NoError(t, err)
			require.Equal(t, cv, gotCV)
			require.Equal(t, value, gotValue)
		}
	}
}

func TestParseRegisterAndMMRequests(t *testing.T) {
	for _, reg := range []byte{1, 4, 8} {
		got, err := ParseReadDCCRegister(ReadDCCRegister(reg).Data)
		require.NoError(t, err)
		require.Equal(t, reg, got)

		gotReg, gotValue, err := ParseWriteDCCRegister(WriteDCCRegister(reg, 0xA5).Data)
		require.NoError(t, err)
		require.Equal(t, reg, gotReg)
		require.Equal(t, byte(0xA5), gotValue)

		gotReg, gotValue, err = ParseWriteMMByte(WriteMMByte(reg, 0x5A).Data)
		require.NoError(t, err)
		require.Equal(t, reg, gotReg)
		require.Equal(t, byte(0x5A), gotValue)
	}
}

func TestParseProgrammingRequestsRejectOtherCommands(t *testing.T) {
	readCV := ReadCV(0).Data
	writeCV := WriteCV(0, 1).Data
	writeMM := WriteMMByte(1, 1).Data
	writeReg := WriteDCCRegister(1, 1).Data

	_, err := ParseReadCV(writeReg)
	require.Error(t, err, "CV read and DCC write register share X-header 0x23")
	_, _, err = ParseWriteDCCRegister(readCV)
	require.Error(t, err)
	_, _, err = ParseWriteCV(writeMM)
	require.Error(t, err, "CV write and MM write byte share X-header 0x24")
	_, _, err = ParseWriteMMByte(writeCV)
	require.Error(t, err)
	_, err = ParseReadDCCRegister(readCV)
	require.Error(t, err)
}

func TestParsePOMLoco(t *testing.T) {
	cv := CVAddressFromNumber(29)
	for _, addr := range []uint16{3, 128, 10239} {
		got, err := ParsePOMLoco(POMLocoWriteByte(addr, cv, 0x06).Data)
		require.NoError(t, err)
		require.Equal(t, POMRequest{Address: addr, CV: cv, Operation: POMWriteByte, Value: 0x06}, got)

		got, err = ParsePOMLoco(POMLocoWriteBit(addr, cv, 5, true).Data)
		require.NoError(t, err)
		require.Equal(t, POMRequest{Address: addr, CV: cv, Operation: POMWriteBit, Bit: 5, BitValue: true}, got)

		got, err = ParsePOMLoco(POMLocoReadByte(addr, 1023).Data)
		require.NoError(t, err)
		require.Equal(t, POMRequest{Address: addr, CV: 1023, Operation: POMReadByte}, got)
	}
}

func TestParsePOMAccessory(t *testing.T) {
	cv := CVAddressFromNumber(1)
	output := uint8(5)
	for _, addr := range []uint16{0, 42, 511} {
		got, err := ParsePOMAccessory(POMAccessoryWriteByte(addr, cv, 0x0A, &output).Data)
		require.NoError(t, err)
		require.Equal(t, POMRequest{Address: addr, HasOutput: true, Output: 5, CV: cv, Operation: POMWriteByte, Value: 0x0A}, got)

		got, err = ParsePOMAccessory(POMAccessoryWriteBit(addr, cv, 0, false, nil).Data)
		require.NoError(t, err)
		require.Equal(t, POMRequest{Address: addr, CV: cv, Operation: POMWriteBit}, got)

		got, err = ParsePOMAccessory(POMAccessoryReadByte(addr, cv, nil).Data)
		require.NoError(t, err)
		require.Equal(t, POMRequest{Address: addr, CV: cv, Operation: POMReadByte}, got)
	}
}

func TestParsePOMErrors(t *testing.T) {
	_, err := ParsePOMLoco(POMAccessoryReadByte(1, 0, nil).Data)
	require.Error(t, err, "loco parser must reject accessory POM")

	_, err = ParsePOMAccessory(POMLocoReadByte(3, 0).Data)
	require.Error(t, err, "accessory parser must reject loco POM")

	_, err = ParsePOMLoco(appendLANXXOR([]byte{0xE6, 0x30, 0x00, 0x03, 0xF0, 0x00, 0x00}))
	require.ErrorContains(t, err, "option")

	_, err = ParsePOMLoco([]byte{0xE6, 0x30, 0x00, 0x03, 0xEC, 0x00, 0x05})
	require.ErrorContains(t, err, "too short")
}
