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

	state     string = ""
	slot_in   int32  = 1
	slot_out  int32  = 1
	requests  [leaderNum]chan *pb.Command
	proposals map[int32]*pb.Proposal
	decisions map[int32]*pb.Decision

	leaderPorts = []string{"127.0.0.1:50055", "127.0.0.1:50056"}

	responses []*pb.Response

	replicaStateUpdateChannel chan *replicaStateUpdateRequest
)

type replicaServer struct {
	pb.UnimplementedClientReplicaServer
}

type replicaStateUpdateRequest struct {
	updateType   int
	newDecisions []*pb.Decision
	serial       int
	c            pb.ReplicaLeaderClient
}

func main() {
	temp, _ := strconv.Atoi(os.Args[1])
	replicaId = int32(temp)

	replicaStateUpdateChannel = make(chan *replicaStateUpdateRequest)

	for i := 0; i < leaderNum; i++ {
		go MessengerRoutine(i)
		go CollectorRoutine(i)
	}

	serve(replicaPorts[replicaId])
}

func serve(port string) {
	// listen client on port
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
func ReplicaStateUpdateRoutine() {
	for {
		update := <-replicaStateUpdateChannel
		if update.updateType == 1 {
			// reference sudo code perform()
			for _, decision := range update.newDecisions {
				decisions[decision.SlotNumber] = decision
				for _, proposal := range proposals {
					if decision.SlotNumber == proposal.SlotNumber {
						delete(proposals, proposal.SlotNumber)
						if decision.Command != proposal.Command {
							requests[update.serial] <- proposal.Command
						}
					}
				}
				// atomic
				state = state + decision.Command.Operation
				slot_out++
				// end atomic
				log.Printf("Operation %s is performed", decision.Command.Operation)
				responses = append(responses, &pb.Response{CommandId: decision.Command.CommandId})
			}
		} else if update.updateType == 2 {
			// reference sudo code propose()
			for slot_in < slot_out+WINDOW {
				_, ok := decisions[slot_in]
				if !ok {
					request := <-requests[update.serial]
					proposals[slot_in] = &pb.Proposal{SlotNumber: slot_in, Command: request}
					ctx, cancel := context.WithTimeout(context.Background(), time.Second)
					_, err := update.c.Propose(ctx, &pb.Proposal{SlotNumber: slot_in, Command: request})
					if err != nil {
						log.Printf("failed to propose: %v", err)
						cancel()
					}
				}
				slot_in++
			}
		}
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
	for {
		replicaStateUpdateChannel <- &replicaStateUpdateRequest{updateType: 2, newDecisions: nil, serial: serial, c: c}
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

	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		r, err := c.Collect(ctx, &pb.Empty{Content: "checking responses"})
		if err != nil {
			log.Printf("failed to collect: %v", err)
			cancel()
		}
		if r.Valid {
			replicaStateUpdateChannel <- &replicaStateUpdateRequest{updateType: 1, newDecisions: r.Decisions, serial: serial}
		}
	}
}

// handlers
func (s *replicaServer) Request(ctx context.Context, in *pb.Command) (*pb.Empty, error) {
	for i := 0; i < leaderNum; i++ {
		requests[i] <- in
	}
	log.Printf("Request with command id %s received", in.CommandId)
	return &pb.Empty{Content: "success"}, nil
}

func (s *replicaServer) Collect(ctx context.Context, in *pb.Empty) (*pb.Responses, error) {
	return &pb.Responses{Valid: true, Responses: responses}, nil
}
