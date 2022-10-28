package main

import (
	pb "gopaxos/gopaxos"

	"google.golang.org/grpc"
)

const (
	replicaNum = 3
	WINDOW     = 5
)

var (
	server       *grpc.Server
	replicaId    int32
	replicaPorts = []string{"127.0.0.1:50053", "127.0.0.1:50054"}

	state       string
	slot_in     int32 = 1
	slot_out    int32 = 1
	requests    [leaderNum]chan *pb.Command
	proposals   [leaderNum]chan *pb.Proposal
	decisions   chan *pb.Decision
	leaderPorts = []string{"127.0.0.1:50055", "127.0.0.1:50056"}

	responses []*pb.Response
)
