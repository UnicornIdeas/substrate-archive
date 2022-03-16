package connection

import (
	"errors"
	"log"

	"github.com/gorilla/websocket"
	"github.com/itering/substrate-api-rpc/rpc"
)

type WsClient struct {
	*websocket.Conn
	ReceiversList chan *rpc.JsonRpcResult
}

func InitWSClient(endpoint string) (*WsClient, error) {
	wsClient := &WsClient{}
	var err error
	wsClient.Conn, _, err = websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		return &WsClient{}, errors.New("could not start websocket")
	}
	return wsClient, nil
}

func (c *WsClient) ReadWSMessages() {
	v := &rpc.JsonRpcResult{}
	for {
		err := c.ReadJSON(v)
		if err != nil {
			log.Println("read:", err)
			return
		}
		// c.receiversList[v.Id] <- v
	}
}
