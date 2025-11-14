package z21

const (
	FREE_NOVOLT    uint16 = 0x0000
	FREE           uint16 = 0x0100
	BUSY_NOVOLT    uint16 = 0x1000
	BUSY           uint16 = 0x1100
	BUSY_OVERLOAD1 uint16 = 0x1201
	BUSY_OVERLOAD2 uint16 = 0x1202
	BUSY_OVERLOAD3 uint16 = 0x1203
)

type Detector struct {
	NetworkID uint16
	Address   uint16
	Ports     []DetectorPort
}

type DetectorPort struct {
	Index  uint8
	Status uint16
}

type Mask8 uint8
type Mask32 uint32

func (m Mask8) Has(flag uint8) bool {
	return uint8(m)&flag != 0
}

func (m Mask32) Has(flag uint32) bool {
	return uint32(m)&flag != 0
}

func isLocoDriveSpeedStep(db0 byte) bool {
	steps := []uint8{
		LAN_X_SET_LOCO_DRIVE_S0,
		LAN_X_SET_LOCO_DRIVE_S2,
		LAN_X_SET_LOCO_DRIVE_S3,
	}
	for _, s := range steps {
		if s == db0 {
			return true
		}
	}
	return false
}

func isLocoFunctionGroup(db0 byte) bool {
	groups := []uint8{
		LAN_X_SET_LOCO_FUNCTION_GROUP1,
		LAN_X_SET_LOCO_FUNCTION_GROUP2,
		LAN_X_SET_LOCO_FUNCTION_GROUP3,
		LAN_X_SET_LOCO_FUNCTION_GROUP4,
		LAN_X_SET_LOCO_FUNCTION_GROUP5,
		LAN_X_SET_LOCO_FUNCTION_GROUP6,
		LAN_X_SET_LOCO_FUNCTION_GROUP7,
		LAN_X_SET_LOCO_FUNCTION_GROUP8,
		LAN_X_SET_LOCO_FUNCTION_GROUP9,
		LAN_X_SET_LOCO_FUNCTION_GROUP10,
	}
	for _, g := range groups {
		if g == db0 {
			return true
		}
	}
	return false
}

const (
	// groups
	LAN_X       uint16 = 0x40
	LAN_X_21    uint8  = 0x21
	LAN_X_23    uint8  = 0x23
	LAN_X_24    uint8  = 0x24
	LAN_X_61    uint8  = 0x61
	LAN_X_63    uint8  = 0x63
	LAN_X_E3    uint8  = 0xE3
	LAN_X_E4    uint8  = 0xE4
	LAN_X_E6    uint8  = 0xE6
	LAN_X_E6_30 uint8  = 0x30
	LAN_X_E6_31 uint8  = 0x31
	LAN_X_F1    uint8  = 0xF1
	LAN_X_F3    uint8  = 0xF3
	LAN_ZLINK   uint16 = 0xE8

	// client to Z21
	LAN_GET_SERIAL_NUMBER             uint16 = 0x10
	LAN_GET_CODE                      uint16 = 0x18
	LAN_GET_HWINFO                    uint16 = 0x1A
	LAN_LOGOFF                        uint16 = 0x30
	LAN_X_GET_VERSION                 uint8  = 0x21 // LAN_X_21 and LAN_X_63
	LAN_X_GET_STATUS                  uint8  = 0x24 // LAN_X_21
	LAN_X_SET_TRACK_POWER_OFF         uint8  = 0x80 // LAN_X_21
	LAN_X_SET_TRACK_POWER_ON          uint8  = 0x81 // LAN_X_21
	LAN_X_DCC_READ_REGISTER           uint8  = 0x22
	LAN_X_CV_READ                     uint8  = 0x11 // LAN_X_23
	LAN_X_DCC_WRITE_REGISTER          uint8  = 0x12 // LAN_X_23
	LAN_X_CV_WRITE                    uint8  = 0x12 // LAN_X_24
	LAN_X_MM_WRITE_BYTE               uint8  = 0xFF // LAN_X_24
	LAN_X_GET_TURNOUT_INFO            uint8  = 0x43
	LAN_X_GET_EXT_ACCESSORY_INFO      uint8  = 0x44
	LAN_X_SET_TURNOUT                 uint8  = 0x53
	LAN_X_SET_EXT_ACCESSORY           uint8  = 0x54
	LAN_X_SET_STOP                    uint8  = 0x80
	LAN_X_SET_LOCO_E_STOP             uint8  = 0x92
	LAN_X_PURGE_LOCO                  uint8  = 0x44 // LAN_X_E3
	LAN_X_GET_LOCO_INFO               uint8  = 0xF0 // LAN_X_E3
	LAN_X_SET_LOCO_DRIVE_S0           uint8  = 0x10 // LAN_X_E4
	LAN_X_SET_LOCO_DRIVE_S2           uint8  = 0x12 // LAN_X_E4
	LAN_X_SET_LOCO_DRIVE_S3           uint8  = 0x13 // LAN_X_E4
	LAN_X_SET_LOCO_FUNCTION           uint8  = 0xF8 // LAN_X_E4
	LAN_X_SET_LOCO_FUNCTION_GROUP1    uint8  = 0x20 // LAN_X_E4
	LAN_X_SET_LOCO_FUNCTION_GROUP2    uint8  = 0x21 // LAN_X_E4
	LAN_X_SET_LOCO_FUNCTION_GROUP3    uint8  = 0x22 // LAN_X_E4
	LAN_X_SET_LOCO_FUNCTION_GROUP4    uint8  = 0x23 // LAN_X_E4
	LAN_X_SET_LOCO_FUNCTION_GROUP5    uint8  = 0x28 // LAN_X_E4
	LAN_X_SET_LOCO_FUNCTION_GROUP6    uint8  = 0x29 // LAN_X_E4
	LAN_X_SET_LOCO_FUNCTION_GROUP7    uint8  = 0x2A // LAN_X_E4
	LAN_X_SET_LOCO_FUNCTION_GROUP8    uint8  = 0x2B // LAN_X_E4
	LAN_X_SET_LOCO_FUNCTION_GROUP9    uint8  = 0x50 // LAN_X_E4
	LAN_X_SET_LOCO_FUNCTION_GROUP10   uint8  = 0x51 // LAN_X_E4
	LAN_X_SET_LOCO_BINARY_STATE       uint8  = 0x5F // LAN_X_E4
	LAN_X_CV_POM_WRITE_BYTE           uint8  = 0xEC // LAN_X_E6_30
	LAN_X_CV_POM_WRITE_BIT            uint8  = 0xE8 // LAN_X_E6_30
	LAN_X_CV_POM_READ_BYTE            uint8  = 0xE4 // LAN_X_E6_30
	LAN_X_CV_POM_ACCESSORY_WRITE_BYTE uint8  = 0xEC // LAN_X_E6_31
	LAN_X_CV_POM_ACCESSORY_WRITE_BIT  uint8  = 0xE8 // LAN_X_E6_31
	LAN_X_CV_POM_ACCESSORY_READ_BYTE  uint8  = 0xE4 // LAN_X_E6_31
	LAN_X_GET_FIRMWARE_VERSION        uint8  = 0x0A // LAN_X_F1
	LAN_SET_BROADCASTFLAGS            uint16 = 0x50
	LAN_GET_BROADCASTFLAGS            uint16 = 0x51
	LAN_GET_LOCOMODE                  uint16 = 0x60
	LAN_SET_LOCOMODE                  uint16 = 0x61
	LAN_GET_TURNOUTMODE               uint16 = 0x70
	LAN_SET_TURNOUTMODE               uint16 = 0x71
	LAN_RMBUS_GETDATA                 uint16 = 0x81
	LAN_RMBUS_PROGRAMMODULE           uint16 = 0x82
	LAN_SYSTEMSTATE_GETDATA           uint16 = 0x85
	LAN_RAILCOM_GETDATA               uint16 = 0x89
	LAN_LOCONET_FROM_LAN              uint16 = 0xA2
	LAN_LOCONET_DETECTOR              uint16 = 0xA3
	LAN_LOCONET_DISPATCH_ADDR         uint16 = 0xA4
	LAN_CAN_DETECTOR                  uint16 = 0xC4
	LAN_CAN_DEVICE_GET_DESCRIPTION    uint16 = 0xC8
	LAN_CAN_DEVICE_SET_DESCRIPTION    uint16 = 0xC9
	LAN_CAN_BOOSTER_SET_TRACKPOWER    uint16 = 0xCB
	LAN_FAST_CLOCK_CONTROL            uint16 = 0xCC
	LAN_FAST_CLOCK_SETTINGS_GET       uint16 = 0xCE
	LAN_FAST_CLOCK_SETTINGS_SET       uint16 = 0xCF
	LAN_BOOSTER_SET_POWER             uint16 = 0xB2
	LAN_BOOSTER_GET_DESCRIPTION       uint16 = 0xB8
	LAN_BOOSTER_SET_DESCRIPTION       uint16 = 0xB9
	LAN_BOOSTER_SYSTEMSTATE_GETDATA   uint16 = 0xBB
	LAN_DECODER_GET_DESCRIPTION       uint16 = 0xD8
	LAN_DECODER_SET_DESCRIPTION       uint16 = 0xD9
	LAN_DECODER_SYSTEMSTATE_GETDATA   uint16 = 0xDB
	LAN_ZLINK_GET_HWINFO              uint8  = 0x06

	// Z21 to client
	LAN_X_BC_TRACK_POWER_OFF            uint8  = 0x00 // LAN_X_61
	LAN_X_BC_TRACK_POWER_ON             uint8  = 0x01 // LAN_X_61
	LAN_X_BC_PROGRAMMING_MODE           uint8  = 0x02 // LAN_X_61
	LAN_X_BC_TRACK_SHORT_CIRCUIT        uint8  = 0x08 // LAN_X_61
	LAN_X_CV_NACK_SC                    uint8  = 0x12 // LAN_X_61
	LAN_X_CV_NACK                       uint8  = 0x13 // LAN_X_61
	LAN_X_UNKNOWN_COMMAND               uint8  = 0x82 // LAN_X_61
	LAN_X_STATUS_CHANGED                uint8  = 0x62
	LAN_X_CV_RESULT                     uint8  = 0x64
	LAN_X_BC_STOPPED                    uint8  = 0x81
	LAN_X_LOCO_INFO                     uint8  = 0xEF
	LAN_RMBUS_DATACHANGED               uint16 = 0x80
	LAN_SYSTEMSTATE_DATACHANGED         uint16 = 0x84
	LAN_RAILCOM_DATACHANGED             uint16 = 0x88
	LAN_LOCONET_Z21_RX                  uint16 = 0xA0
	LAN_LOCONET_Z21_TX                  uint16 = 0xA1
	LAN_BOOSTER_SYSTEMSTATE_DATACHANGED uint16 = 0xBA
	LAN_CAN_BOOSTER_SYSTEMSTATE_CHGD    uint16 = 0xCA
	LAN_FAST_CLOCK_DATA                 uint16 = 0xCD
	LAN_DECODER_SYSTEMSTATE_DATACHANGED uint16 = 0xDA
)
