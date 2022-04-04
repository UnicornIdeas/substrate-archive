package db

import (
	"context"
	"fmt"
	"go-dictionary/models"
	"log"
	"os"
	"sync"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresClient struct {
	Pool            *pgxpool.Pool
	WorkersChannels WorkersChannels
}

type WorkersChannels struct {
	EventsChannel          chan *models.Event
	EvmLogsChannel         chan *models.EvmLog
	EvmTransactionsChannel chan *models.EvmTransaction
	ExtrinsicsChannel      chan *models.Extrinsic
	SpecVersionsChannel    chan *models.SpecVersion
}

type PostgresConfig struct {
	Host string
	Port uint16
	Name string
	User string
	Pwd  string
}

func CreatePostgresPool() (PostgresClient, error) {
	pc := PostgresClient{}

	wc := WorkersChannels{}
	wc.EventsChannel = make(chan *models.Event, 10000000)
	wc.EvmLogsChannel = make(chan *models.EvmLog, 10000000)
	wc.EvmTransactionsChannel = make(chan *models.EvmTransaction, 10000000)
	wc.ExtrinsicsChannel = make(chan *models.Extrinsic, 10000000)
	wc.SpecVersionsChannel = make(chan *models.SpecVersion, 10000000)
	pc.WorkersChannels = wc

	err := pc.InitializePostgresDB()
	if err != nil {
		return PostgresClient{}, err
	}
	return pc, nil
}

func (pc *PostgresClient) InitializePostgresDB() error {
	user := os.Getenv("POSTGRES_USER")
	pwd := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	dbname := os.Getenv("POSTGRES_DB")
	pool_max_conns := os.Getenv("POSTGRES_CONN_POOL")

	connString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable&pool_max_conns=%s",
		user, pwd, host, port, dbname, pool_max_conns)

	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return err
	}
	p, err := pgxpool.ConnectConfig(context.Background(), cfg)
	if err != nil {
		return err
	}
	pc.Pool = p
	return nil
}

func (pc *PostgresClient) Close() {
	pc.Pool.Close()
}

func (pc *PostgresClient) EventsWorker(wg *sync.WaitGroup) {
	log.Println("[+] Starting EventsWorker!")
	defer wg.Done()
	maxBatch := 100000
	counter := 0
	insertItems := [][]interface{}{}
	for event := range pc.WorkersChannels.EventsChannel {
		insertItems = append(insertItems, []interface{}{event.Id, event.Module, event.Event, event.BlockHeight})
		counter++
		if counter == maxBatch {
			pc.Pool.CopyFrom(
				context.Background(),
				pgx.Identifier{"events"},
				[]string{"id", "module", "event", "block_height"},
				pgx.CopyFromRows(insertItems),
			)
			insertItems = nil
			counter = 0
		}
	}
	pc.Pool.CopyFrom(
		context.Background(),
		pgx.Identifier{"events"},
		[]string{"id", "module", "event", "block_height"},
		pgx.CopyFromRows(insertItems),
	)
	log.Println("[-] Exited EventsWorker...")
}

func (pc *PostgresClient) EvmLogsWorker(wg *sync.WaitGroup) {
	log.Println("[+] Started EvmLogsWorker!")
	defer wg.Done()
	maxBatch := 100000
	counter := 0
	insertItems := [][]interface{}{}
	for evmLog := range pc.WorkersChannels.EvmLogsChannel {
		insertItems = append(insertItems, []interface{}{evmLog.Id, evmLog.Address, evmLog.BlockHeight, evmLog.Topics0, evmLog.Topics1, evmLog.Topics2, evmLog.Topics3})
		counter++
		if counter == maxBatch {
			pc.Pool.CopyFrom(
				context.Background(),
				pgx.Identifier{"evm_logs"},
				[]string{"id", "address", "block_height", "topics0", "topics1", "topics2", "topics3"},
				pgx.CopyFromRows(insertItems),
			)
			insertItems = nil
			counter = 0
		}
	}
	pc.Pool.CopyFrom(
		context.Background(),
		pgx.Identifier{"evm_logs"},
		[]string{"id", "address", "block_height", "topics0", "topics1", "topics2", "topics3"},
		pgx.CopyFromRows(insertItems),
	)
	log.Println("[-] Exited EvmLogsWorker...")
}

func (pc *PostgresClient) EvmTransactionsWorker(wg *sync.WaitGroup) {
	log.Println("[+] Started EvmTransactionsWorker!")
	defer wg.Done()
	maxBatch := 100000
	counter := 0
	insertItems := [][]interface{}{}
	for evmTransaction := range pc.WorkersChannels.EvmTransactionsChannel {
		insertItems = append(insertItems, []interface{}{evmTransaction.Id, evmTransaction.TxHash, evmTransaction.From, evmTransaction.To, evmTransaction.Func, evmTransaction.BlockHeight, evmTransaction.Success})
		counter++
		if counter == maxBatch {
			pc.Pool.CopyFrom(
				context.Background(),
				pgx.Identifier{"evm_transactions"},
				[]string{"id", "tx_hash", "from", "to", "func", "block_height", "success"},
				pgx.CopyFromRows(insertItems),
			)
			insertItems = nil
			counter = 0
		}
	}
	_, err := pc.Pool.CopyFrom(
		context.Background(),
		pgx.Identifier{"evm_transactions"},
		[]string{"id", "tx_hash", "from", "to", "func", "block_height", "success"},
		pgx.CopyFromRows(insertItems),
	)
	if err != nil {
		log.Println("[ERR]", err, "- could not insert items with CopyFrom!")
	}
	log.Println("[-] Exited EvmTransactionsWorker...")
}

func (pc *PostgresClient) ExtrinsicsWorker(wg *sync.WaitGroup) {
	log.Println("[+] Started ExtrinsicsWorker!")
	defer wg.Done()
	maxBatch := 100000
	counter := 0
	insertItems := [][]interface{}{}
	for extrinsic := range pc.WorkersChannels.ExtrinsicsChannel {
		insertItems = append(insertItems, []interface{}{extrinsic.Id, extrinsic.TxHash, extrinsic.Module, extrinsic.Call, extrinsic.BlockHeight, extrinsic.Success, extrinsic.IsSigned})
		counter++
		if counter == maxBatch {
			pc.Pool.CopyFrom(
				context.Background(),
				pgx.Identifier{"extrinsics"},
				[]string{"id", "tx_hash", "module", "call", "block_height", "success", "is_signed"},
				pgx.CopyFromRows(insertItems),
			)
			insertItems = nil
			counter = 0
		}
	}
	pc.Pool.CopyFrom(
		context.Background(),
		pgx.Identifier{"extrinsics"},
		[]string{"id", "tx_hash", "module", "call", "block_height", "success", "is_signed"},
		pgx.CopyFromRows(insertItems),
	)
	log.Println("[-] Exited ExtrinsicsWorker...")
}

func (pc *PostgresClient) SpecVersionsWorker(wg *sync.WaitGroup) {
	log.Println("[+] Started SpecVersionWorker!")
	defer wg.Done()
	maxBatch := 100000
	counter := 0
	insertItems := [][]interface{}{}
	for specVersion := range pc.WorkersChannels.SpecVersionsChannel {
		insertItems = append(insertItems, []interface{}{specVersion.Id, specVersion.BlockHeight})
		counter++
		if counter == maxBatch {
			pc.Pool.CopyFrom(
				context.Background(),
				pgx.Identifier{"spec_versions"},
				[]string{"id", "block_height"},
				pgx.CopyFromRows(insertItems),
			)
			insertItems = nil
			counter = 0
		}
	}
	pc.Pool.CopyFrom(
		context.Background(),
		pgx.Identifier{"spec_versions"},
		[]string{"id", "block_height"},
		pgx.CopyFromRows(insertItems),
	)
	log.Println("[-] Exited SpecVersionsWorker...")
}

func (pc *PostgresClient) InsertByQuery(query string) error {
	_, err := pc.Pool.Exec(context.Background(), query)
	if err != nil {
		log.Println("[ERR]", err, "- could not insert to postgres by query!")
		return err
	}
	return nil
}
