package internal

import (
	"log"
	"sync"
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
			log.Println("Got job in body queue")
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
				job.Process()
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
func (job *BodyJob) Process() {

}
