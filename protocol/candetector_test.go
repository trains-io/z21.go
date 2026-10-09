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

func TestEncodeCANDetector(t *testing.T) {
	msg := EncodeCANDetector(CANDetectorReport{
		NetID:  0xC101,
		Addr:   1,
		Port:   0,
		Type:   CANDetectorTypeOccupancy,
		Value1: 0x0100,
	})
	if msg.Header != HeaderLANCANDetector {
		t.Fatalf("Header = %#x, want %#x", msg.Header, HeaderLANCANDetector)
	}
	want := []byte{0x01, 0xc1, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00}
	if string(msg.Data) != string(want) {
		t.Fatalf("EncodeCANDetector(occupancy free).Data = % x, want % x", msg.Data, want)
	}
	if !IsCANDetectorReport(msg) {
		t.Fatal("IsCANDetectorReport(EncodeCANDetector()) = false, want true")
	}
}

func TestEncodeCANDetectorRoundTrip(t *testing.T) {
	tests := []CANDetectorReport{
		{},
		{NetID: 0xC101, Addr: 1, Port: 7, Type: CANDetectorTypeOccupancy, Value1: 0x1100},
		{NetID: 0xC102, Addr: 256, Port: 3, Type: CANDetectorTypeLocoAddressBase, Value1: 0x4003, Value2: 0x8000 | 1234},
		{NetID: 0xffff, Addr: 0xffff, Port: 0xff, Type: 0xff, Value1: 0xffff, Value2: 0xffff},
	}
	for _, in := range tests {
		got, err := CANDetectorReportsFromMessages([]Message{EncodeCANDetector(in)})
		if err != nil {
			t.Fatalf("CANDetectorReportsFromMessages(EncodeCANDetector(%+v)) error = %v", in, err)
		}
		if len(got) != 1 || got[0] != in {
			t.Fatalf("CANDetectorReportsFromMessages(EncodeCANDetector(%+v)) = %+v", in, got)
		}
	}
}

func TestParseGetCANDetector(t *testing.T) {
	for _, in := range []uint16{0x0000, 0xC101, CANDetectorPollAll, 0xffff} {
		got, err := ParseGetCANDetector(GetCANDetector(in).Data)
		if err != nil {
			t.Fatalf("ParseGetCANDetector(GetCANDetector(%#04x)) error = %v", in, err)
		}
		if got != in {
			t.Fatalf("ParseGetCANDetector(GetCANDetector(%#04x)) = %#04x", in, got)
		}
	}
}

func TestParseGetCANDetectorErrors(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "empty", data: nil},
		{name: "too short", data: []byte{0x00, 0x01}},
		{name: "unknown poll type", data: []byte{0x01, 0x00, 0xd0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseGetCANDetector(tt.data); err == nil {
				t.Fatalf("ParseGetCANDetector(% x) error = nil, want error", tt.data)
			}
		})
	}
}
