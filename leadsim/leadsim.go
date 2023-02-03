package main

import (
	pb "gopaxos/gopaxos"
)

const (
	acceptorNum = 3
	leaderNum   = 2

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

	// constants
	leaderId int32
}

type PartialState struct {
	adoptedBallottNumber int32
	newProposal          *pb.Proposal
	decisions            []*pb.Decision
}

func Apply(s State, msg pb.Message) (State, pb.Message) {
	reply := pb.Message{}
	switch msg.Type {
	case PROPOSAL:
		s.ProposalTransformation(msg)
	case P1B:
		s.P1BTransformation(msg)
	case P2B:
		s.P2BTransformation(msg)
	}
	return s, reply
}

func (s *State) ProposalTransformation(msg pb.Message) {
	_, ok := s.proposals[msg.SlotNumber]
	if !ok {
		p := make(map[int32]*pb.Proposal)
		mapCopy(p, s.proposals)
		s.proposals[msg.SlotNumber] = &pb.Proposal{SlotNumber: msg.SlotNumber, Command: msg.Command}
		s.ongoingCommanders = append(s.ongoingCommanders, &CommanderState{ballotNumber: s.adoptedBallotNumber, bsc: &pb.BSC{BallotNumber: s.adoptedBallotNumber, SlotNumber: msg.SlotNumber, Command: msg.Command}, ackCount: 0})
	}
}

/*	KEY IDEAS:

	We DO NOT launch a scout actively here, since we allow the leader to launch a scout at any time,
	which doesn't affect correctness. Rather, we add a scout state when receive a P1B acknowledgement
	to this leader's higher ballot.

	On the other hand, we know that leaders can only launch commanders after receiving proposals and
	after adoption, so we add commander states at those points.
*/

func (s *State) P1BTransformation(msg pb.Message) {
	// DROP STALE P1B
	if msg.BallotNumber <= s.adoptedBallotNumber {
		return
	}

	registered := false
	copied := false
	// of course there should be only 1 ongoing scout at a time, but I don't restrict it here
	for i, scout := range s.ongoingScouts {
		// Case: stale scout (AVOID ADDING SMALLER SCOUTS WHEN HIGHER SCOUT EXIST)
		if msg.BallotNumber > scout.ballotNumber || (msg.BallotNumber == scout.ballotNumber && msg.BallotLeader != s.leaderId) {
			if !copied {
				scouts := make([]*ScoutState, len(s.ongoingScouts))
				copy(scouts, s.ongoingScouts)
				s.ongoingScouts = scouts
				copied = true
			}
			s.ongoingScouts = append(s.ongoingScouts[:i], s.ongoingScouts[i+1:]...)
			continue
		} else if msg.BallotNumber < scout.ballotNumber {
			// AVOID ADDING SMALLER SCOUTS WHEN HIGHER SCOUT EXIST
			registered = true
		}

		// Case: scout can be updated
		if msg.BallotNumber == scout.ballotNumber && msg.BallotLeader == s.leaderId && !scout.ackAcceptors[msg.AcceptorId] {
			registered = true
			// UPDATE THIS SCOUT STATE
			if !copied {
				scouts := make([]*ScoutState, len(s.ongoingScouts))
				copy(scouts, s.ongoingScouts)
				s.ongoingScouts = scouts
				copied = true
			}
			scout.ackAcceptors[msg.AcceptorId] = true
			scout.pvalues = append(scout.pvalues, msg.Accepted...)
			scout.ackCount++
			// TRIGGER ADOPTION
			if scout.ackCount >= acceptorNum/2+1 && scout.ballotNumber > s.adoptedBallotNumber {
				adoption(s, scout)
				// clean-up this scout
				s.ongoingScouts = append(s.ongoingScouts[:i], s.ongoingScouts[i+1:]...)
			}
		}
	}
	// Case: the scout isn't registered
	if !registered && msg.BallotLeader == s.leaderId {
		if !copied {
			scouts := make([]*ScoutState, len(s.ongoingScouts))
			copy(scouts, s.ongoingScouts)
			s.ongoingScouts = scouts
		}
		originalLen := len(s.ongoingScouts)
		newScout := ScoutStateConstructor(msg.BallotNumber)
		newScout.ackAcceptors[msg.AcceptorId] = true
		newScout.ackCount = 1
		s.ongoingScouts = append(s.ongoingScouts, newScout)
		if newScout.ackCount >= acceptorNum/2+1 && newScout.ballotNumber > s.adoptedBallotNumber {
			adoption(s, newScout)
			// clean-up this scout
			s.ongoingScouts = append(s.ongoingScouts[:originalLen])
		}
	}
}

func (s *State) P2BTransformation(msg pb.Message) {
	copied := false
	for i, commander := range s.ongoingCommanders {
		// since we don't know which commander is this P2B responding, just find all commander that can take this P2B
		if msg.BallotNumber == commander.ballotNumber && msg.BallotLeader == s.leaderId && !commander.ackAcceptors[msg.AcceptorId] {
			if !copied {
				// copy on write
				commanders := make([]*CommanderState, len(s.ongoingCommanders))
				copy(commanders, s.ongoingCommanders)
				s.ongoingCommanders = commanders
				copied = true
			}
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

// func (s *State) Equal(s1 *State) bool {
// 	return s.adoptedBallotNumber == s1.adoptedBallotNumber && reflect.DeepEqual(s.proposals, s1.proposals) &&
// }

func DecisionInference(msg pb.Message) PartialState {
	return PartialState{adoptedBallottNumber: -1, decisions: msg.Decisions}
}

// we don't make inference from P1As, a leader can run a scout at anytime without breaking correctness
// what to do in this case?
func P1AInference(msg pb.Message) PartialState {
	return PartialState{adoptedBallottNumber: -1}
}

func P2AInference(msg pb.Message) PartialState {
	newProposal := &pb.Proposal{SlotNumber: msg.Bsc.SlotNumber, Command: msg.Bsc.Command}
	return PartialState{adoptedBallottNumber: msg.Bsc.BallotNumber, newProposal: newProposal}
}

func PartialStateMatched(end_s *interface{}, s State) (State, bool) {
	ps := (*end_s).(PartialState)

	if ps.adoptedBallottNumber != -1 && ps.adoptedBallottNumber != s.adoptedBallotNumber {
		return s, false
	}

	if ps.newProposal != nil {
		prop, ok := s.proposals[ps.newProposal.SlotNumber]
		if !ok || !commandMatched(prop.Command, ps.newProposal.Command) {
			return s, false
		}
	}

	if ps.decisions != nil {
		for _, decision := range ps.decisions {
			command, ok := s.decisions[decision.SlotNumber]
			if !ok || !commandMatched(decision.Command, command) {
				return s, false
			}
		}
	}
	return s, true
}

// --------- HELPER FUNCTIONS BELOW -------------

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

func mapCopy[T any](dst map[int32]*T, src map[int32]*T) {
	for key, val := range src {
		dst[key] = val
	}
}

func adoption(s *State, scout *ScoutState) {
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
			// copy on write
			commanders := make([]*CommanderState, len(s.ongoingCommanders))
			copy(commanders, s.ongoingCommanders)
			s.ongoingCommanders = append(commanders, CommanderStateConstructor(s.adoptedBallotNumber, &pb.BSC{BallotNumber: s.adoptedBallotNumber, SlotNumber: proposal.SlotNumber, Command: proposal.Command}))
		}
	}
}

func commandMatched(c1 *pb.Command, c2 *pb.Command) bool {
	return c1.ClientId == c2.ClientId && c1.CommandId == c1.CommandId
}