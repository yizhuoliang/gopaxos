package main

import (
	pb "gopaxos/gopaxos"
)

type State struct {
	ballotNumber int32
	ballotLeader int32
	accepted     map[int32][]*pb.BSC
}

type PartialState struct {
	ballotNumber int32
	ballotLeader int32
	accepted     map[int32][]*pb.BSC
}

func (s *State) P2ATransformation(msg *pb.Message) {
	if s.ballotNumber == msg.Bsc.BallotNumber && s.ballotLeader == msg.LeaderId {
		s.accepted[msg.Bsc.BallotNumber] = append(s.accepted[msg.Bsc.BallotNumber], msg.Bsc)
	}
}

func (s *State) P1ATransformation(msg *pb.Message) {
	if msg.BallotNumber > s.ballotNumber || (msg.BallotNumber == s.ballotNumber && msg.LeaderId != s.ballotNumber) {
		s.ballotNumber = msg.BallotNumber
		s.ballotLeader = msg.LeaderId
	}
}

func P1BInference(msg *pb.Message) PartialState {
	accepted = make(map[int32][]*pb.BSC)
	for _, bsc := range msg.Accepted {
		_, ok := accepted[bsc.BallotNumber]
		if !ok {
			accepted[bsc.BallotNumber] = make([]*pb.BSC, 0)
		}
		accepted[bsc.BallotNumber] = append(accepted[bsc.BallotNumber], bsc)
	}
	return PartialState{ballotNumber: msg.BallotNumber, ballotLeader: msg.BallotLeader, accepted: accepted}
}

func P2BInference(msg *pb.Message) PartialState {
	return PartialState{ballotNumber: msg.BallotNumber, ballotLeader: msg.BallotLeader, accepted: nil}
}
