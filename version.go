package z21

import "fmt"

// LAN_X_GET_VERSION
type Version struct {
	XBusProtoVersion string
	CommandStationID CommandStation
}

type CommandStation struct {
	ID   uint8
	Name string
}

// ---------- Message interface ----------

func (m *Version) Pack() ([]byte, error) {
	return []byte{0x21, 0x21, 0x00}, nil
}

func (m *Version) Unpack(data []byte) error {
	var proto uint8
	var id uint8
	if err := UnpackFields(data[2:], &proto, &id); err != nil {
		return err
	}
	m.decodeProto(proto)
	m.decodeID(id)
	return nil
}

func (m *Version) EncapType() uint16 {
	return LAN_X
}

func (c CommandStation) String() string {
	cs := c.Name
	if cs == "" {
		cs = fmt.Sprintf("0x%02x", c.ID)
	}
	return cs
}

// ---------- helpers ----------

func (m *Version) decodeProto(proto uint8) {
	major := proto >> 4
	minor := proto & 0x0F
	m.XBusProtoVersion = fmt.Sprintf("V%d.%d", major, minor)
}

func (m *Version) decodeID(id uint8) {
	cs := CommandStation{ID: id}
	if id == 0x12 {
		cs.Name = "Z21"
	}
	m.CommandStationID = cs
}

// ---------- Correlatable interface ----------

func (m *Version) Key() (string, bool) {
	d := []byte{byte(LAN_X), byte(LAN_X_GET_VERSION)}
	f, err := fingerprint(d)
	if err != nil {
		return "", false
	}
	return f, true
}
