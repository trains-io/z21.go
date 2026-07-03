package protocol

import "fmt"

const (
	xHeaderCVRead        byte = 0x23
	xCommandCVRead       byte = 0x11
	xHeaderCVWrite       byte = 0x24
	xCommandCVWrite      byte = 0x12
	xHeaderCVResult      byte = 0x64
	xCommandCVResult     byte = 0x14
	xHeaderDCCReadReg    byte = 0x22
	xCommandDCCReadReg   byte = 0x11
	xCommandDCCWriteReg  byte = 0x12
	xCommandMMWriteByte  byte = 0xFF
)

// CVAddress is a 0-based CV index (0=CV1, 255=CV256) per spec §6.
type CVAddress uint16

// CVAddressFromNumber converts a human CV number (1=CV1) to a spec CV address.
func CVAddressFromNumber(cvNumber uint16) CVAddress {
	if cvNumber == 0 {
		return 0
	}
	return CVAddress(cvNumber - 1)
}

// Number returns the human CV number (CV1=1).
func (cv CVAddress) Number() uint16 {
	return uint16(cv) + 1
}

func encodeCVAddress(cv CVAddress) (msb, lsb byte) {
	v := uint16(cv)
	return byte(v >> 8), byte(v)
}

func parseCVAddress(msb, lsb byte) CVAddress {
	return CVAddress(uint16(msb)<<8 | uint16(lsb))
}

// CVResult is parsed from LAN_X_CV_RESULT (spec §6.5).
type CVResult struct {
	Address CVAddress
	Value   byte
}

// ReadCV returns LAN_X_CV_READ (spec §6.1).
func ReadCV(cv CVAddress) Message {
	msb, lsb := encodeCVAddress(cv)
	return Message{
		Header: HeaderLANX,
		Data:   appendLANXXOR([]byte{xHeaderCVRead, xCommandCVRead, msb, lsb}),
	}
}

// WriteCV returns LAN_X_CV_WRITE (spec §6.2).
func WriteCV(cv CVAddress, value byte) Message {
	msb, lsb := encodeCVAddress(cv)
	return Message{
		Header: HeaderLANX,
		Data: appendLANXXOR([]byte{
			xHeaderCVWrite, xCommandCVWrite, msb, lsb, value,
		}),
	}
}

// WriteMMByte returns LAN_X_MM_WRITE_BYTE (spec §6.12).
func WriteMMByte(register byte, value byte) Message {
	return Message{
		Header: HeaderLANX,
		Data: appendLANXXOR([]byte{
			xHeaderCVWrite, xCommandMMWriteByte, 0x00, register, value,
		}),
	}
}

// ReadDCCRegister returns LAN_X_DCC_READ_REGISTER (spec §6.13).
func ReadDCCRegister(reg byte) Message {
	return Message{
		Header: HeaderLANX,
		Data:   appendLANXXOR([]byte{xHeaderDCCReadReg, xCommandDCCReadReg, reg}),
	}
}

// WriteDCCRegister returns LAN_X_DCC_WRITE_REGISTER (spec §6.14).
func WriteDCCRegister(reg, value byte) Message {
	return Message{
		Header: HeaderLANX,
		Data: appendLANXXOR([]byte{
			xHeaderCVRead, xCommandDCCWriteReg, reg, value,
		}),
	}
}

// CVResultFromMessages extracts LAN_X_CV_RESULT from programming replies.
func CVResultFromMessages(msgs []Message) (CVResult, error) {
	for _, msg := range msgs {
		if msg.Header != HeaderLANX {
			continue
		}
		if IsCVNackSC(msg.Data) {
			return CVResult{}, fmt.Errorf("z21: CV programming short circuit")
		}
		if IsCVNack(msg.Data) {
			return CVResult{}, fmt.Errorf("z21: CV programming NACK")
		}
		result, err := ParseCVResult(msg.Data)
		if err != nil {
			continue
		}
		return result, nil
	}
	return CVResult{}, fmt.Errorf("z21: no CV programming reply")
}

// ParseCVResult decodes LAN_X_CV_RESULT (spec §6.5).
func ParseCVResult(data []byte) (CVResult, error) {
	if len(data) < 6 {
		return CVResult{}, fmt.Errorf("z21: CV result too short (%d bytes)", len(data))
	}
	if data[0] != xHeaderCVResult || data[1] != xCommandCVResult {
		return CVResult{}, fmt.Errorf("z21: not a LAN_X_CV_RESULT reply")
	}
	if data[len(data)-1] != lanXXOR(data[:len(data)-1]) {
		return CVResult{}, fmt.Errorf("z21: invalid LAN_X_CV_RESULT checksum")
	}
	return CVResult{
		Address: parseCVAddress(data[2], data[3]),
		Value:   data[4],
	}, nil
}

// IsCVResult reports whether data is LAN_X_CV_RESULT.
func IsCVResult(data []byte) bool {
	_, err := ParseCVResult(data)
	return err == nil
}

// ParseCVNackSC decodes LAN_X_CV_NACK_SC (spec §6.3).
func ParseCVNackSC(data []byte) error {
	return parseXBusBC(data, xBCDB0CVNackSC)
}

// IsCVNackSC reports whether data is LAN_X_CV_NACK_SC.
func IsCVNackSC(data []byte) bool {
	return ParseCVNackSC(data) == nil
}

// ParseCVNack decodes LAN_X_CV_NACK (spec §6.4).
func ParseCVNack(data []byte) error {
	return parseXBusBC(data, xBCDB0CVNack)
}

// IsCVNack reports whether data is LAN_X_CV_NACK.
func IsCVNack(data []byte) bool {
	return ParseCVNack(data) == nil
}
