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

	// Message types
	COMMAND   = 1
	READ      = 2
	WRITE     = 3
	RESPONSES = 4
	PROPOSAL  = 5
	DECISIONS = 6
	BEAT      = 7
	P1A       = 8
	P1B       = 9
	P2A       = 10
	P2B       = 11
	EMPTY     = 12
)

var (
	clientId       int32
	commandCount   = 0
	replicaPorts   = []string{"127.0.0.1:50053", "127.0.0.1:50054"}
	messageBuffers [replicaNum]chan *pb.Message
	IOBlockChannel chan int // preserve input output order

	simon int // 1 = on, 0 = off
)

func main() {
	// input client id
	temp, _ := strconv.Atoi(os.Args[1])
	clientId = int32(temp)
	simon, _ = strconv.Atoi(os.Args[2])

	if simon == 1 {
		replicaPorts = []string{"172.17.0.7:50050", "172.17.0.8:50050"}
	} else if simon == 2 {
		replicaPorts = []string{"172.17.0.6:50050", "172.17.0.5:50050"}
	}

	// initialize command channels for messenger routines
	for i := 0; i < replicaNum; i++ {
		messageBuffers[i] = make(chan *pb.Message, 1)
	}

	// launch messenger and collector routines
	for i := 0; i < replicaNum; i++ {
		go MessengerRoutine(i)
	}

	// start handling user's operations
	var input string
	IOBlockChannel = make(chan int, replicaNum)
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
			cid := "client" + strconv.Itoa(int(clientId)) + "-W" + strconv.Itoa(commandCount)
			// push client commands to command buffers
			for i := 0; i < replicaNum; i++ {
				messageBuffers[i] <- &pb.Message{Type: WRITE, Command: &pb.Command{Type: WRITE, CommandId: cid, ClientId: clientId, Key: key, Value: value}, CommandId: cid, ClientId: clientId, Key: key, Value: value}
			}
			for i := 0; i < replicaNum; i++ {
				<-IOBlockChannel
			}
		} else if input == "read" {
			var key string
			fmt.Printf("Enter the key you want to read: ")
			fmt.Scanf("%s", &key)
			// generate a new commandID
			commandCount += 1
			cid := "client" + strconv.Itoa(int(clientId)) + "-R" + strconv.Itoa(commandCount)
			for i := 0; i < replicaNum; i++ {
				messageBuffers[i] <- &pb.Message{Type: READ, Command: &pb.Command{Type: READ, CommandId: cid, ClientId: clientId, Key: key}}
			}
			for i := 0; i < replicaNum; i++ {
				<-IOBlockChannel
			}
		} else if input == "check" {
			for i := 0; i < replicaNum; i++ {
				messageBuffers[i] <- &pb.Message{Type: EMPTY, Content: "checking responses"}
			}
			for i := 0; i < replicaNum; i++ {
				<-IOBlockChannel
			}
		}
	}
}

func MessengerRoutine(serial int) {
	for {
		// reset connection for each message
		msg := <-messageBuffers[serial]
		conn, err := grpc.Dial(replicaPorts[serial], grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			log.Printf("failed to connect: %v", err)
		} else {
			c := pb.NewClientReplicaClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			switch msg.Type {
			case WRITE:
				_, err := c.Write(ctx, msg)
				if err != nil {
					log.Printf("failed to request: %v", err)
					cancel()
				}
			case READ:
				r, err := c.Read(ctx, msg)
				if err != nil {
					log.Printf("failed to read: %v", err)
					cancel()
				} else {
					log.Printf("key: %s, value: %s", msg.Command.Key, r.Content)
				}
			case EMPTY:
				r, err := c.Collect(ctx, msg)
				if err != nil {
					log.Printf("failed to collect: %v", err)
					cancel()
				} else {
					for _, response := range r.Responses {
						log.Printf("%s is responded. key: %s, value: %s", response.Command.CommandId, response.Command.Key, response.Command.Value)
					}
				}
			}
		}
		conn.Close()
		IOBlockChannel <- 1
	}
}
