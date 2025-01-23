package taskmanager

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type IWorkerPool interface {
	RegisterHandler(taskType string, handler func(context.Context, Task) error)
	Enqueue(ctx context.Context, task Task) error
	Shutdown()
}

type WorkerPool struct {
	ctx        context.Context
	cancel     context.CancelFunc
	taskQueue  chan Task
	handlers   map[string]func(context.Context, Task) error
	wg         sync.WaitGroup
	numWorkers int
	shutdown   sync.Once
}

type MonitoringData struct {
	QueueLength   int
	Processed     uint64
	Errors        uint64
	ActiveWorkers int
}

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

func (p *WorkerPool) RegisterHandler(taskType string, handler func(context.Context, Task) error) {
	p.handlers[taskType] = handler
}

func (p *WorkerPool) Enqueue(ctx context.Context, task Task) error {
	if _, exists := p.handlers[task.TaskType()]; !exists {
		return fmt.Errorf("no handler registered for task type: %s", task.TaskType())
	}
	p.taskQueue <- task
	return nil
}

func (p *WorkerPool) Shutdown() {
	p.shutdown.Do(func() {
		p.cancel()
		p.wg.Wait()
		close(p.taskQueue)
	})
}

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
