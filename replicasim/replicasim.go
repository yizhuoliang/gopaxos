package main

import (
	pb "gopaxos/gopaxos"
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
	state     string
	slot_in   int32
	slot_out  int32
	requests  []*pb.Command
	proposals []*pb.Proposal
	decisions []*pb.Decision
}

type PartialState struct {
	slot_in     int32
	slot_out    int32
	request     *pb.Command
	newProposal *pb.Proposal
	decisions   []*pb.Decision
}

func Apply(s State, msg pb.Message) (State, pb.Message) {
	reply := pb.Message{}
	switch msg.Type {
	case COMMAND:
		reply = s.ReqTransformation(msg)
	case DECISIONS:
		s.DecTransformation(msg)
	}
	return s, reply
}

func (s *State) ReqTransformation(msg pb.Message) pb.Message {
	c := &pb.Command{ClientId: msg.ClientId, CommandId: msg.CommandId, Operation: msg.Operation}
	r := make([]*pb.Command, len(s.requests))
	copy(r, s.requests)
	s.requests = r
	s.requests = append(s.requests, c)
	if s.decisions[s.slot_in] == nil {
		s.proposals[s.slot_in] = &pb.Proposal{SlotNumber: s.slot_in, Command: c}
	}
	return pb.Message{Type: EMPTY, Content: "success"}
}

func (s *State) DecTransformation(msg pb.Message) {
	if msg.Valid {
		d := make([]*pb.Decision, len(msg.Decisions))
		copy(d, s.decisions)
		s.decisions = d
		for _, dec := range msg.Decisions {
			s.decisions[dec.SlotNumber] = dec
			if s.decisions[s.slot_out] != nil {
				if s.proposals[s.slot_out] != nil {
					p := s.proposals[s.slot_out]
					s.proposals[s.slot_out] = nil
					if s.decisions[s.slot_out].Command.CommandId != p.Command.CommandId {
						if s.decisions[s.slot_in] == nil {
							s.proposals[s.slot_in] = &pb.Proposal{SlotNumber: s.slot_in, Command: p.Command}
						}
					}
				}
				s.state = s.state + dec.Command.Operation
				s.slot_out++
			}
		}
	}
}

func Inference(msg pb.Message) (interface{}, pb.Message) {
	switch msg.Type {
	case RESPONSES:
		return RespInference(msg), pb.Message{}
	case PROPOSAL:
		return PropInference(msg), pb.Message{}
	default:
		return nil, pb.Message{}
	}
}

func RespInference(msg pb.Message) PartialState {
	if msg.Valid {
		decs := make([]*pb.Decision, len(msg.Responses))
		for i := 0; i < len(msg.Responses); i++ {
			decs[i] = &pb.Decision{SlotNumber: int32(i), Command: msg.Responses[i].Command}
		}
		// slot number up to len(msg.Responses) - 1 decided, save to assume applied
		return PartialState{slot_in: -1, slot_out: int32(len(msg.Responses)), request: nil, newProposal: nil, decisions: decs}
	}
	return PartialState{slot_in: -1, slot_out: -1, request: nil, newProposal: nil, decisions: nil}
}

func PropInference(msg pb.Message) PartialState {
	r := &pb.Command{ClientId: msg.Command.ClientId, CommandId: msg.Command.CommandId, Operation: msg.Command.Operation}
	p := &pb.Proposal{SlotNumber: msg.SlotNumber, Command: msg.Command}
	return PartialState{slot_in: msg.SlotNumber, slot_out: -1, request: r, newProposal: p, decisions: nil}
}

func PartialStateEnabled(ps PartialState, from State, msg pb.Message) (bool, bool, State) {
	return false, false, State{}
}

func PartialStateMatched(ps PartialState, s State) bool {
	return false
}

func (s *State) Equal(s1 *State) bool {
	return false
}

func Drop(msg pb.Message, s State) bool {
	return false
}
