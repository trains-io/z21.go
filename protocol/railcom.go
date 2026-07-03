package protocol

import (
	"encoding/binary"
	"fmt"
)

const (
	// RailComGetDataTypeLoco polls RailCom data for a locomotive (spec §8.2).
	RailComGetDataTypeLoco byte = 0x01

	// RailComPollNextAddress requests the next loco in the circular buffer (spec §8.2).
	RailComPollNextAddress uint16 = 0x0000

	RailComOptionSpeed1 byte = 0x01
	RailComOptionSpeed2 byte = 0x02
	RailComOptionQoS    byte = 0x04

	railComDataMinLen = 13
)

// RailComData is a LAN_RAILCOM_DATACHANGED payload (spec §8.1).
type RailComData struct {
	LocoAddress    uint16
	ReceiveCounter uint32
	ErrorCounter   uint16
	Options        byte
	Speed          byte
	QoS            byte
}

// GetRailComData returns LAN_RAILCOM_GETDATA (spec §8.2).
// locoAddress 0 polls the next loco in the Z21 circular buffer.
func GetRailComData(locoAddress uint16) Message {
	data := make([]byte, 3)
	data[0] = RailComGetDataTypeLoco
	binary.LittleEndian.PutUint16(data[1:3], locoAddress)
	return Message{Header: HeaderLANRailComGetData, Data: data}
}

// HasRailComOption reports whether opt includes flag.
func HasRailComOption(opt, flag byte) bool {
	return opt&flag != 0
}

// ParseRailComData decodes LAN_RAILCOM_DATACHANGED (spec §8.1).
// Future Z21 versions may extend the structure; callers should rely on DataLen.
func ParseRailComData(data []byte) (RailComData, error) {
	if len(data) < railComDataMinLen {
		return RailComData{}, fmt.Errorf("z21: RailCom data too short (%d bytes)", len(data))
	}
	return RailComData{
		LocoAddress:    binary.LittleEndian.Uint16(data[0:2]),
		ReceiveCounter: binary.LittleEndian.Uint32(data[2:6]),
		ErrorCounter:   binary.LittleEndian.Uint16(data[6:8]),
		Options:        data[9],
		Speed:          data[10],
		QoS:            data[11],
	}, nil
}

// RailComDataFromMessages extracts LAN_RAILCOM_DATACHANGED replies.
func RailComDataFromMessages(msgs []Message) ([]RailComData, error) {
	var out []RailComData
	for _, msg := range msgs {
		if msg.Header != HeaderLANRailComDataChanged && msg.Header != HeaderLANRailComGetData {
			continue
		}
		rc, err := ParseRailComData(msg.Data)
		if err != nil {
			return nil, err
		}
		out = append(out, rc)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("z21: no LAN_RAILCOM_DATACHANGED reply")
	}
	return out, nil
}

// IsRailComDataChanged reports whether msg is a RailCom update.
func IsRailComDataChanged(msg Message) bool {
	return msg.Header == HeaderLANRailComDataChanged
}

// FormatRailComOptions renders active RailCom option flags.
func FormatRailComOptions(opt byte) string {
	var flags []string
	if HasRailComOption(opt, RailComOptionSpeed1) {
		flags = append(flags, "Speed1")
	}
	if HasRailComOption(opt, RailComOptionSpeed2) {
		flags = append(flags, "Speed2")
	}
	if HasRailComOption(opt, RailComOptionQoS) {
		flags = append(flags, "QoS")
	}
	if len(flags) == 0 {
		return "none"
	}
	out := flags[0]
	for _, f := range flags[1:] {
		out += "|" + f
	}
	return out
}
