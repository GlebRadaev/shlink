// Package taskmanager defines a worker pool system for processing tasks asynchronously.
// The pool manages task handlers and the distribution of tasks to worker goroutines.
package taskmanager

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// IWorkerPool is an interface for managing a worker pool that processes tasks.
type IWorkerPool interface {
	// RegisterHandler registers a handler for a specific task type.
	// The handler is a function that takes a context and a task, and returns an error if something goes wrong.
	RegisterHandler(taskType string, handler func(context.Context, Task) error)

	// Enqueue adds a task to the worker pool's task queue.
	// The task must have a registered handler, and the pool will attempt to process the task.
	Enqueue(ctx context.Context, task Task) error

	// Shutdown gracefully shuts down the worker pool, stopping all active workers and closing the task queue.
	Shutdown()
}

// WorkerPool manages a pool of workers that process tasks concurrently.
type WorkerPool struct {
	ctx        context.Context                              // The context that controls the lifetime of the pool.
	cancel     context.CancelFunc                           // The cancel function to signal shutdown.
	taskQueue  chan Task                                    // Channel holding the tasks to be processed.
	handlers   map[string]func(context.Context, Task) error // Registered task handlers.
	wg         sync.WaitGroup                               // Wait group to track workers and ensure graceful shutdown.
	numWorkers int                                          // The number of workers in the pool.
	shutdown   sync.Once                                    // Ensures that shutdown occurs once.
}

// MonitoringData holds statistics about the worker pool's state, such as task queue length,
// number of processed tasks, number of errors, and active workers.
type MonitoringData struct {
	QueueLength   int    // The current number of tasks in the queue.
	Processed     uint64 // Total number of tasks that have been processed.
	Errors        uint64 // Total number of errors encountered during task processing.
	ActiveWorkers int    // The number of active workers currently processing tasks.
}

// NewWorkerPool creates a new WorkerPool instance with the specified parameters:
// - ctx: The context used to control the lifetime of the pool.
// - queueSize: The size of the task queue.
// - numWorkers: The number of workers in the pool.
func NewWorkerPool(ctx context.Context, queueSize, numWorkers int) *WorkerPool {
	ctx, cancel := context.WithCancel(ctx)
	pool := &WorkerPool{
		ctx:        ctx,
		cancel:     cancel,
		taskQueue:  make(chan Task, queueSize),
		handlers:   make(map[string]func(context.Context, Task) error),
		numWorkers: numWorkers,
	}
	for i := 0; i < numWorkers; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}
	return pool
}

// RegisterHandler registers a handler function for a specific task type.
func (p *WorkerPool) RegisterHandler(taskType string, handler func(context.Context, Task) error) {
	p.handlers[taskType] = handler
}

// Enqueue adds a task to the task queue for processing by the workers.
func (p *WorkerPool) Enqueue(ctx context.Context, task Task) error {
	if _, exists := p.handlers[task.TaskType()]; !exists {
		return fmt.Errorf("no handler registered for task type: %s", task.TaskType())
	}
	p.taskQueue <- task
	return nil
}

// Shutdown gracefully shuts down the worker pool by signaling the workers to stop.
func (p *WorkerPool) Shutdown() {
	p.shutdown.Do(func() {
		p.cancel()
		p.wg.Wait()
		close(p.taskQueue)
	})
}

// worker is a goroutine that listens for tasks in the pool's task queue and processes them using the appropriate handler.
func (p *WorkerPool) worker(workerID int) {
	defer p.wg.Done()
	log.Printf("Worker %d started", workerID)

	for {
		select {
		case <-p.ctx.Done():
			log.Printf("Worker %d received shutdown signal", workerID)
			return
		case task, ok := <-p.taskQueue:
			if !ok {
				log.Printf("Worker %d: task queue is closed, stopping", workerID)
				return
			}
			handler := p.handlers[task.TaskType()]
			if err := handler(p.ctx, task); err != nil {
				log.Printf("Error processing task of type %s: %v", task.TaskType(), err)
			}
		}
	}
}
