package db

import (
	"context"
	"fmt"
	"go-dictionary/models"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresClient struct {
	Pool            *pgxpool.Pool
	WorkersChannels WorkersChannels
}

type WorkersChannels struct {
	EventsChannel          chan models.Event
	EvmLogsChannel         chan models.EvmLog
	EvmTransactionsChannel chan models.EvmTransaction
	ExtrinsicsChannel      chan models.Extrinsic
	SpecVersionsChannel    chan models.SpecVersion
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
	wc.EventsChannel = make(chan models.Event, 10000000)
	wc.EvmLogsChannel = make(chan models.EvmLog, 10000000)
	wc.EvmTransactionsChannel = make(chan models.EvmTransaction, 10000000)
	wc.ExtrinsicsChannel = make(chan models.Extrinsic, 10000000)
	wc.SpecVersionsChannel = make(chan models.SpecVersion, 10000000)
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
	portString := os.Getenv("POSTGRES_PORT")
	port, err := strconv.Atoi(portString)
	if err != nil {
		return err
	}
	dbname := os.Getenv("POSTGRES_DB")

	connString := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		user, pwd, host, port, dbname)

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
	writing := false
	counter := 0
	exitLoop := false
	used := false
	query := `INSERT INTO events (id, module, event, block_height) VALUES `
	for !exitLoop {
		select {
		case event, ok := <-pc.WorkersChannels.EventsChannel:
			if ok {
				query += fmt.Sprintf(`('%s', '%s', '%s', '%d'), `, event.Id, event.Module, event.Event, event.BlockHeight)
				if counter < 700 {
					counter++
					used = true
				} else {
					writing = true
					pc.InsertByQuery(query[:len(query)-2])
					query = `INSERT INTO events (id, module, event, block_height) VALUES `
					counter = 0
					writing = false
				}
			}
		default:
			if counter != 0 && !writing {
				writing = true
				pc.InsertByQuery(query[:len(query)-2])
				query = `INSERT INTO events (id, module, event, block_height) VALUES `
				counter = 0
				writing = false
			} else if !writing && used {
				// exitLoop = true
			}
		}
	}
	log.Println("Exited EventsWorker...")
	wg.Done()
}

func (pc *PostgresClient) EvmLogsWorker(wg *sync.WaitGroup) {
	writing := false
	counter := 0
	exitLoop := false
	used := false
	query := `INSERT INTO evm_logs (id, address, block_height, topics0, topics1, topics2, topics3) VALUES `
	for !exitLoop {
		select {
		case evmLog, ok := <-pc.WorkersChannels.EvmLogsChannel:
			if ok {
				query += fmt.Sprintf(`('%s', '%s', '%d', '%s', '%s', '%s', '%s'), `, evmLog.Id, evmLog.Address, evmLog.BlockHeight, evmLog.Topics0, evmLog.Topics1, evmLog.Topics2, evmLog.Topics3)
				if counter < 700 {
					counter++
					used = true
				} else {
					writing = true
					pc.InsertByQuery(query[:len(query)-2])
					query = `INSERT INTO evm_logs (id, address, block_height, topics0, topics1, topics2, topics3) VALUES `
					counter = 0
					writing = false
				}
			}
		default:
			if counter != 0 && !writing {
				writing = true
				pc.InsertByQuery(query[:len(query)-2])
				query = `INSERT INTO evm_logs (id, address, block_height, topics0, topics1, topics2, topics3) VALUES `
				counter = 0
				writing = false
			} else if !writing && used {
				// exitLoop = true
			}
		}
	}
	log.Println("Exited EvmLogsWorker...")
	wg.Done()
}

func (pc *PostgresClient) EvmTransactionsWorker(wg *sync.WaitGroup) {
	writing := false
	counter := 0
	exitLoop := false
	used := false
	query := `INSERT INTO evm_transactions(id, tx_hash, "from", "to", func, block_height, success) VALUES `
	for !exitLoop {
		select {
		case evmTransaction, ok := <-pc.WorkersChannels.EvmTransactionsChannel:
			if ok {
				query += fmt.Sprintf(`('%s', '%s', '%s', '%s', '%s', '%d', '%t'), `, evmTransaction.Id, evmTransaction.TxHash, evmTransaction.From, evmTransaction.To, evmTransaction.Func, evmTransaction.BlockHeight, evmTransaction.Success)
				if counter < 700 {
					counter++
					used = true
				} else {
					writing = true
					pc.InsertByQuery(query[:len(query)-2])
					query = `INSERT INTO evm_transactions(id, tx_hash, "from", "to", func, block_height, success) VALUES `
					counter = 0
					writing = false
				}
			}
		default:
			if counter != 0 && !writing {
				writing = true
				pc.InsertByQuery(query[:len(query)-2])
				query = `INSERT INTO evm_transactions (id, tx_hash, "from", "to", func, block_height, success) VALUES `
				counter = 0
				writing = false
			} else if !writing && used {
				// exitLoop = true
			}
		}
	}
	log.Println("Exited EvmTransactionsWorker...")
	wg.Done()
}

func (pc *PostgresClient) ExtrinsicsWorker(wg *sync.WaitGroup) {
	writing := false
	counter := 0
	exitLoop := false
	used := false
	query := `INSERT INTO extrinsics (id, tx_hash, module, call, block_height, success, is_signed) VALUES `
	for !exitLoop {
		select {
		case extrinsic, ok := <-pc.WorkersChannels.ExtrinsicsChannel:
			if ok {
				query += fmt.Sprintf(`('%s', '%s', '%s', '%s', '%d', '%t', '%t'), `, extrinsic.Id, extrinsic.TxHash, extrinsic.Module, extrinsic.Call, extrinsic.BlockHeight, extrinsic.Success, extrinsic.IsSigned)
				if counter < 700 {
					counter++
					used = true
				} else {
					writing = true
					pc.InsertByQuery(query[:len(query)-2])
					query = `INSERT INTO extrinsics (id, tx_hash, module, call, block_height, success, is_signed) VALUES `
					counter = 0
					writing = false
				}
			}
		default:
			if counter != 0 && !writing {
				writing = true
				pc.InsertByQuery(query[:len(query)-2])
				query = `INSERT INTO extrinsics (id, tx_hash, module, call, block_height, success, is_signed) VALUES `
				counter = 0
				writing = false
			} else if !writing && used {
				// exitLoop = true
			}
		}
	}
	log.Println("Exited ExtrinsicsWorker...")
	wg.Done()
}

func (pc *PostgresClient) SpecVersionsWorker(wg *sync.WaitGroup) {
	writing := false
	counter := 0
	exitLoop := false
	used := false
	query := `INSERT INTO spec_versions (id, block_height) VALUES `
	for !exitLoop {
		select {
		case specVersion, ok := <-pc.WorkersChannels.SpecVersionsChannel:
			if ok {
				query += fmt.Sprintf(`('%s', '%d'), `, specVersion.Id, specVersion.BlockHeight)
				if counter < 9301 {
					counter++
					used = true
				} else {
					writing = true
					pc.InsertByQuery(query[:len(query)-2])
					query = `INSERT INTO spec_versions (id, block_height) VALUES `
					counter = 0
					writing = false
				}
			}
		default:
			if counter != 0 && !writing {
				writing = true
				pc.InsertByQuery(query[:len(query)-2])
				query = `INSERT INTO spec_versions (id, block_height) VALUES `
				counter = 0
				writing = false
			} else if !writing && used {
				// exitLoop = true
			}
		}
	}
	log.Println("Exited SpecVersionsWorker...")
	wg.Done()
}

func (pc *PostgresClient) InsertByQuery(query string) error {
	_, err := pc.Pool.Exec(context.Background(), query)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
