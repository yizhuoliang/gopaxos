package main

import (
	pb "gopaxos/gopaxos"
)

const (
	acceptorNum = 3
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
	pvalues      []*pb.BSC
}

type State struct {
	adoptedBallotNumber int32
	proposals           map[int32]*pb.Proposal
	ongoingCommanders   []*CommanderState
	ongoingScouts       []*ScoutState
	decisions           map[int32]*pb.Command
	highestSlot         int32
	leaderId            int32
}

type PartialState struct {
	adoptedBallottNumber int32
	newProposal          *pb.Proposal
	decisions            map[int32]*pb.Command
	highestSlot          int32
}

func (s *State) ProposalTransformation(msg *pb.Message) {
	_, ok := s.proposals[msg.SlotNumber]
	if !ok {
		p := make(map[int32]*pb.Proposal)
		copy(p, s.proposals) // write a func
		s.proposals[msg.SlotNumber] = &pb.Proposal{SlotNumber: msg.SlotNumber, Command: msg.Command}
		s.ongoingCommanders = append(s.ongoingCommanders, &CommanderState{ballotNumber: s.adoptedBallotNumber, bsc: &pb.BSC{BallotNumber: s.adoptedBallotNumber, SlotNumber: msg.SlotNumber, Command: msg.Command}, ackCount: 0})
	}
}

func (s *State) P1BTransformation(msg *pb.P1B) {
	// of course there should be only 1 ongoing scout at a time, but I don't restrict it here
	for i, scout := range s.ongoingScouts {
		if msg.BallotNumber == scout.ballotNumber && msg.BallotLeader == s.leaderId && !scout.ackAcceptors[msg.AcceptorId] {
			scout.ackAcceptors[msg.AcceptorId] = true
			scout.pvalues = append(scout.pvalues, msg.Accepted...)
			scout.ackCount++
			// TRIGGER ADOPTION
			if scout.ackCount >= acceptorNum/2+1 && scout.ballotNumber > s.adoptedBallotNumber {
				s.adoptedBallotNumber = scout.ballotNumber
				slotToBallot := make(map[int32]int32) // map from slot number to ballot number to satisfy pmax
				for _, bsc := range scout.pvalues {
					proposal, okProp := s.proposals[bsc.SlotNumber]
					if okProp {
						originalBallot, okBall := slotToBallot[proposal.SlotNumber]
						if (okBall && originalBallot < bsc.BallotNumber) || !okBall {
							// the previous proposal to that slot has a lower ballot number, or this is the first proposal to that slot
							s.proposals[bsc.SlotNumber] = &pb.Proposal{SlotNumber: bsc.SlotNumber, Command: bsc.Command}
							slotToBallot[proposal.SlotNumber] = bsc.BallotNumber
						}
					} else {
						// there was originally no proposal for that slot
						s.proposals[bsc.SlotNumber] = &pb.Proposal{SlotNumber: bsc.SlotNumber, Command: bsc.Command}
						slotToBallot[proposal.SlotNumber] = bsc.BallotNumber
					}
					// Spawn commanders
					for _, proposal := range s.proposals {
						s.ongoingCommanders = append(s.ongoingCommanders, CommanderStateConstructor(s.adoptedBallotNumber, &pb.BSC{BallotNumber: s.adoptedBallotNumber, SlotNumber: proposal.SlotNumber, Command: proposal.Command}))
					}
				}
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
			if commander.ackCount >= acceptorNum/2+1 {
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
	return PartialState{adoptedBallottNumber: -1, decisions: decisions, highestSlot: highestSlot}
}

// we don't make inference from P1As, a leader can run a scout at anytime without breaking correctness
func P1AInference(msg *pb.P1A) PartialState {
	return PartialState{adoptedBallottNumber: -1, highestSlot: -1}
}

func P2AInference(msg *pb.P2A) PartialState {
	newProposal := &pb.Proposal{SlotNumber: msg.Bsc.SlotNumber, Command: msg.Bsc.Command}
	return PartialState{adoptedBallottNumber: msg.Bsc.BallotNumber, newProposal: newProposal, highestSlot: -1}
}

func ScoutStateConstructor(ballotNumber int32) *ScoutState {
	scout := new(ScoutState)
	scout.ackAcceptors = make(map[int32]bool)
	for i := int32(0); i < acceptorNum; i++ {
		scout.ackAcceptors[i] = false
	}
	scout.ballotNumber = ballotNumber
	scout.pvalues = make([]*pb.BSC, 0)
	return scout
}

func CommanderStateConstructor(ballotNumber int32, bsc *pb.BSC) *CommanderState {
	commander := new(CommanderState)
	commander.ackAcceptors = make(map[int32]bool)
	for i := int32(0); i < acceptorNum; i++ {
		commander.ackAcceptors[i] = false
	}
	commander.ballotNumber = ballotNumber
	return commander
}
