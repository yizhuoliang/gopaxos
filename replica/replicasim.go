package main

import (
	pb "gopaxos/gopaxos"
)

type State struct {
	slot_in   int32
	slot_out  int32
	requests  []*pb.Command
	proposals []*pb.Proposal
	decisions []*pb.Decision
}

type PartialState struct {
	decisions []*pb.Decision
}

func Apply(s State, msg pb.Message) (State, pb.Message) {
	reply := pb.Message{}
	switch msg.Type {
	case COMMAND:
		s.ReqTransformation(msg)
	case DECISIONS:
		s.DecTransformation(msg)
	}
	return s, reply
}

func (s *State) ReqTransformation(msg pb.Message) {

}

func (s *State) DecTransformation(msg pb.Message) {

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
		for i := 1; i < len(msg.Responses); i++ {
			decs[i] = &pb.Decision{SlotNumber: int32(i), Command: msg.Responses[i].Command}
		}
		return PartialState{decisions: decs}
	}
	return PartialState{decisions: nil}
}

func PropInference(msg pb.Message) PartialState {
	return PartialState{decisions: nil}
}
