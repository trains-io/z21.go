package protocol

import "testing"

func TestParseSystemState(t *testing.T) {
	data := []byte{
		0x64, 0x00, // main current 100 mA
		0x00, 0x00,
		0x50, 0x00, // filtered 80 mA
		0x2A, 0x00, // temperature 42 °C
		0x10, 0x27, // supply 10000 mV
		0xE8, 0x03, // VCC 1000 mV
		0x00, // central state
		0x00, // central state ex
		0x00, // reserved
		0x31, // capabilities: DCC + loco + accessory
	}

	state, err := ParseSystemState(data)
	if err != nil {
		t.Fatalf("ParseSystemState() error = %v", err)
	}
	if state.MainCurrent != 100 {
		t.Fatalf("MainCurrent = %d, want 100", state.MainCurrent)
	}
	if state.Temperature != 42 {
		t.Fatalf("Temperature = %d, want 42", state.Temperature)
	}
	if got := FormatXStatusFlags(state.CentralState); got != "ok" {
		t.Fatalf("FormatXStatusFlags() = %q, want ok", got)
	}
	if got := FormatCapabilities(state.Capabilities); got != "DCC, loco commands, accessory commands" {
		t.Fatalf("FormatCapabilities() = %q", got)
	}
}

func TestSystemStateFromMessages(t *testing.T) {
	data := make([]byte, 16)
	data[12] = 0x02 // track voltage off
	msgs := []Message{{
		Header: HeaderLANSystemStateDataChanged,
		Data:   data,
	}}

	state, err := SystemStateFromMessages(msgs)
	if err != nil {
		t.Fatalf("SystemStateFromMessages() error = %v", err)
	}
	if got := FormatXStatusFlags(state.CentralState); got != "track voltage off" {
		t.Fatalf("FormatXStatusFlags() = %q", got)
	}
}

func TestEncodeSystemState(t *testing.T) {
	msg := EncodeSystemState(SystemState{
		MainCurrent:         100,
		FilteredMainCurrent: 80,
		Temperature:         42,
		SupplyVoltage:       10000,
		VCCVoltage:          1000,
		Capabilities:        capabilityDCC | capabilityLocoCmds | capabilityAccessoryCmds,
	})
	if msg.Header != HeaderLANSystemStateDataChanged {
		t.Fatalf("Header = %#x, want %#x", msg.Header, HeaderLANSystemStateDataChanged)
	}
	want := []byte{
		0x64, 0x00, 0x00, 0x00, 0x50, 0x00, 0x2A, 0x00,
		0x10, 0x27, 0xE8, 0x03, 0x00, 0x00, 0x00, 0x31,
	}
	if string(msg.Data) != string(want) {
		t.Fatalf("EncodeSystemState().Data = % x, want % x", msg.Data, want)
	}
	if !IsSystemStateDataChanged(msg) {
		t.Fatal("IsSystemStateDataChanged(EncodeSystemState()) = false, want true")
	}
}

func TestEncodeSystemStateRoundTrip(t *testing.T) {
	tests := []SystemState{
		{},
		{
			MainCurrent:         -1,
			ProgCurrent:         -32768,
			FilteredMainCurrent: 32767,
			Temperature:         -20,
			SupplyVoltage:       0xffff,
			VCCVoltage:          18000,
			CentralState:        xStatusEmergencyStop | xStatusTrackVoltageOff,
			CentralStateEx:      centralStateExHighTemperature | centralStateExRCN213,
			Capabilities:        0xff,
		},
	}
	for _, in := range tests {
		got, err := SystemStateFromMessages([]Message{EncodeSystemState(in)})
		if err != nil {
			t.Fatalf("SystemStateFromMessages(EncodeSystemState(%+v)) error = %v", in, err)
		}
		if got != in {
			t.Fatalf("SystemStateFromMessages(EncodeSystemState(%+v)) = %+v", in, got)
		}
	}
}
