package client

import (
	"context"
	"errors"
	"sync/atomic"
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
	DECISION  = 6
	BEAT      = 7
	P1A       = 8
	P1B       = 9
	P2A       = 10
	P2B       = 11
	EMPTY     = 12
)

type Client struct {
	clientId     int32
	commandCount uint32
	replicaPorts []string
	// commandChannels  map[int][]chan *pb.Message
	// replyChannels    map[int][]chan *reply
	connectedClients []pb.ReplicaClient

	incrementCommandNumberChannel chan int
	commandNumberReplyChannel     chan int
}

type reply struct {
	value string
	err   error
}

func NewPaxosClient(clientId int, simon int, replicaPorts []string) *Client {
	client := &Client{}
	client.clientId = int32(clientId)
	client.commandCount = 0
	if simon == 0 {
		client.replicaPorts = []string{"127.0.0.1:50053", "127.0.0.1:50054"}
	} else {
		client.replicaPorts = []string{"172.17.0.7:50050", "172.17.0.8:50050"}
	}

	if replicaPorts != nil {
		for i := 0; i < replicaNum; i++ {
			client.replicaPorts[i] = replicaPorts[i]
		}
	}

	client.connectedClients = make([]pb.ReplicaClient, replicaNum)
	for i := 0; i < replicaNum; i++ {
		conn, err := grpc.Dial(client.replicaPorts[i], grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			client.connectedClients[i] = pb.NewReplicaClient(conn)
		} else {
			client.connectedClients[i] = nil
		}
	}

	client.incrementCommandNumberChannel = make(chan int, 1)
	client.commandNumberReplyChannel = make(chan int, 1)

	return client
}

func (client *Client) Store(key string, value string) error {

	cid := (int64(client.clientId) << 54) + int64(atomic.AddUint32(&client.commandCount, 1))

	replyChannels := make([]chan *reply, replicaNum)
	for i := 0; i < replicaNum; i++ {
		replyChannels[i] = make(chan *reply, 1)
	}

	// push client commands to command buffers
	for i := 0; i < replicaNum; i++ {
		go client.TempMessengerRoutine(&pb.Message{Type: WRITE, Command: &pb.Command{Type: WRITE, CommandId: cid, ClientId: client.clientId, Key: key, Value: value}, CommandId: cid, ClientId: client.clientId, Key: key, Value: value}, replyChannels[i], i)
	}

	errCount := 0
	for i := 0; i < replicaNum; i++ {
		reply := <-replyChannels[i]
		if reply.err != nil {
			errCount++
		}
	}
	// NOTE: only return an error if all replica failed to handle
	if errCount >= replicaNum {
		return errors.New("all replica failed to handle")
	}
	return nil
}

func (client *Client) Read(key string) (string, error) {

	cid := (int64(client.clientId) << 54) + int64(atomic.AddUint32(&client.commandCount, 1))

	replyChannels := make([]chan *reply, replicaNum)
	for i := 0; i < replicaNum; i++ {
		replyChannels[i] = make(chan *reply, 1)
	}

	// push client commands to command buffers
	for i := 0; i < replicaNum; i++ {
		go client.TempMessengerRoutine(&pb.Message{Type: READ, Command: &pb.Command{Type: READ, CommandId: cid, ClientId: client.clientId, Key: key}}, replyChannels[i], i)
	}

	errCount := 0
	value := ""
	for i := 0; i < replicaNum; i++ {
		reply := <-replyChannels[i]
		if reply.err != nil {
			errCount++
		} else {
			value = reply.value
		}
	}
	// NOTE: only return an error if all replica failed to handle
	if errCount >= replicaNum {
		return "", errors.New("all replica failed to handle")
	}
	return value, nil
}

func (client *Client) TempMessengerRoutine(msg *pb.Message, replyChannel chan *reply, replicaSerial int) {
	// reset connection for each message
	// conn, err := grpc.Dial(client.replicaPorts[replicaSerial], grpc.WithTransportCredentials(insecure.NewCredentials()))

	if client.connectedClients[replicaSerial] == nil {
		replyChannel <- &reply{err: errors.New("connection unavailable")}
	} else {
		c := client.connectedClients[replicaSerial]
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		switch msg.Type {
		case WRITE:
			_, err := c.Write(ctx, msg)
			if err != nil {
				replyChannel <- &reply{err: err}
				cancel()
			} else {
				replyChannel <- &reply{err: nil}
			}
		case READ:
			r, err := c.Read(ctx, msg)
			if err != nil {
				replyChannel <- &reply{err: err}
				cancel()
			} else {
				// // fmt.Printf("key: %s, value: %s", msg.Command.Key, r.Content)
				replyChannel <- &reply{value: r.Content, err: nil}
			}
		default:
			replyChannel <- &reply{err: errors.New("unknown command type")}
		}
	}
}
