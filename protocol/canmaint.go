package protocol

import (
	"encoding/binary"
	"fmt"
)

const (
	canMaintSessionPrefix0 = 0x00
	canMaintSessionPrefix1 = 0x08
	canMaintSessionPrefix2 = 0x01

	canDetectorDevTypeHi = 0x18

	canDetectorDevTypeLoBase     = 0x21
	canDetectorDevTypeLoRailComCh2 = 0x29

	canMaintCmdSetModuleAddress byte = 0x14
)

// SetCANDetectorModuleAddress builds LAN header 0xC2 maintenance commit for module address.
// moduleAddr is the user-visible address (1-based); wire encoding uses moduleAddr-1.
// railcomCh2 sets the RailCom channel 2 forward flag on the final commit packet only.
func SetCANDetectorModuleAddress(netID uint16, moduleAddr uint16, railcomCh2 bool) (Message, error) {
	if moduleAddr == 0 {
		return Message{}, fmt.Errorf("z21: module address must be >= 1")
	}
	if moduleAddr > 256 {
		return Message{}, fmt.Errorf("z21: module address must be <= 256")
	}

	devLo := byte(canDetectorDevTypeLoBase)
	if railcomCh2 {
		devLo = canDetectorDevTypeLoRailComCh2
	}

	data := []byte{
		canMaintSessionPrefix0, canMaintSessionPrefix1, canMaintSessionPrefix2, 0x00, 0x00, 0x00,
		devLo, canDetectorDevTypeHi,
		byte(netID), byte(netID >> 8),
		canMaintCmdSetModuleAddress, 0x00, byte(moduleAddr - 1), 0x00, 0x00, 0x00,
	}
	return Message{Header: HeaderLANCANMaintenance, Data: data}, nil
}

// ParseCANMaintenanceSetAddress decodes a module-address commit packet (for tests/trace).
func ParseCANMaintenanceSetAddress(data []byte) (netID uint16, moduleAddr uint16, railcomCh2 bool, err error) {
	if len(data) < 16 {
		return 0, 0, false, fmt.Errorf("z21: CAN maintenance packet too short (%d bytes)", len(data))
	}
	if data[0] != canMaintSessionPrefix0 || data[1] != canMaintSessionPrefix1 || data[2] != canMaintSessionPrefix2 {
		return 0, 0, false, fmt.Errorf("z21: unexpected CAN maintenance session prefix")
	}
	railcomCh2 = data[6] == canDetectorDevTypeLoRailComCh2
	if data[7] != canDetectorDevTypeHi {
		return 0, 0, false, fmt.Errorf("z21: unexpected CAN device type")
	}
	netID = binary.LittleEndian.Uint16(data[8:10])
	if data[10] != canMaintCmdSetModuleAddress {
		return 0, 0, false, fmt.Errorf("z21: not a module address commit (cmd=0x%02x)", data[10])
	}
	moduleAddr = uint16(data[12]) + 1
	return netID, moduleAddr, railcomCh2, nil
}
