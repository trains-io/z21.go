package protocol

import (
	"reflect"
	"testing"
)

func FuzzParseAll(f *testing.F) {
	seeds := []Message{
		GetHWInfo(),
		EncodeHWInfo(HWInfo{HwType: HwTypeZ21New, FirmwareVersion: 0x0143}),
		EncodeLocoInfo(LocoInfo{Address: 1234, SpeedSteps: 4, Forward: true, Speed: 42}),
		EncodeSystemState(SystemState{MainCurrent: 100, Temperature: 42}),
		EncodeCANDetector(CANDetectorReport{NetID: 0xC101, Addr: 1, Type: CANDetectorTypeOccupancy, Value1: 0x0100}),
	}
	for _, msg := range seeds {
		b, err := msg.Marshal()
		if err != nil {
			f.Fatal(err)
		}
		f.Add(b)
	}
	multi, err := MarshalAll(GetTurnoutInfo(4), GetTurnoutInfo(5), GetRMBusData(0))
	if err != nil {
		f.Fatal(err)
	}
	f.Add(multi)
	f.Add([]byte{0x14, 0x00, 0x40, 0x00, 0x21}) // DataLen longer than the packet (Z21Posix quirk)

	f.Fuzz(func(t *testing.T, b []byte) {
		msgs, err := ParseAll(b)
		if err != nil || len(msgs) == 0 {
			return
		}
		out, err := MarshalAll(msgs...)
		if err != nil {
			t.Fatalf("MarshalAll(ParseAll(% x)) error = %v", b, err)
		}
		again, err := ParseAll(out)
		if err != nil {
			t.Fatalf("ParseAll(MarshalAll(...)) error = %v", err)
		}
		if !reflect.DeepEqual(msgs, again) {
			t.Fatalf("ParseAll round trip changed messages:\n got %+v\nwant %+v", again, msgs)
		}
	})
}

// FuzzParsers feeds arbitrary dataset payloads to every parser. None may
// panic, and values decoded by parsers that have an encoder or builder must
// survive a re-encode.
func FuzzParsers(f *testing.F) {
	output := uint8(3)
	seeds := []Message{
		EncodeLocoInfo(LocoInfo{Address: 3, SpeedSteps: 4, Headlight: true, FunctionsF5F12: 0xFF}),
		EncodeTurnoutInfo(TurnoutInfo{Address: 4, Position: TurnoutPositionOutput2}),
		EncodeExtAccessoryInfo(ExtAccessoryInfo{Address: 4, Value: 5}),
		EncodeCVResult(CVResult{Address: 28, Value: 6}),
		EncodeXVersion(XVersion{XBusVersion: 0x30, CommandStationID: 0x12}),
		EncodeXStatus(XStatus{CentralState: 0x02}),
		EncodeXFirmware(XFirmware{VersionMSB: 0x01, VersionLSB: 0x43}),
		EncodeTrackPowerBC(true),
		EncodeBCStopped(),
		EncodeCVNack(),
		EncodeSystemState(SystemState{MainCurrent: -1, Temperature: 42}),
		EncodeRailComData(RailComData{LocoAddress: 3, ReceiveCounter: 7}),
		EncodeRMBusStatus(RMBusStatus{GroupIndex: 1}),
		EncodeCANBoosterSystemState(CANBoosterSystemState{NetID: 0xC101, OutputPort: 1}),
		GetLocoInfo(3),
		SetLocoDrive(1234, LocoSpeedSteps128, true, 42),
		SetLocoFunction(3, 4, LocoFunctionToggle),
		SetLocoFunctionGroup(3, LocoFunctionGroupF0F4, 0x1F),
		SetLocoBinaryState(3, 300, true),
		SetLocoEStop(3),
		PurgeLoco(3),
		SetTurnout(4, TurnoutSwitch{Activate: true, Output2: true}),
		SetExtAccessory(4, 5),
		ReadCV(28),
		WriteCV(28, 6),
		POMLocoWriteBit(3, 28, 5, true),
		POMAccessoryWriteByte(42, 1, 0x0A, &output),
		SetBroadcastFlags(DefaultBroadcastFlags),
		GetRailComData(3),
		SetLocoMode(3, OutputModeMM),
	}
	for _, msg := range seeds {
		f.Add(msg.Data)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		callAllParsers(data)

		if v, err := ParseLocoInfo(data); err == nil {
			roundTrip(t, "LocoInfo", v, func() (LocoInfo, error) { return ParseLocoInfo(EncodeLocoInfo(v).Data) })
		}
		if v, err := ParseTurnoutInfo(data); err == nil {
			roundTrip(t, "TurnoutInfo", v, func() (TurnoutInfo, error) { return ParseTurnoutInfo(EncodeTurnoutInfo(v).Data) })
		}
		if v, err := ParseExtAccessoryInfo(data); err == nil {
			roundTrip(t, "ExtAccessoryInfo", v, func() (ExtAccessoryInfo, error) {
				return ParseExtAccessoryInfo(EncodeExtAccessoryInfo(v).Data)
			})
		}
		if v, err := ParseCVResult(data); err == nil {
			roundTrip(t, "CVResult", v, func() (CVResult, error) { return ParseCVResult(EncodeCVResult(v).Data) })
		}
		if v, err := ParseSystemState(data); err == nil {
			roundTrip(t, "SystemState", v, func() (SystemState, error) { return ParseSystemState(EncodeSystemState(v).Data) })
		}
		if v, err := ParseRailComData(data); err == nil {
			roundTrip(t, "RailComData", v, func() (RailComData, error) { return ParseRailComData(EncodeRailComData(v).Data) })
		}
		if v, err := ParseRMBusStatus(data); err == nil {
			roundTrip(t, "RMBusStatus", v, func() (RMBusStatus, error) { return ParseRMBusStatus(EncodeRMBusStatus(v).Data) })
		}
		if v, err := ParseCANDetector(data); err == nil {
			roundTrip(t, "CANDetectorReport", v, func() (CANDetectorReport, error) {
				return ParseCANDetector(EncodeCANDetector(v).Data)
			})
		}
		if v, err := ParseCANBoosterSystemState(data); err == nil {
			roundTrip(t, "CANBoosterSystemState", v, func() (CANBoosterSystemState, error) {
				return ParseCANBoosterSystemState(EncodeCANBoosterSystemState(v).Data)
			})
		}
		if v, err := ParseSetLocoDrive(data); err == nil {
			roundTrip(t, "LocoDrive", v, func() (LocoDrive, error) {
				return ParseSetLocoDrive(SetLocoDrive(v.Address, v.SpeedSteps, v.Forward, v.Speed).Data)
			})
		}
		if addr, sw, err := ParseSetTurnout(data); err == nil {
			gotAddr, gotSw, err := ParseSetTurnout(SetTurnout(addr, sw).Data)
			if err != nil || gotAddr != addr || gotSw != sw {
				t.Fatalf("SetTurnout round trip: got (%d, %+v, %v), want (%d, %+v)", gotAddr, gotSw, err, addr, sw)
			}
		}
		if v, err := ParsePOMLoco(data); err == nil {
			roundTrip(t, "POMRequest (loco)", v, func() (POMRequest, error) { return ParsePOMLoco(buildPOMLoco(v).Data) })
		}
		if v, err := ParsePOMAccessory(data); err == nil {
			roundTrip(t, "POMRequest (accessory)", v, func() (POMRequest, error) {
				return ParsePOMAccessory(buildPOMAccessory(v).Data)
			})
		}
	})
}

func roundTrip[T any](t *testing.T, name string, want T, again func() (T, error)) {
	t.Helper()
	got, err := again()
	if err != nil {
		t.Fatalf("%s round trip error = %v (value %+v)", name, err, want)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s round trip = %+v, want %+v", name, got, want)
	}
}

func buildPOMLoco(r POMRequest) Message {
	switch r.Operation {
	case POMWriteByte:
		return POMLocoWriteByte(r.Address, r.CV, r.Value)
	case POMWriteBit:
		return POMLocoWriteBit(r.Address, r.CV, r.Bit, r.BitValue)
	default:
		return POMLocoReadByte(r.Address, r.CV)
	}
}

func buildPOMAccessory(r POMRequest) Message {
	var output *uint8
	if r.HasOutput {
		output = &r.Output
	}
	switch r.Operation {
	case POMWriteByte:
		return POMAccessoryWriteByte(r.Address, r.CV, r.Value, output)
	case POMWriteBit:
		return POMAccessoryWriteBit(r.Address, r.CV, r.Bit, r.BitValue, output)
	default:
		return POMAccessoryReadByte(r.Address, r.CV, output)
	}
}

// callAllParsers runs every byte-slice parser and predicate; results are
// discarded because only panics matter here.
func callAllParsers(data []byte) {
	_, _ = ParseBroadcastFlags(data)
	_, _ = ParseSetBroadcastFlags(data)
	_, _ = ParseCANBoosterSystemState(data)
	_, _, _ = ParseCANDeviceDescription(data)
	_, _ = ParseGetCANDeviceDescription(data)
	_, _, _ = ParseSetCANBoosterTrackPower(data)
	_, _, _ = ParseSetCANDeviceDescription(data)
	_, _ = ParseCANDetector(data)
	_, _ = ParseGetCANDetector(data)
	_, _, _, _ = ParseCANMaintenanceSetAddress(data)
	_, _ = ParseCode(data)
	_ = IsCVNack(data)
	_ = IsCVNackSC(data)
	_ = IsCVResult(data)
	_, _ = ParseCVResult(data)
	_, _ = ParseReadCV(data)
	_, _ = ParseReadDCCRegister(data)
	_, _, _ = ParseWriteCV(data)
	_, _, _ = ParseWriteDCCRegister(data)
	_, _, _ = ParseWriteMMByte(data)
	_, _ = ParseHWInfo(data)
	_, _ = ParseSerialNumber(data)
	_ = IsLocoInfo(data)
	_, _ = ParseGetLocoInfo(data)
	_, _ = ParseLocoInfo(data)
	_, _ = ParsePurgeLoco(data)
	_, _, _, _ = ParseSetLocoBinaryState(data)
	_, _ = ParseSetLocoDrive(data)
	_, _ = ParseSetLocoEStop(data)
	_, _, _, _ = ParseSetLocoFunction(data)
	_, _, _, _ = ParseSetLocoFunctionGroup(data)
	_, _ = ParseGetLocoMode(data)
	_, _ = ParseLocoMode(data)
	_, _ = ParseSetLocoMode(data)
	_, _ = ParsePOMAccessory(data)
	_, _ = ParsePOMLoco(data)
	_, _ = ParseGetRailComData(data)
	_, _ = ParseRailComData(data)
	_, _ = ParseGetRMBusData(data)
	_, _ = ParseProgramRMBusModule(data)
	_, _ = ParseRMBusStatus(data)
	_ = IsBCStopped(data)
	_ = ParseBCStopped(data)
	_, _ = ParseSystemState(data)
	_ = IsTrackPowerBC(data)
	_, _ = ParseTrackPowerBC(data)
	_ = IsExtAccessoryInfo(data)
	_ = IsTurnoutInfo(data)
	_, _ = ParseExtAccessoryInfo(data)
	_, _ = ParseGetExtAccessoryInfo(data)
	_, _ = ParseGetTurnoutInfo(data)
	_, _, _ = ParseSetExtAccessory(data)
	_, _, _ = ParseSetTurnout(data)
	_, _ = ParseTurnoutInfo(data)
	_, _ = ParseGetTurnoutMode(data)
	_, _ = ParseSetTurnoutMode(data)
	_, _ = ParseTurnoutMode(data)
	_ = IsBCProgrammingMode(data)
	_ = IsBCShortCircuit(data)
	_ = IsUnknownCommand(data)
	_ = ParseBCProgrammingMode(data)
	_ = ParseBCShortCircuit(data)
	_ = ParseUnknownCommand(data)
	_, _ = ParseXFirmware(data)
	_ = IsXStatusChanged(data)
	_, _ = ParseXStatus(data)
	_, _ = ParseXVersion(data)
	_, _, _ = ParseLANX(Message{Header: HeaderLANX, Data: data})
}
