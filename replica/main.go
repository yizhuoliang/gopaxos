package main

import (
	"context"
	pb "gopaxos/gopaxos"
	"log"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	leaderNum = 2
	WINDOW    = 5
)

var (
	replicaId int64

	state       string
	slot_in     int32 = 1
	slot_out    int32 = 1
	requests    []*pb.Command
	proposals   []*pb.Proposal
	decisions   []*pb.Decision
	leaderPorts = []string{"127.0.0.1:50055", "127.0.0.1:50056"}

	responses []*pb.Response
)

type replicaServer struct {
	pb.UnimplementedClientReplicaServer
}

func main() {
	replicaId, _ = strconv.ParseInt(os.Args[1], 10, 32)

	for i := 0; i < leaderNum; i++ {
		go MessengerRoutine(i)
		go CollectorRoutine(i)
	}
}

func MessengerRoutine(serial int) {
	conn, err := grpc.Dial(leaderPorts[serial], grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect: %v", err)
		return
	}
	defer conn.Close()

	c := pb.NewReplicaLeaderClient(conn)

	// reference sudo code
	for slot_in < slot_out+WINDOW && len(requests) != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := c.Propose(ctx, &pb.Proposal{SlotNumber: slot_in, Command: nil})
	}
}

func CollectorRoutine(serial int) {

}

func (s *replicaServer) Request(ctx context.Context, in *pb.Command) (*pb.Empty, error) {
	requests = append(requests, in)
	return &pb.Empty{Content: "success"}, nil
}

func (s *replicaServer) Collect(ctx context.Context, in *pb.Empty) (*pb.Responses, error) {
	return &pb.Responses{Valid: true, Responses: responses}, nil
}
