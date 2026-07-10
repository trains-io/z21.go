package protocol

// RequestExpectsReply reports whether the command station normally answers req.
func RequestExpectsReply(req Message) bool {
	switch req.Header {
	case HeaderLANSetBroadcastFlags, HeaderLANLogoff:
		return false
	default:
		return true
	}
}

// RequestsExpectingReplies returns the subset of reqs that expect a reply.
func RequestsExpectingReplies(reqs []Message) []Message {
	if len(reqs) == 0 {
		return nil
	}
	out := make([]Message, 0, len(reqs))
	for _, req := range reqs {
		if RequestExpectsReply(req) {
			out = append(out, req)
		}
	}
	return out
}

// ReplyMatchesRequest reports whether reply is a valid response to req.
func ReplyMatchesRequest(req, reply Message) bool {
	if req.Header == HeaderLANSystemStateGetData {
		return reply.Header == HeaderLANSystemStateDataChanged && IsSystemStateDataChanged(reply)
	}
	if req.Header != reply.Header {
		return false
	}
	if req.Header != HeaderLANX {
		return true
	}
	return lanXReplyMatchesRequest(req.Data, reply.Data)
}

// MatchedReplies returns the subset of replies paired to reqs, in request order.
// Unmatched datasets such as broadcast events are omitted.
func MatchedReplies(reqs []Message, replies []Message) []Message {
	if len(reqs) == 0 {
		return nil
	}

	used := make([]bool, len(replies))
	matched := make([]Message, 0, len(reqs))
	for _, req := range reqs {
		for i, reply := range replies {
			if used[i] {
				continue
			}
			if ReplyMatchesRequest(req, reply) {
				used[i] = true
				matched = append(matched, reply)
				break
			}
		}
	}
	return matched
}

// AllRequestsAnswered reports whether replies contain a matching response for
// every request, ignoring unrelated datasets such as broadcast events.
func AllRequestsAnswered(reqs []Message, replies []Message) bool {
	return len(MatchedReplies(reqs, replies)) == len(reqs)
}

func lanXReplyMatchesRequest(reqData, replyData []byte) bool {
	if len(reqData) < 1 || len(replyData) < 1 {
		return false
	}

	switch reqData[0] {
	case xHeaderGetVersion:
		if len(reqData) < 2 {
			return false
		}
		switch reqData[1] {
		case xHeaderGetVersion:
			return len(replyData) >= 2 &&
				replyData[0] == xHeaderGetVersionReply &&
				replyData[1] == xHeaderGetVersion
		case xCommandGetStatus:
			return len(replyData) >= 2 &&
				replyData[0] == xHeaderStatusReply &&
				replyData[1] == xStatusDB0
		case xCommandSetTrackPowerOff, xCommandSetTrackPowerOn:
			_, err := ParseTrackPowerBC(replyData)
			return err == nil
		default:
			return false
		}
	case xHeaderSetStop:
		return ParseBCStopped(replyData) == nil
	case xCommandGetFirmware:
		return len(replyData) >= 2 &&
			replyData[0] == xHeaderFirmwareReply &&
			replyData[1] == xFirmwareDB0
	default:
		return false
	}
}
