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

	// reference sudo code propose()
	for slot_in < slot_out+WINDOW {
		request := <-requests[serial]
		proposals[serial] <- &pb.Proposal{SlotNumber: slot_in, Command: request}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		c.Propose(ctx, &pb.Proposal{SlotNumber: slot_in, Command: request})
		slot_in++
	}
}
func CollectorRoutine(serial int) {
	conn, err := grpc.Dial(leaderPorts[serial], grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect: %v", err)
		return
	}
	defer conn.Close()

	c := pb.NewReplicaLeaderClient(conn)

	// reference sudo code perform()
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := c.Collect(ctx, &pb.Empty{Content: "checking responses"})
		if err != nil {
			log.Printf("failed to collect: %v", err)
		}
		if r.Valid {
			for _, decision := range r.Decisions {
				decisions <- decision
				// sudo code
				for _, proposal := range proposals {
					if decision.SlotNumber == proposal.SlotNumber {
						proposals.remove(proposal)
						if decision.Command != proposal.Command {
							requests[serial] <- proposal
						}
					}
				}
				// perform(decision) atomically
				state = state + decision.Command.Operation
				slot_out++
				log.Printf("Operation %s is performed", decision.Command.Operation)
			}
		}
	}
}

// handlers
func (s *replicaServer) Request(ctx context.Context, in *pb.Command) (*pb.Empty, error) {
	for i := 0; i < leaderNum; i++ {
		requests[i] <- in
	}
	return &pb.Empty{Content: "success"}, nil
}

func (s *replicaServer) Collect(ctx context.Context, in *pb.Empty) (*pb.Responses, error) {
	return &pb.Responses{Valid: true, Responses: responses}, nil
}
