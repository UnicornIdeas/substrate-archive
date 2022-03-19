package main

import (
	"go-dictionary/internal/downloader"
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
	// endpoint := "wss://polkadot.api.onfinality.io/public-ws"
	// wsClient, err := connection.InitWSClient(endpoint)
	// if err != nil {
	// 	log.Println(err)
	// }
	// go wsClient.ReadWSMessages()

	// internal.GetBlockHashes(wsClient, 100000)

	downServer := downloader.InitDownServer("https://dot-rocksdb.polkashots.io/snapshot", "polkadot-9493950.RocksDb.tar.lz4")
	// err := downServer.DownloadSnapshot()
	// if err != nil {
	// 	log.Println(err)
	// }

	// err := downServer.DecompressSnapshot()
	// if err != nil {
	// 	log.Println(err)
	// }

	err := downServer.UntarSnapshot()
	if err != nil {
		log.Println(err)
	}

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

// func DecodeEvents() {
// 	m := scalecodec.MetadataDecoder{}
// 	m.Init(utiles.HexToBytes(kusamaV14))
// 	_ = m.Process()
// 	c, err := ioutil.ReadFile(fmt.Sprintf("%s.json", "network/crab"))
// 	if err != nil {
// 		panic(err)
// 	}
// 	types.RegCustomTypes(source.LoadTypeRegistry(c))
// 	e := scalecodec.EventsDecoder{}
// 	option := types.ScaleDecoderOption{Metadata: &m.Metadata}

// 	eventRaw := "0x180000000000000080e36a090000000002000000010000000000000000000000000002000000020000000e022ac9219ace40f5846ed675dded4e25a1997da7eabdea2f78597a71d6f38031487089d481874e06bd4026807fd0464f5c7a1c691c21237ed9b912ac0443a2bc2200ca9a3b000000000000000000000000000002000000150600b4c4040000000000000000000000000000020000000e04b4f7f03bebc56ebe96bc52ea5ed3159d45a0ce3a8d7f082983c33ef133274747002d31010000000000000000000000000000020000000000401b5f1300000000000000"
// 	e.Init(types.ScaleBytes{Data: utiles.HexToBytes(eventRaw)}, &option)
// 	e.Process()
// 	b, _ := json.Marshal(e.Value)
// 	fmt.Println(string(b))
// }
