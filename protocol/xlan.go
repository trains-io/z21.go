package protocol

import "fmt"

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
