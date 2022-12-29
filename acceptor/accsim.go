package main

import (
	pb "gopaxos/gopaxos"
)

type State struct {
	ballotNumber int32
	ballotLeader int32
	accepted     map[int32][]*pb.BSC
}

func (s *State) P2ATransformation(message *pb.P2A) {
	if s.ballotNumber == message.Bsc.BallotNumber && s.ballotLeader == message.LeaderId {
		s.accepted[message.Bsc.BallotNumber] = append(accepted[message.Bsc.BallotNumber], message.Bsc)
	}
}

func (s *State) P1ATransformation(message *pb.P1A) {
	if message.BallotNumber > s.ballotNumber || (message.BallotNumber == s.ballotNumber && message.LeaderId != s.ballotNumber) {
		s.ballotNumber = message.BallotNumber
		s.ballotLeader = message.LeaderId
	}
}
