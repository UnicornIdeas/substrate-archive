package internal

import (
	"encoding/hex"
	"fmt"
	"go-dictionary/models"
	"log"
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
	BlockHeight    int
	BlockHash      string
	BlockLookupKey []byte
	BlockBody      []map[string]interface{}
}

// WorkerBody - the worker threads that actually process the jobs for Body data
type WorkerBody struct {
	done             sync.WaitGroup
	readyPool        chan chan BodyJob
	assignedJobQueue chan BodyJob
	quit             chan bool
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
func NewJobQueueBody(maxWorkers int) *JobQueueBody {
	workersStopped := sync.WaitGroup{}
	readyPool := make(chan chan BodyJob, maxWorkers)
	workers := make([]*WorkerBody, maxWorkers, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		workers[i] = NewWorkerBody(readyPool, workersStopped)
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
func NewWorkerBody(readyPool chan chan BodyJob, done sync.WaitGroup) *WorkerBody {
	return &WorkerBody{
		done:             done,
		readyPool:        readyPool,
		assignedJobQueue: make(chan BodyJob),
		quit:             make(chan bool),
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
				job.ProcessBody()
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

// Processing function
func (job *BodyJob) ProcessBody() {
	// log.Println(job.BlockBody)
	job.ProcessExtrinsics()
}

func (job *BodyJob) ProcessExtrinsics() {
	extrinsicsResult := make([]models.Extrinsic, len(job.BlockBody))
	transactionsResult := []models.EvmTransaction{}
	transactionId := 0
	for i, ex := range job.BlockBody {
		extrinsicsResult[i].Id = strconv.Itoa(job.BlockHeight) + "-" + strconv.Itoa(i)
		extrinsicsResult[i].Module = strings.ToLower(ex["call_module"].(string))
		extrinsicsResult[i].Call = ex["call_module_function"].(string)
		extrinsicsResult[i].BlockHeight = job.BlockHeight
		extrinsicsResult[i].Success = true
		if txHash := job.BlockBody[i]["extrinsic_hash"]; txHash != nil {
			extrinsicsResult[i].TxHash = txHash.(string)
		}
		if signature := job.BlockBody[i]["signature"]; signature != nil {
			extrinsicsResult[i].IsSigned = true
		}
		if extrinsicsResult[i].Module == "utility" {
			if transactions := job.BlockBody[i]["params"].([]scalecodec.ExtrinsicParam); transactions != nil {
				tempTransactions, tid := job.ProcessUtilityTransaction(transactions, extrinsicsResult[i].TxHash, transactionId)
				transactionId = tid
				transactionsResult = append(transactionsResult, tempTransactions...)
			}
		} else if transactions := job.BlockBody[i]["params"].([]scalecodec.ExtrinsicParam); transactions != nil {
			for _, transaction := range transactions {
				if transaction.Name == "call" {
					if transaction.Value.(map[string]interface{})["call_module"] == "Balances" {
						tempTransaction := job.ProcessBalancesTransaction(transaction, extrinsicsResult[i].TxHash, transactionId)
						transactionId++
						transactionsResult = append(transactionsResult, tempTransaction)
					}
				}
			}
		}
	}
	fmt.Println("EXTRINSICS:")
	for _, e := range extrinsicsResult {
		fmt.Println(e)
	}
	fmt.Println("TRANSACTIONS:")
	for _, t := range transactionsResult {
		fmt.Println(t)
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
			log.Println(param.Value)
			tempTransaction.To = EncodeAddressId(param.Value.(string))
		}
	}
	tempTransaction.Func = transactionData["call_module"].(string)
	tempTransaction.BlockHeight = job.BlockHeight
	tempTransaction.Success = true

	return tempTransaction
}

func EncodeAddressId(txHash string) string {
	checksumString := SS58PRE + polkaAddressPrefix + txHash
	checksumBytes, _ := hex.DecodeString(checksumString)
	checksum := blake2b.Sum512(checksumBytes)
	checksumEnd := hex.EncodeToString(checksum[:2])
	finalString := polkaAddressPrefix + txHash + checksumEnd
	finalBytes, _ := hex.DecodeString(finalString)
	return base58.Encode(finalBytes)
}

func (job *BodyJob) ProcessUtilityTransaction(transactions []scalecodec.ExtrinsicParam, txHash string, transactionId int) ([]models.EvmTransaction, int) {
	tempTransactions := []models.EvmTransaction{}
	log.Println("utility for txHash:", txHash)
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
	// log.Println(tempTransactions)
	return tempTransactions, tid
}
