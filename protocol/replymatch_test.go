package protocol

import "testing"

func TestAllRequestsAnsweredIgnoresBroadcasts(t *testing.T) {
	reqs := []Message{
		GetSerialNumber(),
		GetHWInfo(),
		GetXVersion(),
		GetXFirmware(),
		GetCode(),
	}
	replies := []Message{
		{Header: HeaderLANGetSerialNumber, Data: []byte{0x01, 0x02, 0x03, 0x04}},
		{Header: HeaderLANSystemStateDataChanged, Data: make([]byte, 16)},
		{Header: HeaderLANGetHWInfo, Data: make([]byte, 8)},
		{Header: HeaderLANX, Data: []byte{0x63, 0x21, 0x30, 0x12, 0x60}},
		{Header: HeaderLANX, Data: []byte{0xF3, 0x0A, 0x01, 0x42, 0xB8}},
	}

	if AllRequestsAnswered(reqs, replies[:2]) {
		t.Fatal("expected missing replies after serial and broadcast")
	}
	if AllRequestsAnswered(reqs, replies[:4]) {
		t.Fatal("expected missing firmware and code replies")
	}
	if !AllRequestsAnswered(reqs, append(replies, Message{Header: HeaderLANGetCode, Data: []byte{0x00}})) {
		t.Fatal("expected all info requests to be answered")
	}
}

func TestRequestExpectsReply(t *testing.T) {
	if RequestExpectsReply(SetBroadcastFlags(0x101)) {
		t.Fatal("expected SetBroadcastFlags not to expect a reply")
	}
	if !RequestExpectsReply(GetHWInfo()) {
		t.Fatal("expected GetHWInfo to expect a reply")
	}

	reqs := RequestsExpectingReplies([]Message{
		SetBroadcastFlags(0x101),
		GetHWInfo(),
		SetStop(),
	})
	if len(reqs) != 2 {
		t.Fatalf("RequestsExpectingReplies() len = %d, want 2", len(reqs))
	}
}

func TestReplyMatchesRequestXStatus(t *testing.T) {
	req := GetXStatus()
	reply := Message{
		Header: HeaderLANX,
		Data:   []byte{0x62, 0x22, 0x00, 0x40},
	}
	if !ReplyMatchesRequest(req, reply) {
		t.Fatal("expected status reply to match GetXStatus request")
	}

	matched := MatchedReplies([]Message{req}, []Message{reply})
	if len(matched) != 1 {
		t.Fatalf("MatchedReplies() len = %d, want 1", len(matched))
	}
}

func TestReplyMatchesRequestSystemState(t *testing.T) {
	req := SystemStateGetData()
	reply := Message{Header: HeaderLANSystemStateDataChanged, Data: make([]byte, 16)}
	if !ReplyMatchesRequest(req, reply) {
		t.Fatal("expected system state reply to match GetData request")
	}

	matched := MatchedReplies([]Message{req}, []Message{reply})
	if len(matched) != 1 {
		t.Fatalf("MatchedReplies() len = %d, want 1", len(matched))
	}
}

func TestMatchedRepliesFiltersBroadcasts(t *testing.T) {
	reqs := []Message{GetSerialNumber(), GetHWInfo()}
	replies := []Message{
		{Header: HeaderLANGetSerialNumber, Data: []byte{0x01, 0x02, 0x03, 0x04}},
		{Header: HeaderLANSystemStateDataChanged, Data: make([]byte, 16)},
		{Header: HeaderLANGetHWInfo, Data: make([]byte, 8)},
	}

	matched := MatchedReplies(reqs, replies)
	if len(matched) != 2 {
		t.Fatalf("MatchedReplies() len = %d, want 2", len(matched))
	}
	if matched[0].Header != HeaderLANGetSerialNumber || matched[1].Header != HeaderLANGetHWInfo {
		t.Fatalf("MatchedReplies() = %#v", matched)
	}
}

func TestAllRequestsAnsweredDuplicateRequests(t *testing.T) {
	reqs := []Message{GetHWInfo(), GetHWInfo()}
	replies := []Message{{Header: HeaderLANGetHWInfo, Data: make([]byte, 8)}}
	if AllRequestsAnswered(reqs, replies) {
		t.Fatal("expected two HW replies for two HW requests")
	}
	replies = append(replies, Message{Header: HeaderLANGetHWInfo, Data: make([]byte, 8)})
	if !AllRequestsAnswered(reqs, replies) {
		t.Fatal("expected duplicate HW requests to be satisfied")
	}
}
