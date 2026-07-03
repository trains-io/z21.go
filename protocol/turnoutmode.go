package protocol

import "fmt"

// GetTurnoutMode returns a request for the output format of an accessory decoder address (spec §3.3).
func GetTurnoutMode(address uint16) Message {
	return Message{
		Header: HeaderLANGetTurnoutMode,
		Data:   encodeAddressBE(address),
	}
}

// SetTurnoutMode sets the persisted output format for an accessory decoder address (spec §3.4).
// Addresses >= 256 remain DCC regardless of mode. MM accessories require Z21 FW 1.20+.
func SetTurnoutMode(address uint16, mode byte) Message {
	data := encodeAddressBE(address)
	data = append(data, mode)
	return Message{
		Header: HeaderLANSetTurnoutMode,
		Data:   data,
	}
}

// TurnoutModeFromMessages extracts a LAN_GET_TURNOUTMODE reply.
func TurnoutModeFromMessages(msgs []Message) (AddressMode, error) {
	for _, msg := range msgs {
		if msg.Header != HeaderLANGetTurnoutMode {
			continue
		}
		mode, err := ParseTurnoutMode(msg.Data)
		if err != nil {
			continue
		}
		return mode, nil
	}
	return AddressMode{}, fmt.Errorf("z21: no LAN_GET_TURNOUTMODE reply")
}

// ParseTurnoutMode decodes a LAN_GET_TURNOUTMODE reply payload.
func ParseTurnoutMode(data []byte) (AddressMode, error) {
	return parseAddressModeReply(data)
}
