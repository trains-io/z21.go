package protocol

import "testing"

func TestGetAllCANDetectorsWireFormat(t *testing.T) {
	msg := GetAllCANDetectors()
	if msg.Header != HeaderLANCANDetector {
		t.Fatalf("Header = %#x, want %#x", msg.Header, HeaderLANCANDetector)
	}
	want := []byte{0x00, 0x00, 0xd0}
	if string(msg.Data) != string(want) {
		t.Fatalf("Data = % x, want % x", msg.Data, want)
	}

	wire, err := msg.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	wantWire := []byte{0x07, 0x00, 0xc4, 0x00, 0x00, 0x00, 0xd0}
	if string(wire) != string(wantWire) {
		t.Fatalf("wire = % x, want % x", wire, wantWire)
	}
}

func TestParseCANDetector(t *testing.T) {
	data := []byte{
		0x04, 0xdb, // netid
		0x1f, 0x00, // addr 31
		0x00,       // port 0
		0x01,       // type occupancy
		0x00, 0x01, // value1 free with voltage
		0x00, 0x00,
	}
	report, err := ParseCANDetector(data)
	if err != nil {
		t.Fatal(err)
	}
	if report.NetID != 0xDB04 || report.Addr != 31 || report.Port != 0 || report.Type != 0x01 {
		t.Fatalf("report = %+v", report)
	}
	if got := OccupancyStatusLabel(report.Value1); got != "free" {
		t.Fatalf("OccupancyStatusLabel() = %q, want free", got)
	}
}

func TestCANDetectorReportsFromMessages(t *testing.T) {
	msgs := []Message{{
		Header: HeaderLANGetHWInfo,
		Data:   []byte{1, 2, 3},
	}, {
		Header: HeaderLANCANDetector,
		Data: []byte{
			0x04, 0xdb, 0x1f, 0x00, 0x01, 0x01, 0x11, 0x00, 0x00, 0x00,
		},
	}}

	reports, err := CANDetectorReportsFromMessages(msgs)
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 || reports[0].Port != 1 {
		t.Fatalf("reports = %+v", reports)
	}
}
