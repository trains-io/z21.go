package z21

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Message defines methods to serialize and deserialize Z21 LAN messages
// to and from their binary wire representation.
type Message interface {
	// Pack encodes the message into its protocol-specific binary form.
	Pack() ([]byte, error)

	// Unpack decodes the provided binary data into the message struct.
	Unpack([]byte) error

	// EncapType returns the protocol encapsulation type used to identify
	// the kind of encapsulation of this message on the wire.
	EncapType() uint16
}

// Correlatable defines how to correlate requests and responses.
type Correlatable interface {
	// Unique key that identifies a pair of message request and response.
	Key() (string, bool)
}

type Serializable interface {
	Message
	Correlatable
}

func PackFields(fields ...any) ([]byte, error) {
	buf := new(bytes.Buffer)
	for _, f := range fields {
		if err := binary.Write(buf, binary.LittleEndian, f); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func UnpackFields(data []byte, fields ...any) error {
	buf := bytes.NewReader(data)
	for _, f := range fields {
		if err := binary.Read(buf, binary.LittleEndian, f); err != nil {
			return err
		}
	}
	return nil
}

type Frame struct {
	Length  uint16
	Header  uint16
	Payload []byte
}

func (f *Frame) Name() string {
	switch f.Header {
	case LAN_GET_SERIAL_NUMBER:
		return "LAN_GET_SERIAL_NUMBER"
	case LAN_GET_CODE:
		return "LAN_GET_CODE"
	case LAN_GET_HWINFO:
		return "LAN_GET_HWINFO"
	case LAN_LOGOFF:
		return "LAN_LOGOFF"
	case LAN_X:
		xHeader := f.Payload[0]
		switch xHeader {
		case LAN_X_GET_FIRMWARE_VERSION:
			return "LAN_X_GET_FIRMWARE_VERSION"
		case LAN_X_21:
			db0 := f.Payload[1]
			switch db0 {
			case LAN_X_GET_VERSION:
				return "LAN_X_GET_VERSION"
			case LAN_X_GET_STATUS:
				return "LAN_X_GET_STATUS"
			case LAN_X_SET_TRACK_POWER_OFF:
				return "LAN_X_SET_TRACK_POWER_OFF"
			case LAN_X_SET_TRACK_POWER_ON:
				return "LAN_X_SET_TRACK_POWER_ON"
			default:
				return fmt.Sprintf("UNKNOWN DB0: (%02x)", db0)
			}
		case LAN_X_DCC_READ_REGISTER:
			return "LAN_X_DCC_READ_REGISTER"
		case LAN_X_23:
			db0 := f.Payload[1]
			switch db0 {
			case LAN_X_CV_READ:
				return "LAN_X_CV_READ"
			case LAN_X_DCC_WRITE_REGISTER:
				return "LAN_X_DCC_WRITE_REGISTER"
			default:
				return fmt.Sprintf("UNKNOWN DB0: (%02x)", db0)
			}
		case LAN_X_24:
			db0 := f.Payload[1]
			switch db0 {
			case LAN_X_CV_WRITE:
				return "LAN_X_CV_WRITE"
			case LAN_X_MM_WRITE_BYTE:
				return "LAN_X_MM_WRITE_BYTE"
			default:
				return fmt.Sprintf("UNKNOWN DB0: (%02x)", db0)
			}
		case LAN_X_GET_TURNOUT_INFO:
			return "LAN_X_GET_TURNOUT_INFO"
		case LAN_X_GET_EXT_ACCESSORY_INFO:
			return "LAN_X_GET_EXT_ACCESSORY_INFO"
		case LAN_X_SET_TURNOUT:
			return "LAN_X_SET_TURNOUT"
		case LAN_X_SET_EXT_ACCESSORY:
			return "LAN_X_SET_EXT_ACCESSORY"
		case LAN_X_61:
			db0 := f.Payload[1]
			switch db0 {
			case LAN_X_BC_TRACK_POWER_OFF:
				return "LAN_X_BC_TRACK_POWER_OFF"
			case LAN_X_BC_TRACK_POWER_ON:
				return "LAN_X_BC_TRACK_POWER_ON"
			case LAN_X_BC_PROGRAMMING_MODE:
				return "LAN_X_BC_PROGRAMMING_MODE"
			case LAN_X_BC_TRACK_SHORT_CIRCUIT:
				return "LAN_X_BC_TRACK_SHORT_CIRCUIT"
			case LAN_X_CV_NACK_SC:
				return "LAN_X_CV_NACK_SC"
			case LAN_X_CV_NACK:
				return "LAN_X_CV_NACK"
			case LAN_X_UNKNOWN_COMMAND:
				return "LAN_X_UNKNOWN_COMMAND"
			default:
				return fmt.Sprintf("UNKNOWN DB0: (%02x)", db0)
			}
		case LAN_X_STATUS_CHANGED:
			return "LAN_X_STATUS_CHANGED"
		case LAN_X_63:
			db0 := f.Payload[1]
			switch db0 {
			case LAN_X_GET_VERSION:
				return "LAN_X_GET_VERSION"
			default:
				return fmt.Sprintf("UNKNOWN DB0: (%02x)", db0)
			}
		case LAN_X_CV_RESULT:
			return "LAN_X_CV_RESULT"
		case LAN_X_SET_STOP:
			return "LAN_X_SET_STOP"
		case LAN_X_BC_STOPPED:
			return "LAN_X_BC_STOPPED"
		case LAN_X_SET_LOCO_E_STOP:
			return "LAN_X_SET_LOCO_E_STOP"
		case LAN_X_E3:
			db0 := f.Payload[1]
			switch db0 {
			case LAN_X_PURGE_LOCO:
				return "LAN_X_PURGE_LOCO"
			case LAN_X_GET_LOCO_INFO:
				return "LAN_X_GET_LOCO_INFO"
			default:
				return fmt.Sprintf("UNKNOWN DB0: (%02x)", db0)
			}
		case LAN_X_E4:
			db0 := f.Payload[1]
			switch db0 {
			case LAN_X_SET_LOCO_FUNCTION:
				return "LAN_X_SET_LOCO_FUNCTION"
			case LAN_X_SET_LOCO_BINARY_STATE:
				return "LAN_X_SET_LOCO_BINARY_STATE"
			}
			switch {
			case isLocoDriveSpeedStep(db0):
				return "LAN_X_SET_LOCO_DRIVE"
			case isLocoFunctionGroup(db0):
				return "LAN_X_SET_LOCO_FUNCTION_GROUP"
			default:
				return fmt.Sprintf("UNKNOWN DB0: (%02x)", db0)
			}
		case LAN_X_E6:
			db0 := f.Payload[1]
			switch db0 {
			case LAN_X_E6_30:
				db3 := f.Payload[4]
				switch db3 {
				case LAN_X_CV_POM_WRITE_BYTE:
					return "LAN_X_CV_POM_WRITE_BYTE"
				case LAN_X_CV_POM_WRITE_BIT:
					return "LAN_X_CV_POM_WRITE_BIT"
				case LAN_X_CV_POM_READ_BYTE:
					return "LAN_X_CV_POM_READ_BYTE"
				default:
					return fmt.Sprintf("UNKNOWN DB3: (%02x)", db3)
				}
			case LAN_X_E6_31:
				db3 := f.Payload[4]
				switch db3 {
				case LAN_X_CV_POM_ACCESSORY_WRITE_BYTE:
					return "LAN_X_CV_POM_ACCESSORY_WRITE_BYTE"
				case LAN_X_CV_POM_ACCESSORY_WRITE_BIT:
					return "LAN_X_CV_POM_ACCESSORY_WRITE_BIT"
				case LAN_X_CV_POM_ACCESSORY_READ_BYTE:
					return "LAN_X_CV_POM_ACCESSORY_READ_BYTE"
				default:
					return fmt.Sprintf("UNKNOWN DB3: (%02x)", db3)
				}
			default:
				return fmt.Sprintf("UNKNOWN DB0: (%02x)", db0)
			}
		case LAN_X_LOCO_INFO:
			return "LAN_X_LOCO_INFO"
		case LAN_X_F3:
			db0 := f.Payload[1]
			switch db0 {
			case LAN_X_GET_FIRMWARE_VERSION:
				return "LAN_X_GET_FIRMWARE_VERSION"
			default:
				return fmt.Sprintf("UNKNOWN DB0: (%02x)", db0)
			}
		default:
			return fmt.Sprintf("UNKNOWN X-Header: (%02x)", xHeader)
		}
	case LAN_SET_BROADCASTFLAGS:
		return "LAN_SET_BROADCASTFLAGS"
	case LAN_GET_BROADCASTFLAGS:
		return "LAN_GET_BROADCASTFLAGS"
	case LAN_GET_LOCOMODE:
		return "LAN_GET_LOCOMODE"
	case LAN_SET_LOCOMODE:
		return "LAN_SET_LOCOMODE"
	case LAN_GET_TURNOUTMODE:
		return "LAN_GET_TURNOUTMODE"
	case LAN_SET_TURNOUTMODE:
		return "LAN_SET_TURNOUTMODE"
	case LAN_RMBUS_DATACHANGED:
		return "LAN_RMBUS_DATACHANGED"
	case LAN_RMBUS_GETDATA:
		return "LAN_RMBUS_GETDATA"
	case LAN_RMBUS_PROGRAMMODULE:
		return "LAN_RMBUS_PROGRAMMODULE"
	case LAN_SYSTEMSTATE_DATACHANGED:
		return "LAN_SYSTEMSTATE_DATACHANGED"
	case LAN_SYSTEMSTATE_GETDATA:
		return "LAN_SYSTEMSTATE_GETDATA"
	case LAN_RAILCOM_DATACHANGED:
		return "LAN_RAILCOM_DATACHANGED"
	case LAN_RAILCOM_GETDATA:
		return "LAN_RAILCOM_GETDATA"
	case LAN_LOCONET_Z21_RX:
		return "LAN_LOCONET_Z21_RX"
	case LAN_LOCONET_Z21_TX:
		return "LAN_LOCONET_Z21_TX"
	case LAN_LOCONET_FROM_LAN:
		return "LAN_LOCONET_FROM_LAN"
	case LAN_LOCONET_DISPATCH_ADDR:
		return "LAN_LOCONET_DISPATCH_ADDR"
	case LAN_LOCONET_DETECTOR:
		return "LAN_LOCONET_DETECTOR"
	case LAN_CAN_DETECTOR:
		return "LAN_CAN_DETECTOR"
	case LAN_CAN_DEVICE_GET_DESCRIPTION:
		return "LAN_CAN_DEVICE_GET_DESCRIPTION"
	case LAN_CAN_DEVICE_SET_DESCRIPTION:
		return "LAN_CAN_DEVICE_SET_DESCRIPTION"
	case LAN_CAN_BOOSTER_SET_TRACKPOWER:
		return "LAN_CAN_BOOSTER_SET_TRACKPOWER"
	case LAN_FAST_CLOCK_CONTROL:
		return "LAN_FAST_CLOCK_CONTROL"
	case LAN_FAST_CLOCK_DATA:
		return "LAN_FAST_CLOCK_DATA"
	case LAN_FAST_CLOCK_SETTINGS_GET:
		return "LAN_FAST_CLOCK_SETTINGS_GET"
	case LAN_FAST_CLOCK_SETTINGS_SET:
		return "LAN_FAST_CLOCK_SETTINGS_SET"
	case LAN_BOOSTER_SET_POWER:
		return "LAN_BOOSTER_SET_POWER"
	case LAN_BOOSTER_GET_DESCRIPTION:
		return "LAN_BOOSTER_GET_DESCRIPTION"
	case LAN_BOOSTER_SET_DESCRIPTION:
		return "LAN_BOOSTER_SET_DESCRIPTION"
	case LAN_BOOSTER_SYSTEMSTATE_DATACHANGED:
		return "LAN_BOOSTER_SYSTEMSTATE_DATACHANGED"
	case LAN_BOOSTER_SYSTEMSTATE_GETDATA:
		return "LAN_BOOSTER_SYSTEMSTATE_GETDATA"
	case LAN_CAN_BOOSTER_SYSTEMSTATE_CHGD:
		return "LAN_CAN_BOOSTER_SYSTEMSTATE_CHGD"
	case LAN_DECODER_GET_DESCRIPTION:
		return "LAN_DECODER_GET_DESCRIPTION"
	case LAN_DECODER_SET_DESCRIPTION:
		return "LAN_DECODER_SET_DESCRIPTION"
	case LAN_DECODER_SYSTEMSTATE_DATACHANGED:
		return "LAN_DECODER_SYSTEMSTATE_DATACHANGED"
	case LAN_DECODER_SYSTEMSTATE_GETDATA:
		return "LAN_DECODER_SYSTEMSTATE_GETDATA"
	case LAN_ZLINK:
		xHeader := f.Payload[0]
		switch xHeader {
		case LAN_ZLINK_GET_HWINFO:
			return "LAN_ZLINK_GET_HWINFO"
		default:
			return fmt.Sprintf("UNKNOWN X-Header: (%02x)", xHeader)
		}
	default:
		return fmt.Sprintf("UNKNOWN Header: (%02x)", f.Header)
	}
}

func (f *Frame) Pack() ([]byte, error) {
	buf := new(bytes.Buffer)

	f.Length = uint16(2 + 2 + len(f.Payload))
	if err := binary.Write(buf, binary.LittleEndian, f.Length); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.LittleEndian, f.Header); err != nil {
		return nil, err
	}

	if _, err := buf.Write(f.Payload); err != nil {
		return nil, err
	}

	bytes := buf.Bytes()
	return bytes, nil
}

func (f *Frame) Unpack(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("truncated frame")
	}

	f.Length = binary.LittleEndian.Uint16(data[0:2])
	if int(f.Length) > len(data) {
		return fmt.Errorf("frame too short")
	}

	f.Header = binary.LittleEndian.Uint16(data[2:4])
	f.Payload = data[4:int(f.Length)]

	return nil
}

func WrapMessage(m Serializable) (*Frame, error) {
	payload, err := m.Pack()
	if err != nil {
		return nil, err
	}

	return &Frame{
		Header:  m.EncapType(),
		Payload: payload,
	}, nil
}

func ParseFrames(datagram []byte) ([]Frame, error) {
	var frames []Frame
	offset := 0

	for offset < len(datagram) {
		f := Frame{}
		if err := f.Unpack(datagram[offset:]); err != nil {
			return nil, err
		}

		frames = append(frames, f)

		offset += int(f.Length)
	}

	return frames, nil
}

func DecodeFrame(f Frame) (Serializable, error) {
	var m Serializable
	var err error

	switch f.Header {
	case LAN_GET_SERIAL_NUMBER:
		m = &SerialNumber{}
	case LAN_GET_CODE:
		m = &Code{}
	case LAN_GET_HWINFO:
		m = &HwInfo{}
	case LAN_X:
		m, err = DecodeXHeader(f.Payload)
		if err != nil {
			return nil, err
		}
	case LAN_GET_BROADCASTFLAGS:
		m = &SubscribedBroadcastFlags{}
	case LAN_SYSTEMSTATE_DATACHANGED:
		m = &SysData{}
	default:
		return nil, fmt.Errorf("unknown frame header %d", f.Header)
	}

	return m, nil
}

func DecodeXHeader(p []byte) (Serializable, error) {
	var m Serializable
	var err error
	xhdr := uint8(p[0])

	switch xhdr {
	case LAN_X_61, LAN_X_63:
		m, err = DecodeDB0(p)
		if err != nil {
			return nil, err
		}
	case LAN_X_STATUS_CHANGED:
		m = &Status{}
	case LAN_X_BC_STOPPED:
		m = &Stop{}
	case LAN_X_LOCO_INFO:
		m = &LocoInfo{}
	default:
		return nil, fmt.Errorf("unknown x-bus header %d", xhdr)
	}

	return m, nil
}

func DecodeDB0(p []byte) (Serializable, error) {
	var m Serializable
	db0 := uint8(p[1])

	switch db0 {
	case LAN_X_BC_TRACK_POWER_OFF, LAN_X_BC_TRACK_POWER_ON:
		m = &TrackPower{}
	case LAN_X_GET_VERSION:
		m = &Version{}
	default:
		return nil, fmt.Errorf("unknown x-bus db0 %d", db0)
	}

	return m, nil
}
