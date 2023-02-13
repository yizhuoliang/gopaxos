package main

import (
	pb "gopaxos/gopaxos"
	"reflect"
)

const (
	COMMAND   = 1
	RESPONSES = 2
	PROPOSAL  = 3
	DECISIONS = 4
	BEAT      = 5
	P1A       = 6
	P1B       = 7
	P2A       = 8
	P2B       = 9
	EMPTY     = 10
)

type State struct {
	ballotNumber int32
	ballotLeader int32
	accepted     [][]*pb.BSC
}

type PartialState struct {
	ballotNumber int32
	ballotLeader int32
	accepted     [][]*pb.BSC
}

func Apply(s State, msg pb.Message) (State, pb.Message) {
	reply := pb.Message{}
	switch msg.Type {
	case P1A:
		reply = s.P1ATransformation(msg)
	case P2A:
		reply = s.P2ATransformation(msg)
	}
	return s, reply
}

func (s *State) P1ATransformation(msg pb.Message) pb.Message {
	if msg.BallotNumber > s.ballotNumber || (msg.BallotNumber == s.ballotNumber && msg.LeaderId != s.ballotNumber) {
		s.ballotNumber = msg.BallotNumber
		s.ballotLeader = msg.LeaderId
	}
	var acceptedList []*pb.BSC
	for i := 1; i < int(s.ballotNumber); i++ {
		if s.accepted[int32(i)] != nil {
			acceptedList = append(acceptedList, s.accepted[int32(i)]...)
		}
	}
	return pb.Message{Type: P1B, AcceptorId: -1, BallotNumber: s.ballotNumber, BallotLeader: s.ballotLeader, Accepted: acceptedList}
}

func (s *State) P2ATransformation(msg pb.Message) pb.Message {
	if msg.Bsc.BallotNumber >= s.ballotNumber && s.ballotLeader == msg.LeaderId {
		s.ballotNumber = msg.Bsc.BallotNumber
		m := make([][]*pb.BSC, len(s.accepted))
		copy(m, s.accepted)
		s.accepted = m
		s.accepted[msg.Bsc.BallotNumber] = append(s.accepted[msg.Bsc.BallotNumber], msg.Bsc)
	}
	return pb.Message{Type: P2B, AcceptorId: -1, BallotNumber: s.ballotNumber, BallotLeader: s.ballotLeader}
}

func Inference(msg pb.Message) (interface{}, pb.Message) {
	switch msg.Type {
	case P1B:
		return P1BInference(msg), pb.Message{}
	case P2B:
		return P2BInference(msg), pb.Message{}
	default:
		return nil, pb.Message{}
	}
}

func P1BInference(msg pb.Message) PartialState {
	accepted := make([][]*pb.BSC, msg.BallotNumber)
	for _, bsc := range msg.Accepted {
		if accepted[bsc.BallotNumber] == nil {
			accepted[bsc.BallotNumber] = make([]*pb.BSC, 0)
		}
		accepted[bsc.BallotNumber] = append(accepted[bsc.BallotNumber], bsc)
	}
	return PartialState{ballotNumber: msg.BallotNumber, ballotLeader: msg.BallotLeader, accepted: accepted}
}

func P2BInference(msg pb.Message) PartialState {
	return PartialState{ballotNumber: msg.BallotNumber, ballotLeader: msg.BallotLeader, accepted: nil}
}

func PartialStateEnabled(ps PartialState, from State, msg pb.Message) (bool, bool, State) {
	switch msg.Type {
	case P1A:
		if msg.BallotNumber > ps.ballotNumber || (msg.BallotNumber <= from.ballotNumber && (msg.BallotNumber != from.ballotNumber || msg.LeaderId == from.ballotLeader)) {
			return false, false, State{}
		}
	case P2A:
		if msg.Bsc.BallotNumber > ps.ballotNumber || (msg.Bsc.BallotNumber < from.ballotNumber) {
			return false, false, State{}
		}
		found := false // msg must exist in ps
		for i := from.ballotNumber; i <= ps.ballotNumber; i++ {
			for _, bsc := range ps.accepted[i] {
				if reflect.DeepEqual(msg.Bsc, bsc) {
					return true, false, State{}
				}
			}
		}
		if !found {
			return false, false, State{}
		}
	}
	s, _ := Apply(from, msg)
	return true, true, s
}

func PartialStateMatched(ps PartialState, s State) bool {
	return ps.ballotLeader == s.ballotLeader && ps.ballotNumber == s.ballotNumber && AcceptedMatch(ps.accepted, s.accepted)
}

func (s *State) Equal(s1 *State) bool {
	return s.ballotLeader == s1.ballotLeader && s.ballotNumber == s1.ballotNumber && AcceptedMatch(s.accepted, s1.accepted)
}

func Drop(msg pb.Message, s State) bool {
	switch msg.Type {
	case P1A:
		return msg.BallotNumber <= s.ballotNumber && (msg.BallotNumber != s.ballotNumber || msg.LeaderId == s.ballotLeader)
	case P2A:
		return msg.Bsc.BallotNumber < s.ballotNumber || msg.LeaderId != s.ballotLeader
	default:
		return true
	}
}

// invariant L1 ensures unique ballot-slot -> command
// for any slot, bsc with the highest ballot matters
// a1 matches a2 iff their slots map to the same highest ballot
func AcceptedMatch(a1 [][]*pb.BSC, a2 [][]*pb.BSC) bool {
	if len(a1) != len(a2) {
		return false
	}
	a1SlotToBallot := make([]int32, len(a1))
	for ballot, bscs := range a1 {
		for _, bsc := range bscs {
			if a1SlotToBallot[bsc.SlotNumber] < int32(ballot) {
				a1SlotToBallot[bsc.SlotNumber] = int32(ballot)
			}
		}
	}
	a2SlotToBallot := make([]int32, len(a2))
	for ballot, bscs := range a2 {
		for _, bsc := range bscs {
			if a2SlotToBallot[bsc.SlotNumber] < int32(ballot) {
				a2SlotToBallot[bsc.SlotNumber] = int32(ballot)
			}
		}
	}
	for slot, ballot := range a1SlotToBallot {
		if a2SlotToBallot[slot] != ballot {
			return false
		}
	}
	return true
}
