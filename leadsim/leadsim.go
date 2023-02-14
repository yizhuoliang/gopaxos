package main

import (
	pb "gopaxos/gopaxos"
	"reflect"
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
	role         int32
	ballotNumber int32
	bsc          *pb.BSC
	ackCount     int32
	ackAcceptors []bool
}

type ScoutState struct {
	role         int32
	ballotNumber int32
	ackCount     int32
	ackAcceptors []bool
	pvalues      []*pb.BSC
}

type State struct {
	adoptedBallotNumber int32
	proposals           map[int32]*pb.Proposal
	ongoingCommanders   []*CommanderState
	ongoingScout        *ScoutState
	decisions           []*pb.Decision

	// constants
	leaderId int32
}

type PartialState struct {
	adoptedBallotNumber int32
	newProposal         *pb.Proposal
	decisions           []*pb.Decision
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

	if s.ongoingScout != nil {
		// Case - Stale Scout
		if msg.BallotNumber > s.ongoingScout.ballotNumber || (msg.BallotNumber == s.ongoingScout.ballotNumber && msg.BallotLeader != s.leaderId) {
			s.ongoingScout = nil
		}

		// Case - Current Scout Update
		if msg.BallotNumber == s.ongoingScout.ballotNumber && msg.BallotLeader == s.leaderId && !s.ongoingScout.ackAcceptors[msg.AcceptorId] {
			newScout := new(ScoutState)
			scoutStateCopy(newScout, s.ongoingScout)
			s.ongoingScout = newScout
			s.ongoingScout.ackAcceptors[msg.AcceptorId] = true
			s.ongoingScout.ackCount++
			s.ongoingScout.pvalues = append(s.ongoingScout.pvalues, msg.Accepted...)
		}
	}

	// Case - Register new Scout
	if s.ongoingScout == nil && msg.BallotLeader == s.leaderId {
		newScout := scoutStateConstructor(msg.BallotNumber)
		s.ongoingScout = newScout
		s.ongoingScout.ackAcceptors[msg.AcceptorId] = true
		s.ongoingScout.ackCount++
		s.ongoingScout.pvalues = append(s.ongoingScout.pvalues, msg.Accepted...)
	}

	// TRIGGER ADOPTION
	if s.ongoingScout.ackCount >= acceptorNum/2+1 && s.ongoingScout.ballotNumber > s.adoptedBallotNumber {
		adoption(s)
		// clean-up this scout
		s.ongoingScout = nil
	}
}

func (s *State) P2BTransformation(msg pb.Message) {
	copiedCom := false
	copiedDec := false
	for i, commander := range s.ongoingCommanders {
		// since we don't know which commander is this P2B responding, just find all commander that can take this P2B
		if msg.BallotNumber == commander.ballotNumber && msg.BallotLeader == s.leaderId && !commander.ackAcceptors[msg.AcceptorId] {
			if !copiedCom {
				// copy on write
				commanders := make([]*CommanderState, len(s.ongoingCommanders))
				copy(commanders, s.ongoingCommanders)
				s.ongoingCommanders = commanders
				copiedCom = true
			}
			commander.ackAcceptors[msg.AcceptorId] = true
			commander.ackCount++
			// RECORD THIS SLOT BEING DECIDED (assuming that the bsc for the slot in "proposals" won't change latter)
			if commander.ackCount >= acceptorNum/2+1 {
				if !copiedDec {
					// copy on write
					decisions := make([]*pb.Decision, len(s.decisions))
					copy(decisions, s.decisions)
					s.decisions = decisions
					copiedDec = true
				}
				if len(s.decisions) <= int(commander.bsc.SlotNumber) {
					newChunk := make([]*pb.Decision, int(commander.bsc.SlotNumber)-len(s.decisions)+1)
					s.decisions = append(s.decisions, newChunk...) // TODO: test this part
				}
				s.decisions[commander.bsc.SlotNumber] = &pb.Decision{SlotNumber: commander.bsc.SlotNumber, Command: commander.bsc.Command}
				// clean-up this commander
				s.ongoingCommanders = append(s.ongoingCommanders[:i], s.ongoingCommanders[i+1:]...)
			}
			break
		}
	}
}

func DecisionInference(msg pb.Message) PartialState {
	decisions := make([]*pb.Decision, len(msg.Decisions))
	for _, decision := range msg.Decisions {
		decisions[decision.SlotNumber] = decision
	}
	return PartialState{adoptedBallotNumber: -1, decisions: decisions}
}

// Note that a leader can run a scout at anytime without breaking correctness
// and we actually don't make any inference from P1A
func P1AInference(msg pb.Message) PartialState {
	return PartialState{adoptedBallotNumber: -1}
}

func P2AInference(msg pb.Message) PartialState {
	newProposal := &pb.Proposal{SlotNumber: msg.Bsc.SlotNumber, Command: msg.Bsc.Command}
	return PartialState{adoptedBallotNumber: msg.Bsc.BallotNumber, newProposal: newProposal}
}

func Inference(msg pb.Message) (interface{}, pb.Message) {
	switch msg.Type {
	case DECISIONS:
		return DecisionInference(msg), pb.Message{}
	case P1A:
		return P1AInference(msg), pb.Message{}
	case P2A:
		return P2AInference(msg), pb.Message{}
	default:
		return nil, pb.Message{}
	}
}

func PartialStateMatched(end_s *interface{}, s State) (State, bool) {
	ps := (*end_s).(PartialState)

	if ps.adoptedBallotNumber != -1 && ps.adoptedBallotNumber != s.adoptedBallotNumber {
		return s, false
	}

	if ps.newProposal != nil {
		prop, ok := s.proposals[ps.newProposal.SlotNumber]
		if !ok || !reflect.DeepEqual(prop.Command, ps.newProposal.Command) {
			return s, false
		}
	}

	if ps.decisions != nil {
		for _, decision := range ps.decisions {
			if len(s.decisions) <= int(decision.SlotNumber) {
				return s, false
			}

			if s.decisions[decision.SlotNumber] == nil || !reflect.DeepEqual(decision.Command, s.decisions[decision.SlotNumber]) {
				return s, false
			}
		}
	}
	return s, true
}

func (s *State) Equal(s1 *State) bool {
	if s.leaderId != s1.leaderId {
		return false
	}

	if s.adoptedBallotNumber != s1.adoptedBallotNumber {
		return false
	}

	if !reflect.DeepEqual(s.proposals, s1.proposals) {
		return false
	}

	if !reflect.DeepEqual(s.ongoingScout, s1.ongoingScout) {
		return false
	}

	if !reflect.DeepEqual(s.decisions, s1.decisions) {
		return false
	}

	// TODO: think of this condition more carefully
	if len(s.ongoingCommanders) != len(s1.ongoingCommanders) {
		return false
	}

	commanderMap := make(map[int32]*CommanderState, len(s.ongoingCommanders)+1)

	for _, commander := range s.ongoingCommanders {
		originalCommander, ok := commanderMap[commander.bsc.SlotNumber]
		if !ok || originalCommander.bsc.BallotNumber < commander.ballotNumber {
			commanderMap[commander.ballotNumber] = commander
		}
	}

	for _, commadner := range s1.ongoingCommanders {
		concreteCommander, ok := commanderMap[commadner.bsc.SlotNumber]
		if !ok || !reflect.DeepEqual(commadner, concreteCommander) {
			return false
		}
	}

	return true
}

// first bool - if this partial state is reachable from "from"
// secon bool - if the transformation function is already applied to the returned state strcut
func PartialStateEnabled(end_s *interface{}, from State, msg pb.Message) (bool, bool, State) {
	ps := (*end_s).(PartialState)
	if ps.adoptedBallotNumber != -1 {
		if from.adoptedBallotNumber > ps.adoptedBallotNumber {
			return false, false, from
		}
	}

	if ps.newProposal != nil {
		slotPs := ps.newProposal.SlotNumber
		if len(from.decisions) > int(slotPs) && from.decisions[slotPs] != nil && from.decisions[slotPs].Command.CommandId != ps.newProposal.Command.CommandId {
			return false, false, from
		}
	}

	if ps.decisions != nil {
		// "from" state decisions must be the subset of partial state
		if len(from.decisions) > len(ps.decisions) {
			return false, false, from
		}

		for i, _ := range from.decisions {
			if !reflect.DeepEqual(from.decisions[i], ps.decisions[i]) {
				return false, false, from
			}
		}
	}

	return true, false, from
}

func Drop(msg pb.Message, s *State) bool {
	switch msg.Type {
	case PROPOSAL:
		_, ok := s.proposals[msg.SlotNumber]
		if !ok {
			return false
		}
	case P1B:
		if msg.BallotNumber > s.adoptedBallotNumber {
			return false
		}
	case P2B:
		if len(s.ongoingCommanders) != 0 {
			return false
		}
	}
	return true
}

// --------- HELPER FUNCTIONS BELOW -------------

func scoutStateConstructor(ballotNumber int32) *ScoutState {
	scout := new(ScoutState)
	scout.ackAcceptors = make([]bool, acceptorNum)
	for i := int32(0); i < acceptorNum; i++ {
		scout.ackAcceptors[i] = false
	}
	scout.ballotNumber = ballotNumber
	scout.pvalues = make([]*pb.BSC, 0)
	return scout
}

// note that, this is deep copy, but not too deep to copy the bsc structs
func scoutStateCopy(dst *ScoutState, src *ScoutState) {
	dst.ballotNumber = src.ballotNumber
	dst.ackCount = src.ackCount
	dst.ackAcceptors = make([]bool, acceptorNum)
	copy(dst.ackAcceptors, src.ackAcceptors)
	dst.pvalues = make([]*pb.BSC, len(src.pvalues))
	copy(dst.pvalues, src.pvalues)
}

func commanderStateConstructor(ballotNumber int32, bsc *pb.BSC) *CommanderState {
	commander := new(CommanderState)
	commander.ackAcceptors = make([]bool, acceptorNum)
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

func adoption(s *State) {
	s.adoptedBallotNumber = s.ongoingScout.ballotNumber
	slotToBallot := make(map[int32]int32) // map from slot number to ballot number to satisfy pmax
	for _, bsc := range s.ongoingScout.pvalues {
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
			s.ongoingCommanders = append(commanders, commanderStateConstructor(s.adoptedBallotNumber, &pb.BSC{BallotNumber: s.adoptedBallotNumber, SlotNumber: proposal.SlotNumber, Command: proposal.Command}))
		}
	}
}
