package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	pb "gopaxos/gopaxos"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	replicaNum = 2
)

var (
	clientId       int32
	commandCount   = 0
	replicaPorts   = []string{"amd167.utah.cloudlab.us:50053", "amd159.utah.cloudlab.us:50054"}
	commandBuffers [replicaNum]chan *pb.Command
	responseBuffer chan *pb.Responses
)

func main() {
	// input client id
	temp, _ := strconv.Atoi(os.Args[1])
	clientId = int32(temp)

	// initialize command channels for messenger routines
	for i := 0; i < replicaNum; i++ {
		commandBuffers[i] = make(chan *pb.Command, 1)
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
			// generate a new commandID
			commandCount += 1
			cid := strconv.Itoa(int(clientId)) + "-" + strconv.Itoa(commandCount)
			// push client commands to command buffers
			for i := 0; i < replicaNum; i++ {
				commandBuffers[i] <- &pb.Command{ClientId: clientId, CommandId: cid, Operation: input}
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
		command := <-commandBuffers[serial]
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err = c.Request(ctx, command)
		if err != nil {
			log.Printf("failed to request: %v", err)
		} else {
			log.Printf("Request sent")
		}
	}
}

func CollectorRoutine(serial int) {
	conn, err := grpc.Dial(replicaPorts[serial], grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect: %v", err)
		return
	}
	defer conn.Close()

	c := pb.NewClientReplicaClient(conn)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		r, err := c.Collect(ctx, &pb.Empty{Content: "checking responses"})
		if err != nil {
			log.Printf("failed to collect: %v", err)
		}
		// print commandId of responded requests
		if r.Valid {
			for _, response := range r.Responses {
				log.Printf("Command %s is responded", response.CommandId)
			}
		}
	}
}
