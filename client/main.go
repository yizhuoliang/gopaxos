package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "gopaxos/gopaxos"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	replicaNum = 2
)

var (
	clientId       string
	replicaPorts   = []string{"127.0.0.1:50053", "127.0.0.1:50054"}
	commandBuffer  [replicaNum]chan *pb.Command
	responseBuffer chan *pb.Responses
)

func main() {
	// input client id
	clientId = os.Args[1]

	// initialize command channels for messenger routines
	for i := 0; i < replicaNum; i++ {
		commandBuffer[i] = make(chan *pb.Command, 1)
	}

	// launch messenger and collector routines
	for i := 0; i < replicaNum; i++ {
		go MessengerRoutine(i)
		go CollectorRoutine(i)
	}

	// start handling user's operations
	var input string
	for {
		fmt.Printf("Enter 'Operate' or 'Get' (^C to quit): ")
		fmt.Scanf("%s", &input)

		if input == "Operate" {
			fmt.Printf("Enter the operation you want to perform")
			fmt.Scanf("%s", &input)
			for i := 0; i < replicaNum; i++ {

			}
		}
	}
}

func MessengerRoutine(serial int) {
	conn, err := grpc.Dial(replicaPorts[serial], grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect: %v", err)
		return
	}
	defer conn.Close()

	c := pb.NewClientReplicaClient(conn)

	for {

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	}
}

func CollectorRoutine(serial int) {

}
