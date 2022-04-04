package main

import (
	"fmt"
	"go-dictionary/db"
	"go-dictionary/internal"
	"go-dictionary/utils"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/itering/scale.go/source"
	"github.com/itering/scale.go/types"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("[+] Loading .env variables!")
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("[ERR] Failed to load environment variables:", err)
		return
	}
	rocksDbPath := os.Getenv("ROCKSDB_PATH")
	log.Println("[+] Initializing Postgres Database Pool")
	// Postgres database initialize
	postgresClient, err := db.CreatePostgresPool()
	if err != nil {
		log.Println("[ERR]", err, "- could not initialize postgres!")
	}
	defer postgresClient.Pool.Close()
	//LOAD ranges for spec versions
	log.Println("[+] Loading config info from files...")
	specVRanges, err := utils.GetSpecVersionsFromFile()
	if err != nil {
		log.Println("[ERR] Failed to load configs from file!")
		return
	}

	log.Println("[+] Initializing Pool Workers for Header and Body processing")
	// Pool Workers routines for Header and Body
	jobQueueHeader := internal.NewJobQueueHeader(1, postgresClient.WorkersChannels.EvmLogsChannel)
	jobQueueHeader.Start()

	jobQueueBody := internal.NewJobQueueBody(10, specVRanges, postgresClient.WorkersChannels.ExtrinsicsChannel, postgresClient.WorkersChannels.EvmTransactionsChannel)
	jobQueueBody.Start()

	log.Println("[+] Initializing Rocksdb")
	rc, err := internal.OpenRocksdb(rocksDbPath)
	if err != nil {
		log.Println("[ERR]", err, "- could not open Rocksdb!")
	}
	//Register decoder custom types
	log.Println("[+] Registering decoder custom types...")
	c, err := ioutil.ReadFile(fmt.Sprintf("%s.json", "./network/polkadot"))
	if err != nil {
		log.Println("[ERR] Failed to register types for network Polkadot:", err)
		return
	}
	types.RegCustomTypes(source.LoadTypeRegistry(c))

	// Postgres Insert Workers
	var workersWG sync.WaitGroup
	workersWG.Add(1)
	go postgresClient.EvmLogsWorker(&workersWG)
	workersWG.Add(1)
	go postgresClient.EvmTransactionsWorker(&workersWG)
	workersWG.Add(1)
	go postgresClient.ExtrinsicsWorker(&workersWG)

	t := time.Now()
	rc.StartProcessing(jobQueueBody, jobQueueHeader, specVRanges[len(specVRanges)-1].Last)

	workersWG.Wait()

	log.Println("[INFO] All the processing took:", time.Since(t))
	log.Println("[-] Exiting program...")
}
