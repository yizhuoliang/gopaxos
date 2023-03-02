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
	"google.golang.org/grpc/credentials/insecure"
	// "google.golang.org/protobuf/proto"
)

const (
	WINDOW = 5

	acceptorNum = 3
	leaderNum   = 2

	// Message types
	COMMAND   = 1
	READ      = 2
	WRITE     = 3
	RESPONSES = 4
	PROPOSAL  = 5
	DECISIONS = 6
	BEAT      = 7
	P1A       = 8
	P1B       = 9
	P2A       = 10
	P2B       = 11
	EMPTY     = 12

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

	// this is for replying
	readReplyMap map[string]chan string

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
		// simc = new(comm.RPCConnection).Init(uint64(replicaId), )
		replicaPorts = []string{"172.17.0.7:50050", "172.17.0.8:50050"}
		leaderPorts = []string{"172.17.0.5:50050", "172.17.0.6:50050"}
	}

	// initialization
	proposals = make(map[int32]*pb.Proposal)
	decisions = make(map[int32]*pb.Decision)
	keyValueLog = make(map[string]string)
	readReplyMap = make(map[string]chan string)
	replicaStateUpdateChannel = make(chan *replicaStateUpdateRequest, 1)
	requests = make(chan *pb.Command, 1)
	for i := 0; i < leaderNum; i++ {
		notificationChannel[i] = make(chan *pb.Message, 1)
		slotInUpdateChannel[i] = make(chan int, 1)
	}

	// go!
	go ReplicaStateUpdateRoutine()
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
							// TODO: possible deadlock
							requests <- p.Command
							replicaStateUpdateChannel <- &replicaStateUpdateRequest{updateType: 2, newDecisions: nil}
						}
					}
					switch d.Command.Type {
					case WRITE:
						// update log and update slot_out
						keyValueLog[d.Command.Key] = d.Command.Value
						slot_out++
						log.Printf("Log updated - key: %s, val: %s", d.Command.Key, d.Command.Value)
						d, ok = decisions[slot_out]
					case READ:
						// reply to clients
						// TODO: mark this read request as completed
						replyChan, chanOk := readReplyMap[d.Command.CommandId]
						val, valOk := keyValueLog[d.Command.Key]
						if chanOk {
							if valOk {
								replyChan <- val
							} else {
								replyChan <- "ERROR: key is not in log"
							}
						}
						slot_out++
						d, ok = decisions[slot_out]
					}
				}
			}
		} else if update.updateType == 2 {
			// PROPOSE NEW COMMAND
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
					slot_in++
				}
			}
		}
	}
}

func MessengerRoutine(serial int) {
	for {
		// reset connection for each message
		msgTosend := <-notificationChannel[serial]
		conn, err := grpc.Dial(leaderPorts[serial], grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			log.Printf("failed to connect: %v", err)
		} else {
			c := pb.NewReplicaLeaderClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			_, err := c.Propose(ctx, msgTosend)
			if err != nil {
				log.Printf("failed to propose: %v", err)
				cancel()
			}
		}
		conn.Close()
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
func (s *replicaServer) Write(ctx context.Context, in *pb.Message) (*pb.Message, error) {
	log.Printf("Request with command id %s received", in.CommandId)
	requests <- in.Command
	replicaStateUpdateChannel <- &replicaStateUpdateRequest{updateType: 2, newDecisions: nil}
	return &pb.Message{Type: EMPTY, Content: "success"}, nil
}

func (s *replicaServer) Read(ctx context.Context, in *pb.Message) (*pb.Message, error) {
	log.Printf("Read command received, key: %s", in.Command.Key)
	readReplyMap[in.Command.CommandId] = make(chan string, 1)
	requests <- in.Command
	replicaStateUpdateChannel <- &replicaStateUpdateRequest{updateType: 2, newDecisions: nil}
	value := <-readReplyMap[in.Command.CommandId]
	delete(readReplyMap, in.Command.CommandId)
	return &pb.Message{Type: EMPTY, Content: value}, nil
}

func (s *replicaServer) Collect(ctx context.Context, in *pb.Message) (*pb.Message, error) {
	var responseList []*pb.Response
	var i int32 = 0
	_, ok := decisions[i]
	for ok { // concurrent access (fatal)
		responseList = append(responseList, &pb.Response{Command: decisions[i].Command})
		i++
		_, ok = decisions[i]
	}
	return &pb.Message{Type: RESPONSES, Responses: responseList}, nil
}
