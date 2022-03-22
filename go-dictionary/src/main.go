package main

import (
	"encoding/hex"
	"fmt"

	"github.com/linxGnu/grocksdb"
)

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

	//lookup key
	// [(n >> 24) as u8, ((n >> 16) & 0xff) as u8, ((n >> 8) & 0xff) as u8, (n & 0xff) as u8]
	resp, _ := db.GetCF(grocksdb.NewDefaultReadOptions(), handles[4], []byte{0, 0, 0, 1})
	// resp, _ := db.GetCF(grocksdb.NewDefaultReadOptions(), handles[4], []byte{192, 9, 99, 88, 83, 78, 200, 210, 29, 1, 211, 75, 131, 110, 237, 71, 106, 28, 52, 63, 135, 36, 250, 33, 83, 220, 7, 37, 173, 121, 122, 144})
	dataString := hex.EncodeToString(resp.Data())
	fmt.Println(dataString)

	// header, _ := db.GetCF(grocksdb.NewDefaultReadOptions(), handles[5], resp.Data())
	// fmt.Println(string(header.Data()))

	body, _ := db.GetCF(grocksdb.NewDefaultReadOptions(), handles[6], resp.Data())
	fmt.Println(body.Data())
}
