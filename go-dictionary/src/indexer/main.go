package main

import (
	"go-dictionary/internal"
	"log"
	"sync"
)

func main() {
	// wsClient, err := connection.InitWSClient("wss://polkadot.api.onfinality.io/public-ws")
	// if err != nil {
	// 	log.Println(err)
	// }
	// log.Println(wsClient)
	// wsClient2, err := connection.InitWSClient("ws://localhost:9944")
	// if err != nil {
	// 	log.Println(err)
	// }
	// // meta, _ := ioutil.ReadFile("/mnt/hgfs/metas/0")
	// var wg sync.WaitGroup
	// wg.Add(1)
	// wg.Add(1)
	// wg.Add(1)
	// go wsClient.ReadWSMessages(&wg, wsClient2)
	// go wsClient2.ReadWSMessages(&wg, wsClient2)
	// // go wsClient.GetEvents(&wg, string(meta))
	// go wsClient.SendMessages(&wg)
	// wg.Wait()

	// websocket.SetEndpoint("ws://localhost:9944")

	// v := &rpc.JsonRpcResult{}
	// websocket.SendWsRequest(nil, v, rpc.ChainGetBlockHash(1, 210000))
	// hash, _ := v.ToString()
	// log.Println(hash)

	jobQueueHeader := internal.NewJobQueueHeader(10)
	jobQueueHeader.Start()

	jobQueueBody := internal.NewJobQueueBody(10)
	jobQueueBody.Start()

	rc, err := internal.OpenRocksdb("/mnt/hgfs/ArchivedRocksdb/chains/polkadot/db/full")
	if err != nil {
		log.Println(err)
	}
	// log.Println(rc)
	rc.ProcessLookupKey(jobQueueBody, jobQueueHeader)

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
	// // header, _ := db.GetCF(grocksdb.NewDefaultReadOptions(), handles[5], resp.Data())
	// // fmt.Println(string(header.Data()))

	// body, _ := db.GetCF(grocksdb.NewDefaultReadOptions(), handles[6], key)
	// fmt.Println(body.Data())

	// m := scale.MetadataDecoder{}
	// m.Init(utiles.HexToBytes(``)) // Todo: rpc state_getMetadata
	// _ = m.Process()

	// e := scale.EventsDecoder{}
	// option := types.ScaleDecoderOption{Metadata: &m.Metadata}

	// eventRaw := "0x180000000000000080e36a090000000002000000010000000000000000000000000002000000020000000e022ac9219ace40f5846ed675dded4e25a1997da7eabdea2f78597a71d6f38031487089d481874e06bd4026807fd0464f5c7a1c691c21237ed9b912ac0443a2bc2200ca9a3b000000000000000000000000000002000000150600b4c4040000000000000000000000000000020000000e04b4f7f03bebc56ebe96bc52ea5ed3159d45a0ce3a8d7f082983c33ef133274747002d31010000000000000000000000000000020000000000401b5f1300000000000000"
	// e.Init(types.ScaleBytes{Data: utiles.HexToBytes(eventRaw)}, &option)
	// e.Process()
	// b, _ := json.Marshal(e.Value)
	// fmt.Println(string(b))

}
