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
	acceptorPorts = []string{"127.0.0.1:50057", "127.0.0.1:50058", "127.0.0.1:50059"}

	ballotNumber int32
	accepted     map[int32][]*pb.BSC

	ballotNumberUpdateChannel chan int32
)

type acceptorServer struct {
	pb.UnimplementedLeaderAcceptorServer
}

func main() {
	temp, _ := strconv.Atoi(os.Args[1])
	acceptorId = int32(temp)

	ballotNumberUpdateChannel = make(chan int32)
	go ballotNumberUpdateRoutine()

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

func ballotNumberUpdateRoutine() {
	for {
		newNum := <-ballotNumberUpdateChannel

		// CONCERN: need more consideration on the concurrency of ballotNumber
		if newNum > ballotNumber {
			ballotNumber = newNum
		}
	}
}

// gRPC handlers
func (s *acceptorServer) Scouting(ctx context.Context, in *pb.P1A) (*pb.P1B, error) {
	ballotNumberUpdateChannel <- in.BallotNumber // considerations on why not check her
	var acceptedList []*pb.BSC
	for i := 1; i < int(ballotNumber); i++ {
		if bscs, ok := accepted[int32(i)]; ok {
			for _, bsc := range bscs {
				acceptedList = append(acceptedList, bsc)
			}
		}
	}
	return &pb.P1B{AcceptorId: acceptorId, BallotNumber: ballotNumber, Accepted: acceptedList}, nil
}

func (s *acceptorServer) Commanding(ctx context.Context, in *pb.P2A) (*pb.P2B, error) {
	currentBallotNumber := ballotNumber // concurrency concern, avoid ballot number update during execution
	if in.Bsc.BallotNumber == currentBallotNumber {
		accepted[currentBallotNumber] = append(accepted[currentBallotNumber], in.Bsc)
	}
	return &pb.P2B{AcceptorId: acceptorId, BallotNumber: currentBallotNumber}, nil
}
