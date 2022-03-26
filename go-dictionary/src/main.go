package main

import (
	"encoding/hex"
	"fmt"

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
	opts := grocksdb.NewDefaultOptions()
	opts.SetMaxOpenFiles(-1)
	db, handles, _ := grocksdb.OpenDbAsSecondaryColumnFamilies(
		opts,
		"/tmp/rocksdb-polkadot/chains/polkadot/db/full",
		"/tmp/rocksdb-secondary",
		[]string{"default", "col0", "col1", "col2", "col3", "col4", "col5", "col6", "col7", "col8", "col9", "col10", "col11", "col12"},
		[]*grocksdb.Options{opts, opts, opts, opts, opts, opts, opts, opts, opts, opts, opts, opts, opts, opts},
	)

	// resp, _ := db.GetCF(grocksdb.NewDefaultReadOptions(), handles[4], []byte{192, 9, 99, 88, 83, 78, 200, 210, 29, 1, 211, 75, 131, 110, 237, 71, 106, 28, 52, 63, 135, 36, 250, 33, 83, 220, 7, 37, 173, 121, 122, 144})

	// client := dbClient{
	// 	db,
	// 	handles,
	// }
	iter := db.NewIteratorCF(grocksdb.NewDefaultReadOptions(), handles[COL_STATE])
	// iter.SeekToFirst()

	for iter.SeekToFirst(); iter.Valid(); iter.Next() {
		key := iter.Value().Data()
		s := hex.EncodeToString(key)
		// fmt.Println(key)
		// if strings.Contains(s, "80d41e5e16056765bc8461851072c9d7") {
		fmt.Println(s)
		// }
	}

	// key, _ = client.GetLookupKeyForBlockHeight(i)
	// s = hex.EncodeToString(key)
	// fmt.Println(s)

	// // header, _ := db.GetCF(grocksdb.NewDefaultReadOptions(), handles[5], resp.Data())
	// // fmt.Println(string(header.Data()))

	// body, _ := db.GetCF(grocksdb.NewDefaultReadOptions(), handles[6], key)
	// fmt.Println(body.Data())
}

func (dbc *dbClient) GetLookupKeyForBlockHeight(blockHeight int) ([]byte, error) {
	blockKey := BlockHeightToKey(blockHeight)
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
