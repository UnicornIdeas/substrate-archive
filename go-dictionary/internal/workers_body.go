package internal

import (
	"encoding/hex"
	"go-dictionary/models"
	"go-dictionary/utils"
	"log"
	"strconv"
	"strings"
	"sync"

	scalecodec "github.com/itering/scale.go"
	"github.com/itering/scale.go/types"
	"github.com/itering/substrate-api-rpc"
	"github.com/mr-tron/base58"
	"golang.org/x/crypto/blake2b"
)

// BodyJob - structure for job processing
type BodyJob struct {
	BlockHeight    int
	BlockHash      string
	BlockLookupKey []byte
	BlockBody      []byte
}

// WorkerBody - the worker threads that actually process the jobs for Body data
type WorkerBody struct {
	specVersionList        utils.SpecVersionRangeList
	extrinsicsChannel      chan *models.Extrinsic
	evmTransactionsChannel chan *models.EvmTransaction
	done                   sync.WaitGroup
	readyPool              chan chan BodyJob
	assignedJobQueue       chan BodyJob
	quit                   chan bool
}

// JobQueueBody - a queue for enqueueing jobs to be processed
type JobQueueBody struct {
	internalQueue          chan BodyJob
	readyPool              chan chan BodyJob
	workers                []*WorkerBody
	workersStopped         sync.WaitGroup
	extrinsicsChannel      chan *models.Extrinsic
	evmTransactionsChannel chan *models.EvmTransaction
}

// NewJobQueueBody - creates a new job queue
func NewJobQueueBody(maxWorkers int, specVersionList utils.SpecVersionRangeList, extrinsicsChannel chan *models.Extrinsic, evmTransactionsChannel chan *models.EvmTransaction) *JobQueueBody {
	workersStopped := sync.WaitGroup{}
	readyPool := make(chan chan BodyJob, maxWorkers)
	workers := make([]*WorkerBody, maxWorkers, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		workers[i] = NewWorkerBody(specVersionList, extrinsicsChannel, evmTransactionsChannel, readyPool, workersStopped)
	}
	return &JobQueueBody{
		internalQueue:          make(chan BodyJob),
		readyPool:              readyPool,
		workers:                workers,
		workersStopped:         workersStopped,
		extrinsicsChannel:      extrinsicsChannel,
		evmTransactionsChannel: evmTransactionsChannel,
	}
}

// Start - starts the worker routines and dispatcher routine
func (q *JobQueueBody) Start() {
	for i := 0; i < len(q.workers); i++ {
		q.workers[i].Start()
	}
	go q.dispatch()
}

func (q *JobQueueBody) dispatch() {
	for job := range q.internalQueue {
		workerChannel := <-q.readyPool // Check out an available worker
		workerChannel <- job           // Send the request to the channel
	}
	for i := 0; i < len(q.workers); i++ {
		q.workers[i].Stop()
	}
	q.workersStopped.Wait()
	close(q.extrinsicsChannel)
	close(q.evmTransactionsChannel)
	log.Println("[-] Closing Body Pool...")
}

// Submit - adds a new BodyJob to be processed
func (q *JobQueueBody) Submit(job *BodyJob) {
	q.internalQueue <- *job
}

// NewWorkerBody - creates a new workerBody
func NewWorkerBody(specVersionList utils.SpecVersionRangeList, extrinsicsChannel chan *models.Extrinsic, evmTransactionsChannel chan *models.EvmTransaction, readyPool chan chan BodyJob, done sync.WaitGroup) *WorkerBody {
	return &WorkerBody{
		specVersionList:        specVersionList,
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
				job.ProcessBody(w.extrinsicsChannel, w.evmTransactionsChannel, w.specVersionList)
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

func (job *BodyJob) ProcessBody(extrinsicsChannel chan *models.Extrinsic, evmTransactionsChannel chan *models.EvmTransaction, specVersionList utils.SpecVersionRangeList) {
	bodyDecoder := types.ScaleDecoder{}
	bodyDecoder.Init(types.ScaleBytes{Data: job.BlockBody}, nil)
	decodedBody := bodyDecoder.ProcessAndUpdateData("Vec<Bytes>")
	if bodyList, ok := decodedBody.([]interface{}); ok {
		extrinsics := []string{}
		for _, bodyL := range bodyList {
			if bl, ok := bodyL.(string); ok {
				extrinsics = append(extrinsics, bl)

			}
		}
		specV, instant := specVersionList.GetBlockSpecVersionAndInstant(job.BlockHeight)

		decodedExtrinsics, err := substrate.DecodeExtrinsic(extrinsics, instant, specV)
		if err != nil {
			log.Println("[ERR]", err, job.BlockHeight, "- could not decode extrinsic!")
		}

		transactionId := 0
		for i, ex := range decodedExtrinsics {
			extrinsicsResult := models.Extrinsic{}
			extrinsicsResult.Id = strconv.Itoa(job.BlockHeight) + "-" + strconv.Itoa(i)
			if callModule, ok := ex["call_module"].(string); ok {
				extrinsicsResult.Module = strings.ToLower(callModule)
			}
			if callModuleFunction, ok := ex["call_module_function"].(string); ok {
				extrinsicsResult.Call = callModuleFunction
			}
			extrinsicsResult.BlockHeight = job.BlockHeight
			extrinsicsResult.Success = true

			if txHash, ok := ex["extrinsic_hash"].(string); ok {
				extrinsicsResult.TxHash = txHash
			}
			if _, ok := ex["signature"]; ok {
				extrinsicsResult.IsSigned = true
			}
			extrinsicsChannel <- &extrinsicsResult
			if extrinsicsResult.Module == "utility" {
				if transactions, ok := ex["params"].([]scalecodec.ExtrinsicParam); ok {
					tempTransactions, tid := job.ProcessUtilityTransaction(transactions, extrinsicsResult.TxHash, transactionId)
					transactionId += tid
					for i := range tempTransactions {
						evmTransactionsChannel <- &tempTransactions[i]
					}
				}
			} else if transactions, ok := ex["params"].([]scalecodec.ExtrinsicParam); ok {
				for _, transaction := range transactions {
					if transaction.Name == "call" {
						if value, ok := transaction.Value.(map[string]interface{}); ok {
							if value["call_module"] == "Balances" {
								transaction := job.ProcessBalancesTransaction(transaction, extrinsicsResult.TxHash, transactionId)
								transactionId++
								evmTransactionsChannel <- &transaction
							}
						}
					}
				}
			}
		}
	}
}

func (job *BodyJob) ProcessBalancesTransaction(transaction scalecodec.ExtrinsicParam, txHash string, transactionId int) models.EvmTransaction {
	tempTransaction := models.EvmTransaction{}
	if transactionData, ok := transaction.Value.(map[string]interface{}); ok {
		tempTransaction.Id = strconv.Itoa(job.BlockHeight) + "-" + strconv.Itoa(transactionId)
		tempTransaction.TxHash = txHash
		if params, ok := transactionData["params"].([]types.ExtrinsicParam); ok {
			for _, param := range params {
				if param.Name == "source" {
					if source, ok := param.Value.(string); ok {
						tempTransaction.From = EncodeAddressId(source)
					}
				}
				if param.Name == "dest" {
					if dest, ok := param.Value.(string); ok {
						tempTransaction.To = EncodeAddressId(dest)
					}
				}
			}
			if callModule, ok := transactionData["call_module"].(string); ok {
				tempTransaction.Func = callModule
			}
			tempTransaction.BlockHeight = job.BlockHeight
			tempTransaction.Success = true
		}
	}

	return tempTransaction
}

func (job *BodyJob) ProcessUtilityTransaction(transactions []scalecodec.ExtrinsicParam, txHash string, transactionId int) ([]models.EvmTransaction, int) {
	tempTransactions := []models.EvmTransaction{}
	tid := transactionId
	for _, transaction := range transactions {
		if transaction.Name == "calls" {
			if transactionValue, ok := transaction.Value.([]interface{}); ok {
				for _, tv := range transactionValue {
					if params, ok := tv.(map[string]interface{})["params"].([]types.ExtrinsicParam); ok {
						for _, param := range params {
							if param.Name == "call" {
								tempTransaction := models.EvmTransaction{}
								if transactionData, ok := param.Value.(map[string]interface{}); ok {
									tempTransaction.Id = strconv.Itoa(job.BlockHeight) + "-" + strconv.Itoa(tid)
									tempTransaction.TxHash = txHash
									if transactionParams, ok := transactionData["params"].([]types.ExtrinsicParam); ok {
										for _, tparam := range transactionParams {
											if callData, ok := tparam.Value.(map[string]interface{}); ok {
												callParams := callData["params"].([]types.ExtrinsicParam)
												for _, callParam := range callParams {
													if callParam.Name == "source" {
														if source, ok := callParam.Value.(string); ok {
															tempTransaction.From = EncodeAddressId(source)
														}
													}
													if callParam.Name == "dest" {
														if dest, ok := callParam.Value.(string); ok {
															tempTransaction.To = EncodeAddressId(dest)
														}
													}
												}
												if CallModule, ok := callData["call_module"].(string); ok {
													tempTransaction.Func = CallModule
												}
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
				}
			}
		}
	}
	return tempTransactions, tid
}
