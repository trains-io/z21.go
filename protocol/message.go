package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Message is a single Z21 LAN dataset: Header + payload bytes.
// On the wire each dataset is prefixed with a 16-bit little-endian DataLen
// covering Header (2 bytes) and Data.
type Message struct {
	Header uint16
	Data   []byte
}

// Marshal encodes m as one Z21 dataset (DataLen + Header + Data).
// DataLen covers the entire dataset including the length field (spec §1.2).
func (m Message) Marshal() ([]byte, error) {
	if len(m.Data) > 0xFFFC-4 {
		return nil, fmt.Errorf("z21: dataset data too large (%d bytes)", len(m.Data))
	}

	dataLen := uint16(4 + len(m.Data))
	out := make([]byte, dataLen)
	binary.LittleEndian.PutUint16(out[0:2], dataLen)
	binary.LittleEndian.PutUint16(out[2:4], m.Header)
	copy(out[4:], m.Data)
	return out, nil
}

// MarshalAll encodes multiple datasets into one UDP payload (spec §1.3).
func MarshalAll(msgs ...Message) ([]byte, error) {
	if len(msgs) == 0 {
		return nil, fmt.Errorf("z21: no datasets to marshal")
	}

	var out []byte
	for i, msg := range msgs {
		encoded, err := msg.Marshal()
		if err != nil {
			return nil, fmt.Errorf("z21: marshal dataset %d: %w", i, err)
		}
		out = append(out, encoded...)
	}
	return out, nil
}

// Unmarshal decodes one dataset from b and returns any trailing bytes.
func Unmarshal(b []byte) (Message, []byte, error) {
	if len(b) < 4 {
		return Message{}, b, errors.New("z21: dataset too short")
	}

	dataLen := binary.LittleEndian.Uint16(b[0:2])
	if dataLen < 4 {
		return Message{}, b, fmt.Errorf("z21: invalid data length %d", dataLen)
	}

	total := int(dataLen)
	if len(b) < total {
		if len(b) < 4 {
			return Message{}, b, fmt.Errorf("z21: dataset truncated (need %d bytes, have %d)", total, len(b))
		}
		// Some implementations (e.g. Z21Posix) declare a longer DataLen than they send.
		total = len(b)
	}

	payloadLen := total - 4
	msg := Message{
		Header: binary.LittleEndian.Uint16(b[2:4]),
		Data:   nil,
	}
	if payloadLen > 0 {
		msg.Data = make([]byte, payloadLen)
		copy(msg.Data, b[4:total])
	}

	return msg, b[total:], nil
}

// ParseAll decodes every dataset in a UDP payload.
func ParseAll(b []byte) ([]Message, error) {
	var msgs []Message
	for len(b) > 0 {
		msg, rest, err := Unmarshal(b)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
		b = rest
	}
	return msgs, nil
}

// GetHWInfo returns a request for hardware type and firmware version (spec §2.20).
func GetHWInfo() Message {
	return Message{Header: HeaderLANGetHWInfo}
}

// Logoff ends the client session with the command station (spec §2.4).
func Logoff() Message {
	return Message{Header: HeaderLANLogoff}
}

// SetBroadcastFlags subscribes the client to broadcast groups (spec §2.16).
func SetBroadcastFlags(flags uint32) Message {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, flags)
	return Message{Header: HeaderLANSetBroadcastFlags, Data: data}
}
