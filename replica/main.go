package main

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	pb "github.com/yizhuoliang/gopaxos"
	// "github.com/yizhuoliang/gopaxos/comm"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	// "google.golang.org/protobuf/proto"
)

const (
	WINDOW = 5

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

	// Roles
	LEADER   = 0
	ACCEPTOR = 1
)

var (
	server    *grpc.Server
	replicaId int32

	// for no-sim tests
	replicaPorts = []string{"127.0.0.1:50053", "127.0.0.1:50054"}
	leaderPorts  = []string{"127.0.0.1:50055", "127.0.0.1:50056"}

	slot_in   int32 = 0
	slot_out  int32 = 0
	requests  chan *pb.Command
	proposals map[int32]*pb.Proposal
	decisions map[int32]*pb.Decision

	replicaStateUpdateChannel chan *replicaStateUpdateRequest
	notificationChannel       [leaderNum]chan *pb.Message
	slotInUpdateChannel       [leaderNum]chan int

	// Yeah, this is the key-value map
	keyValueLog map[string]string

	mutexChannel chan int32

	// simc *comm.RPCConnection
	simon int // 1 = on, 0 = off
)

type replicaServer struct {
	pb.UnimplementedClientReplicaServer
}

type replicaStateUpdateRequest struct {
	updateType   int
	newDecisions []*pb.Decision
}

func main() {
	temp, _ := strconv.Atoi(os.Args[1])
	replicaId = int32(temp)
	simon, _ = strconv.Atoi(os.Args[2])

	// connect sim
	if simon == 1 {
		// simc = new(comm.RPCConnection).Init(uint64(leaderId+5), )
		replicaPorts = []string{"172.17.0.7:50050", "172.17.0.8:50050"}
		leaderPorts = []string{"172.17.0.5:50050", "172.17.0.6:50050"}
	}

	// initialization
	proposals = make(map[int32]*pb.Proposal)
	decisions = make(map[int32]*pb.Decision)
	keyValueLog = make(map[string]string)
	replicaStateUpdateChannel = make(chan *replicaStateUpdateRequest, 1)
	requests = make(chan *pb.Command, 1)
	for i := 0; i < leaderNum; i++ {
		notificationChannel[i] = make(chan *pb.Message, 1)
		slotInUpdateChannel[i] = make(chan int, 1)
	}
	mutexChannel = make(chan int32, 1)
	mutexChannel <- 1

	// go!
	go ReplicaStateUpdateRoutine()
	go SlotInUpdateRoutine()
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
		<-mutexChannel
		if update.updateType == 1 {
			// RECEIVE DECISION + PERFORM
			// log.Printf("messenger %d slot_out %d processing update type 1...", update.serial, slot_out)
			// reference sudo code receive() -> perform()
			for _, decision := range update.newDecisions {
				decisions[decision.SlotNumber] = decision
				d, ok := decisions[slot_out]
				// PERFORM - UPDATING THE KEY_VALUE MAP
				for ok {
					p, okProp := proposals[slot_out]
					if okProp {
						delete(proposals, slot_out)
						if d.Command.CommandId != p.Command.CommandId {
							requests <- p.Command
							replicaStateUpdateChannel <- &replicaStateUpdateRequest{updateType: 2, newDecisions: nil}
						}
					}
					// update log and update slot_out
					keyValueLog[d.Command.Key] = d.Command.Value
					slot_out++
					log.Printf("Log updated - key: %s, val: %s", d.Command.Key, d.Command.Value)
					d, ok = decisions[slot_out]
				}
			}
		} else if update.updateType == 2 {
			log.Printf("processing new request...")
			// reference sudo code propose()
			if slot_in < slot_out+WINDOW {
				_, ok := decisions[slot_in]
				if !ok {
					request := <-requests
					proposals[slot_in] = &pb.Proposal{SlotNumber: slot_in, Command: request}
					msgTosend := &pb.Message{Type: PROPOSAL, SlotNumber: slot_in, Command: request}
					for i := 0; i < leaderNum; i++ {
						notificationChannel[i] <- msgTosend
					}
				}
			}
		}
		mutexChannel <- 1
	}
}

func SlotInUpdateRoutine() {
	for {
		log.Print("updating slot_in...")
		for i := 0; i < leaderNum; i++ {
			<-slotInUpdateChannel[i]
		}
		slot_in++
		log.Print("slot_in updated successfully")
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
		msgTosend := <-notificationChannel[serial]
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, err := c.Propose(ctx, msgTosend)
		if err != nil {
			log.Printf("failed to propose: %v", err)
			cancel()
		}
		slotInUpdateChannel[serial] <- 1
	}
}

func CollectorRoutine(serial int) {
	logOutput := true
	conn, err := grpc.Dial(leaderPorts[serial], grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect: %v", err)
		return
	}
	defer conn.Close()

	c := pb.NewReplicaLeaderClient(conn)

	for {
		time.Sleep(time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		r, err := c.Collect(ctx, &pb.Message{Type: EMPTY, Content: "checking responses"})
		if err != nil {
			if logOutput {
				log.Printf("failed to collect: %v", err)
			}
			cancel()
			logOutput = false
			continue
		}

		logOutput = true
		replicaStateUpdateChannel <- &replicaStateUpdateRequest{updateType: 1, newDecisions: r.Decisions}
	}
}

// handlers
func (s *replicaServer) Request(ctx context.Context, in *pb.Message) (*pb.Message, error) {
	log.Printf("Request with command id %s received", in.CommandId)
	requests <- in.Command
	replicaStateUpdateChannel <- &replicaStateUpdateRequest{updateType: 2, newDecisions: nil}
	return &pb.Message{Type: EMPTY, Content: "success"}, nil
}

// TODO: finish this
func (s *replicaServer) Read(ctx context.Context, in *pb.Message) (*pb.Message, error) {
	log.Printf("Read request received, key: %s", in.Key)
	value, ok := keyValueLog[in.Key]
	if !ok {
		return nil, status.Error(codes.NotFound, "KEY IS NOT IN LOG")
	}
	return &pb.Message{Valid: true, Value: value}, nil
}

func (s *replicaServer) Collect(ctx context.Context, in *pb.Message) (*pb.Message, error) {
	var responseList []*pb.Response
	var i int32 = 0
	<-mutexChannel
	_, ok := decisions[i]
	for ok { // concurrent access (fatal)
		responseList = append(responseList, &pb.Response{Command: decisions[i].Command})
		i++
		_, ok = decisions[i]
	}
	mutexChannel <- 1
	return &pb.Message{Type: RESPONSES, Valid: true, Responses: responseList}, nil
}
