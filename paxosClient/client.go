package client

import (
	"context"
	"errors"
	"log"
	"strconv"
	"sync"
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

type Client struct {
	clientId        int32
	commandCount    int
	replicaPorts    []string
	messageChannels [replicaNum]chan *pb.Message
	replyChannels   [replicaNum]chan *reply

	simon int // 1 = on, 0 = off

	mu sync.Mutex // YES! MUTEX!
}

type reply struct {
	value string
	err   error
}

func NewPaxosClient(clientId int, simon int) *Client {
	client := &Client{}
	client.clientId = int32(clientId)
	client.commandCount = 0
	if simon == 0 {
		client.replicaPorts = []string{"127.0.0.1:50053", "127.0.0.1:50054"}
	} else {
		client.replicaPorts = []string{"172.17.0.7:50050", "172.17.0.8:50050"}
	}

	// initialize command channels for messenger routines
	for i := 0; i < replicaNum; i++ {
		client.messageChannels[i] = make(chan *pb.Message, 1)
		client.replyChannels[i] = make(chan *reply)
	}

	// launch messenger and collector routines
	for i := 0; i < replicaNum; i++ {
		go client.MessengerRoutine(i)
	}
	return client
}

func (client *Client) Store(key string, value string) error {
	client.mu.Lock()
	defer client.mu.Unlock()

	client.commandCount += 1
	cid := "client" + strconv.Itoa(int(client.clientId)) + "-W" + strconv.Itoa(client.commandCount)
	// push client commands to command buffers
	for i := 0; i < replicaNum; i++ {
		client.messageChannels[i] <- &pb.Message{Type: WRITE, Command: &pb.Command{Type: WRITE, CommandId: cid, ClientId: client.clientId, Key: key, Value: value}, CommandId: cid, ClientId: client.clientId, Key: key, Value: value}
	}
	var err error = nil
	errCount := 0
	for i := 0; i < replicaNum; i++ {
		reply := <-client.replyChannels[i]
		if reply.err != nil {
			errCount++
			err = reply.err
		}
	}
	// NOTE: only return an error if all replica failed to handle
	if errCount >= replicaNum {
		return err
	}
	return nil
}

func (client *Client) Read(key string) (string, error) {
	client.mu.Lock()
	defer client.mu.Unlock()

	client.commandCount += 1
	cid := "client" + strconv.Itoa(int(client.clientId)) + "-R" + strconv.Itoa(client.commandCount)
	for i := 0; i < replicaNum; i++ {
		client.messageChannels[i] <- &pb.Message{Type: READ, Command: &pb.Command{Type: READ, CommandId: cid, ClientId: client.clientId, Key: key}}
	}
	var err error = nil
	errCount := 0
	value := ""
	for i := 0; i < replicaNum; i++ {
		reply := <-client.replyChannels[i]
		if reply.err != nil {
			errCount++
			err = reply.err
		} else {
			value = reply.value
		}
	}
	// NOTE: only return an error if all replica failed to handle
	if errCount >= replicaNum {
		return value, err
	}
	return value, nil
}

// func (client *Client)

func (client *Client) MessengerRoutine(serial int) {
	for {
		// reset connection for each message
		msg := <-client.messageChannels[serial]
		conn, err := grpc.Dial(client.replicaPorts[serial], grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			client.replyChannels[serial] <- &reply{err: err}
		} else {
			c := pb.NewClientReplicaClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			switch msg.Type {
			case WRITE:
				_, err := c.Write(ctx, msg)
				if err != nil {
					client.replyChannels[serial] <- &reply{err: err}
					cancel()
				}
			case READ:
				r, err := c.Read(ctx, msg)
				if err != nil {
					client.replyChannels[serial] <- &reply{err: err}
					cancel()
				} else {
					log.Printf("key: %s, value: %s", msg.Command.Key, r.Content)
					client.replyChannels[serial] <- &reply{value: r.Content, err: nil}
				}
			default:
				client.replyChannels[serial] <- &reply{err: errors.New("unknown command type")}
			}
		}
		conn.Close()
	}
}
