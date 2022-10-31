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
	acceptorNum = 3
)

var (
	server   *grpc.Server
	leaderId int32 // also considered as the serial of this acceptor

	leaderPorts   = []string{"127.0.0.1:50055", "127.0.0.1:50056"}
	acceptorPorts = []string{"127.0.0.1:50057", "127.0.0.1:50058", "127.0.0.1:50059"}

	// leader states
	ballotNumber int32
	active       bool = false
	proposals    map[int32]*pb.Proposal

	ballotNumberUpdateChannel chan int32
	proposalsUpdateChannel    chan *pb.Proposal

	decisions []*pb.Decision

	feedbackChannel chan int32
)

type leaderServer struct {
	pb.UnimplementedReplicaLeaderServer
}

func main() {
	temp, _ := strconv.Atoi(os.Args[1])
	leaderId = int32(temp)

	ballotNumberUpdateChannel = make(chan int32)
	go ballotNumberUpdateRoutine()

	feedbackChannel = make(chan int32)
	// TODO: implement feedback, need a struct for Preemption and adoption
	// TODO: update decisions[] during adoption for sending feedback to replicas

	serve(leaderPorts[leaderId])

	// TODO: more leader routine
}

func serve(port string) {
	// listen client on port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("server listening at %v", lis.Addr())
	server = grpc.NewServer()
	pb.RegisterReplicaLeaderServer(server, &leaderServer{})
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// UPDATER ROUTINES
func ballotNumberUpdateRoutine() {
	for {
		newNum := <-ballotNumberUpdateChannel

		// CONCERN: need more consideration on the concurrency of ballotNumber
		if newNum > ballotNumber {
			ballotNumber = newNum
		}
	}
}

// TODO: expand this
func proposalsUpdateRoutine() {
	for {
		newProposal := <-proposalsUpdateChannel
		if _, ok := proposals[newProposal.SlotNumber]; !ok {
			proposals[newProposal.SlotNumber] = newProposal
			go CommanderRoutine(&pb.BSC{
				BallotNumber: ballotNumber, SlotNumber: newProposal.SlotNumber, Command: newProposal.Command})
		}
	}
}

// SUB-ROUTINES
func ScoutRoutine() {

}

func CommanderRoutine(bsc *pb.BSC) {
	commanderCollectChannel := make(chan *pb.P2B)
	// send messages
	for i := 0; i < acceptorNum; i++ {
		go CommanderMessenger(i, bsc, commanderCollectChannel)
	}

	// collect messages
	replyCount := 0
	for {
		r := <-commanderCollectChannel
		if r.BallotNumber == bsc.BallotNumber {
			// waitfor:=waitfor-{α};
			replyCount++
			if replyCount > acceptorNum/2 {
				decisions = append(decisions, &pb.Decision{SlotNumber: bsc.SlotNumber, Command: bsc.Command})
			}
		}
	}
}

func CommanderMessenger(serial int, bsc *pb.BSC, commanderCollectChannel chan (*pb.P2B)) {
	conn, err := grpc.Dial(acceptorPorts[serial], grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect: %v", err)
		return
	}
	defer conn.Close()

	c := pb.NewLeaderAcceptorClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p2b, err := c.Commanding(ctx, &pb.P2A{LeaderId: leaderId, Bsc: bsc})
	if err != nil {
		commanderCollectChannel <- &pb.P2B{AcceptorId: -1, BallotNumber: -1}
		return
	}
	commanderCollectChannel <- p2b
}

// gRPC HANDLERS
func (s *leaderServer) Propose(ctx context.Context, in *pb.Proposal) (*pb.Empty, error) {
	proposalsUpdateChannel <- in // Weird? Yes! Things can only be down after chekcing proposals
	return &pb.Empty{Content: "success"}, nil
}

func (s *leaderServer) Collect(ctx context.Context, in *pb.Empty) (*pb.Decisions, error) {
	if len(decisions) == 0 {
		return &pb.Decisions{Valid: false, Decisions: nil}, nil
	} else {
		r := decisions
		decisions = decisions[:0]
		return &pb.Decisions{Valid: true, Decisions: r}, nil
	}
}
