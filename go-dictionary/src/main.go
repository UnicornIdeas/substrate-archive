package main

import (
	"go-dictionary/internal/connection"
	"io/ioutil"
	"log"
	"sync"

	"github.com/linxGnu/grocksdb"
)

const (
	/// Metadata about chain
	COL_META = iota + 1
	COL_STATE
	COL_STATE_META
	/// maps hashes -> lookup keys and numbers to canon hashes
	COL_KEY_LOOKUP
	/// Part of Block
	COL_HEADER
	COL_BODY
	COL_JUSTIFICATION
	/// Stores the changes tries for querying changed storage of a block
	COL_CHANGES_TRIE
	COL_AUX
	/// Off Chain workers local storage
	COL_OFFCHAIN
	COL_CACHE
	COL_TRANSACTION
)

type dbClient struct {
	db            *grocksdb.DB
	columnHandles []*grocksdb.ColumnFamilyHandle
}

func main() {
	wsClient, err := connection.InitWSClient("ws://localhost:9944")
	if err != nil {
		log.Println(err)
	}
	meta, _ := ioutil.ReadFile("/mnt/hgfs/metas/0")
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Add(1)
	go wsClient.ReadWSMessages(&wg)
	go wsClient.GetEvents(&wg, string(meta))
	// go wsClient.SendMessages(&wg)
	wg.Wait()

	// websocket.SetEndpoint("ws://localhost:9944")

	// v := &rpc.JsonRpcResult{}
	// websocket.SendWsRequest(nil, v, rpc.ChainGetBlockHash(1, 210000))
	// hash, _ := v.ToString()
	// log.Println(hash)

	// opts := grocksdb.NewDefaultOptions()
	// opts.SetMaxOpenFiles(-1)
	// db, handles, _ := grocksdb.OpenDbAsSecondaryColumnFamilies(
	// 	opts,
	// 	"/mnt/hgfs/ArchivedRocksdb/chains/polkadot/db/full",
	// 	"/mnt/hgfs/ArchivedRocksdb/rocksdb-secondary/",
	// 	[]string{"default", "col0", "col1", "col2", "col3", "col4", "col5", "col6", "col7", "col8", "col9", "col10", "col11", "col12"},
	// 	[]*grocksdb.Options{opts, opts, opts, opts, opts, opts, opts, opts, opts, opts, opts, opts, opts, opts},
	// )
	// log.Println(db, handles)
	// resp, _ := db.GetCF(grocksdb.NewDefaultReadOptions(), handles[4], []byte{192, 9, 99, 88, 83, 78, 200, 210, 29, 1, 211, 75, 131, 110, 237, 71, 106, 28, 52, 63, 135, 36, 250, 33, 83, 220, 7, 37, 173, 121, 122, 144})

	// client := dbClient{
	// 	db,
	// 	handles,
	// }

	// key, _ := client.GetLookupKeyForBlockHeight(1463)
	// key, _ := client.GetLookupKeyForBlockHeight(1)
	// s := hex.EncodeToString(key)
	// fmt.Println(len(s))

	// fuckme := "00001aaba87ad8a51efa86d19e5ce6076346d1a999bf2d2d769336e4facb403a"
	// xxx, _ := hex.DecodeString(fuckme)
	// log.Println(xxx)
	// blockMeta, err := db.GetCF(grocksdb.NewDefaultReadOptions(), handles[COL_JUSTIFICATION], key)
	// fmt.Println(blockMeta.Data())
	// fmt.Println(err)

	// header, _ := db.GetCF(grocksdb.NewDefaultReadOptions(), handles[5], resp.Data())
	// fmt.Println(string(header.Data()))

	// body, _ := db.GetCF(grocksdb.NewDefaultReadOptions(), handles[6], key)
	// fmt.Println(body.Data())

	// cheva := "26aa394eea5630e07c48ae0c9558cef780d41e5e16056765bc8461851072c9d7" + s
	// cheva := s + "26aa394eea5630e07c48ae0c9558cef780d41e5e16056765bc8461851072c9d7"

	// testStorage := "200000000000000080e36a09000000000200000001000000000000000000000000000200000002000000000000ca9a3b0000000002000000030000000406647a9b6cc579aafd3b0b1a04ae88f2dc5724b89a8fc45042679e87665e17720000e067901d040100000000000000000000000300000013010000030000000e0600038b170000000000000000000000000000030000000304dc9974cdb3cebfac4d31333c30865ff66c35c1bf898df5c5dd2924d3280e7201c0c0e20500000000000000000000000000000300000000004032084e00000000000000"
	// store, _ := hex.DecodeString(testStorage)
	// log.Println(string(store))
	// pos := "000005b7"
	// testHash := "ad176c97e085102c55a6080c6a3797f9162e6c86953e30038b6c4d134b8de35b"
	// log.Println(testHash)
	// // cheva := "26aa394eea5630e07c48ae0c9558cef780d41e5e16056765bc8461851072c9d7"
	// cheva := "00001aaba87ad8a51efa86d19e5ce6076346d1a999bf2d2d769336e4facb403a"
	// cheva := testHash + "26aa394eea5630e07c48ae0c9558cef780d41e5e16056765bc8461851072c9d7"
	// cheva := "26aa394eea5630e07c48ae0c9558cef780d41e5e16056765bc8461851072c9d7" + testHash
	// fmt.Println(cheva)
	// byteCheva, err := hex.DecodeString(cheva)
	// if err != nil {
	// 	log.Println(err)
	// }
	// // log.Println(byteCheva)
	// ro := grocksdb.NewDefaultReadOptions()
	// storage11, _ := db.GetCF(ro, handles[1], byteCheva)
	// // log.Println(string(storage.Data()))
	// fmt.Println("1", storage11)
	// storage12, _ := db.GetCF(ro, handles[2], byteCheva)
	// // log.Println(string(storage.Data()))
	// fmt.Println("2", storage12)
	// storage, _ := db.GetCF(ro, handles[8], byteCheva)
	// // log.Println(string(storage.Data()))
	// fmt.Println(storage)
	// storage2, _ := db.GetCF(ro, handles[9], byteCheva)
	// // log.Println(string(storage2.Data()))
	// fmt.Println(storage2)
	// storage3, _ := db.GetCF(ro, handles[10], byteCheva)
	// // log.Println(string(storage3.Data()))
	// fmt.Println(storage3)
	// storage4, _ := db.GetCF(ro, handles[11], byteCheva)
	// // log.Println(string(storage4.Data()))
	// fmt.Println(storage4)
	// storage5, _ := db.GetCF(ro, handles[12], byteCheva)
	// // log.Println(string(storage5.Data()))
	// fmt.Println(storage5)

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

func (dbc *dbClient) GetLookupKeyForBlockHeight(blockHeight int) ([]byte, error) {
	blockKey := BlockHeightToKey(blockHeight)
	// fmt.Println(hex.EncodeToString(blockKey))
	ro := grocksdb.NewDefaultReadOptions()
	response, err := dbc.db.GetCF(ro, dbc.columnHandles[COL_KEY_LOOKUP], blockKey)
	if err != nil {
		return []byte{}, err
	}
	return response.Data(), nil
}

func BlockHeightToKey(blockHeight int) []byte {
	return []byte{
		byte(blockHeight >> 24),
		byte((blockHeight >> 16) & 0xff),
		byte((blockHeight >> 8) & 0xff),
		byte(blockHeight & 0xff),
	}
}
