package main

import (
	"context"
	pb "gopaxos/gopaxos"
	"log"
	"net"
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
	replicaId int32
	server    *grpc.Server

	state        string
	slot_in      int32 = 1
	slot_out     int32 = 1
	requests     []*pb.Command
	proposals    []*pb.Proposal
	decisions    []*pb.Decision
	leaderPorts  = []string{"127.0.0.1:50055", "127.0.0.1:50056"}
	replicaPorts = []string{"127.0.0.1:50053", "127.0.0.1:50054"}

	responses []*pb.Response
)

type replicaServer struct {
	pb.UnimplementedClientReplicaServer
}

func main() {
	temp, _ := strconv.Atoi(os.Args[1])
	replicaId = int32(temp)

	for i := 0; i < leaderNum; i++ {
		go MessengerRoutine(i)
		go CollectorRoutine(i)
	}

	go serve(replicaPorts[replicaId])

	preventExit := make(chan int32, 1)
	<-preventExit
}

func serve(port string) {
	// listen leader on port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("server listening at %v", lis.Addr())
	server = grpc.NewServer()
	pb.RegisterClientReplicaServer(server, &replicaServer{})
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// routines
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
		c.Propose(ctx, &pb.Proposal{SlotNumber: slot_in, Command: nil})
	}
}
func CollectorRoutine(serial int) {

}

// handlers
func (s *replicaServer) Request(ctx context.Context, in *pb.Command) (*pb.Empty, error) {
	requests = append(requests, in)
	return &pb.Empty{Content: "success"}, nil
}
func (s *replicaServer) Collect(ctx context.Context, in *pb.Empty) (*pb.Responses, error) {
	return &pb.Responses{Valid: true, Responses: responses}, nil
}
