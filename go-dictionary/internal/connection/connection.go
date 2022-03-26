package connection

import (
	"errors"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/itering/substrate-api-rpc/rpc"
)

type WsClient struct {
	wsPool           []*websocket.Conn
	receivedMessages chan *[]byte
	sender           chan *rpc.JsonRpcResult //chan of the function calling the ws
}

func InitWSClient(endpoint string, sender chan *rpc.JsonRpcResult) (*WsClient, error) {
	numSockets := 20
	messagesChanSize := 10000
	receivedMessages := make(chan *[]byte, messagesChanSize)

	wsClient := &WsClient{
		sender:           sender,
		receivedMessages: receivedMessages,
	}

	err := wsClient.connectPool(endpoint, numSockets)
	if err != nil {
		return &WsClient{}, errors.New(fmt.Sprintf("Could not start websocket: %v", err))
	}

	for i := 0; i < numSockets; i++ {
		go wsClient.workerGetMessage(i)
		go wsClient.readWSMessages(i)
	}

	return wsClient, nil
}

//connect the ws pool to the endpoint
func (c *WsClient) connectPool(endpoint string, numSockets int) error {
	for i := 0; i < numSockets; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
		if err != nil {
			return err
		}
		c.wsPool = append(c.wsPool, conn)
	}
	return nil
}

func (c *WsClient) workerGetMessage(workerId int) {
	for {
		msg := <-c.receivedMessages
		c.wsPool[workerId].WriteMessage(1, *msg)
	}
}

func (c *WsClient) readWSMessages(workerId int) {
	for {
		v := &rpc.JsonRpcResult{}
		err := c.wsPool[workerId].ReadJSON(v)
		if err != nil {
			log.Println(err)
		}
		c.sender <- v
	}
}

func (c *WsClient) SendMessage(message []byte) {
	c.receivedMessages <- &message
}
