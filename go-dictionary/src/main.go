package main

import (
	process "go-dictionary/internal"
	"go-dictionary/internal/connection"
	"log"
)

// "github.com/itering/substrate-api-rpc/websocket"

// func main() {
// 	api, _ := gsrpc.NewSubstrateAPI("wss://polkadot.api.onfinality.io/public-ws")
// 	hash, _ := api.RPC.Chain.GetBlockHash(9429341)

// 	meta, _ := api.RPC.State.GetMetadata(hash)

// 	fmt.Println(meta.AsMetadataV14.Pallets)
// 	// fmt.Println(meta.AsMetadataV14.)

// 	key, _ := types.CreateStorageKey(meta, "System", "Events", nil, nil)

// 	fmt.Println(string(key))

// 	rawEvents, _ := api.RPC.State.GetStorageRaw(key, hash)

// 	fmt.Println(rawEvents)

// 	// block, _ := api.RPC.Chain.GetBlock(hash)

// 	// v, _ := api.RPC.State.GetRuntimeVersion(hash)

// 	// fmt.Println(v.SpecName)
// 	// fmt.Println(v.SpecVersion)

// }

func main() {
	endpoint := "wss://polkadot.api.onfinality.io/public-ws"
	wsClient, err := connection.InitWSClient(endpoint)
	if err != nil {
		log.Println(err)
	}
	go wsClient.ReadWSMessages()

	process.GetBlockHashes(wsClient, 100000)

	// v := &rpc.JsonRpcResult{}
	// websocket.SendWsRequest(nil, v, rpc.ChainGetBlockHash(1, 9429341))

	// hash, _ := v.ToString()

	// v2 := &rpc.JsonRpcResult{}
	// websocket.SendWsRequest(nil, v2, rpc.ChainGetBlock(1, hash))

	// meta := &rpc.JsonRpcResult{}
	// err := websocket.SendWsRequest(nil, meta, rpc.StateGetMetadata(1, hash))
	// fmt.Println(err)
	// // fmt.Println(meta)

	// metaString, _ := meta.ToString()
	// rawMeta := metadata.RuntimeRaw{Spec: 14, Raw: metaString}
	// instant := metadata.Process(&rawMeta)
	// fmt.Println(instant)
}
