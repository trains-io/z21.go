package protocol

import "fmt"

const (
	// LocoInfoSpeedSteps values for KKK in LAN_X_LOCO_INFO (spec §4.4). They
	// differ from the LocoSpeedSteps* values of LAN_X_SET_LOCO_DRIVE for 128
	// speed steps.
	LocoInfoSpeedSteps14  byte = 0x00
	LocoInfoSpeedSteps28  byte = 0x02
	LocoInfoSpeedSteps128 byte = 0x04

	// LocoSpeedEStop is the speed of an emergency stop in every speed step
	// mode (spec §4.2).
	LocoSpeedEStop byte = 0x01
)

// LocoInfoSpeedSteps converts speedSteps (LocoSpeedSteps*) to the KKK value of
// LAN_X_LOCO_INFO. Unknown values are returned unchanged.
func LocoInfoSpeedSteps(speedSteps byte) byte {
	if speedSteps == LocoSpeedSteps128 {
		return LocoInfoSpeedSteps128
	}
	return speedSteps
}

// LocoDriveSpeedSteps converts the KKK value of LAN_X_LOCO_INFO to the
// LocoSpeedSteps* value of LAN_X_SET_LOCO_DRIVE. Unknown values are returned
// unchanged.
func LocoDriveSpeedSteps(kkk byte) byte {
	if kkk == LocoInfoSpeedSteps128 {
		return LocoSpeedSteps128
	}
	return kkk
}

// MaxLocoSpeedStep returns the highest speed step of speedSteps
// (LocoSpeedSteps*): 14, 28 or 126; 0 for unknown values.
func MaxLocoSpeedStep(speedSteps byte) uint8 {
	switch speedSteps {
	case LocoSpeedSteps14:
		return 14
	case LocoSpeedSteps28:
		return 28
	case LocoSpeedSteps128:
		return 126
	default:
		return 0
	}
}

// EncodeLocoSpeed returns the speed of LAN_X_SET_LOCO_DRIVE and
// LAN_X_LOCO_INFO for speed step step (0 stops) with speedSteps
// (LocoSpeedSteps*), without the direction bit (spec §4.2). Use
// LocoSpeedEStop for an emergency stop.
func EncodeLocoSpeed(speedSteps byte, step uint8) (byte, error) {
	maxStep := MaxLocoSpeedStep(speedSteps)
	if maxStep == 0 {
		return 0, fmt.Errorf("z21: unknown speed steps %#02x", speedSteps)
	}
	if step > maxStep {
		return 0, fmt.Errorf("z21: speed step %d out of range [0, %d]", step, maxStep)
	}
	if step == 0 {
		return 0, nil
	}
	if speedSteps == LocoSpeedSteps28 {
		// Steps 1–28 are 4–31 with the lowest bit moved to bit 4.
		v := step + 3
		return v>>1 | (v&1)<<4, nil
	}
	return step + 1, nil
}

// DecodeLocoSpeed returns the speed step encoded in speed with speedSteps
// (LocoSpeedSteps*), ignoring the direction bit, and whether speed is an
// emergency stop, whose step is 0 (spec §4.2). Unknown speed step modes are
// decoded like 128 speed steps.
func DecodeLocoSpeed(speedSteps byte, speed byte) (step uint8, eStop bool) {
	switch speedSteps {
	case LocoSpeedSteps14:
		speed &= 0x0F
	case LocoSpeedSteps28:
		v := (speed&0x0F)<<1 | (speed>>4)&1
		switch {
		case v <= 1:
			return 0, false
		case v <= 3:
			return 0, true
		default:
			return v - 3, false
		}
	default:
		speed &= 0x7F
	}
	switch speed {
	case 0:
		return 0, false
	case LocoSpeedEStop:
		return 0, true
	default:
		return speed - 1, false
	}
}

// SpeedStep returns the speed step of info and whether it is an emergency
// stop, see DecodeLocoSpeed.
func (info LocoInfo) SpeedStep() (step uint8, eStop bool) {
	return DecodeLocoSpeed(LocoDriveSpeedSteps(info.SpeedSteps), info.Speed)
}
