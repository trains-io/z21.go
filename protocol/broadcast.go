package protocol

import (
	"encoding/binary"
	"fmt"
)

// Z21 LAN broadcast flag bits (spec §2.16).
const (
	BroadcastFlagXpressNet       uint32 = 0x00000001
	BroadcastFlagRBus            uint32 = 0x00000002
	BroadcastFlagRailComLegacy   uint32 = 0x00000004
	BroadcastFlagFastClock       uint32 = 0x00000010
	BroadcastFlagSystemState     uint32 = 0x00000100
	BroadcastFlagAllLocoInfo     uint32 = 0x00010000
	BroadcastFlagCANBooster      uint32 = 0x00020000
	BroadcastFlagRailCom         uint32 = 0x00040000
	BroadcastFlagCANDetector     uint32 = 0x00080000
	BroadcastFlagLocoNetGeneral uint32 = 0x01000000
	BroadcastFlagLocoNetLoco     uint32 = 0x02000000
	BroadcastFlagLocoNetTurnout uint32 = 0x04000000
	BroadcastFlagLocoNetOcc     uint32 = 0x08000000
)

// BroadcastFlag describes one LAN_SET_BROADCASTFLAGS bit.
type BroadcastFlag struct {
	Value       uint32
	Name        string
	Description string
}

// KnownBroadcastFlags lists user-facing broadcast flag toggles.
var KnownBroadcastFlags = []BroadcastFlag{
	{
		Value:       BroadcastFlagXpressNet,
		Name:        "XpressNet",
		Description: "Track power, programming mode, short circuit, stop, locomotive and turnout messages.",
	},
	{
		Value:       BroadcastFlagRBus,
		Name:        "R-Bus feedback",
		Description: "Changes from R-Bus feedback modules.",
	},
	{
		Value:       BroadcastFlagRailComLegacy,
		Name:        "RailCom (legacy)",
		Description: "Deprecated RailCom updates for subscribed locomotives.",
	},
	{
		Value:       BroadcastFlagFastClock,
		Name:        "Fast clock",
		Description: "Fast clock time messages.",
	},
	{
		Value:       BroadcastFlagSystemState,
		Name:        "System state",
		Description: "Z21 system status such as track voltage.",
	},
	{
		Value:       BroadcastFlagAllLocoInfo,
		Name:        "All locomotive info",
		Description: "Locomotive updates without per-address subscription. High traffic; for PC software only.",
	},
	{
		Value:       BroadcastFlagCANBooster,
		Name:        "CAN booster status",
		Description: "CAN booster status messages.",
	},
	{
		Value:       BroadcastFlagRailCom,
		Name:        "RailCom",
		Description: "RailCom updates for all controlled locomotives.",
	},
	{
		Value:       BroadcastFlagCANDetector,
		Name:        "CAN detector",
		Description: "CAN occupancy detector messages.",
	},
	{
		Value:       BroadcastFlagLocoNetGeneral,
		Name:        "LocoNet (general)",
		Description: "General LocoNet bus messages excluding locomotives and turnouts.",
	},
	{
		Value:       BroadcastFlagLocoNetLoco,
		Name:        "LocoNet locomotives",
		Description: "Locomotive-specific LocoNet messages.",
	},
	{
		Value:       BroadcastFlagLocoNetTurnout,
		Name:        "LocoNet turnouts",
		Description: "Turnout-specific LocoNet messages.",
	},
	{
		Value:       BroadcastFlagLocoNetOcc,
		Name:        "LocoNet occupancy",
		Description: "LocoNet track occupancy detector status.",
	},
}

// HasBroadcastFlag reports whether mask includes flag.
func HasBroadcastFlag(mask, flag uint32) bool {
	return mask&flag != 0
}

// FormatBroadcastFlags renders a mask as 0x-prefixed hex.
func FormatBroadcastFlags(mask uint32) string {
	return fmt.Sprintf("0x%08x", mask)
}

// GetBroadcastFlags returns a request for the client's broadcast subscription mask (spec §2.17).
func GetBroadcastFlags() Message {
	return Message{Header: HeaderLANGetBroadcastFlags}
}

// BroadcastFlagsFromMessages extracts broadcast flags from a Call reply.
func BroadcastFlagsFromMessages(msgs []Message) (uint32, error) {
	for _, msg := range msgs {
		if msg.Header != HeaderLANGetBroadcastFlags {
			continue
		}
		flags, err := ParseBroadcastFlags(msg.Data)
		if err != nil {
			continue
		}
		return flags, nil
	}
	return 0, fmt.Errorf("z21: no LAN_GET_BROADCASTFLAGS reply")
}

// ParseBroadcastFlags decodes the 32-bit little-endian broadcast mask.
func ParseBroadcastFlags(data []byte) (uint32, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("z21: broadcast flags reply too short (%d bytes)", len(data))
	}
	return binary.LittleEndian.Uint32(data[0:4]), nil
}
