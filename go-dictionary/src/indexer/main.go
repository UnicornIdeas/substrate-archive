package main

import (
	"fmt"
	"go-dictionary/db"
	"go-dictionary/internal"
	"go-dictionary/utils"
	"log"
	"sync"
	"time"

	"github.com/itering/scale.go/types"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("[+] Loading .env variables!")
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Failed to load environment variables:", err)
		return
	}

	log.Println("[+] Initializing Postgres Database Pool")
	// Postgres database initialize
	postgresClient, err := db.CreatePostgresPool()
	if err != nil {
		log.Println("PostgresClient Error:", err)
	}

	//LOAD ranges for spec versions
	log.Println("[+] Loading config info from files...")
	specVRanges, err := utils.GetSpecVersionsFromFile()
	if err != nil {
		fmt.Println("Failed to load configs from file")
		return
	}

	log.Println("[+] Initializing Pool Workers for Header and Body processing")
	// Pool Workers routines for Header and Body
	jobQueueHeader := internal.NewJobQueueHeader(1, postgresClient.WorkersChannels.EvmLogsChannel)
	jobQueueHeader.Start()

	jobQueueBody := internal.NewJobQueueBody(10, specVRanges, postgresClient.WorkersChannels.ExtrinsicsChannel, postgresClient.WorkersChannels.EvmTransactionsChannel)
	jobQueueBody.Start()

	log.Println("[+] Initializing Rocksdb")
	rc, err := internal.OpenRocksdb("/mnt/hgfs/ArchivedRocksdb/chains/polkadot/db/full")
	if err != nil {
		log.Println(err)
	}

	headerDecoder := types.ScaleDecoder{}
	headerDecoder.ProcessAndUpdateData("Header")

	// Postgres Insert Workers
	var workersWG sync.WaitGroup
	workersWG.Add(1)
	go postgresClient.EvmLogsWorker(&workersWG)
	workersWG.Add(1)
	go postgresClient.EvmTransactionsWorker(&workersWG)
	workersWG.Add(1)
	go postgresClient.ExtrinsicsWorker(&workersWG)

	t := time.Now()
	rc.StartProcessing(jobQueueBody, jobQueueHeader)

	workersWG.Wait()
	log.Println("[INFO] All the processing took: ", time.Since(t))

	// Closing all remaining items
	postgresClient.Close()

	log.Println("[-] Exiting program...")
}
