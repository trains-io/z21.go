package z21

import "encoding/binary"

const (
	TRACK_UPDATES            uint32 = 0x00000001 // bit 0
	FEEDBACK_UPDATES         uint32 = 0x00000002 // bit 1
	RAILCOM_SUB_UPDATES      uint32 = 0x00000004 // bit 2
	FAST_CLOCK_UPDATES       uint32 = 0x00000010 // bit 4
	SYSTEM_UPDATES           uint32 = 0x00000100 // bit 8
	LOCO_UPDATES             uint32 = 0x00010000 // bit 16
	CAN_BOOSTER_UPDATES      uint32 = 0x00020000 // bit 17
	RAILCOM_UPDATES          uint32 = 0x00040000 // bit 18
	CAN_DETECTOR_UPDATES     uint32 = 0x00080000 // bit 19
	LOCONET_UPDATES          uint32 = 0x01000000 // bit 24
	LOCONET_LOCO_UPDATES     uint32 = 0x02000000 // bit 25
	LOCONET_SWITCH_UPDATES   uint32 = 0x04000000 // bit 26
	LOCONET_DETECTOR_UPDATES uint32 = 0x08000000 // bit 27
)

// LAN_GET_BROADCASTFLAGS
type SubscribedBroadcastFlags struct {
	Flags Mask32
}

// ---------- Message interface ----------

func (m *SubscribedBroadcastFlags) Pack() ([]byte, error) {
	return []byte{}, nil
}

func (m *SubscribedBroadcastFlags) Unpack(data []byte) error {
	return UnpackFields(data, &m.Flags)
}

func (m *SubscribedBroadcastFlags) EncapType() uint16 {
	return LAN_GET_BROADCASTFLAGS
}

// ---------- Correlatable interface ----------

func (m *SubscribedBroadcastFlags) Key() (string, bool) {
	d := []byte{byte(LAN_GET_BROADCASTFLAGS)}
	f, err := fingerprint(d)
	if err != nil {
		return "", false
	}
	return f, true
}

// LAN_SET_BROADCASTFLAGS
type BroadcastFlags struct {
	Flags Mask32
}

// ---------- Message interface ----------

func (m *BroadcastFlags) Pack() ([]byte, error) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(m.Flags))
	return b, nil
}

func (m *BroadcastFlags) Unpack(data []byte) error {
	return nil
}

func (m *BroadcastFlags) EncapType() uint16 {
	return LAN_SET_BROADCASTFLAGS
}

// ---------- Correlatable interface ----------

func (m *BroadcastFlags) Key() (string, bool) {
	return "", false
}
