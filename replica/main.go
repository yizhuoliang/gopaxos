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

	replicaStateUpdateChannel chan *replicaStateUpdateRequest
	notificationChannel       [leaderNum]chan int
	slotInUpdateChannel       [leaderNum]chan int

	mutexChannel chan int32
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

	// initialization
	proposals = make(map[int32]*pb.Proposal)
	decisions = make(map[int32]*pb.Decision)
	replicaStateUpdateChannel = make(chan *replicaStateUpdateRequest, 1)
	for i := 0; i < leaderNum; i++ {
		requests[i] = make(chan *pb.Command, 1)
		notificationChannel[i] = make(chan int, 1)
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
				if ok {
					p, ok := proposals[slot_out]
					if ok {
						delete(proposals, slot_out)
						if d.Command.CommandId != p.Command.CommandId {
							requests[update.serial] <- p.Command
							notificationChannel[update.serial] <- 1
						}
					}
					// atomic
					state = state + decision.Command.Operation
					slot_out++
					// end atomic
					log.Printf("Operation %s is performed", decision.Command.Operation)
				}
			}
		} else if update.updateType == 2 {
			log.Printf("messenger %d slot_in %d processing update type 2...", update.serial, slot_in)
			// reference sudo code propose()
			if slot_in < slot_out+WINDOW {
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
				slotInUpdateChannel[update.serial] <- 1
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
		<-notificationChannel[serial]
		replicaStateUpdateChannel <- &replicaStateUpdateRequest{updateType: 2, newDecisions: nil, serial: serial, c: c}
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
		r, err := c.Collect(ctx, &pb.Empty{Content: "checking responses"})
		if err != nil {
			if logOutput {
				log.Printf("failed to collect: %v", err)
			}
			cancel()
			logOutput = false
			continue
		}
		if r.Valid {
			logOutput = true
			replicaStateUpdateChannel <- &replicaStateUpdateRequest{updateType: 1, newDecisions: r.Decisions, serial: serial}
		}
	}
}

// handlers
func (s *replicaServer) Request(ctx context.Context, in *pb.Command) (*pb.Empty, error) {
	log.Printf("Request with command id %s received", in.CommandId)
	for i := 0; i < leaderNum; i++ {
		requests[i] <- in
		notificationChannel[i] <- 1
	}
	return &pb.Empty{Content: "success"}, nil
}

func (s *replicaServer) Collect(ctx context.Context, in *pb.Empty) (*pb.Responses, error) {
	var responseList []*pb.Response
	var i int32 = 1
	<-mutexChannel
	_, ok := decisions[i]
	for ok { // concurrent access (fatal)
		responseList = append(responseList, &pb.Response{Command: decisions[i].Command})
		i++
		_, ok = decisions[i]
	}
	mutexChannel <- 1
	return &pb.Responses{Valid: true, Responses: responseList}, nil
}
