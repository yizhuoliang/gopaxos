package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	pb "github.com/yizhuoliang/gopaxos"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	replicaNum = 2
)

const (
	// Message types
	COMMAND   = 1
	READ      = 2
	RESPONSES = 3
	PROPOSAL  = 4
	DECISIONS = 5
	BEAT      = 6
	P1A       = 7
	P1B       = 8
	P2A       = 9
	P2B       = 10
	EMPTY     = 11
)

var (
	clientId              int32
	commandCount          = 0
	replicaPorts          = []string{"127.0.0.1:50053", "127.0.0.1:50054"}
	messageBuffers        [replicaNum]chan *pb.Message
	responded             []*pb.Response
	responseUpdateChannel chan []*pb.Response
	IOBlockChannel        chan int // preserve input output order
	checkSignal           chan int

	simon int // 1 = on, 0 = off
)

func main() {
	// input client id
	temp, _ := strconv.Atoi(os.Args[1])
	clientId = int32(temp)
	simon, _ = strconv.Atoi(os.Args[2])

	if simon == 1 {
		replicaPorts = []string{"172.17.0.7:50050", "172.17.0.8:50050"}
	}

	// initialize command channels for messenger routines
	for i := 0; i < replicaNum; i++ {
		messageBuffers[i] = make(chan *pb.Message, 1)
	}
	responseUpdateChannel = make(chan []*pb.Response, 1)
	checkSignal = make(chan int, 1)

	// launch messenger and collector routines
	for i := 0; i < replicaNum; i++ {
		go MessengerRoutine(i)
		go CollectorRoutine(i)
	}
	go responseUpdateRoutine()

	// start handling user's operations
	var input string
	IOBlockChannel = make(chan int, 1)
	for {
		fmt.Printf("Enter 'store' or 'read' or 'check' (^C to quit): ")
		fmt.Scanf("%s", &input)

		if input == "store" {
			var key string
			var value string
			fmt.Printf("Enter the key you want to store: ")
			fmt.Scanf("%s", &key)
			fmt.Printf("Enter the value you want to store: ")
			fmt.Scanf("%s", &value)
			// generate a new commandID
			commandCount += 1
			cid := "client" + strconv.Itoa(int(clientId)) + "-" + strconv.Itoa(commandCount)
			// push client commands to command buffers
			for i := 0; i < replicaNum; i++ {
				messageBuffers[i] <- &pb.Message{Type: COMMAND, Command: &pb.Command{CommandId: cid, ClientId: clientId, Key: key, Value: value}, CommandId: cid, ClientId: clientId, Key: key, Value: value}
			}
		} else if input == "read" {
			var key string
			fmt.Printf("Enter the key you want to read: ")
			fmt.Scanf("%s", &key)
			for i := 0; i < replicaNum; i++ {
				messageBuffers[i] <- &pb.Message{Type: READ, Key: key}
			}
			for i := 0; i < replicaNum; i++ {
				<-IOBlockChannel
			}
		} else if input == "check" {
			checkSignal <- 1
			<-IOBlockChannel
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
		msg := <-messageBuffers[serial]
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		switch msg.Type {
		case COMMAND:
			_, err = c.Request(ctx, msg)
			if err != nil {
				log.Printf("failed to request: %v", err)
				cancel()
			}
		case READ:
			value, err := c.Read(ctx, msg)
			if err != nil {
				log.Printf("failed to read: %v", err)
				cancel()
			} else {
				log.Printf("key: %s, value: %s", msg.Key, value)
			}
			IOBlockChannel <- 1
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
		<-checkSignal
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		r, err := c.Collect(ctx, &pb.Message{Type: EMPTY, Content: "checking responses\n"})
		if err != nil {
			log.Printf("failed to collect: %v", err)
			cancel()
		}
		// print commandId of responded requests
		if r.Valid {
			responseUpdateChannel <- r.Responses
		}
	}
}

func responseUpdateRoutine() {
	for {
		r := <-responseUpdateChannel
		responded = r
		for _, response := range responded {
			log.Printf("%s is responded. key: %s, value: %s\n", response.Command.CommandId, response.Command.Key, response.Command.Value)
		}
		IOBlockChannel <- 1
	}
}
