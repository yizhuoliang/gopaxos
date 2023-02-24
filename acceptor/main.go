package main

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"

	pb "github.com/yizhuoliang/gopaxos"
	"github.com/yizhuoliang/gopaxos/comm"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

const (
	COMMAND   = 1
	RESPONSES = 2
	PROPOSAL  = 3
	DECISIONS = 4
	BEAT      = 5
	P1A       = 6
	P1B       = 7
	P2A       = 8
	P2B       = 9
	EMPTY     = 10
)

var (
	server        *grpc.Server
	acceptorId    int32 // also considered as the serial of this acceptor
	acceptorPorts       = []string{"172.17.0.2:50050", "172.17.0.3:50050", "172.17.0.4:50050"}

	ballotNumber int32 = -1
	ballotLeader int32 = -1
	accepted     map[int32][]*pb.BSC

	mutexChannel chan int32

	simc *comm.RPCConnection
)

type acceptorServer struct {
	pb.UnimplementedLeaderAcceptorServer
}

type ReaderWriter interface {
	Write(b []byte) (n int, err error)
	Read(b []byte) (n int, err error)
	Close() error
}

func main() {
	temp, _ := strconv.Atoi(os.Args[1])
	acceptorId = int32(temp)

	// connect sim
	simc = new(comm.RPCConnection).Init(uint64(acceptorId), 3)

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
	log.Printf("server listening at %v", lis.Addr())
	server = grpc.NewServer()
	pb.RegisterLeaderAcceptorServer(server, &acceptorServer{})
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// gRPC handlers
func (s *acceptorServer) Scouting(ctx context.Context, in *pb.Message) (*pb.Message, error) {

	// P1A received
	tosend, err := proto.Marshal(in)
	if err != nil {
		log.Fatalf("marshal err:%v\n", err)
	}
	_, err = simc.OutConn.Write(tosend)
	if err != nil {
		log.Fatalf("Write to simulator failed, err:%v\n", err)
	}

	<-mutexChannel
	log.Printf("Scouting received")
	if in.BallotNumber > ballotNumber || (in.BallotNumber == ballotNumber && in.LeaderId != ballotLeader) {
		ballotNumber = in.BallotNumber
		ballotLeader = in.LeaderId
	}
	var acceptedList []*pb.BSC
	for i := 1; i < int(ballotNumber); i++ {
		if bscs, ok := accepted[int32(i)]; ok {
			acceptedList = append(acceptedList, bscs...)
		}
	}
	currentBallotNumber := ballotNumber
	currentBallotLeader := ballotLeader
	log.Printf("Scouting received, current states: ballot number %d, ballot leader %d", ballotNumber, ballotLeader)
	mutexChannel <- 1

	// P1B sent
	tosend, err = proto.Marshal(&pb.Message{Type: P1B, AcceptorId: acceptorId, BallotNumber: currentBallotNumber, BallotLeader: currentBallotLeader, Accepted: acceptedList, Req: in, Send: true})
	if err != nil {
		log.Fatalf("marshal err:%v\n", err)
	}
	_, err = simc.OutConn.Write(tosend)
	if err != nil {
		log.Fatalf("Write to simulator failed, err:%v\n", err)
	}

	return &pb.Message{Type: P1B, AcceptorId: acceptorId, BallotNumber: currentBallotNumber, BallotLeader: currentBallotLeader, Accepted: acceptedList}, nil
}

func (s *acceptorServer) Commanding(ctx context.Context, in *pb.Message) (*pb.Message, error) {

	// P2A received
	tosend, err := proto.Marshal(in)
	if err != nil {
		log.Fatalf("marshal err:%v\n", err)
	}
	_, err = simc.OutConn.Write(tosend)
	if err != nil {
		log.Fatalf("Write to simulator failed, err:%v\n", err)
	}

	<-mutexChannel
	ballotNumber := ballotNumber // concurrency concern, avoid ballot number update during execution
	log.Printf("Commanding received")
	if in.Bsc.BallotNumber >= ballotNumber && in.LeaderId == ballotLeader {
		ballotNumber = in.Bsc.BallotNumber
		accepted[ballotNumber] = append(accepted[ballotNumber], in.Bsc)
		log.Printf("Commanding accepted: ballot number %d, ballot leader %d", ballotNumber, ballotLeader)
	}
	currentBallotNumber := ballotNumber
	currentBallotLeader := ballotLeader
	mutexChannel <- 1

	// P2B sent
	tosend, err = proto.Marshal(&pb.Message{Type: P2B, AcceptorId: acceptorId, BallotNumber: currentBallotNumber, BallotLeader: currentBallotLeader, Req: in, Send: true})
	if err != nil {
		log.Fatalf("marshal err:%v\n", err)
	}
	_, err = simc.OutConn.Write(tosend)
	if err != nil {
		log.Fatalf("Write to simulator failed, err:%v\n", err)
	}

	return &pb.Message{Type: P2B, AcceptorId: acceptorId, BallotNumber: currentBallotNumber, BallotLeader: currentBallotLeader}, nil
}
