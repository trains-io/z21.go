package protocol

import "fmt"

// GetLocoMode returns a request for the output format of a locomotive address (spec §3.1).
func GetLocoMode(address uint16) Message {
	return Message{
		Header: HeaderLANGetLocoMode,
		Data:   encodeAddressBE(address),
	}
}

// SetLocoMode sets the persisted output format for a locomotive address (spec §3.2).
// Addresses >= 256 remain DCC regardless of mode.
func SetLocoMode(address uint16, mode byte) Message {
	data := encodeAddressBE(address)
	data = append(data, mode)
	return Message{
		Header: HeaderLANSetLocoMode,
		Data:   data,
	}
}

// LocoModeFromMessages extracts a LAN_GET_LOCOMODE reply.
func LocoModeFromMessages(msgs []Message) (AddressMode, error) {
	for _, msg := range msgs {
		if msg.Header != HeaderLANGetLocoMode {
			continue
		}
		mode, err := ParseLocoMode(msg.Data)
		if err != nil {
			continue
		}
		return mode, nil
	}
	return AddressMode{}, fmt.Errorf("z21: no LAN_GET_LOCOMODE reply")
}

// ParseLocoMode decodes a LAN_GET_LOCOMODE reply payload.
func ParseLocoMode(data []byte) (AddressMode, error) {
	return parseAddressModeReply(data)
}
