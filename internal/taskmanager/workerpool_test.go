package taskmanager

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

type DummyTask struct {
	Type string
}

func (t *DummyTask) TaskType() string {
	return t.Type
}

func TestWorkerPool_RegisterHandler(t *testing.T) {
	tests := []struct {
		name         string
		taskType     string
		handler      func(ctx context.Context, task Task) error
		expectExists bool
	}{
		{
			name:     "Valid handler registration",
			taskType: "test_task",
			handler: func(ctx context.Context, task Task) error {
				return nil
			},
			expectExists: true,
		},
		{
			name:         "No handler registered",
			taskType:     "non_existent_task",
			handler:      nil,
			expectExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			pool := NewWorkerPool(ctx, 10, 3)
			defer pool.Shutdown()

			if tt.handler != nil {
				pool.RegisterHandler(tt.taskType, tt.handler)
			}

			_, exists := pool.handlers[tt.taskType]
			if exists != tt.expectExists {
				t.Fatalf("Expected handler existence: %v, got: %v", tt.expectExists, exists)
			}
		})
	}
}

func TestWorkerPool_Enqueue(t *testing.T) {
	tests := []struct {
		name        string
		taskType    string
		register    bool
		expectError bool
	}{
		{
			name:        "Task with registered handler",
			taskType:    "test_task",
			register:    true,
			expectError: false,
		},
		{
			name:        "Task without registered handler",
			taskType:    "unregistered_task",
			register:    false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			pool := NewWorkerPool(ctx, 10, 3)
			defer pool.Shutdown()

			if tt.register {
				pool.RegisterHandler(tt.taskType, func(ctx context.Context, task Task) error {
					return nil
				})
			}

			task := &DummyTask{Type: tt.taskType}
			err := pool.Enqueue(ctx, task)
			if (err != nil) != tt.expectError {
				t.Fatalf("Expected error: %v, got: %v", tt.expectError, err)
			}
		})
	}
}

func TestWorkerPool_worker(t *testing.T) {
	tests := []struct {
		name              string
		numTasks          int
		handler           func(context.Context, Task) error
		expectedErrors    uint64
		expectedProcessed uint64
		afterEnqueue      func(*WorkerPool)
	}{
		{
			name:     "All tasks processed successfully",
			numTasks: 5,
			handler: func(ctx context.Context, task Task) error {
				return nil
			},
			expectedErrors:    0,
			expectedProcessed: 5,
		},
		{
			name:     "Tasks with errors",
			numTasks: 3,
			handler: func(ctx context.Context, task Task) error {
				return fmt.Errorf("error processing task")
			},
			expectedErrors:    3,
			expectedProcessed: 0,
		},
		{
			name:              "No handler registered",
			numTasks:          2,
			handler:           nil,
			expectedErrors:    0,
			expectedProcessed: 0,
		},
		{
			name:              "Channel closed",
			numTasks:          0,
			handler:           nil,
			expectedErrors:    0,
			expectedProcessed: 0,
			afterEnqueue: func(pool *WorkerPool) {
				pool.Shutdown()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			pool := NewWorkerPool(ctx, 10, 1)
			defer func() {
				if tt.afterEnqueue == nil {
					pool.Shutdown()
				}
			}()
			processed := uint64(0)
			errors := uint64(0)

			if tt.handler != nil {
				pool.RegisterHandler("test_task", func(ctx context.Context, task Task) error {
					err := tt.handler(ctx, task)
					if err != nil {
						atomic.AddUint64(&errors, 1)
					} else {
						atomic.AddUint64(&processed, 1)
					}
					return err
				})
			}

			for i := 0; i < tt.numTasks; i++ {
				task := &DummyTask{Type: "test_task"}
				if tt.handler == nil {
					continue
				}
				if err := pool.Enqueue(ctx, task); err != nil {
					atomic.AddUint64(&errors, 1)
				}
			}

			if tt.name == "Channel closed" {
				close(pool.taskQueue)
				time.Sleep(100 * time.Millisecond)
				return
			}

			if tt.afterEnqueue != nil {
				tt.afterEnqueue(pool)
			} else {
				time.Sleep(500 * time.Millisecond)
			}
			if atomic.LoadUint64(&processed) != tt.expectedProcessed {
				t.Fatalf("Expected %d tasks processed, got %d", tt.expectedProcessed, processed)
			}

			if atomic.LoadUint64(&errors) != tt.expectedErrors {
				t.Fatalf("Expected %d errors, got %d", tt.expectedErrors, errors)
			}
		})
	}
}
