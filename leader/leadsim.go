package main

import (
	pb "gopaxos/gopaxos"
)

type CommanderState struct {
	ballotNumber int32
	bsc          *pb.BSC
	ackCount     int32
	ackAcceptors map[int32]bool
}

type ScoutState struct {
	ballotNumber int32
	ackCount     int32
	ackAcceptors map[int32]bool
}

type State struct {
	ballotNumber      int32
	proposals         map[int32]*pb.Proposal
	ongoingCommanders []*CommanderState
	ongoingScouts     []*ScoutState
	decidedSlots      map[int32]bool
	adoptedBallots    map[int32]bool

	// constants
	acceptorNum int32
	leaderId    int32
}

func (s *State) ProposalTransformation(message *pb.Proposal) {
	_, ok := s.proposals[message.SlotNumber]
	if !ok {
		s.proposals[message.SlotNumber] = message
		s.ongoingCommanders = append(s.ongoingCommanders, &CommanderState{ballotNumber: s.ballotNumber, bsc: &pb.BSC{BallotNumber: s.ballotNumber, SlotNumber: message.SlotNumber, Command: message.Command}, ackCount: 0})
	}
}

func (s *State) P1BTransformation(message *pb.P1B) {
	// of course there should be only 1 ongoing scout at a time, but I don't restrict it here
	for i, scout := range s.ongoingScouts {
		if message.BallotNumber == scout.ballotNumber && message.BallotLeader == s.leaderId && !scout.ackAcceptors[message.AcceptorId] {
			scout.ackAcceptors[message.AcceptorId] = true
			scout.ackCount++
			// TRIGGER ADOPTION
			if scout.ackCount >= s.acceptorNum/2+1 && scout.ballotNumber == s.ballotNumber {
				s.adoptedBallots[s.ballotNumber] = true
				// clean-up this scout
				s.ongoingScouts = append(s.ongoingScouts[:i], s.ongoingScouts[i+1:]...)
			}
			break
		}
	}
}

func (s *State) P2BTransformation(message *pb.P2B) {
	for i, commander := range s.ongoingCommanders {
		// since we don't know which commander is this P2B responding, just find a commander that can take this P2B
		// yes there is a concern, I put it in conceern1.png
		if message.BallotNumber == commander.ballotNumber && message.BallotLeader == s.leaderId && !commander.ackAcceptors[message.AcceptorId] {
			commander.ackAcceptors[message.AcceptorId] = true
			commander.ackCount++
			// RECORD THIS SLOT BEING DECIDED (assuming that the bsc for the slot in "proposals" won't change latter)
			if commander.ackCount >= s.acceptorNum/2+1 {
				s.decidedSlots[commander.bsc.SlotNumber] = true
				// clean-up this commander
				s.ongoingCommanders = append(s.ongoingCommanders[:i], s.ongoingCommanders[i+1:]...)
			}
			break
		}
	}
}
