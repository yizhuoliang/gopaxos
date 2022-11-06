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
	ballotNumber int32 = 0
	active       bool  = false
	proposals    map[int32]*pb.Proposal

	leaderStateUpdateChannel chan *leaderStateUpdateRequest

	decisions []*pb.Decision
)

type leaderServer struct {
	pb.UnimplementedReplicaLeaderServer
}

// REQUEST TYPES:
// 1 - new proposal, asked by handlers
// 2 - adoption
// 3 - preemption
type leaderStateUpdateRequest struct {
	updateType             int
	newProposal            *pb.Proposal
	pvalues                []*pb.BSC
	adoptionBallowNumber   int32
	preemptionBallotNumber int32
}

func main() {
	temp, _ := strconv.Atoi(os.Args[1])
	leaderId = int32(temp)

	// initialization
	proposals = make(map[int32]*pb.Proposal)
	leaderStateUpdateChannel = make(chan *leaderStateUpdateRequest, 1)

	go leaderStateUpdateRoutine()

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

// TODO: expand this
func leaderStateUpdateRoutine() {
	for {
		update := <-leaderStateUpdateChannel
		if update.updateType == 1 {
			// gRPC handler add new proposals
			newProposal := update.newProposal
			if _, ok := proposals[newProposal.SlotNumber]; !ok {
				proposals[newProposal.SlotNumber] = newProposal
				if active {
					go CommanderRoutine(&pb.BSC{
						BallotNumber: ballotNumber, SlotNumber: newProposal.SlotNumber, Command: newProposal.Command})
				}
			}
		} else if update.updateType == 2 {
			// ADOPTION
			pvalues := update.pvalues
			slotToBallot := make(map[int32]int32) // map from slot number to ballot number to satisfy pmax
			for _, bsc := range pvalues {
				proposal, okProp := proposals[bsc.SlotNumber]
				if okProp {
					originalBallot, okBall := slotToBallot[proposal.SlotNumber]
					if (okBall && originalBallot < bsc.BallotNumber) || !okBall {
						// the previous proposal to that slot has a lower ballot number, or this is the first proposal to that slot
						proposals[bsc.SlotNumber] = &pb.Proposal{SlotNumber: bsc.SlotNumber, Command: bsc.Command}
						slotToBallot[proposal.SlotNumber] = bsc.BallotNumber
					}
				} else {
					// there was originally no proposal for that slot
					proposals[bsc.SlotNumber] = &pb.Proposal{SlotNumber: bsc.SlotNumber, Command: bsc.Command}
					slotToBallot[proposal.SlotNumber] = bsc.BallotNumber
				}
			}
			// send proposals
			for _, proposal := range proposals {
				go CommanderRoutine(&pb.BSC{BallotNumber: ballotNumber, SlotNumber: proposal.SlotNumber, Command: proposal.Command})
			}
			active = true
		} else if update.updateType == 3 {
			// PREEMPTION
			if update.preemptionBallotNumber > ballotNumber {
				active = false
				ballotNumber = update.preemptionBallotNumber + 1
				go ScoutRoutine(ballotNumber)
			}
		}
	}
}

// SUB-ROUTINES
func ScoutRoutine(scoutBallotNumber int32) {
	log.Printf("Scout spawned with ballot numebr %d", scoutBallotNumber)
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
		if p1b.AcceptorId >= 0 {
			if p1b.BallotNumber != scoutBallotNumber || p1b.BallotLeader != leaderId {
				// do preemption
				leaderStateUpdateChannel <- &leaderStateUpdateRequest{updateType: 3, preemptionBallotNumber: p1b.BallotNumber}
				log.Printf("Scout send preemption")
				return
			}
			acceptCount++
			pvalues = append(pvalues, p1b.Accepted...)
			if acceptCount > acceptorNum/2 {
				leaderStateUpdateChannel <- &leaderStateUpdateRequest{updateType: 2, pvalues: pvalues, adoptionBallowNumber: scoutBallotNumber}
				log.Printf("Scout send adoption")
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
		scoutCollectChannel <- &pb.P1B{AcceptorId: -1}
		return
	}
	scoutCollectChannel <- p1b
}

func CommanderRoutine(bsc *pb.BSC) {
	log.Printf("Commander spawned for ballot number %d, slot number %d, command id %s", bsc.BallotNumber, bsc.SlotNumber, bsc.Command.CommandId)
	commanderCollectChannel := make(chan *pb.P2B)
	// send messages
	for i := 0; i < acceptorNum; i++ {
		go CommanderMessenger(i, bsc, commanderCollectChannel)
	}

	// collect messages
	acceptCount := 0
	for i := 0; i < acceptorNum; i++ {
		p2b := <-commanderCollectChannel
		if p2b.BallotNumber == bsc.BallotNumber && p2b.BallotLeader == leaderId {
			// waitfor:=waitfor-{α};
			acceptCount++
			if acceptCount > acceptorNum/2 {
				decisions = append(decisions, &pb.Decision{SlotNumber: bsc.SlotNumber, Command: bsc.Command})
				log.Printf("The bsc is decided, commander exit")
				return
			}
		} else if p2b.AcceptorId >= 0 {
			// PREEMPTION
			leaderStateUpdateChannel <- &leaderStateUpdateRequest{updateType: 3, preemptionBallotNumber: p2b.BallotNumber}
			log.Printf("Commander send preemption")
			return
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
		commanderCollectChannel <- &pb.P2B{AcceptorId: -1}
		return
	}
	commanderCollectChannel <- p2b
}

// gRPC HANDLERS
func (s *leaderServer) Propose(ctx context.Context, in *pb.Proposal) (*pb.Empty, error) {
	leaderStateUpdateChannel <- &leaderStateUpdateRequest{updateType: 1, newProposal: in} // Weird? Yes! Things can only be down after chekcing proposals
	log.Printf("Received proposal with commandId %s and slot number %d", in.Command.CommandId, in.SlotNumber)
	return &pb.Empty{Content: "success"}, nil
}

func (s *leaderServer) Collect(ctx context.Context, in *pb.Empty) (*pb.Decisions, error) {
	return &pb.Decisions{Valid: true, Decisions: decisions}, nil
}
