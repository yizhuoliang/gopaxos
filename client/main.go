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
	replicaPorts   = []string{"127.0.0.1:50053", "127.0.0.1:50054"}
	commandBuffers [replicaNum]chan *pb.Command
	responded      []*pb.Response
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
		fmt.Printf("Enter 'operate' or 'check' (^C to quit): ")
		fmt.Scanf("%s", &input)

		if input == "operate" {
			fmt.Printf("Enter the operation you want to perform: ")
			fmt.Scanf("%s", &input)
			// generate a new commandID
			commandCount += 1
			cid := strconv.Itoa(int(clientId)) + "-" + strconv.Itoa(commandCount)
			// push client commands to command buffers
			for i := 0; i < replicaNum; i++ {
				commandBuffers[i] <- &pb.Command{ClientId: clientId, CommandId: cid, Operation: input}
			}
		} else if input == "check" {
			for _, response := range responded {
				log.Printf("%s is responded: %s\n", response.Comamnd.CommandId, response.Comamnd.Operation)
			}
		}
	}
}

func MessengerRoutine(serial int) {
	conn, err := grpc.Dial(replicaPorts[serial], grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	c := pb.NewClientReplicaClient(conn)

	for {
		command := <-commandBuffers[serial]
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		_, err = c.Request(ctx, command)
		if err != nil {
			log.Printf("failed to request: %v\n", err)
			cancel()
		}
	}
}

func CollectorRoutine(serial int) {
	conn, err := grpc.Dial(replicaPorts[serial], grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	c := pb.NewClientReplicaClient(conn)

	for {
		time.Sleep(time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		r, err := c.Collect(ctx, &pb.Empty{Content: "checking responses\n"})
		if err != nil {
			log.Printf("failed to collect: %v", err)
			cancel()
		}
		// print commandId of responded requests
		if r.Valid {
			responded = r.Responses
		}
	}
}
