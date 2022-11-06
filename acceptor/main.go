package main

import (
	"context"
	pb "gopaxos/gopaxos"
	"log"
	"net"
	"os"
	"strconv"

	"google.golang.org/grpc"
)

var (
	server        *grpc.Server
	acceptorId    int32 // also considered as the serial of this acceptor
	acceptorPorts       = []string{"127.0.0.1:50057", "127.0.0.1:50058", "127.0.0.1:50059"}

	ballotNumber int32 = -1
	ballotLeader int32 = -1
	accepted     map[int32][]*pb.BSC

	mutexChannel chan int32
)

type acceptorServer struct {
	pb.UnimplementedLeaderAcceptorServer
}

func main() {
	temp, _ := strconv.Atoi(os.Args[1])
	acceptorId = int32(temp)

	// initialization
	accepted = make(map[int32][]*pb.BSC)
	mutexChannel = make(chan int32, 1)
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
func (s *acceptorServer) Scouting(ctx context.Context, in *pb.P1A) (*pb.P1B, error) {
	<-mutexChannel
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
	return &pb.P1B{AcceptorId: acceptorId, BallotNumber: currentBallotNumber, BallotLeader: currentBallotLeader, Accepted: acceptedList}, nil
}

func (s *acceptorServer) Commanding(ctx context.Context, in *pb.P2A) (*pb.P2B, error) {
	<-mutexChannel
	ballotNumber := ballotNumber // concurrency concern, avoid ballot number update during execution
	log.Printf("Commanding received")
	if in.Bsc.BallotNumber == ballotNumber && in.LeaderId == ballotLeader {
		accepted[ballotNumber] = append(accepted[ballotNumber], in.Bsc)
		log.Printf("Commanding accepted: ballot number %d, ballot leader %d", ballotNumber, ballotLeader)
	}
	currentBallotNumber := ballotNumber
	currentBallotLeader := ballotLeader
	mutexChannel <- 1
	return &pb.P2B{AcceptorId: acceptorId, BallotNumber: currentBallotNumber, BallotLeader: currentBallotLeader}, nil
}
