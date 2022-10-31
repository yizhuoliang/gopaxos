package main

import (
	pb "gopaxos/gopaxos"
	"log"
	"net"
	"os"
	"strconv"

	"google.golang.org/grpc"
)

var (
	cmdServer     *grpc.Server
	sctServer     *grpc.Server
	acceptorId    int32
	acceptorPorts = []string{"127.0.0.1:50057", "127.0.0.1:50058", "127.0.0.1:50059"}

	ballotNumber int32
	pvalues      []pb.BSC
)

type commanderServer struct {
	pb.UnimplementedCommanderAcceptorServer
}

type scottServer struct {
	pb.UnimplementedScottAcceptorServer
}

func main() {
	temp, _ := strconv.Atoi(os.Args[1])
	acceptorId = int32(temp)

}

func serveCmd(port string) {
	// listen client on port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("server listening at %v", lis.Addr())
	cmdServer = grpc.NewServer()
	pb.RegisterCommanderAcceptorServer(cmdServer, &commanderServer{})
	if err := cmdServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func serveSct(port string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("server listening at %v", lis.Addr())
	cmdServer = grpc.NewServer()
	pb.RegisterCommanderAcceptorServer(cmdServer, &commanderServer{})
	if err := cmdServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
