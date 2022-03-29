package internal

import (
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/linxGnu/grocksdb"
)

const polkaAddressPrefix = "00"
const SS58PRE = "53533538505245"

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

type RockClient struct {
	db            *grocksdb.DB
	columnHandles []*grocksdb.ColumnFamilyHandle
	opts          *grocksdb.Options
	ro            *grocksdb.ReadOptions
}

func OpenRocksdb(path string) (RockClient, error) {
	opts := grocksdb.NewDefaultOptions()
	opts.SetMaxOpenFiles(-1)
	ro := grocksdb.NewDefaultReadOptions()

	cf, err := grocksdb.ListColumnFamilies(opts, path)
	if err != nil {
		return RockClient{}, err
	}
	cfOpts := []*grocksdb.Options{}
	for range cf {
		cfOpts = append(cfOpts, opts)
	}
	db, handles, err := grocksdb.OpenDbAsSecondaryColumnFamilies(
		opts,
		path,
		"/tmp/secondary",
		cf,
		cfOpts,
		// []string{"default", "col0", "col1", "col2", "col3", "col4", "col5", "col6", "col7", "col8", "col9", "col10", "col11", "col12"},
		// []*grocksdb.Options{opts, opts, opts, opts, opts, opts, opts, opts, opts, opts, opts, opts, opts, opts},
	)
	if err != nil {
		return RockClient{}, err
	}
	rc := RockClient{
		db,
		handles,
		opts,
		ro,
	}
	return rc, nil
}

func (rc *RockClient) GetLookupKeyForBlockHeight(blockHeight int) ([]byte, error) {
	blockKey := BlockHeightToKey(blockHeight)
	ro := grocksdb.NewDefaultReadOptions()
	response, err := rc.db.GetCF(ro, rc.columnHandles[COL_KEY_LOOKUP], blockKey)
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

func (rc *RockClient) GetHeaderForBlockLookupKey(key []byte) ([]byte, error) {
	header, err := rc.db.GetCF(rc.ro, rc.columnHandles[COL_HEADER], key)
	return header.Data(), err
}

func (rc *RockClient) GetBodyForBlockLookupKey(key []byte) ([]byte, error) {
	body, err := rc.db.GetCF(rc.ro, rc.columnHandles[COL_BODY], key)
	return body.Data(), err
}

func (rc *RockClient) GetLastBlockSynced() (int, error) {
	lastElement, err := rc.db.GetCF(rc.ro, rc.columnHandles[COL_META], []byte("final"))
	if err != nil {
		return 0, err
	}
	hexIndex := hex.EncodeToString(lastElement.Data()[0:4])
	maxBlockHeight, err := strconv.ParseInt(hexIndex, 16, 64)
	if err != nil {
		return 0, err
	}
	return int(maxBlockHeight), nil
}

func (rc *RockClient) ProcessLookupKey(bq *JobQueueBody, hq *JobQueueHeader, b chan *BodyJob, h chan *HeaderJob) {
	maxBlockHeight, err := rc.GetLastBlockSynced()
	if err != nil {
		log.Println(err)
	}
	fmt.Println("MAX BLOCK HEIGHT:", maxBlockHeight)

	t := time.Now()
	// testBlockHeight := maxBlockHeight
	// testBlockHeight := 198073
	// testBlockHeight := 29259
	// testBlockHeight := 287353
	// testBlockHeight := 200866
	testBlockHeight := 20000
	var wg sync.WaitGroup
	for i := 0; i < testBlockHeight; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			rc.TestFunction(i, bq, hq, b, h)
		}(i)
	}
	wg.Wait()

	log.Println("Done with", testBlockHeight, "after", time.Now().Sub(t))
}

func (rc *RockClient) GetHeaderRaw(wg *sync.WaitGroup, headerJob *HeaderJob, h chan *HeaderJob, key []byte) {
	defer wg.Done()
	header, err := rc.GetHeaderForBlockLookupKey(key)
	if err != nil {
		log.Println(err)
	}
	headerJob.BlockHeader = header
	h <- headerJob
}

func (rc *RockClient) GetBodyRaw(wg *sync.WaitGroup, bodyJob *BodyJob, b chan *BodyJob, key []byte) {
	defer wg.Done()
	body, err := rc.GetBodyForBlockLookupKey(key)
	if err != nil {
		log.Println(err)
	}
	bodyJob.BlockBody = body
	b <- bodyJob
}

func (rc *RockClient) TestFunction(blockHeight int, bq *JobQueueBody, hq *JobQueueHeader, b chan *BodyJob, h chan *HeaderJob) {
	bodyJob := BodyJob{}
	headerJob := HeaderJob{}

	bodyJob.BlockHeight = blockHeight
	headerJob.BlockHeight = blockHeight

	bodyJob.PoolChannel = bq
	headerJob.PoolChannel = hq

	key, err := rc.GetLookupKeyForBlockHeight(blockHeight)
	if err != nil {
		log.Println(err)
	}
	bodyJob.BlockLookupKey = key
	headerJob.BlockLookupKey = key

	hash := hex.EncodeToString(key[4:])
	bodyJob.BlockHash = hash
	headerJob.BlockHash = hash

	var rawWg sync.WaitGroup
	rawWg.Add(1)
	go rc.GetHeaderRaw(&rawWg, &headerJob, h, key)
	rawWg.Add(1)
	go rc.GetBodyRaw(&rawWg, &bodyJob, b, key)

	rawWg.Wait()

	// header, err := rc.GetHeaderForBlockLookupKey(key)
	// if err != nil {
	// 	log.Println(err)
	// }
	// headerDecoder := types.ScaleDecoder{}
	// headerDecoder.Init(types.ScaleBytes{Data: job.BlockHeader.([]byte)}, nil)
	// // decodedHeader := headerDecoder.ProcessAndUpdateData(`Struct<Struct<Vec>,CompactU32,H256,H256,Compact,BlockNumber,H256>`)
	// decodedHeader := headerDecoder.ProcessAndUpdateData("Header")
	// log.Println(decodedHeader)
	// headerJob.BlockHeader = header

	// hq.Submit(headerJob)

	// body, err := rc.GetBodyForBlockLookupKey(key)
	// if err != nil {
	// 	log.Println(err)
	// }

	// bodyDecoder := types.ScaleDecoder{}
	// bodyDecoder.Init(types.ScaleBytes{Data: body}, nil)
	// decodedBody := bodyDecoder.ProcessAndUpdateData("Vec<Bytes>")
	// // log.Println(decodedBody)
	// bodyList := decodedBody.([]interface{})
	// extrinsics := []string{}
	// for _, bodyL := range bodyList {
	// 	extrinsics = append(extrinsics, bodyL.(string))
	// }
	// specV := 0
	// metaString, _ := ioutil.ReadFile("./meta_files/" + strconv.Itoa(specV))
	// rawMeta := metadata.RuntimeRaw{Spec: specV, Raw: string(metaString)}
	// instant := metadata.Process(&rawMeta)

	// decodedExtrinsics, err := substrate.DecodeExtrinsic(extrinsics, instant, specV)
	// if err != nil {
	// 	log.Println(err)
	// }
	// bodyJob.BlockBody = body

	// bq.Submit(bodyJob)
}

func (rc *RockClient) ProcessHash(key []byte) {

}
