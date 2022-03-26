package connection

import (
	"errors"
	"fmt"
	"log"
	"sync"

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

func (c *WsClient) ReadWSMessages(wg *sync.WaitGroup, c2 *WsClient) {
	v := &rpc.JsonRpcResult{}
	// count := 1
	log.Println("started")
	// t := time.Now()
	for {
		err := c.ReadJSON(v)
		log.Println(v)
		if v.Id == 1 {
			wg.Add(1)
			go c2.GetRange(v.Result.(string))
			wg.Done()
		} else {
			fmt.Println(v.Result)
		}
		// fmt.Println(count)
		// count++
		// if count == 1500000 {
		// 	log.Println(count, time.Now().Sub(t))
		// 	wg.Done()
		// }
		if err != nil {
			log.Println(err)
			return
		}

		// c.receiversList[v.Id] <- v
	}
}

func (c *WsClient) SendMessages(wg *sync.WaitGroup) {
	log.Println("sending")
	for i := 0; i < 1095689; i++ {
		err := c.Conn.WriteMessage(1, rpc.ChainGetBlockHash(1, i))
		if err != nil {
			log.Println(err)
		}
	}
	// wg.Done()
}

func (c *WsClient) GetRange(hash string) {
	// log.Println("asd")
	c.Conn.WriteMessage(1, rpc.ChainGetRuntimeVersion(2, hash))
}

func (c *WsClient) GetEvents(wg *sync.WaitGroup, meta string) {
	// c.Conn.WriteMessage(1, rpc.ChainGetBlockHash(1, 100))
	// hash := "c0096358534ec8d21d01d34b836eed476a1c343f8724fa2153dc0725ad797a90"
	// key := "26aa394eea5630e07c48ae0c9558cef780d41e5e16056765bc8461851072c9d7"
	// c.Conn.WriteMessage(1, rpc.StateGetStorage(1, key, hash))
	// wg.Done()
	// c.Conn.WriteMessage(1, rpc.StateGetStorage(1, key))
	for i := 0; i < 9595689; i++ {
		c.Conn.WriteMessage(1, rpc.ChainGetBlockHash(1, i))
	}
}
