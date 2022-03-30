package main

import (
	"encoding/hex"
	"fmt"
	"go-dictionary/internal"
	"go-dictionary/utils"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/itering/substrate-api-rpc/rpc"
	"github.com/joho/godotenv"
)

type EventsClient struct {
	conn          *websocket.Conn
	rocksdbClient internal.RockClient
}

const (
	eventQuery = `{"id":%d,"method":"state_getStorage","params":["0x26aa394eea5630e07c48ae0c9558cef780d41e5e16056765bc8461851072c9d7","%s"],"jsonrpc":"2.0"}`
	blockHash  = "0x490cd542b4a40ad743183c7d1088a4fe7b1edf21e50c850b86f29e389f31c5c1"
)

func main() {
	//LOAD env
	fmt.Println("* Loading env variables from .env...")
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Failed to load environment variables:", err)
		return
	}
	rpcEndpoint := os.Getenv("WS_RPC_ENDPOINT")
	rocksDbPath := os.Getenv("ROCKSDB_PATH")

	//LOAD ranges for spec versions
	fmt.Println("* Loading config info from files...")
	specVRanges, err := utils.GetSpecVersionsFromFile()
	if err != nil {
		fmt.Println("Failed to load configs from file")
		return
	}
	fmt.Println("Successfuly loaded configs from files")
	lastIndexedBlock := specVRanges[len(specVRanges)-1].Last

	//Connect to ws
	fmt.Println("* Connecting to ws rpc endpoint", rpcEndpoint)
	wsConn, _, err := websocket.DefaultDialer.Dial(rpcEndpoint, nil)
	if err != nil {
		fmt.Printf("Error connecting ws %d. Trying to reconnect...\n", err)
		return
	}
	fmt.Println("Connected to ws endpoint")

	//CONNECT to rocksdb
	fmt.Println("* Connecting to rocksdb at", rocksDbPath, "...")
	rdbClient, err := internal.OpenRocksdb(rocksDbPath)
	if err != nil {
		fmt.Println("Error opening rocksdb:", err)
		return
	}
	defer rdbClient.Close()
	fmt.Println("Rocksdb connected")

	evClient := EventsClient{
		conn:          wsConn,
		rocksdbClient: rdbClient,
	}

	var wg sync.WaitGroup
	lastIndexedBlock = 100000 //dbg
	t := time.Now()
	wg.Add(1)
	go evClient.readWs(&wg, lastIndexedBlock+1)

	fmt.Println("* Getting raw events...")
	for i := 0; i <= lastIndexedBlock; i++ {
		hash, err := evClient.getBlockHash(i)
		if err != nil {
			fmt.Println("Failed to get hash for block", i)
		}
		msg := fmt.Sprintf(eventQuery, i, hash)
		evClient.conn.WriteMessage(1, []byte(msg))
	}

	wg.Wait()
	fmt.Println(time.Now().Sub(t))
}

func (ev *EventsClient) readWs(wg *sync.WaitGroup, total int) {
	defer wg.Done()
	c := 0
	for {
		v := &rpc.JsonRpcResult{}
		err := ev.conn.ReadJSON(v)
		if err != nil {
			fmt.Println("Failed to get events for a block:", err)
			return
		}
		c++
		if c == total {
			fmt.Println("Finished getting events for", total, "blocks")
			break
		}
	}
}

func (ev *EventsClient) getBlockHash(height int) (string, error) {
	lk, err := ev.rocksdbClient.GetLookupKeyForBlockHeight(height)
	if err != nil {
		return "", err
	}
	hash := hex.EncodeToString(lk[4:])
	return hash, nil
}
