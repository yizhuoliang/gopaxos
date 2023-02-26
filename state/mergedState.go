package main

import (
	"log"
	"reflect"

	pb "github.com/yizhuoliang/gopaxos"
)

const (
	acceptorNum = 3
	leaderNum   = 2

	// Message types
	COMMAND   = 1
	READ      = 2
	RESPONSES = 3
	PROPOSAL  = 4
	DECISIONS = 5
	BEAT      = 6
	P1A       = 7
	P1B       = 8
	P2A       = 9
	P2B       = 10
	EMPTY     = 11

	// Roles
	LEADER   = 0
	ACCEPTOR = 1
)

/*
NOTE: Changed names
leaderID / AccID -> me
*/
type State struct {
	me   int32
	role int32

	// For leader
	adoptedBallotNumber int32
	proposals           map[int32]*pb.Proposal
	ongoingCommanders   []*CommanderState
	ongoingScout        *ScoutState
	decisions           []*pb.Decision

	// For acceptor
	ballotNumber int32
	ballotLeader int32
	accepted     [][]*pb.BSC
}

// --- Transformation Functions --

func Apply(s State, msg pb.Message) (State, pb.Message) {
	switch s.role {
	case LEADER:
		return LeaderApply(s, msg)
	case ACCEPTOR:
		return AcceptorApply(s, msg)
	default:
		log.Fatalf("unkonwn role type")
	}
	return State{}, pb.Message{}
}

func LeaderApply(s State, msg pb.Message) (State, pb.Message) {
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

// Leader
func (s *State) ProposalTransformation(msg pb.Message) {
	_, ok := s.proposals[msg.SlotNumber]
	if !ok {
		p := make(map[int32]*pb.Proposal)
		mapCopy(p, s.proposals)
		s.proposals[msg.SlotNumber] = &pb.Proposal{SlotNumber: msg.SlotNumber, Command: msg.Command}
		s.ongoingCommanders = append(s.ongoingCommanders, &CommanderState{ballotNumber: s.adoptedBallotNumber, bsc: &pb.BSC{BallotNumber: s.adoptedBallotNumber, SlotNumber: msg.SlotNumber, Command: msg.Command}, ackCount: 0})
	}
}

// Leader
func (s *State) P1BTransformation(msg pb.Message) {
	// DROP STALE P1B
	if msg.BallotNumber <= s.adoptedBallotNumber {
		return
	}

	if s.ongoingScout != nil {
		// Case - Stale Scout
		if msg.BallotNumber > s.ongoingScout.ballotNumber || (msg.BallotNumber == s.ongoingScout.ballotNumber && msg.BallotLeader != s.me) {
			s.ongoingScout = nil
		}

		// Case - Current Scout Update
		if msg.BallotNumber == s.ongoingScout.ballotNumber && msg.BallotLeader == s.me && !s.ongoingScout.ackAcceptors[msg.AcceptorId] {
			newScout := new(ScoutState)
			scoutStateCopy(newScout, s.ongoingScout)
			s.ongoingScout = newScout
			s.ongoingScout.ackAcceptors[msg.AcceptorId] = true
			s.ongoingScout.ackCount++
			s.ongoingScout.pvalues = append(s.ongoingScout.pvalues, msg.Accepted...)
		}
	}

	// Case - Register new Scout
	if s.ongoingScout == nil && msg.BallotLeader == s.me {
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

// Leader
func (s *State) P2BTransformation(msg pb.Message) {
	copiedCom := false
	copiedDec := false
	for i, commander := range s.ongoingCommanders {
		// since we don't know which commander is this P2B responding, just find all commander that can take this P2B
		if msg.BallotNumber == commander.ballotNumber && msg.BallotLeader == s.me && !commander.ackAcceptors[msg.AcceptorId] {
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

// Acc
func AcceptorApply(s State, msg pb.Message) (State, pb.Message) {
	reply := pb.Message{}
	switch msg.Type {
	case P1A:
		reply = s.P1ATransformation(msg)
	case P2A:
		reply = s.P2ATransformation(msg)
	}
	return s, reply
}

// Acc
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
	return pb.Message{Type: P1B, AcceptorId: s.me, BallotNumber: s.ballotNumber, BallotLeader: s.ballotLeader, Accepted: acceptedList}
}

// Acc
func (s *State) P2ATransformation(msg pb.Message) pb.Message {
	if msg.Bsc.BallotNumber >= s.ballotNumber && s.ballotLeader == msg.LeaderId {
		s.ballotNumber = msg.Bsc.BallotNumber
		m := make([][]*pb.BSC, len(s.accepted))
		copy(m, s.accepted)
		s.accepted = m
		s.accepted[msg.Bsc.BallotNumber] = append(s.accepted[msg.Bsc.BallotNumber], msg.Bsc)
	}
	return pb.Message{Type: P2B, AcceptorId: s.me, BallotNumber: s.ballotNumber, BallotLeader: s.ballotLeader}
}

type CommanderState struct {
	ballotNumber int32
	bsc          *pb.BSC
	ackCount     int32
	ackAcceptors []bool
}

type ScoutState struct {
	ballotNumber int32
	ackCount     int32
	ackAcceptors []bool
	pvalues      []*pb.BSC
}

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

// OTHER REQUIRED FUNCTIONS
func (s *State) Equal(s1 *State) bool {
	if s.role != s1.role {
		return false
	}

	switch s.role {
	case LEADER:
		return s.LeaderEqual(s1)
	case ACCEPTOR:
		return s.AcceptorEqual(s1)
	default:
		log.Fatalf("pannnnnic: unknown role type")
		return false
	}
}

func (s *State) AcceptorEqual(s1 *State) bool {
	if s.me != s1.me {
		return false
	}
	return s.ballotLeader == s1.ballotLeader && s.ballotNumber == s1.ballotNumber && AcceptedMatch(s.accepted, s1.accepted)
}

func (s *State) LeaderEqual(s1 *State) bool {
	if s.me != s1.me {
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

// --- HELPER FUNCTIONS ---

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
