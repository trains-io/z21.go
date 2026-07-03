package protocol

import "fmt"

const (
	xCommandSetTrackPowerOff byte = 0x80
	xCommandSetTrackPowerOn  byte = 0x81
	xHeaderXCommand          byte = 0x21
)

// SetTrackPower returns LAN_X_SET_TRACK_POWER_ON or LAN_X_SET_TRACK_POWER_OFF (spec §2.5 / §2.6).
func SetTrackPower(on bool) Message {
	cmd := xCommandSetTrackPowerOff
	if on {
		cmd = xCommandSetTrackPowerOn
	}
	return Message{
		Header: HeaderLANX,
		Data:   []byte{xHeaderXCommand, cmd, xHeaderXCommand ^ cmd},
	}
}

// TrackPowerFromMessages extracts the track power state from a Call reply.
func TrackPowerFromMessages(msgs []Message) (bool, error) {
	for _, msg := range msgs {
		if msg.Header != HeaderLANX {
			continue
		}
		on, err := ParseTrackPowerBC(msg.Data)
		if err != nil {
			continue
		}
		return on, nil
	}
	return false, fmt.Errorf("z21: no LAN_X_BC_TRACK_POWER reply")
}

// ParseTrackPowerBC decodes LAN_X_BC_TRACK_POWER_ON/OFF (spec §2.7 / §2.8).
func ParseTrackPowerBC(data []byte) (bool, error) {
	if err := parseXBusBC(data, xBCDB0TrackPowerOff); err == nil {
		return false, nil
	}
	if err := parseXBusBC(data, xBCDB0TrackPowerOn); err == nil {
		return true, nil
	}
	if len(data) < 2 || data[0] != xHeaderXBusBC {
		return false, fmt.Errorf("z21: not a LAN_X_BC_TRACK_POWER reply")
	}
	if data[1] != xBCDB0TrackPowerOff && data[1] != xBCDB0TrackPowerOn {
		return false, fmt.Errorf("z21: not a LAN_X_BC_TRACK_POWER reply")
	}
	return data[1] != xBCDB0TrackPowerOff, nil
}

// IsTrackPowerBC reports whether data is LAN_X_BC_TRACK_POWER_ON or OFF.
func IsTrackPowerBC(data []byte) bool {
	_, err := ParseTrackPowerBC(data)
	return err == nil
}
