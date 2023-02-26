package main

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	pb "github.com/yizhuoliang/gopaxos"
	"github.com/yizhuoliang/gopaxos/comm"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const (
	acceptorNum = 3
	leaderNum   = 2

	// Message types
	COMMAND   = 1
	READ      = 2
	RESPONSES = 3
	PROPOSAL  = 4
	DECISIONS = 5
	BEAT      = 6
	P1A       = 7
	P1B       = 8
	P2A       = 9
	P2B       = 10
	EMPTY     = 11
)

var (
	server   *grpc.Server
	leaderId int32 // also considered as the serial of this acceptor

	// for no-sim tests
	leaderPorts   = []string{"127.0.0.1:50055", "127.0.0.1:50056"}
	acceptorPorts = []string{"127.0.0.1:50057", "127.0.0.1:50058", "127.0.0.1:50059"}

	heartbeatClients [leaderNum]*pb.ReplicaLeaderClient

	// leader states
	ballotNumber int32 = 0
	active       bool  = false
	proposals    map[int32]*pb.Proposal

	leaderStateUpdateChannel chan *leaderStateUpdateRequest

	decisions []*pb.Decision

	simc  *comm.RPCConnection
	simon int // 1 = on, 0 = off
)

type leaderServer struct {
	pb.UnimplementedReplicaLeaderServer
}

// REQUEST TYPES:
// 1 - new proposal, asked by handlers
// 2 - adoption
// 3 - preemption
// 4 - returned preemption
type leaderStateUpdateRequest struct {
	updateType             int
	newProposal            *pb.Proposal
	pvalues                []*pb.BSC
	adoptionBallowNumber   int32
	preemptionBallotNumber int32
	preemptionLeader       int32
}

func main() {
	temp, _ := strconv.Atoi(os.Args[1])
	leaderId = int32(temp)
	simon, _ = strconv.Atoi(os.Args[2])

	// connect sim
	if simon == 1 {
		simc = new(comm.RPCConnection).Init(uint64(leaderId+3), 2)
		leaderPorts = []string{"172.17.0.5:50050", "172.17.0.6:50050"}
		acceptorPorts = []string{"172.17.0.2:50050", "172.17.0.3:50050", "172.17.0.4:50050"}
	}

	// initialization
	proposals = make(map[int32]*pb.Proposal)
	leaderStateUpdateChannel = make(chan *leaderStateUpdateRequest, 1)

	go leaderStateUpdateRoutine()

	go serve(leaderPorts[leaderId])
	setupHeartbeat()

	// spawn the initial Scout
	go ScoutRoutine(ballotNumber, false)

	preventExit := make(chan int)
	<-preventExit
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

func setupHeartbeat() {
	log.Printf("waiting for other leaders...")
	readyCount := 0
	for {
		for i, port := range leaderPorts {
			if i != int(leaderId) && heartbeatClients[i] == nil {
				conn, err := grpc.Dial(port, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					continue
				}

				c := pb.NewReplicaLeaderClient(conn)
				heartbeatClients[i] = &c
				readyCount++
				if readyCount == leaderNum-1 {
					log.Printf("heartbeat clients ready")
					return
				}
			}
		}
	}
}

// CORE FUNCTIONS
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
			if update.preemptionBallotNumber > ballotNumber || update.preemptionLeader != leaderId {
				active = false
				ballotNumber = update.preemptionBallotNumber + 1
				go ScoutRoutine(ballotNumber, false)
			}
		} else if update.updateType == 4 {
			// RETURNED PREEMPTION
			go ScoutRoutine(ballotNumber, true)
		}
	}
}

// SUB-ROUTINES
func ScoutRoutine(scoutBallotNumber int32, returned bool) {

	if !returned {
		log.Printf("Scout spawned with ballot numebr %d", scoutBallotNumber)
	}

	if returned {
		time.Sleep(time.Second) // avoid heartbeat too frequent
	}

	beatCollectChannle := make(chan bool)

	// testing if an active leader is already working
	for i, heartbeatClient := range heartbeatClients {
		if heartbeatClient != nil {
			go BeatMessenger(i, beatCollectChannle)
		}
	}

	// collecting beat results
	hasActive := false
	for i := 0; i < leaderNum-1; i++ {
		if <-beatCollectChannle {
			hasActive = true
		}
	}

	// return this preemption if an active leader already exists
	if hasActive {
		// return preemption
		leaderStateUpdateChannel <- &leaderStateUpdateRequest{updateType: 4}
		return
	}

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
				// do adoption
				leaderStateUpdateChannel <- &leaderStateUpdateRequest{updateType: 2, pvalues: pvalues, adoptionBallowNumber: scoutBallotNumber}
				log.Printf("Scout send adoption")
				return
			}
		}
	}
}

func BeatMessenger(serial int, beatCollectChannel chan bool) {
	heartbeatClient := heartbeatClients[serial]
	if heartbeatClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, err := (*heartbeatClient).Heartbeat(ctx, &pb.Message{Type: EMPTY, Content: "checking heartbeat"})
		if err == nil {
			beatCollectChannel <- r.GetActive()
			return
		}
	}
	beatCollectChannel <- false
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

	// P1A sent
	if simon == 1 {
		m := pb.Message{Type: P1A, LeaderId: leaderId, BallotNumber: scoutBallotNumber, Send: true}
		tosend, offset := simc.AllocateRequest((uint64)(proto.Size(&m)))
		b, err := proto.Marshal(&m)
		if err != nil {
			log.Fatalf("marshal err:%v\n", err)
		}
		copy(tosend[offset:], b)
		_, err = simc.OutConn.Write(tosend)
		if err != nil {
			log.Fatalf("Write to simulator failed, err:%v\n", err)
		}
	}

	r, err := c.Scouting(ctx, &pb.Message{Type: P1A, LeaderId: leaderId, BallotNumber: scoutBallotNumber})
	if err != nil {
		log.Printf("scouting failed: %v", err)
		scoutCollectChannel <- &pb.P1B{AcceptorId: -1}
		return
	}

	// P1B received
	if simon == 1 {
		tosend, offset := simc.AllocateRequest((uint64)(proto.Size(r)))
		b, err := proto.Marshal(r)
		if err != nil {
			log.Fatalf("marshal err:%v\n", err)
		}
		copy(tosend[offset:], b)
		_, err = simc.OutConn.Write(tosend)
		if err != nil {
			log.Fatalf("Write to simulator failed, err:%v\n", err)
		}
	}

	scoutCollectChannel <- &pb.P1B{AcceptorId: r.AcceptorId, BallotNumber: r.BallotNumber, BallotLeader: r.BallotLeader, Accepted: r.Accepted}
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
			log.Printf("Commander send preemption, commander exit")
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

	// P2A sent
	if simon == 1 {
		m := pb.Message{Type: P2A, LeaderId: leaderId, Bsc: bsc, Send: true}
		tosend, offset := simc.AllocateRequest((uint64)(proto.Size(&m)))
		b, err := proto.Marshal(&m)
		if err != nil {
			log.Fatalf("marshal err:%v\n", err)
		}
		copy(tosend[offset:], b)
		_, err = simc.OutConn.Write(tosend)
		if err != nil {
			log.Fatalf("Write to simulator failed, err:%v\n", err)
		}
	}

	r, err := c.Commanding(ctx, &pb.Message{Type: P2A, LeaderId: leaderId, Bsc: bsc, Send: false})
	if err != nil {
		commanderCollectChannel <- &pb.P2B{AcceptorId: -1}
		return
	}

	// P2B received
	if simon == 1 {
		tosend, offset := simc.AllocateRequest((uint64)(proto.Size(r)))
		b, err := proto.Marshal(r)
		if err != nil {
			log.Fatalf("marshal err:%v\n", err)
		}
		copy(tosend[offset:], b)
		_, err = simc.OutConn.Write(tosend)
		if err != nil {
			log.Fatalf("Write to simulator failed, err:%v\n", err)
		}
	}

	commanderCollectChannel <- &pb.P2B{AcceptorId: r.AcceptorId, BallotNumber: r.BallotNumber, BallotLeader: r.BallotLeader}
}

// gRPC HANDLERS
func (s *leaderServer) Propose(ctx context.Context, in *pb.Message) (*pb.Message, error) {

	// Proposal received
	if simon == 1 {
		tosend, offset := simc.AllocateRequest((uint64)(proto.Size(in)))
		b, err := proto.Marshal(in)
		if err != nil {
			log.Fatalf("marshal err:%v\n", err)
		}
		copy(tosend[offset:], b)
		_, err = simc.OutConn.Write(tosend)
		if err != nil {
			log.Fatalf("Write to simulator failed, err:%v\n", err)
		}
	}

	leaderStateUpdateChannel <- &leaderStateUpdateRequest{updateType: 1, newProposal: &pb.Proposal{SlotNumber: in.SlotNumber, Command: in.Command}} // Weird? Yes! Things can only be down after chekcing proposals
	log.Printf("Received proposal with commandId %s and slot number %d", in.Command.CommandId, in.SlotNumber)
	return &pb.Message{Type: EMPTY, Content: "success"}, nil
}

func (s *leaderServer) Collect(ctx context.Context, in *pb.Message) (*pb.Message, error) {

	// Collection received
	if simon == 3 {
		tosend, offset := simc.AllocateRequest((uint64)(proto.Size(in)))
		b, err := proto.Marshal(in)
		if err != nil {
			log.Fatalf("marshal err:%v\n", err)
		}
		copy(tosend[offset:], b)
		// debug
		log.Print("here")
		_, err = simc.OutConn.Write(tosend)
		if err != nil {
			log.Fatalf("Write to simulator failed, err:%v\n", err)
		}
		// debug
		log.Print("here")
	}

	// Decisions sent
	if simon == 1 {
		m := pb.Message{Type: DECISIONS, Decisions: decisions, Req: in, Send: true}
		tosend, offset := simc.AllocateRequest((uint64)(proto.Size(&m)))
		b, err := proto.Marshal(&m)
		if err != nil {
			log.Fatalf("marshal err:%v\n", err)
		}
		copy(tosend[offset:], b)
		_, err = simc.OutConn.Write(tosend)
		if err != nil {
			log.Fatalf("Write to simulator failed, err:%v\n", err)
		}
	}

	if !active {
		return nil, status.Error(codes.FailedPrecondition, "inactive leader")
	}

	return &pb.Message{Type: DECISIONS, Decisions: decisions}, nil
}

func (s *leaderServer) Heartbeat(ctx context.Context, in *pb.Message) (*pb.Message, error) {
	return &pb.Message{Type: BEAT, Active: active}, nil
}
