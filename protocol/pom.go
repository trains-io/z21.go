package protocol

import "fmt"

const (
	xHeaderPOMLoco       byte = 0xE6
	xCommandPOMLoco      byte = 0x30
	xCommandPOMAccessory byte = 0x31

	pomOptionWriteByte byte = 0xEC
	pomOptionWriteBit  byte = 0xE8
	pomOptionReadByte  byte = 0xE4
)

// POMOperation is the option in DB3 of a POM request (spec §6.6–§6.11).
type POMOperation byte

const (
	POMWriteByte POMOperation = POMOperation(pomOptionWriteByte)
	POMWriteBit  POMOperation = POMOperation(pomOptionWriteBit)
	POMReadByte  POMOperation = POMOperation(pomOptionReadByte)
)

// POMRequest is decoded from a LAN_X_CV_POM_* request (spec §6.6–§6.11).
type POMRequest struct {
	// Address is the loco address (ParsePOMLoco) or the accessory decoder
	// address (ParsePOMAccessory).
	Address uint16
	// HasOutput and Output select a single accessory output; loco requests
	// never set them.
	HasOutput bool
	Output    uint8
	CV        CVAddress
	Operation POMOperation
	// Value is set for POMWriteByte.
	Value byte
	// Bit and BitValue are set for POMWriteBit.
	Bit      uint8
	BitValue bool
}

// ParsePOMLoco decodes LAN_X_CV_POM_WRITE_BYTE, LAN_X_CV_POM_WRITE_BIT and
// LAN_X_CV_POM_READ_BYTE (spec §6.6–§6.8).
func ParsePOMLoco(data []byte) (POMRequest, error) {
	req, d, err := parsePOM(data, xCommandPOMLoco, "LAN_X_CV_POM")
	if err != nil {
		return POMRequest{}, err
	}
	req.Address = parseLocoAddressBytes(d[1], d[2])
	return req, nil
}

// ParsePOMAccessory decodes LAN_X_CV_POM_ACCESSORY_WRITE_BYTE, _WRITE_BIT and
// _READ_BYTE (spec §6.9–§6.11).
func ParsePOMAccessory(data []byte) (POMRequest, error) {
	req, d, err := parsePOM(data, xCommandPOMAccessory, "LAN_X_CV_POM_ACCESSORY")
	if err != nil {
		return POMRequest{}, err
	}
	packed := uint16(d[1])<<8 | uint16(d[2])
	req.Address = (packed >> 4) & 0x1FF
	req.HasOutput = packed&0x08 != 0
	if req.HasOutput {
		req.Output = uint8(packed & 0x07)
	}
	return req, nil
}

func parsePOM(data []byte, db0 byte, name string) (POMRequest, []byte, error) {
	d, err := parseLANXData(data, xHeaderPOMLoco, 6, name)
	if err != nil {
		return POMRequest{}, nil, err
	}
	if d[0] != db0 {
		return POMRequest{}, nil, fmt.Errorf("z21: not a %s (DB0 %#02x)", name, d[0])
	}
	req := POMRequest{
		Operation: POMOperation(d[3] & 0xFC),
		CV:        CVAddress(uint16(d[3]&0x03)<<8 | uint16(d[4])),
	}
	switch req.Operation {
	case POMWriteByte:
		req.Value = d[5]
	case POMWriteBit:
		req.Bit = d[5] & 0x07
		req.BitValue = d[5]&0x08 != 0
	case POMReadByte:
	default:
		return POMRequest{}, nil, fmt.Errorf("z21: unknown %s option %#02x", name, d[3]&0xFC)
	}
	return req, d, nil
}

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

// encodePOMBitParam returns the 0000VPPP byte of the POM write-bit commands (spec §6.7 / §6.10).
func encodePOMBitParam(bitPos uint8, on bool) byte {
	b := bitPos & 0x07
	if on {
		b |= 0x08
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
