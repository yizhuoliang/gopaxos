package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	pb "github.com/yizhuoliang/gopaxos"
	// "github.com/yizhuoliang/gopaxos/comm"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	// "google.golang.org/protobuf/proto"
)

const (
	WINDOW = 2000

	acceptorNum = 3
	leaderNum   = 2

	// Message types
	COMMAND   = 1
	READ      = 2
	WRITE     = 3
	RESPONSES = 4
	PROPOSAL  = 5
	DECISION  = 6
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
	replicaId        int32
	serverForClients *grpc.Server
	serverForLeaders *grpc.Server

	// for no-sim tests
	replicaPorts = []string{"127.0.0.1:50053", "127.0.0.1:50054"}
	leaderPorts  = []string{"127.0.0.1:50055", "127.0.0.1:50056"}

	slot_in            int32 = 0
	slot_out           int32 = 0
	newCommandsChannel chan *pb.Command
	proposals          map[int32]*pb.Proposal
	decisions          map[int32]*pb.Decision

	replicaStateUpdateChannel chan *replicaStateUpdateRequest
	notificationChannel       [leaderNum]chan *pb.Message
	slotInUpdateChannel       [leaderNum]chan int

	// Yeah, this is the key-value map
	keyValueLog map[string]string

	// this is for replying
	clientReplyMap map[string]chan string
	replyMapLock   sync.RWMutex

	// simc *comm.RPCConnection
	simon int // 1 = on, 0 = off
)

type replicaServer struct {
	pb.UnimplementedReplicaServer
}

type replicaStateUpdateRequest struct {
	updateType  int
	newDecision *pb.Decision
}

func main() {
	temp, _ := strconv.Atoi(os.Args[1])
	replicaId = int32(temp)
	simon, _ = strconv.Atoi(os.Args[2])

	// connect sim
	if simon == 1 {
		// simc = new(comm.RPCConnection).Init(uint64(replicaId), )
		replicaPorts = []string{"172.17.0.3:50050", "172.17.0.2:50050"}
		leaderPorts = []string{"172.17.0.5:50050", "172.17.0.4:50050"}
	} else if simon == 2 {
		// simc = new(comm.RPCConnection).Init(uint64(replicaId), )
		replicaPorts = []string{"172.17.0.6:50050", "172.17.0.5:50050"}
		leaderPorts = []string{"172.17.0.8:50050", "172.17.0.7:50050"}
	}

	// overwrite ports with the file
	readPortsFile()

	// initialization
	proposals = make(map[int32]*pb.Proposal)
	decisions = make(map[int32]*pb.Decision)
	keyValueLog = make(map[string]string)
	clientReplyMap = make(map[string]chan string)
	replicaStateUpdateChannel = make(chan *replicaStateUpdateRequest, 1)
	newCommandsChannel = make(chan *pb.Command, 1)
	for i := 0; i < leaderNum; i++ {
		notificationChannel[i] = make(chan *pb.Message, 1)
		slotInUpdateChannel[i] = make(chan int, 1)
	}

	// go!
	go ReplicaStateUpdateRoutine()
	for i := 0; i < leaderNum; i++ {
		go MessengerRoutine(i)
	}

	serve(replicaPorts[replicaId])
}

func serve(port string) {
	// listen client on port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("serverForClients listening at %v", lis.Addr())
	serverForClients = grpc.NewServer()
	pb.RegisterReplicaServer(serverForClients, &replicaServer{})
	if err := serverForClients.Serve(lis); err != nil {
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
			decision := update.newDecision
			decisions[decision.SlotNumber] = decision
			d, ok := decisions[slot_out]
			// PERFORM - UPDATING THE KEY_VALUE MAP
			for ok {
				p, okProp := proposals[slot_out]
				if okProp {
					delete(proposals, slot_out)
					if d.Command.CommandId != p.Command.CommandId {
						go RetryRequestRoutine(p.Command, &replicaStateUpdateRequest{updateType: 2})
					}
				}
				switch d.Command.Type {
				case WRITE:
					// update log and reply to clients and update slot_out
					keyValueLog[d.Command.Key] = d.Command.Value
					replyMapLock.RLock()
					replyChan, chanOk := clientReplyMap[d.Command.CommandId]
					replyMapLock.RUnlock()
					if chanOk {
						replyChan <- ""
					}
					slot_out++
					log.Printf("Log updated - key: %s\n", d.Command.Key)
					fmt.Printf("slot_out: %d, slot_in: %d\n", slot_out, slot_in)
					d, ok = decisions[slot_out]
				case READ:
					// reply to clients
					// TODO: mark this read request as completed
					replyMapLock.RLock()
					replyChan, chanOk := clientReplyMap[d.Command.CommandId]
					replyMapLock.RUnlock()
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

		} else if update.updateType == 2 {
			// PROPOSE NEW COMMAND
			// reference sudo code propose()
			if slot_in < slot_out+WINDOW {
				_, ok := decisions[slot_in]
				if !ok {
					request := <-newCommandsChannel
					proposals[slot_in] = &pb.Proposal{SlotNumber: slot_in, Command: request}
					msgTosend := &pb.Message{Type: PROPOSAL, SlotNumber: slot_in, Command: request}
					for i := 0; i < leaderNum; i++ {
						notificationChannel[i] <- msgTosend
					}
					slot_in++
				}
			} else {
				go RetryRequestPreserveOrderRoutine()
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

// when a proposed command's slot is decided for another proposal, we need to find a new slot for this command
func RetryRequestRoutine(command *pb.Command, update *replicaStateUpdateRequest) {
	newCommandsChannel <- command
	replicaStateUpdateChannel <- update
}

func RetryRequestPreserveOrderRoutine() {
	replicaStateUpdateChannel <- &replicaStateUpdateRequest{updateType: 2}
}

func readPortsFile() {
	f, err := os.Open("./ports.txt")
	// if reading failed, do nothing
	if err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for i := 0; i < 7; i++ {
			scanner.Scan()
			if i <= 1 {
				replicaPorts[i] = scanner.Text()
			} else if i >= 2 && i <= 3 {
				leaderPorts[i-2] = scanner.Text()
			}
		}
	}
}

// gRPC handlers for Clients
func (s *replicaServer) Write(ctx context.Context, in *pb.Message) (*pb.Message, error) {
	log.Printf("Request with command id %s received", in.CommandId)
	myChan := make(chan string, 1)
	replyMapLock.Lock()
	clientReplyMap[in.Command.CommandId] = myChan
	replyMapLock.Unlock()
	newCommandsChannel <- in.Command
	replicaStateUpdateChannel <- &replicaStateUpdateRequest{updateType: 2, newDecision: nil}
	<-myChan
	replyMapLock.Lock()
	delete(clientReplyMap, in.Command.CommandId)
	replyMapLock.Unlock()
	return &pb.Message{Type: EMPTY, Content: "success"}, nil
}

func (s *replicaServer) Read(ctx context.Context, in *pb.Message) (*pb.Message, error) {
	log.Printf("Read command received, key: %s", in.Command.Key)
	myChan := make(chan string, 1)
	replyMapLock.Lock()
	clientReplyMap[in.Command.CommandId] = myChan
	replyMapLock.Unlock()
	newCommandsChannel <- in.Command
	replicaStateUpdateChannel <- &replicaStateUpdateRequest{updateType: 2, newDecision: nil}
	value := <-myChan
	replyMapLock.Lock()
	delete(clientReplyMap, in.Command.CommandId)
	replyMapLock.Unlock()
	return &pb.Message{Type: EMPTY, Content: value}, nil
}

// gRPC handlers for Leader
func (s *replicaServer) Decide(ctx context.Context, in *pb.Message) (*pb.Message, error) {
	log.Printf("Decision with command id %s received", in.Decision.Command.CommandId)
	replicaStateUpdateChannel <- &replicaStateUpdateRequest{updateType: 1, newDecision: in.Decision}
	return &pb.Message{Type: EMPTY}, nil
}
