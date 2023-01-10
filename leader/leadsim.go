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

// TODO: if we don't check P1A, we can combine ballotNumber and adoptedBallotNumber
type State struct {
	ballotNumber        int32
	proposals           map[int32]*pb.Proposal
	ongoingCommanders   []*CommanderState
	ongoingScouts       []*ScoutState
	adoptedBallotNumber int32
	decisions           map[int32]*pb.Command
	highestSlot         int32

	// constants
	acceptorNum int32
	leaderId    int32
}

type PartialState struct {
	ballottNumber       int32
	newProposal         *pb.Proposal
	newCommander        *CommanderState
	newScout            *ScoutState
	adoptedBallotNumber int32
	decisions           map[int32]*pb.Command
	highestSlot         int32
}

func (s *State) ProposalTransformation(msg *pb.Proposal) {
	_, ok := s.proposals[msg.SlotNumber]
	if !ok {
		s.proposals[msg.SlotNumber] = msg
		s.ongoingCommanders = append(s.ongoingCommanders, &CommanderState{ballotNumber: s.ballotNumber, bsc: &pb.BSC{BallotNumber: s.ballotNumber, SlotNumber: msg.SlotNumber, Command: msg.Command}, ackCount: 0})
	}
}

func (s *State) P1BTransformation(msg *pb.P1B) {
	// of course there should be only 1 ongoing scout at a time, but I don't restrict it here
	for i, scout := range s.ongoingScouts {
		if msg.BallotNumber == scout.ballotNumber && msg.BallotLeader == s.leaderId && !scout.ackAcceptors[msg.AcceptorId] {
			scout.ackAcceptors[msg.AcceptorId] = true
			scout.ackCount++
			// TRIGGER ADOPTION
			if scout.ackCount >= s.acceptorNum/2+1 && scout.ballotNumber == s.ballotNumber {
				s.adoptedBallotNumber = s.ballotNumber
				// clean-up this scout
				s.ongoingScouts = append(s.ongoingScouts[:i], s.ongoingScouts[i+1:]...)
			}
			break
		}
	}
}

func (s *State) P2BTransformation(msg *pb.P2B) {
	for i, commander := range s.ongoingCommanders {
		// since we don't know which commander is this P2B responding, just find a commander that can take this P2B
		// yes there is a concern, I put it in conceern1.png
		if msg.BallotNumber == commander.ballotNumber && msg.BallotLeader == s.leaderId && !commander.ackAcceptors[msg.AcceptorId] {
			commander.ackAcceptors[msg.AcceptorId] = true
			commander.ackCount++
			// RECORD THIS SLOT BEING DECIDED (assuming that the bsc for the slot in "proposals" won't change latter)
			if commander.ackCount >= s.acceptorNum/2+1 {
				s.decisions[commander.bsc.SlotNumber] = commander.bsc.Command
				if commander.bsc.SlotNumber > s.highestSlot {
					s.highestSlot = commander.bsc.SlotNumber
				}
				// clean-up this commander
				s.ongoingCommanders = append(s.ongoingCommanders[:i], s.ongoingCommanders[i+1:]...)
			}
			break
		}
	}
}

func DecisionInference(msg *pb.Decisions) PartialState {
	decisions := make(map[int32]*pb.Command)
	highestSlot := int32(0)
	for _, decision := range msg.Decisions {
		decisions[decision.SlotNumber] = decision.Command
		if decision.SlotNumber > highestSlot {
			highestSlot = decision.SlotNumber
		}
	}
	return PartialState{ballottNumber: -1, adoptedBallotNumber: -1, decisions: decisions, highestSlot: highestSlot}
}

func P1AInference(msg *pb.P1A) PartialState {

	return PartialState{newScout: ScoutStateConstructor()}
}

func P2AInference(msg *pb.P2A) PartialState {
	return PartialState{ballottNumber: msg.Bsc.BallotNumber, adoptedBallotNumber: msg.Bsc.BallotNumber, newProposal: msg.Bsc, decisions: nil, highestSlot: -1}
}

// TODO
func ScoutStateConstructor() *ScoutState {
	return nil
}
