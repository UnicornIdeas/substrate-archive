package internal

import (
	"go-dictionary/models"
	"strconv"
	"sync"
)

// Job - structure for job processing
type HeaderJob struct {
	PoolChannel    *JobQueueHeader
	BlockHeight    int
	BlockHash      string
	BlockLookupKey []byte
	BlockHeader    []byte
	DecodedHeader  interface{}
	// BlockHeader interface{}
}

// Worker - the worker threads that actually process the jobs
type WorkerHeader struct {
	evmLogsChannel   chan models.EvmLog
	done             sync.WaitGroup
	readyPool        chan chan HeaderJob
	assignedJobQueue chan HeaderJob
	quit             chan bool
}

// JobQueue - a queue for enqueueing jobs to be processed
type JobQueueHeader struct {
	internalQueue     chan HeaderJob
	readyPool         chan chan HeaderJob
	workers           []*WorkerHeader
	dispatcherStopped sync.WaitGroup
	workersStopped    sync.WaitGroup
	quit              chan bool
}

// NewJobQueue - creates a new job queue
func NewJobQueueHeader(maxWorkers int, evmLogsChannel chan models.EvmLog) *JobQueueHeader {
	workersStopped := sync.WaitGroup{}
	readyPool := make(chan chan HeaderJob, maxWorkers)
	workers := make([]*WorkerHeader, maxWorkers, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		workers[i] = NewWorkerHeader(evmLogsChannel, readyPool, workersStopped)
	}
	return &JobQueueHeader{
		internalQueue:     make(chan HeaderJob),
		readyPool:         readyPool,
		workers:           workers,
		dispatcherStopped: sync.WaitGroup{},
		workersStopped:    workersStopped,
		quit:              make(chan bool),
	}
}

// Start - starts the worker routines and dispatcher routine
func (q *JobQueueHeader) Start() {
	for i := 0; i < len(q.workers); i++ {
		q.workers[i].Start()
	}
	go q.dispatch()
}

// Stop - stops the workers and sispatcher routine
func (q *JobQueueHeader) Stop() {
	q.quit <- true
	q.dispatcherStopped.Wait()
}

func (q *JobQueueHeader) dispatch() {
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

// Submit - adds a new job to be processed
func (q *JobQueueHeader) Submit(job HeaderJob) {
	q.internalQueue <- job
}

// NewWorker - creates a new worker
func NewWorkerHeader(evmLogsChannel chan models.EvmLog, readyPool chan chan HeaderJob, done sync.WaitGroup) *WorkerHeader {
	return &WorkerHeader{
		evmLogsChannel:   evmLogsChannel,
		done:             done,
		readyPool:        readyPool,
		assignedJobQueue: make(chan HeaderJob),
		quit:             make(chan bool),
	}
}

// Start - begins the job processing loop for the worker
func (w *WorkerHeader) Start() {
	go func() {
		w.done.Add(1)
		for {
			w.readyPool <- w.assignedJobQueue // check the job queue in
			select {
			case job := <-w.assignedJobQueue: // see if anything has been assigned to the queue
				job.ProcessHeader(w.evmLogsChannel)
			case <-w.quit:
				w.done.Done()
				return
			}
		}
	}()
}

// Stop - stops the worker
func (w *WorkerHeader) Stop() {
	w.quit <- true
}

// Processing function
func (job *HeaderJob) ProcessHeader(evmLogsChannel chan models.EvmLog) {
	logs := job.DecodedHeader.(map[string]interface{})["digest"].(map[string]interface{})["logs"].([]interface{})
	for i := range logs {
		evmLog := models.EvmLog{}
		evmLog.Id = strconv.Itoa(job.BlockHeight) + "-" + strconv.Itoa(i)
		evmLog.BlockHeight = job.BlockHeight
		evmLogsChannel <- evmLog
	}
}
