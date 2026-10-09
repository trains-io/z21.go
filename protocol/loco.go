package protocol

import "fmt"

const (
	xHeaderGetLocoInfo      byte = 0xE3
	xCommandGetLocoInfo     byte = 0xF0
	xCommandPurgeLoco       byte = 0x44
	xHeaderSetLocoDrive     byte = 0xE4
	xCommandSetLocoFunction byte = 0xF8
	xHeaderSetLocoBinary    byte = 0xE5
	xCommandSetLocoBinary   byte = 0x5F
	xHeaderLocoInfo         byte = 0xEF
	xHeaderSetLocoEStop     byte = 0x92

	// LocoSpeedSteps values for the S nibble in 0x1S (spec §4.2).
	LocoSpeedSteps14  byte = 0x00
	LocoSpeedSteps28  byte = 0x02
	LocoSpeedSteps128 byte = 0x03
)

// LocoFunctionAction is TT in LAN_X_SET_LOCO_FUNCTION (spec §4.3.1).
type LocoFunctionAction byte

const (
	LocoFunctionOff    LocoFunctionAction = 0x00
	LocoFunctionOn     LocoFunctionAction = 0x01
	LocoFunctionToggle LocoFunctionAction = 0x02
)

// LocoFunctionGroup selects a function group in LAN_X_SET_LOCO_FUNCTION_GROUP (spec §4.3.2).
type LocoFunctionGroup byte

const (
	LocoFunctionGroupF0F4   LocoFunctionGroup = 0x20
	LocoFunctionGroupF5F8   LocoFunctionGroup = 0x21
	LocoFunctionGroupF9F12  LocoFunctionGroup = 0x22
	LocoFunctionGroupF13F20 LocoFunctionGroup = 0x23
	LocoFunctionGroupF21F28 LocoFunctionGroup = 0x28
	LocoFunctionGroupF29F36 LocoFunctionGroup = 0x29
	LocoFunctionGroupF37F44 LocoFunctionGroup = 0x2A
	LocoFunctionGroupF45F52 LocoFunctionGroup = 0x2B
	LocoFunctionGroupF53F60 LocoFunctionGroup = 0x50
	LocoFunctionGroupF61F68 LocoFunctionGroup = 0x51
)

// LocoInfo is parsed from LAN_X_LOCO_INFO (spec §4.4).
type LocoInfo struct {
	Address         uint16
	Busy            bool
	MMFormat        bool
	SpeedSteps      byte
	Forward         bool
	Speed           byte
	DoubleTraction  bool
	SmartSearch     bool
	Headlight       bool
	FunctionsF1F4   byte
	FunctionsF5F12  byte
	FunctionsF13F20 byte
	FunctionsF21F28 byte
	FunctionsF29F31 byte
}

// GetLocoInfo returns LAN_X_GET_LOCO_INFO and subscribes to loco updates (spec §4.1).
func GetLocoInfo(address uint16) Message {
	msb, lsb := encodeLocoAddressBytes(address)
	return Message{
		Header: HeaderLANX,
		Data:   appendLANXXOR([]byte{xHeaderGetLocoInfo, xCommandGetLocoInfo, msb, lsb}),
	}
}

// SetLocoDrive returns LAN_X_SET_LOCO_DRIVE (spec §4.2).
func SetLocoDrive(address uint16, speedSteps byte, forward bool, speed byte) Message {
	msb, lsb := encodeLocoAddressBytes(address)
	speedByte := speed & 0x7F
	if forward {
		speedByte |= 0x80
	}
	return Message{
		Header: HeaderLANX,
		Data: appendLANXXOR([]byte{
			xHeaderSetLocoDrive,
			0x10 | (speedSteps & 0x0F),
			msb,
			lsb,
			speedByte,
		}),
	}
}

// SetLocoFunction returns LAN_X_SET_LOCO_FUNCTION (spec §4.3.1).
func SetLocoFunction(address uint16, function uint8, action LocoFunctionAction) Message {
	msb, lsb := encodeLocoAddressBytes(address)
	return Message{
		Header: HeaderLANX,
		Data: appendLANXXOR([]byte{
			xHeaderSetLocoDrive,
			xCommandSetLocoFunction,
			msb,
			lsb,
			byte(action<<6) | (function & 0x3F),
		}),
	}
}

// SetLocoFunctionGroup returns LAN_X_SET_LOCO_FUNCTION_GROUP (spec §4.3.2).
func SetLocoFunctionGroup(address uint16, group LocoFunctionGroup, functions byte) Message {
	msb, lsb := encodeLocoAddressBytes(address)
	return Message{
		Header: HeaderLANX,
		Data: appendLANXXOR([]byte{
			xHeaderSetLocoDrive,
			byte(group),
			msb,
			lsb,
			functions,
		}),
	}
}

// SetLocoBinaryState returns LAN_X_SET_LOCO_BINARY_STATE (spec §4.3.3).
func SetLocoBinaryState(address uint16, binaryAddress uint16, on bool) Message {
	msb, lsb := encodeLocoAddressBytes(address)
	low := byte(binaryAddress & 0x7F)
	high := byte(binaryAddress >> 7)
	if on {
		low |= 0x80
	}
	return Message{
		Header: HeaderLANX,
		Data: appendLANXXOR([]byte{
			xHeaderSetLocoBinary,
			xCommandSetLocoBinary,
			msb,
			lsb,
			low,
			high,
		}),
	}
}

// SetLocoEStop returns LAN_X_SET_LOCO_E_STOP (spec §4.5).
func SetLocoEStop(address uint16) Message {
	msb, lsb := encodeLocoAddressBytes(address)
	return Message{
		Header: HeaderLANX,
		Data:   appendLANXXOR([]byte{xHeaderSetLocoEStop, msb, lsb}),
	}
}

// PurgeLoco returns LAN_X_PURGE_LOCO (spec §4.6).
func PurgeLoco(address uint16) Message {
	msb, lsb := encodeLocoAddressBytes(address)
	return Message{
		Header: HeaderLANX,
		Data:   appendLANXXOR([]byte{xHeaderGetLocoInfo, xCommandPurgeLoco, msb, lsb}),
	}
}

// LocoDrive is decoded from LAN_X_SET_LOCO_DRIVE (spec §4.2).
// SpeedSteps is one of the LocoSpeedSteps* values.
type LocoDrive struct {
	Address    uint16
	SpeedSteps byte
	Forward    bool
	Speed      byte
}

// ParseGetLocoInfo decodes the address from LAN_X_GET_LOCO_INFO (spec §4.1).
func ParseGetLocoInfo(data []byte) (address uint16, err error) {
	d, err := parseLANXData(data, xHeaderGetLocoInfo, 3, "LAN_X_GET_LOCO_INFO")
	if err != nil {
		return 0, err
	}
	if d[0] != xCommandGetLocoInfo {
		return 0, fmt.Errorf("z21: not a LAN_X_GET_LOCO_INFO (DB0 %#02x)", d[0])
	}
	return parseLocoAddressBytes(d[1], d[2]), nil
}

// ParseSetLocoDrive decodes LAN_X_SET_LOCO_DRIVE (spec §4.2).
func ParseSetLocoDrive(data []byte) (LocoDrive, error) {
	d, err := parseLANXData(data, xHeaderSetLocoDrive, 4, "LAN_X_SET_LOCO_DRIVE")
	if err != nil {
		return LocoDrive{}, err
	}
	steps := d[0] & 0x0F
	if d[0]&0xF0 != 0x10 || (steps != LocoSpeedSteps14 && steps != LocoSpeedSteps28 && steps != LocoSpeedSteps128) {
		return LocoDrive{}, fmt.Errorf("z21: not a LAN_X_SET_LOCO_DRIVE (DB0 %#02x)", d[0])
	}
	return LocoDrive{
		Address:    parseLocoAddressBytes(d[1], d[2]),
		SpeedSteps: steps,
		Forward:    d[3]&0x80 != 0,
		Speed:      d[3] & 0x7F,
	}, nil
}

// ParseSetLocoFunction decodes LAN_X_SET_LOCO_FUNCTION (spec §4.3.1).
func ParseSetLocoFunction(data []byte) (address uint16, function uint8, action LocoFunctionAction, err error) {
	d, err := parseLANXData(data, xHeaderSetLocoDrive, 4, "LAN_X_SET_LOCO_FUNCTION")
	if err != nil {
		return 0, 0, 0, err
	}
	if d[0] != xCommandSetLocoFunction {
		return 0, 0, 0, fmt.Errorf("z21: not a LAN_X_SET_LOCO_FUNCTION (DB0 %#02x)", d[0])
	}
	action = LocoFunctionAction(d[3] >> 6)
	if action > LocoFunctionToggle {
		return 0, 0, 0, fmt.Errorf("z21: invalid LAN_X_SET_LOCO_FUNCTION action %d", action)
	}
	return parseLocoAddressBytes(d[1], d[2]), d[3] & 0x3F, action, nil
}

// ParseSetLocoFunctionGroup decodes LAN_X_SET_LOCO_FUNCTION_GROUP (spec §4.3.2).
func ParseSetLocoFunctionGroup(data []byte) (address uint16, group LocoFunctionGroup, functions byte, err error) {
	d, err := parseLANXData(data, xHeaderSetLocoDrive, 4, "LAN_X_SET_LOCO_FUNCTION_GROUP")
	if err != nil {
		return 0, 0, 0, err
	}
	group = LocoFunctionGroup(d[0])
	switch group {
	case LocoFunctionGroupF0F4, LocoFunctionGroupF5F8, LocoFunctionGroupF9F12, LocoFunctionGroupF13F20,
		LocoFunctionGroupF21F28, LocoFunctionGroupF29F36, LocoFunctionGroupF37F44, LocoFunctionGroupF45F52,
		LocoFunctionGroupF53F60, LocoFunctionGroupF61F68:
	default:
		return 0, 0, 0, fmt.Errorf("z21: not a LAN_X_SET_LOCO_FUNCTION_GROUP (DB0 %#02x)", d[0])
	}
	return parseLocoAddressBytes(d[1], d[2]), group, d[3], nil
}

// ParseSetLocoBinaryState decodes LAN_X_SET_LOCO_BINARY_STATE (spec §4.3.3).
func ParseSetLocoBinaryState(data []byte) (address, binaryAddress uint16, on bool, err error) {
	d, err := parseLANXData(data, xHeaderSetLocoBinary, 5, "LAN_X_SET_LOCO_BINARY_STATE")
	if err != nil {
		return 0, 0, false, err
	}
	if d[0] != xCommandSetLocoBinary {
		return 0, 0, false, fmt.Errorf("z21: not a LAN_X_SET_LOCO_BINARY_STATE (DB0 %#02x)", d[0])
	}
	binaryAddress = uint16(d[4])<<7 | uint16(d[3]&0x7F)
	return parseLocoAddressBytes(d[1], d[2]), binaryAddress, d[3]&0x80 != 0, nil
}

// ParseSetLocoEStop decodes the address from LAN_X_SET_LOCO_E_STOP (spec §4.5).
func ParseSetLocoEStop(data []byte) (address uint16, err error) {
	d, err := parseLANXData(data, xHeaderSetLocoEStop, 2, "LAN_X_SET_LOCO_E_STOP")
	if err != nil {
		return 0, err
	}
	return parseLocoAddressBytes(d[0], d[1]), nil
}

// ParsePurgeLoco decodes the address from LAN_X_PURGE_LOCO (spec §4.6).
func ParsePurgeLoco(data []byte) (address uint16, err error) {
	d, err := parseLANXData(data, xHeaderGetLocoInfo, 3, "LAN_X_PURGE_LOCO")
	if err != nil {
		return 0, err
	}
	if d[0] != xCommandPurgeLoco {
		return 0, fmt.Errorf("z21: not a LAN_X_PURGE_LOCO (DB0 %#02x)", d[0])
	}
	return parseLocoAddressBytes(d[1], d[2]), nil
}

// EncodeLocoInfo builds LAN_X_LOCO_INFO with DB0–DB8 (F0–F31), the layout
// sent by firmware 1.42+ and the ZIMO reference server (spec §4.4).
// SpeedSteps is the raw KKK field (0 = 14, 2 = 28, 4 = 128 steps), not the
// LocoSpeedSteps* values used by SetLocoDrive. Fields wider than their wire
// bits are masked.
func EncodeLocoInfo(info LocoInfo) Message {
	msb, lsb := encodeLocoAddressBytes(info.Address)

	db2 := info.SpeedSteps & 0x07
	if info.Busy {
		db2 |= 0x08
	}
	if info.MMFormat {
		db2 |= 0x10
	}

	db3 := info.Speed & 0x7F
	if info.Forward {
		db3 |= 0x80
	}

	db4 := info.FunctionsF1F4 & 0x0F
	if info.Headlight {
		db4 |= 0x10
	}
	if info.SmartSearch {
		db4 |= 0x20
	}
	if info.DoubleTraction {
		db4 |= 0x40
	}

	return EncodeLANX(xHeaderLocoInfo,
		msb, lsb, db2, db3, db4,
		info.FunctionsF5F12,
		info.FunctionsF13F20,
		info.FunctionsF21F28,
		info.FunctionsF29F31&0x07,
	)
}

// LocoInfoFromMessages extracts LAN_X_LOCO_INFO from a Call reply or broadcast.
func LocoInfoFromMessages(msgs []Message) (LocoInfo, error) {
	for _, msg := range msgs {
		if msg.Header != HeaderLANX {
			continue
		}
		info, err := ParseLocoInfo(msg.Data)
		if err != nil {
			continue
		}
		return info, nil
	}
	return LocoInfo{}, fmt.Errorf("z21: no LAN_X_LOCO_INFO reply")
}

// ParseLocoInfo decodes LAN_X_LOCO_INFO (spec §4.4).
func ParseLocoInfo(data []byte) (LocoInfo, error) {
	if len(data) < 7 {
		return LocoInfo{}, fmt.Errorf("z21: loco info too short (%d bytes)", len(data))
	}
	if data[0] != xHeaderLocoInfo {
		return LocoInfo{}, fmt.Errorf("z21: not a LAN_X_LOCO_INFO reply")
	}
	if data[len(data)-1] != lanXXOR(data[:len(data)-1]) {
		return LocoInfo{}, fmt.Errorf("z21: invalid LAN_X_LOCO_INFO checksum")
	}

	info := LocoInfo{
		Address:        parseLocoAddressBytes(data[1], data[2]),
		Busy:           data[3]&0x08 != 0,
		MMFormat:       data[3]&0x10 != 0,
		SpeedSteps:     (data[3] >> 0) & 0x07,
		Forward:        data[4]&0x80 != 0,
		Speed:          data[4] & 0x7F,
		DoubleTraction: data[5]&0x40 != 0,
		SmartSearch:    data[5]&0x20 != 0,
		Headlight:      data[5]&0x10 != 0,
		FunctionsF1F4:  data[5] & 0x0F,
	}
	if len(data) >= 8 {
		info.FunctionsF5F12 = data[6]
	}
	if len(data) >= 9 {
		info.FunctionsF13F20 = data[7]
	}
	if len(data) >= 10 {
		info.FunctionsF21F28 = data[8]
	}
	if len(data) >= 11 {
		info.FunctionsF29F31 = data[9] & 0x07
	}
	return info, nil
}

// IsLocoInfo reports whether data is LAN_X_LOCO_INFO.
func IsLocoInfo(data []byte) bool {
	_, err := ParseLocoInfo(data)
	return err == nil
}
