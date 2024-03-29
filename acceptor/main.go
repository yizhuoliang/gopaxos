package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	pb "github.com/yizhuoliang/gopaxos"
	"github.com/yizhuoliang/gopaxos/comm"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

const (
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
	server     *grpc.Server
	acceptorId int32 // also considered as the serial of this acceptor

	// for no-sim test
	acceptorPorts = []string{"127.0.0.1:50057", "127.0.0.1:50058", "127.0.0.1:50059"}

	ballotNumber int32 = -1
	ballotLeader int32 = -1
	accepted     map[int32][]*pb.BSC

	mutexChannel chan int32

	simc   *comm.RPCConnection
	simon  int // 1 = on, 0 = off
	syncon int // 1 = on, 0 = off
)

type acceptorServer struct {
	pb.UnimplementedLeaderAcceptorServer
}

func main() {
	temp, _ := strconv.Atoi(os.Args[1])
	acceptorId = int32(temp)
	simon, _ = strconv.Atoi(os.Args[2])

	// connect sim
	if simon == 1 {
		simc = new(comm.RPCConnection).Init(uint64(acceptorId), ACCEPTOR)
		acceptorPorts = []string{"172.17.0.8:50050", "172.17.0.7:50050", "172.17.0.6:50050"}
	} else if simon == 2 {
		simc = new(comm.RPCConnection).Init(uint64(acceptorId), ACCEPTOR)
		acceptorPorts = []string{"172.17.0.11:50050", "172.17.0.10:50050", "172.17.0.9:50050"}
	}

	// overwrite ports with file
	readPortsFile()

	// initialization
	accepted = make(map[int32][]*pb.BSC)
	mutexChannel = make(chan int32, 1)
	mutexChannel <- 1
	serve(acceptorPorts[acceptorId])
}

func serve(port string) {
	// listen client on port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Printf("server listening at %v", lis.Addr())
	server = grpc.NewServer()
	pb.RegisterLeaderAcceptorServer(server, &acceptorServer{})
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func readPortsFile() {
	f, err := os.Open("./ports.txt")
	// if reading failed, do nothing
	if err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for i := 0; i < 7; i++ {
			scanner.Scan()
			if i >= 4 && i <= 6 {
				acceptorPorts[i-4] = scanner.Text()
			}
		}
	} else {
		fmt.Printf("%v\n", err)
	}
}

func commandId2String(cid int64) string {
	return strconv.FormatInt(cid>>54, 10) + "-" + strconv.FormatInt(cid%(int64(1)<<53), 10)
}

// gRPC handlers
func (s *acceptorServer) Scouting(ctx context.Context, in *pb.Message) (*pb.Message, error) {

	// P1A received
	if simon >= 1 {
		reqId, tosend, offset := simc.AllocateRequest((uint64)(proto.Size(in)))
		b, err := proto.Marshal(in)
		if err != nil {
			log.Fatalf("marshal err:%v\n", err)
		}
		copy(tosend[offset:], b)
		ch := simc.Request(tosend)
		simc.WaitFor(reqId, ch)
	}

	<-mutexChannel
	fmt.Printf("Scouting received\n")
	oldBallotNumber := ballotNumber
	if in.BallotNumber > ballotNumber || (in.BallotNumber == ballotNumber && in.LeaderId != ballotLeader) {
		ballotNumber = in.BallotNumber
		ballotLeader = in.LeaderId
	}
	var acceptedList []*pb.BSC
	for i := 0; i <= int(oldBallotNumber); i++ {
		if bscs, ok := accepted[int32(i)]; ok {
			acceptedList = append(acceptedList, bscs...)
		}
	}
	currentBallotNumber := ballotNumber
	currentBallotLeader := ballotLeader
	fmt.Printf("Scouting received, current states: ballot number %d, ballot leader %d\n", ballotNumber, ballotLeader)
	mutexChannel <- 1

	// P1B sent
	if simon >= 1 {
		m := pb.Message{Type: P1B, AcceptorId: acceptorId, BallotNumber: currentBallotNumber, BallotLeader: currentBallotLeader, Accepted: acceptedList, Req: in, Send: true}
		reqId, tosend, offset := simc.AllocateRequest((uint64)(proto.Size(&m)))
		b, err := proto.Marshal(&m)
		if err != nil {
			log.Fatalf("marshal err:%v\n", err)
		}
		copy(tosend[offset:], b)
		ch := simc.Request(tosend)
		simc.WaitFor(reqId, ch)
	}

	return &pb.Message{Type: P1B, AcceptorId: acceptorId, BallotNumber: currentBallotNumber, BallotLeader: currentBallotLeader, Accepted: acceptedList}, nil
}

func (s *acceptorServer) Commanding(ctx context.Context, in *pb.Message) (*pb.Message, error) {

	// P2A received
	if simon >= 1 {
		reqId, tosend, offset := simc.AllocateRequest((uint64)(proto.Size(in)))
		b, err := proto.Marshal(in)
		if err != nil {
			log.Fatalf("marshal err:%v\n", err)
		}
		copy(tosend[offset:], b)
		ch := simc.Request(tosend)
		simc.WaitFor(reqId, ch)
	}

	<-mutexChannel
	ballotNumber := ballotNumber // concurrency concern, avoid ballot number update during execution
	fmt.Printf("Commanding received:  ballot number %d, ballot leader %d, comID: %s, slot: %d\n", ballotNumber, ballotLeader, commandId2String(in.Bsc.Command.CommandId), in.Bsc.SlotNumber)
	acceptCid := int64(-1)
	if in.Bsc.BallotNumber >= ballotNumber && in.LeaderId == ballotLeader {
		ballotNumber = in.Bsc.BallotNumber
		acceptCid = in.Bsc.Command.CommandId
		accepted[ballotNumber] = append(accepted[ballotNumber], in.Bsc)
		fmt.Printf("Commanding accepted: ballot number %d, ballot leader %d, comID: %s, slot: %d\n", ballotNumber, ballotLeader, commandId2String(in.Bsc.Command.CommandId), in.Bsc.SlotNumber)
	}
	currentBallotNumber := ballotNumber
	currentBallotLeader := ballotLeader
	mutexChannel <- 1

	// P2B sent
	if simon >= 1 {
		m := pb.Message{Type: P2B, AcceptorId: acceptorId, BallotNumber: currentBallotNumber, BallotLeader: currentBallotLeader, CommandId: acceptCid, Req: in, Send: true}
		reqId, tosend, offset := simc.AllocateRequest((uint64)(proto.Size(&m)))
		b, err := proto.Marshal(&m)
		if err != nil {
			log.Fatalf("marshal err:%v\n", err)
		}
		copy(tosend[offset:], b)
		ch := simc.Request(tosend)
		simc.WaitFor(reqId, ch)
	}

	return &pb.Message{Type: P2B, AcceptorId: acceptorId, BallotNumber: currentBallotNumber, BallotLeader: currentBallotLeader}, nil
}
