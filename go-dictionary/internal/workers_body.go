package internal

import (
	"encoding/hex"
	"go-dictionary/models"
	"strconv"
	"strings"
	"sync"

	scalecodec "github.com/itering/scale.go"
	"github.com/itering/scale.go/types"
	"github.com/mr-tron/base58"
	"golang.org/x/crypto/blake2b"
)

// BodyJob - structure for job processing
type BodyJob struct {
	PoolChannel    *JobQueueBody
	BlockHeight    int
	BlockHash      string
	BlockLookupKey []byte
	BlockBody      []byte
	DecodedBody    []map[string]interface{}
	// BlockBody      []map[string]interface{}
}

// WorkerBody - the worker threads that actually process the jobs for Body data
type WorkerBody struct {
	extrinsicsChannel      chan models.Extrinsic
	evmTransactionsChannel chan models.EvmTransaction
	done                   sync.WaitGroup
	readyPool              chan chan BodyJob
	assignedJobQueue       chan BodyJob
	quit                   chan bool
}

// JobQueueBody - a queue for enqueueing jobs to be processed
type JobQueueBody struct {
	internalQueue     chan BodyJob
	readyPool         chan chan BodyJob
	workers           []*WorkerBody
	dispatcherStopped sync.WaitGroup
	workersStopped    sync.WaitGroup
	quit              chan bool
}

// NewJobQueueBody - creates a new job queue
func NewJobQueueBody(maxWorkers int, extrinsicsChannel chan models.Extrinsic, evmTransactionsChannel chan models.EvmTransaction) *JobQueueBody {
	workersStopped := sync.WaitGroup{}
	readyPool := make(chan chan BodyJob, maxWorkers)
	workers := make([]*WorkerBody, maxWorkers, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		workers[i] = NewWorkerBody(extrinsicsChannel, evmTransactionsChannel, readyPool, workersStopped)
	}
	return &JobQueueBody{
		internalQueue:     make(chan BodyJob),
		readyPool:         readyPool,
		workers:           workers,
		dispatcherStopped: sync.WaitGroup{},
		workersStopped:    workersStopped,
		quit:              make(chan bool),
	}
}

// Start - starts the worker routines and dispatcher routine
func (q *JobQueueBody) Start() {
	for i := 0; i < len(q.workers); i++ {
		q.workers[i].Start()
	}
	go q.dispatch()
}

// Stop - stops the workers and dispatcher routine
func (q *JobQueueBody) Stop() {
	q.quit <- true
	q.dispatcherStopped.Wait()
}

func (q *JobQueueBody) dispatch() {
	q.dispatcherStopped.Add(1)
	for {
		select {
		case job := <-q.internalQueue: // We got something in on our queue
			workerChannel := <-q.readyPool // Check out an available worker
			workerChannel <- job           // Send the request to the channel
		case <-q.quit:
			for i := 0; i < len(q.workers); i++ {
				q.workers[i].Stop()
			}
			q.workersStopped.Wait()
			q.dispatcherStopped.Done()
			return
		}
	}
}

// Submit - adds a new BodyJob to be processed
func (q *JobQueueBody) Submit(job BodyJob) {
	q.internalQueue <- job
}

// NewWorkerBody - creates a new workerBody
func NewWorkerBody(extrinsicsChannel chan models.Extrinsic, evmTransactionsChannel chan models.EvmTransaction, readyPool chan chan BodyJob, done sync.WaitGroup) *WorkerBody {
	return &WorkerBody{
		extrinsicsChannel:      extrinsicsChannel,
		evmTransactionsChannel: evmTransactionsChannel,
		done:                   done,
		readyPool:              readyPool,
		assignedJobQueue:       make(chan BodyJob),
		quit:                   make(chan bool),
	}
}

// Start - begins the job processing loop for the workerBody
func (w *WorkerBody) Start() {
	go func() {
		w.done.Add(1)
		for {
			w.readyPool <- w.assignedJobQueue // check the job queue in
			select {
			case job := <-w.assignedJobQueue: // see if anything has been assigned to the queue
				job.ProcessBody(w.extrinsicsChannel, w.evmTransactionsChannel)
			case <-w.quit:
				w.done.Done()
				return
			}
		}
	}()
}

// Stop - stops the workerBody
func (w *WorkerBody) Stop() {
	w.quit <- true
}

func EncodeAddressId(txHash string) string {
	checksumBytes, _ := hex.DecodeString(SS58PRE + polkaAddressPrefix + txHash)
	checksum := blake2b.Sum512(checksumBytes)
	checksumEnd := hex.EncodeToString(checksum[:2])
	finalBytes, _ := hex.DecodeString(polkaAddressPrefix + txHash + checksumEnd)
	return base58.Encode(finalBytes)
}

func (job *BodyJob) ProcessBody(extrinsicsChannel chan models.Extrinsic, evmTransactionsChannel chan models.EvmTransaction) {
	transactionId := 0
	for i, ex := range job.DecodedBody {
		extrinsicsResult := models.Extrinsic{}
		extrinsicsResult.Id = strconv.Itoa(job.BlockHeight) + "-" + strconv.Itoa(i)
		extrinsicsResult.Module = strings.ToLower(ex["call_module"].(string))
		extrinsicsResult.Call = ex["call_module_function"].(string)
		extrinsicsResult.BlockHeight = job.BlockHeight
		extrinsicsResult.Success = true

		if txHash := ex["extrinsic_hash"]; txHash != nil {
			extrinsicsResult.TxHash = txHash.(string)
		}
		if signature := ex["signature"]; signature != nil {
			extrinsicsResult.IsSigned = true
		}
		extrinsicsChannel <- extrinsicsResult
		if extrinsicsResult.Module == "utility" {
			if transactions := ex["params"].([]scalecodec.ExtrinsicParam); transactions != nil {
				transactions, tid := job.ProcessUtilityTransaction(transactions, extrinsicsResult.TxHash, transactionId)
				transactionId = tid
				for _, transaction := range transactions {
					evmTransactionsChannel <- transaction
				}
			}
		} else if transactions := ex["params"].([]scalecodec.ExtrinsicParam); transactions != nil {
			for _, transaction := range transactions {
				if transaction.Name == "call" {
					if transaction.Value.(map[string]interface{})["call_module"] == "Balances" {
						transaction := job.ProcessBalancesTransaction(transaction, extrinsicsResult.TxHash, transactionId)
						transactionId++
						evmTransactionsChannel <- transaction
					}
				}
			}
		}

	}
}

func (job *BodyJob) ProcessBalancesTransaction(transaction scalecodec.ExtrinsicParam, txHash string, transactionId int) models.EvmTransaction {
	tempTransaction := models.EvmTransaction{}
	transactionData := transaction.Value.(map[string]interface{})
	tempTransaction.Id = strconv.Itoa(job.BlockHeight) + "-" + strconv.Itoa(transactionId)
	tempTransaction.TxHash = txHash
	params := transactionData["params"].([]types.ExtrinsicParam)
	for _, param := range params {
		if param.Name == "source" {
			tempTransaction.From = EncodeAddressId(param.Value.(string))
		}
		if param.Name == "dest" {
			tempTransaction.To = EncodeAddressId(param.Value.(string))
		}
	}
	tempTransaction.Func = transactionData["call_module"].(string)
	tempTransaction.BlockHeight = job.BlockHeight
	tempTransaction.Success = true

	return tempTransaction
}

func (job *BodyJob) ProcessUtilityTransaction(transactions []scalecodec.ExtrinsicParam, txHash string, transactionId int) ([]models.EvmTransaction, int) {
	tempTransactions := []models.EvmTransaction{}
	tid := transactionId
	for _, transaction := range transactions {
		if transaction.Name == "calls" {
			transactionValue := transaction.Value.([]interface{})
			for _, tv := range transactionValue {
				if params := tv.(map[string]interface{})["params"].([]types.ExtrinsicParam); params != nil {
					for _, param := range params {
						if param.Name == "call" {
							tempTransaction := models.EvmTransaction{}
							transactionData := param.Value.(map[string]interface{})
							tempTransaction.Id = strconv.Itoa(job.BlockHeight) + "-" + strconv.Itoa(tid)
							tempTransaction.TxHash = txHash
							transactionParams := transactionData["params"].([]types.ExtrinsicParam)
							for _, tparam := range transactionParams {
								callData := tparam.Value.(map[string]interface{})
								callParams := callData["params"].([]types.ExtrinsicParam)
								for _, callParam := range callParams {
									if callParam.Name == "source" {
										tempTransaction.From = EncodeAddressId(callParam.Value.(string))
									}
									if callParam.Name == "dest" {
										tempTransaction.To = EncodeAddressId(callParam.Value.(string))
									}
								}
								tempTransaction.Func = callData["call_module"].(string)
							}

							tempTransaction.BlockHeight = job.BlockHeight
							tempTransaction.Success = true
							tid++
							tempTransactions = append(tempTransactions, tempTransaction)

						}
					}
				}
			}
		}
	}
	return tempTransactions, tid
}
