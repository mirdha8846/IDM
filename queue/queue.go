package queue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// TaskStatus represents the status of a task
type TaskStatus int

const (
	TaskPending TaskStatus = iota
	TaskRunning
	TaskCompleted
	TaskFailed
	TaskRetrying
)

// Task represents a download task
type Task struct {
	ID          string
	URL         string
	Size        int64
	Status      TaskStatus
	Attempts    int
	MaxAttempts int
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	Error       error
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	Task   *Task
	Error  error
	Output interface{}
}

// Worker represents a worker that processes tasks
type Worker struct {
	ID       int
	TaskChan chan *Task
	Quit     chan bool
	Wg       *sync.WaitGroup
}

// Queue represents a production-ready queue system
type Queue struct {
	tasks        chan *Task
	results      chan *TaskResult
	workers      []*Worker
	numWorkers   int
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	taskMap      map[string]*Task
	taskMapMutex sync.RWMutex

	// Queue statistics
	totalTasks     int64
	completedTasks int64
	failedTasks    int64
	statsMutex     sync.RWMutex
}

// NewQueue creates a new queue with specified number of workers
func NewQueue(numWorkers int, bufferSize int) *Queue {
	ctx, cancel := context.WithCancel(context.Background())

	q := &Queue{
		tasks:      make(chan *Task, bufferSize),
		results:    make(chan *TaskResult, bufferSize),
		numWorkers: numWorkers,
		ctx:        ctx,
		cancel:     cancel,
		taskMap:    make(map[string]*Task),
		workers:    make([]*Worker, numWorkers),
	}

	return q
}

// Start initializes and starts all workers
func (q *Queue) Start() {
	log.Printf("🚀 Starting queue with %d workers", q.numWorkers)

	for i := 0; i < q.numWorkers; i++ {
		worker := &Worker{
			ID:       i + 1,
			TaskChan: make(chan *Task),
			Quit:     make(chan bool),
			Wg:       &q.wg,
		}

		q.workers[i] = worker
		q.wg.Add(1)
		go q.startWorker(worker)
	}

	// Start task distributor
	go q.distributor()

	// Start result processor
	go q.resultProcessor()
}

// AddTask adds a new task to the queue
func (q *Queue) AddTask(task *Task) error {
	if task.ID == "" {
		task.ID = fmt.Sprintf("task_%d_%d", time.Now().Unix(), len(q.taskMap))
	}

	task.Status = TaskPending
	task.CreatedAt = time.Now()
	task.MaxAttempts = 3 // Default retry attempts

	q.taskMapMutex.Lock()
	q.taskMap[task.ID] = task
	q.taskMapMutex.Unlock()

	q.statsMutex.Lock()
	q.totalTasks++
	q.statsMutex.Unlock()

	select {
	case q.tasks <- task:
		log.Printf("📝 Task %s added to queue", task.ID)
		return nil
	case <-q.ctx.Done():
		return fmt.Errorf("queue is shutting down")
	default:
		return fmt.Errorf("queue is full")
	}
}

// distributor distributes tasks to available workers
func (q *Queue) distributor() {
	workerIndex := 0

	for {
		select {
		case task := <-q.tasks:
			// Round-robin distribution
			worker := q.workers[workerIndex]
			workerIndex = (workerIndex + 1) % q.numWorkers

			select {
			case worker.TaskChan <- task:
				log.Printf("📤 Task %s assigned to worker %d", task.ID, worker.ID)
			case <-q.ctx.Done():
				return
			}

		case <-q.ctx.Done():
			log.Println("🛑 Task distributor shutting down")
			return
		}
	}
}

// startWorker starts a worker goroutine
func (q *Queue) startWorker(worker *Worker) {
	defer worker.Wg.Done()
	defer log.Printf("🔚 Worker %d stopped", worker.ID)

	log.Printf("👷 Worker %d started", worker.ID)

	for {
		select {
		case task := <-worker.TaskChan:
			q.processTask(worker, task)

		case <-worker.Quit:
			return

		case <-q.ctx.Done():
			return
		}
	}
}

// processTask processes a single task
func (q *Queue) processTask(worker *Worker, task *Task) {
	log.Printf("🔄 Worker %d processing task %s", worker.ID, task.ID)

	// Update task status
	now := time.Now()
	task.Status = TaskRunning
	task.StartedAt = &now
	task.Attempts++

	// Simulate task processing (replace with actual download logic)
	result := q.executeTask(task)

	// Send result
	select {
	case q.results <- result:
	case <-q.ctx.Done():
		return
	}
}

// executeTask executes the actual task (calls download manager)
func (q *Queue) executeTask(task *Task) *TaskResult {
	// Import download manager - you'll need to add this import
	// import "github.com/mirdha8846/IDM.git/manager"

	log.Printf("🔽 Starting download for %s", task.URL)

	// Call the actual download function from manager
	// result := manager.Download(task.Size, task.URL)

	// For now, simulate the work (replace with actual download call)
	time.Sleep(time.Second * 2) // Simulate download time

	now := time.Now()
	task.Status = TaskCompleted
	task.CompletedAt = &now

	return &TaskResult{
		Task:   task,
		Error:  nil,
		Output: fmt.Sprintf("Downloaded %s successfully", task.URL),
	}
}

// resultProcessor processes task results
func (q *Queue) resultProcessor() {
	for {
		select {
		case result := <-q.results:
			q.handleResult(result)

		case <-q.ctx.Done():
			log.Println("🛑 Result processor shutting down")
			return
		}
	}
}

// handleResult handles task completion results
func (q *Queue) handleResult(result *TaskResult) {
	q.statsMutex.Lock()
	defer q.statsMutex.Unlock()

	if result.Error != nil {
		q.failedTasks++
		log.Printf("❌ Task %s failed: %v", result.Task.ID, result.Error)

		// Retry logic
		if result.Task.Attempts < result.Task.MaxAttempts {
			result.Task.Status = TaskRetrying
			log.Printf("🔄 Retrying task %s (attempt %d/%d)",
				result.Task.ID, result.Task.Attempts+1, result.Task.MaxAttempts)

			// Add back to queue after a delay
			go func() {
				time.Sleep(time.Second * 2) // Retry delay
				q.AddTask(result.Task)
			}()
		}
	} else {
		q.completedTasks++
		log.Printf("✅ Task %s completed successfully", result.Task.ID)
	}
}

// GetStats returns queue statistics
func (q *Queue) GetStats() (total, completed, failed int64) {
	q.statsMutex.RLock()
	defer q.statsMutex.RUnlock()

	return q.totalTasks, q.completedTasks, q.failedTasks
}

// GetTaskStatus returns the status of a specific task
func (q *Queue) GetTaskStatus(taskID string) (*Task, bool) {
	q.taskMapMutex.RLock()
	defer q.taskMapMutex.RUnlock()

	task, exists := q.taskMap[taskID]
	return task, exists
}

// Stop gracefully shuts down the queue
func (q *Queue) Stop() {
	log.Println("🛑 Shutting down queue...")

	// Cancel context to signal shutdown
	q.cancel()

	// Close task channel
	close(q.tasks)

	// Stop all workers
	for _, worker := range q.workers {
		close(worker.Quit)
	}

	// Wait for all workers to finish
	q.wg.Wait()

	// Close results channel
	close(q.results)

	log.Println("✅ Queue shutdown complete")
}

// WaitForCompletion waits for all tasks to complete
func (q *Queue) WaitForCompletion() {
	for {
		total, completed, failed := q.GetStats()
		if completed+failed >= total && total > 0 {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
}
