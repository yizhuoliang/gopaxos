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
	acceptorId    int32
	acceptorPorts = []string{"127.0.0.1:50057", "127.0.0.1:50058", "127.0.0.1:50059"}

	ballotNumber int32
	accepted     []pb.BSC

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

		// ISSUE: need more consideration on the concurrency of ballotNumber
		if newNum > ballotNumber {
			ballotNumber = newNum
		}
	}
}

// gRPC handlers
func (s *acceptorServer) Scouting(ctx context.Context, in *pb.P1A) (*pb.Empty, error) {
	ballotNumberUpdateChannel <- in.BallotNumber // considerations on why not check her
	return &pb.Empty{Content: "success"}, nil
}
