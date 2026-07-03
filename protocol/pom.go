package protocol

const (
	xHeaderPOMLoco      byte = 0xE6
	xCommandPOMLoco     byte = 0x30
	xCommandPOMAccessory byte = 0x31

	pomOptionWriteByte byte = 0xEC
	pomOptionWriteBit  byte = 0xE8
	pomOptionReadByte  byte = 0xE4
)

// POMLocoWriteByte returns LAN_X_CV_POM_WRITE_BYTE (spec §6.6).
func POMLocoWriteByte(locoAddress uint16, cv CVAddress, value byte) Message {
	msb, lsb := encodeLocoAddressBytes(locoAddress)
	cvMSB, cvLSB := encodeCVAddress(cv)
	return Message{
		Header: HeaderLANX,
		Data: appendLANXXOR([]byte{
			xHeaderPOMLoco, xCommandPOMLoco, msb, lsb,
			pomOptionWriteByte | (cvMSB & 0x03), cvLSB, value,
		}),
	}
}

// POMLocoWriteBit returns LAN_X_CV_POM_WRITE_BIT (spec §6.7).
func POMLocoWriteBit(locoAddress uint16, cv CVAddress, bitPos uint8, on bool) Message {
	msb, lsb := encodeLocoAddressBytes(locoAddress)
	cvMSB, cvLSB := encodeCVAddress(cv)
	return Message{
		Header: HeaderLANX,
		Data: appendLANXXOR([]byte{
			xHeaderPOMLoco, xCommandPOMLoco, msb, lsb,
			pomOptionWriteBit | (cvMSB & 0x03), cvLSB, encodePOMBitParam(bitPos, on),
		}),
	}
}

// POMLocoReadByte returns LAN_X_CV_POM_READ_BYTE (spec §6.8).
func POMLocoReadByte(locoAddress uint16, cv CVAddress) Message {
	msb, lsb := encodeLocoAddressBytes(locoAddress)
	cvMSB, cvLSB := encodeCVAddress(cv)
	return Message{
		Header: HeaderLANX,
		Data: appendLANXXOR([]byte{
			xHeaderPOMLoco, xCommandPOMLoco, msb, lsb,
			pomOptionReadByte | (cvMSB & 0x03), cvLSB, 0x00,
		}),
	}
}

// POMAccessoryWriteByte returns LAN_X_CV_POM_ACCESSORY_WRITE_BYTE (spec §6.9).
func POMAccessoryWriteByte(decoderAddress uint16, cv CVAddress, value byte, output *uint8) Message {
	db1, db2 := encodeAccessoryPOMAddress(decoderAddress, output)
	cvMSB, cvLSB := encodeCVAddress(cv)
	return Message{
		Header: HeaderLANX,
		Data: appendLANXXOR([]byte{
			xHeaderPOMLoco, xCommandPOMAccessory, db1, db2,
			pomOptionWriteByte | (cvMSB & 0x03), cvLSB, value,
		}),
	}
}

// POMAccessoryWriteBit returns LAN_X_CV_POM_ACCESSORY_WRITE_BIT (spec §6.10).
func POMAccessoryWriteBit(decoderAddress uint16, cv CVAddress, bitPos uint8, on bool, output *uint8) Message {
	db1, db2 := encodeAccessoryPOMAddress(decoderAddress, output)
	cvMSB, cvLSB := encodeCVAddress(cv)
	return Message{
		Header: HeaderLANX,
		Data: appendLANXXOR([]byte{
			xHeaderPOMLoco, xCommandPOMAccessory, db1, db2,
			pomOptionWriteBit | (cvMSB & 0x03), cvLSB, encodePOMBitParam(bitPos, on),
		}),
	}
}

// POMAccessoryReadByte returns LAN_X_CV_POM_ACCESSORY_READ_BYTE (spec §6.11).
func POMAccessoryReadByte(decoderAddress uint16, cv CVAddress, output *uint8) Message {
	db1, db2 := encodeAccessoryPOMAddress(decoderAddress, output)
	cvMSB, cvLSB := encodeCVAddress(cv)
	return Message{
		Header: HeaderLANX,
		Data: appendLANXXOR([]byte{
			xHeaderPOMLoco, xCommandPOMAccessory, db1, db2,
			pomOptionReadByte | (cvMSB & 0x03), cvLSB, 0x00,
		}),
	}
}

func encodePOMBitParam(bitPos uint8, on bool) byte {
	b := bitPos & 0x1F
	if on {
		b |= 0x20
	}
	return b
}

func encodeAccessoryPOMAddress(decoderAddress uint16, output *uint8) (db1, db2 byte) {
	var cddd byte
	if output != nil {
		cddd = 0x08 | (*output & 0x07)
	}
	packed := (uint16(decoderAddress&0x1FF) << 4) | uint16(cddd)
	return byte(packed >> 8), byte(packed)
}
