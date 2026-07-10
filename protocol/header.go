package protocol

import "fmt"

// Well-known Z21 LAN dataset headers (16-bit little-endian).
// See Roco Z21 LAN Protocol Specification.
const (
	HeaderLANGetSerialNumber    uint16 = 0x0010
	HeaderLANGetCode              uint16 = 0x0018
	HeaderLANGetHWInfo            uint16 = 0x001A
	HeaderLANLogoff               uint16 = 0x0030
	HeaderLANSetBroadcastFlags    uint16 = 0x0050
	HeaderLANGetBroadcastFlags    uint16 = 0x0051
	HeaderLANGetLocoMode          uint16 = 0x0060
	HeaderLANSetLocoMode          uint16 = 0x0061
	HeaderLANGetTurnoutMode       uint16 = 0x0070
	HeaderLANSetTurnoutMode       uint16 = 0x0071
	HeaderLANSystemStateGetData       uint16 = 0x0085
	HeaderLANSystemStateDataChanged   uint16 = 0x0084
	HeaderLANRMBusDataChanged     uint16 = 0x0080
	HeaderLANRMBusGetData         uint16 = 0x0081
	HeaderLANRMBusProgramModule   uint16 = 0x0082
	HeaderLANRailComDataChanged   uint16 = 0x0088
	HeaderLANRailComGetData       uint16 = 0x0089
	HeaderLANCANDetector          uint16 = 0x00C4
	HeaderLANCANMaintenance       uint16 = 0x00C2
	HeaderLANCANDeviceGetDescription uint16 = 0x00C8
	HeaderLANCANDeviceSetDescription uint16 = 0x00C9
	HeaderLANCANBoosterSystemState   uint16 = 0x00CA
	HeaderLANCANBoosterSetTrackPower uint16 = 0x00CB
)

// DefaultBroadcastFlags enables XpressNet and system state broadcasts.
const DefaultBroadcastFlags uint32 = 0x00000101

// HeaderName returns a human-readable name for a dataset header.
func HeaderName(header uint16) string {
	switch header {
	case HeaderLANGetSerialNumber:
		return "LAN_GET_SERIAL_NUMBER"
	case HeaderLANGetCode:
		return "LAN_GET_CODE"
	case HeaderLANGetHWInfo:
		return "LAN_GET_HWINFO"
	case HeaderLANLogoff:
		return "LAN_LOGOFF"
	case HeaderLANSetBroadcastFlags:
		return "LAN_SET_BROADCASTFLAGS"
	case HeaderLANGetBroadcastFlags:
		return "LAN_GET_BROADCASTFLAGS"
	case HeaderLANGetLocoMode:
		return "LAN_GET_LOCOMODE"
	case HeaderLANSetLocoMode:
		return "LAN_SET_LOCOMODE"
	case HeaderLANGetTurnoutMode:
		return "LAN_GET_TURNOUTMODE"
	case HeaderLANSetTurnoutMode:
		return "LAN_SET_TURNOUTMODE"
	case HeaderLANSystemStateGetData:
		return "LAN_SYSTEMSTATE_GETDATA"
	case HeaderLANSystemStateDataChanged:
		return "LAN_SYSTEMSTATE_DATACHANGED"
	case HeaderLANRMBusDataChanged:
		return "LAN_RMBUS_DATACHANGED"
	case HeaderLANRMBusGetData:
		return "LAN_RMBUS_GETDATA"
	case HeaderLANRMBusProgramModule:
		return "LAN_RMBUS_PROGRAMMODULE"
	case HeaderLANRailComDataChanged:
		return "LAN_RAILCOM_DATACHANGED"
	case HeaderLANRailComGetData:
		return "LAN_RAILCOM_GETDATA"
	case HeaderLANCANDetector:
		return "LAN_CAN_DETECTOR"
	case HeaderLANCANMaintenance:
		return "LAN_CAN_MAINTENANCE"
	case HeaderLANCANDeviceGetDescription:
		return "LAN_CAN_DEVICE_GET_DESCRIPTION"
	case HeaderLANCANDeviceSetDescription:
		return "LAN_CAN_DEVICE_SET_DESCRIPTION"
	case HeaderLANCANBoosterSystemState:
		return "LAN_CAN_BOOSTER_SYSTEMSTATE_CHGD"
	case HeaderLANCANBoosterSetTrackPower:
		return "LAN_CAN_BOOSTER_SET_TRACKPOWER"
	case HeaderLANX:
		return "LAN_X"
	default:
		return fmt.Sprintf("0x%04x", header)
	}
}

// MessageName returns a descriptive name for a Z21 dataset, including LAN_X sub-commands.
func MessageName(msg Message) string {
	if msg.Header != HeaderLANX {
		return HeaderName(msg.Header)
	}
	return lanXMessageName(msg.Data)
}

func lanXMessageName(data []byte) string {
	if len(data) == 0 {
		return "LAN_X"
	}

	switch data[0] {
	case xHeaderSetStop:
		if len(data) >= 2 && data[1] == xChecksumSetStop {
			return "LAN_X_SET_STOP"
		}
		return fmt.Sprintf("LAN_X x=0x%02x", data[0])
	case xHeaderBCStopped:
		if len(data) >= 3 && data[1] == xBCStoppedDB0 && data[2] == xChecksumBCStopped {
			return "LAN_X_BC_STOPPED"
		}
		return fmt.Sprintf("LAN_X x=0x%02x", data[0])
	case xHeaderXBusBC:
		return xBusBCMessageName(data)
	case xHeaderGetVersion:
		if len(data) < 2 {
			return "LAN_X"
		}
		switch data[1] {
		case xHeaderGetVersion:
			return "LAN_X_GET_VERSION"
		case xCommandGetStatus:
			return "LAN_X_GET_STATUS"
		case xCommandSetTrackPowerOff:
			return "LAN_X_SET_TRACK_POWER_OFF"
		case xCommandSetTrackPowerOn:
			return "LAN_X_SET_TRACK_POWER_ON"
		default:
			return fmt.Sprintf("LAN_X cmd=0x%02x", data[1])
		}
	case xHeaderGetLocoInfo:
		if len(data) >= 2 && data[1] == xCommandPurgeLoco {
			return "LAN_X_PURGE_LOCO"
		}
		if len(data) >= 2 && data[1] == xCommandGetLocoInfo {
			return "LAN_X_GET_LOCO_INFO"
		}
		return fmt.Sprintf("LAN_X x=0x%02x", data[0])
	case xHeaderSetLocoDrive:
		if len(data) < 2 {
			return "LAN_X"
		}
		switch {
		case data[1] == xCommandSetLocoFunction:
			return "LAN_X_SET_LOCO_FUNCTION"
		case isLocoFunctionGroup(data[1]):
			return "LAN_X_SET_LOCO_FUNCTION_GROUP"
		case data[1]&0xF0 == 0x10:
			return "LAN_X_SET_LOCO_DRIVE"
		default:
			return fmt.Sprintf("LAN_X cmd=0x%02x", data[1])
		}
	case xHeaderSetLocoBinary:
		return "LAN_X_SET_LOCO_BINARY_STATE"
	case xHeaderLocoInfo:
		return "LAN_X_LOCO_INFO"
	case xHeaderSetLocoEStop:
		return "LAN_X_SET_LOCO_E_STOP"
	case xHeaderGetTurnoutInfo:
		if len(data) >= 5 {
			return "LAN_X_TURNOUT_INFO"
		}
		return "LAN_X_GET_TURNOUT_INFO"
	case xHeaderSetTurnout:
		return "LAN_X_SET_TURNOUT"
	case xHeaderGetExtAccessory:
		if len(data) >= 6 {
			return "LAN_X_EXT_ACCESSORY_INFO"
		}
		return "LAN_X_GET_EXT_ACCESSORY_INFO"
	case xHeaderSetExtAccessory:
		return "LAN_X_SET_EXT_ACCESSORY"
	case xHeaderDCCReadReg:
		if len(data) >= 2 && data[1] == xCommandDCCReadReg {
			return "LAN_X_DCC_READ_REGISTER"
		}
		return fmt.Sprintf("LAN_X x=0x%02x", data[0])
	case xHeaderCVRead:
		if len(data) >= 2 {
			switch data[1] {
			case xCommandCVRead:
				return "LAN_X_CV_READ"
			case xCommandDCCWriteReg:
				return "LAN_X_DCC_WRITE_REGISTER"
			}
		}
		return fmt.Sprintf("LAN_X x=0x%02x", data[0])
	case xHeaderCVWrite:
		if len(data) >= 2 {
			switch data[1] {
			case xCommandCVWrite:
				return "LAN_X_CV_WRITE"
			case xCommandMMWriteByte:
				return "LAN_X_MM_WRITE_BYTE"
			}
		}
		return fmt.Sprintf("LAN_X x=0x%02x", data[0])
	case xHeaderCVResult:
		if len(data) >= 2 && data[1] == xCommandCVResult {
			return "LAN_X_CV_RESULT"
		}
		return fmt.Sprintf("LAN_X x=0x%02x", data[0])
	case xHeaderPOMLoco:
		if len(data) >= 2 {
			switch data[1] {
			case xCommandPOMLoco:
				return pomLocoMessageName(data)
			case xCommandPOMAccessory:
				return pomAccessoryMessageName(data)
			}
		}
		return fmt.Sprintf("LAN_X x=0x%02x", data[0])
	case xHeaderStatusReply:
		return "LAN_X_STATUS_CHANGED"
	case xHeaderGetVersionReply:
		return "LAN_X_GET_VERSION"
	case xCommandGetFirmware:
		return "LAN_X_GET_FIRMWARE_VERSION"
	case xHeaderFirmwareReply:
		return "LAN_X_GET_FIRMWARE_VERSION"
	default:
		return fmt.Sprintf("LAN_X x=0x%02x", data[0])
	}
}

func isLocoFunctionGroup(b byte) bool {
	switch b {
	case 0x20, 0x21, 0x22, 0x23, 0x28, 0x29, 0x2A, 0x2B, 0x50, 0x51:
		return true
	default:
		return false
	}
}

func pomLocoMessageName(data []byte) string {
	if len(data) < 6 {
		return "LAN_X_CV_POM"
	}
	switch data[4] & 0xFC {
	case pomOptionWriteByte:
		return "LAN_X_CV_POM_WRITE_BYTE"
	case pomOptionWriteBit:
		return "LAN_X_CV_POM_WRITE_BIT"
	case pomOptionReadByte:
		return "LAN_X_CV_POM_READ_BYTE"
	default:
		return "LAN_X_CV_POM"
	}
}

func pomAccessoryMessageName(data []byte) string {
	if len(data) < 6 {
		return "LAN_X_CV_POM_ACCESSORY"
	}
	switch data[4] & 0xFC {
	case pomOptionWriteByte:
		return "LAN_X_CV_POM_ACCESSORY_WRITE_BYTE"
	case pomOptionWriteBit:
		return "LAN_X_CV_POM_ACCESSORY_WRITE_BIT"
	case pomOptionReadByte:
		return "LAN_X_CV_POM_ACCESSORY_READ_BYTE"
	default:
		return "LAN_X_CV_POM_ACCESSORY"
	}
}
