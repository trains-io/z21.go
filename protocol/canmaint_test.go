package protocol

import "testing"

func TestSetCANDetectorModuleAddress(t *testing.T) {
	msg, err := SetCANDetectorModuleAddress(0xDB04, 60, true)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Header != HeaderLANCANMaintenance {
		t.Fatalf("Header = %#x, want %#x", msg.Header, HeaderLANCANMaintenance)
	}

	netID, addr, railcom, err := ParseCANMaintenanceSetAddress(msg.Data)
	if err != nil {
		t.Fatal(err)
	}
	if netID != 0xDB04 || addr != 60 || !railcom {
		t.Fatalf("parsed = netID %#x addr %d railcom %v", netID, addr, railcom)
	}

	wire, err := msg.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{
		0x14, 0x00, 0xc2, 0x00,
		0x00, 0x08, 0x01, 0x00, 0x00, 0x00,
		0x29, 0x18, 0x04, 0xdb,
		0x14, 0x00, 0x3b, 0x00, 0x00, 0x00,
	}
	if string(wire) != string(want) {
		t.Fatalf("wire = % x, want % x", wire, want)
	}
}

func TestSetCANDetectorModuleAddressRailComOff(t *testing.T) {
	msg, err := SetCANDetectorModuleAddress(0xDB04, 25, false)
	if err != nil {
		t.Fatal(err)
	}
	_, addr, railcom, err := ParseCANMaintenanceSetAddress(msg.Data)
	if err != nil {
		t.Fatal(err)
	}
	if addr != 25 || railcom {
		t.Fatalf("addr = %d railcom = %v", addr, railcom)
	}
	if msg.Data[6] != 0x21 {
		t.Fatalf("dev type lo = %#x, want 0x21", msg.Data[6])
	}
}

func TestSetCANDetectorModuleAddressValidation(t *testing.T) {
	if _, err := SetCANDetectorModuleAddress(0xDB04, 0, true); err == nil {
		t.Fatal("expected error for address 0")
	}
}
