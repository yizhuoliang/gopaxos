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
	proposalsUpdateChannel    chan *proposalsUpdateRequest

	decisions []*pb.Decision
)

type leaderServer struct {
	pb.UnimplementedReplicaLeaderServer
}

// REQUEST TYPES:
// 1 - new proposal, asked by handlers
// 2 - adoption
// 3 - preemption
type proposalsUpdateRequest struct {
	updateType             int
	newProposal            *pb.Proposal
	pvalues                []*pb.BSC
	adoptionBallowNumber   int32
	preemptionBallotNumber int32
}

func main() {
	temp, _ := strconv.Atoi(os.Args[1])
	leaderId = int32(temp)

	ballotNumberUpdateChannel = make(chan int32)
	go ballotNumberUpdateRoutine()

	// spawn the initial Scout
	go ScoutRoutine(ballotNumber)

	serve(leaderPorts[leaderId])
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
		update := <-proposalsUpdateChannel
		if update.updateType == 1 {
			// gRPC handler add new proposals
			newProposal := update.newProposal
			if _, ok := proposals[newProposal.SlotNumber]; !ok {
				proposals[newProposal.SlotNumber] = newProposal
				go CommanderRoutine(&pb.BSC{
					BallotNumber: ballotNumber, SlotNumber: newProposal.SlotNumber, Command: newProposal.Command})
			}
		} else if update.updateType == 2 {
			// ADOPTION
			pvalues := update.pvalues
			var slotToBallot map[int32]int32 // map from slot number to ballot number to satisfy pmax
			for _, bsc := range pvalues {
				proposal, okProp := proposals[bsc.BallotNumber]
				if okProp {
					originalBallot, okBall := slotToBallot[proposal.SlotNumber]
					if (okBall && originalBallot < bsc.BallotNumber) || !okBall {
						proposals[bsc.SlotNumber] = &pb.Proposal{SlotNumber: bsc.SlotNumber, Command: bsc.Command}
						slotToBallot[proposal.SlotNumber] = bsc.BallotNumber
					}
				} else {
					proposals[bsc.SlotNumber] = &pb.Proposal{SlotNumber: bsc.SlotNumber, Command: bsc.Command}
					slotToBallot[proposal.SlotNumber] = bsc.BallotNumber
				}
			}
			// send proposals
			for _, proposal := range proposals {
				go CommanderRoutine(&pb.BSC{BallotNumber: ballotNumber, SlotNumber: proposal.SlotNumber, Command: proposal.Command})
			}
			active = true
		}
	}
}

// SUB-ROUTINES
func ScoutRoutine(scoutBallotNumber int32) {
	scoutCollectChannel := make(chan *pb.P1B)
	// send messages
	for i := 0; i < acceptorNum; i++ {
		go ScoutMessenger(i, scoutCollectChannel, scoutBallotNumber)
	}

	// collect messages
	acceptCount := 0
	var pvalues []*pb.BSC
	for i := 0; i < acceptorNum; i++ {
		p1b := <-scoutCollectChannel
		if p1b.AcceptorId >= 0 && p1b.BallotNumber >= 0 {
			if p1b.BallotNumber != scoutBallotNumber {
				// do preemption

			}
			acceptCount++
			pvalues = append(pvalues, p1b.Accepted...)
			if acceptCount > acceptorNum/2 {
				proposalsUpdateChannel <- &proposalsUpdateRequest{updateType: 2, pvalues: pvalues, adoptionBallowNumber: scoutBallotNumber}
				return
			}
		}
	}
}

func ScoutMessenger(serial int, scoutCollectChannel chan *pb.P1B, scoutBallotNumber int32) {
	conn, err := grpc.Dial(acceptorPorts[serial], grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect: %v", err)
		return
	}
	defer conn.Close()

	c := pb.NewLeaderAcceptorClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p1b, err := c.Scouting(ctx, &pb.P1A{LeaderId: leaderId, BallotNumber: scoutBallotNumber})
	if err != nil {
		scoutCollectChannel <- &pb.P1B{AcceptorId: -1, BallotNumber: -1, Accepted: nil}
		return
	}
	scoutCollectChannel <- p1b
}

func CommanderRoutine(bsc *pb.BSC) {
	commanderCollectChannel := make(chan *pb.P2B)
	// send messages
	for i := 0; i < acceptorNum; i++ {
		go CommanderMessenger(i, bsc, commanderCollectChannel)
	}

	// collect messages
	replyCount := 0
	for i := 0; i < acceptorNum; i++ {
		r := <-commanderCollectChannel
		if r.BallotNumber == bsc.BallotNumber {
			// waitfor:=waitfor-{α};
			replyCount++
			if replyCount > acceptorNum/2 {
				decisions = append(decisions, &pb.Decision{SlotNumber: bsc.SlotNumber, Command: bsc.Command})
				return
			}
		} else if r.BallotNumber > 0 && r.AcceptorId > 0 {
			// PREEMPTION

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
	proposalsUpdateChannel <- &proposalsUpdateRequest{updateType: 1, newProposal: in} // Weird? Yes! Things can only be down after chekcing proposals
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
