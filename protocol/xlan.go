package protocol

import "fmt"

// EncodeLANX builds a LAN_X dataset (header 0x40) from an X-header and its
// data bytes, appending the XOR checksum over all of them (spec §1.2).
func EncodeLANX(xHeader byte, data ...byte) Message {
	out := make([]byte, len(data)+2)
	out[0] = xHeader
	copy(out[1:], data)
	out[len(out)-1] = lanXXOR(out[:len(out)-1])
	return Message{Header: HeaderLANX, Data: out}
}

// ParseLANX validates a LAN_X dataset (header, length and XOR checksum) and
// returns its X-header and the data bytes without the checksum (spec §1.2).
// The returned data is a copy and does not alias msg.Data.
func ParseLANX(msg Message) (xHeader byte, data []byte, err error) {
	if msg.Header != HeaderLANX {
		return 0, nil, fmt.Errorf("z21: not a LAN_X dataset (header %#04x)", msg.Header)
	}
	if len(msg.Data) < 2 {
		return 0, nil, fmt.Errorf("z21: LAN_X dataset too short (%d bytes)", len(msg.Data))
	}
	last := len(msg.Data) - 1
	if msg.Data[last] != lanXXOR(msg.Data[:last]) {
		return 0, nil, fmt.Errorf("z21: invalid LAN_X checksum")
	}
	data = make([]byte, last-1)
	copy(data, msg.Data[1:last])
	return msg.Data[0], data, nil
}

// parseLANXData validates LAN_X data (X-header, at least minData data bytes,
// XOR checksum) and returns the data bytes between X-header and checksum.
// The result aliases data.
func parseLANXData(data []byte, xHeader byte, minData int, name string) ([]byte, error) {
	if len(data) < minData+2 {
		return nil, fmt.Errorf("z21: %s too short (%d bytes)", name, len(data))
	}
	if data[0] != xHeader {
		return nil, fmt.Errorf("z21: not a %s (X-header %#02x)", name, data[0])
	}
	last := len(data) - 1
	if data[last] != lanXXOR(data[:last]) {
		return nil, fmt.Errorf("z21: invalid %s checksum", name)
	}
	return data[1:last], nil
}

func lanXXOR(data []byte) byte {
	var x byte
	for _, b := range data {
		x ^= b
	}
	return x
}

func appendLANXXOR(data []byte) []byte {
	out := make([]byte, len(data)+1)
	copy(out, data)
	out[len(data)] = lanXXOR(data)
	return out
}

func encodeLocoAddressBytes(address uint16) (msb, lsb byte) {
	msb = byte((address >> 8) & 0x3F)
	lsb = byte(address & 0xFF)
	if address >= 128 {
		msb |= 0xC0
	}
	return msb, lsb
}

func parseLocoAddressBytes(msb, lsb byte) uint16 {
	return uint16(msb&0x3F)<<8 | uint16(lsb)
}

func encodeFunctionAddressBytes(address uint16) []byte {
	return []byte{byte(address >> 8), byte(address)}
}

func parseFunctionAddressBytes(data []byte) (uint16, error) {
	if len(data) < 2 {
		return 0, fmt.Errorf("z21: function address too short (%d bytes)", len(data))
	}
	return uint16(data[0])<<8 | uint16(data[1]), nil
}
