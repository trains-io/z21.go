package z21

import (
	"encoding/hex"
	"fmt"
	"hash/fnv"
)

func decodeBCD(b byte) (int, int, error) {
	high := int(b >> 4)
	low := int(b & 0x0F)

	if high > 9 || low > 9 {
		return 0, 0, fmt.Errorf("invalid BCD digit in byte: 0x%x", b)
	}

	return high, low, nil
}

func fingerprint(data []byte) (string, error) {
	h := fnv.New32a()
	if _, err := h.Write(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
