package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

// Task represents a unit of work
type Task struct {
	ID    int
	Data  string
	Retry int
}

// Result represents the outcome of processing a task
type Result struct {
	TaskID   int
	Output   string
	Duration time.Duration
	Error    error
	WorkerID int
}

// WorkerPool manages a pool of worker goroutines
type WorkerPool struct {
	workerCount   int
	taskQueue     chan Task
	resultChan    chan Result
	workerWg      sync.WaitGroup
	mu            sync.Mutex
	activeWorkers int
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workerCount int, taskBuffer int) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		taskQueue:   make(chan Task, taskBuffer),
		resultChan:  make(chan Result, taskBuffer),
	}
}

// Start begins processing tasks with context support
func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 1; i <= wp.workerCount; i++ {
		wp.workerWg.Add(1)
		go wp.worker(ctx, i)
	}

	// Monitor worker status in background
	go wp.monitorWorkers(ctx)
}

// worker processes tasks from the queue
func (wp *WorkerPool) worker(ctx context.Context, workerID int) {
	defer wp.workerWg.Done()

	// Track this worker as active
	wp.mu.Lock()
	wp.activeWorkers++
	wp.mu.Unlock()

	log.Printf("Worker %d started", workerID)

	for {
		select {
		case task, ok := <-wp.taskQueue:
			if !ok {
				log.Printf("Worker %d: task queue closed, shutting down", workerID)

				wp.mu.Lock()
				wp.activeWorkers--
				wp.mu.Unlock()
				return
			}

			// Process the task with timeout
			result := wp.processTask(ctx, task, workerID)
			wp.resultChan <- result

		case <-ctx.Done():
			log.Printf("Worker %d: context cancelled, shutting down", workerID)

			wp.mu.Lock()
			wp.activeWorkers--
			wp.mu.Unlock()
			return
		}
	}
}

// processTask handles individual task processing with retry logic
func (wp *WorkerPool) processTask(ctx context.Context, task Task, workerID int) Result {
	start := time.Now()

	// Create a timeout context for this specific task
	taskCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var lastError error

	// Retry logic
	for attempt := 0; attempt <= task.Retry; attempt++ {
		select {
		case <-taskCtx.Done():
			return Result{
				TaskID:   task.ID,
				Error:    fmt.Errorf("task cancelled after %d attempts: %w", attempt, taskCtx.Err()),
				Duration: time.Since(start),
				WorkerID: workerID,
			}
		default:
			// Simulate processing work
			result, err := wp.simulateWork(taskCtx, task, attempt)
			if err == nil {
				return Result{
					TaskID:   task.ID,
					Output:   result,
					Duration: time.Since(start),
					WorkerID: workerID,
				}
			}
			lastError = err

			if attempt < task.Retry {
				log.Printf("Worker %d: task %d failed (attempt %d), retrying: %v",
					workerID, task.ID, attempt+1, err)
				time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
			}
		}
	}

	return Result{
		TaskID:   task.ID,
		Error:    fmt.Errorf("task failed after %d retries: %w", task.Retry, lastError),
		Duration: time.Since(start),
		WorkerID: workerID,
	}
}

// simulateWork simulates actual work with random failures
func (wp *WorkerPool) simulateWork(ctx context.Context, task Task, attempt int) (string, error) {
	// Simulate variable processing time
	processingTime := time.Duration(rand.Intn(2000)) * time.Millisecond

	select {
	case <-time.After(processingTime):
		// Simulate random failures (20% chance on first attempt, decreasing)
		failureChance := 20 - (attempt * 5)
		if failureChance < 5 {
			failureChance = 5
		}

		if rand.Intn(100) < failureChance {
			return "", fmt.Errorf("simulated processing failure")
		}

		return fmt.Sprintf("Processed: %s (attempt %d, took %v)",
			task.Data, attempt+1, processingTime), nil

	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// monitorWorkers periodically logs worker status
func (wp *WorkerPool) monitorWorkers(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wp.mu.Lock()
			active := wp.activeWorkers
			queueLen := len(wp.taskQueue)
			wp.mu.Unlock()

			log.Printf("Monitor: %d active workers, %d tasks in queue", active, queueLen)

		case <-ctx.Done():
			log.Printf("Monitor: shutting down")
			return
		}
	}
}

// Submit adds tasks to the queue
func (wp *WorkerPool) Submit(task Task) error {
	select {
	case wp.taskQueue <- task:
		return nil
	default:
		return fmt.Errorf("task queue full, cannot submit task %d", task.ID)
	}
}

// Stop gracefully shuts down the worker pool
func (wp *WorkerPool) Stop() {
	log.Println("Shutting down worker pool...")

	// Close task queue (no new tasks accepted)
	close(wp.taskQueue)

	// Wait for all workers to finish
	wp.workerWg.Wait()

	// Close result channel
	close(wp.resultChan)
}

// Results returns the result channel
func (wp *WorkerPool) Results() <-chan Result {
	return wp.resultChan
}

// TaskGenerator generates sample tasks
func TaskGenerator(ctx context.Context, count int) <-chan Task {
	taskChan := make(chan Task)

	go func() {
		defer close(taskChan)

		for i := 1; i <= count; i++ {
			task := Task{
				ID:    i,
				Data:  fmt.Sprintf("Task data %d", i),
				Retry: rand.Intn(3), // 0-2 retries
			}

			select {
			case taskChan <- task:
			case <-ctx.Done():
				log.Println("Task generator: context cancelled")
				return
			}

			// Small delay between task generation
			time.Sleep(100 * time.Millisecond)
		}
	}()

	return taskChan
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Create context with overall timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create worker pool
	workerPool := NewWorkerPool(4, 10) // 4 workers, buffer of 10

	// Start the worker pool
	workerPool.Start(ctx)

	// WaitGroup to track result processing
	var resultWg sync.WaitGroup

	// Start result processor
	resultWg.Add(1)
	go func() {
		defer resultWg.Done()
		processResults(ctx, workerPool.Results())
	}()

	// Generate and submit tasks
	taskGenerator := TaskGenerator(ctx, 20)

	var submittedTasks int
	for task := range taskGenerator {
		if err := workerPool.Submit(task); err != nil {
			log.Printf("Failed to submit task %d: %v", task.ID, err)
		} else {
			submittedTasks++
		}
	}

	log.Printf("Submitted %d tasks to worker pool", submittedTasks)

	// Wait for context cancellation or timeout
	<-ctx.Done()
	log.Printf("Main: context done - %v", ctx.Err())

	// Graceful shutdown
	workerPool.Stop()

	// Wait for result processing to complete
	resultWg.Wait()

	log.Println("Application shutdown complete")
}

// processResults handles results from the worker pool
func processResults(ctx context.Context, results <-chan Result) {
	successCount := 0
	failureCount := 0

	for {
		select {
		case result, ok := <-results:
			if !ok {
				log.Printf("Result processor: channel closed. Success: %d, Failures: %d",
					successCount, failureCount)
				return
			}

			if result.Error != nil {
				log.Printf("Task %d (Worker %d) FAILED after %v: %v",
					result.TaskID, result.WorkerID, result.Duration, result.Error)
				failureCount++
			} else {
				log.Printf("Task %d (Worker %d) SUCCESS: %s (took %v)",
					result.TaskID, result.WorkerID, result.Output, result.Duration)
				successCount++
			}

		case <-ctx.Done():
			log.Printf("Result processor: context cancelled. Success: %d, Failures: %d",
				successCount, failureCount)
			return
		}
	}
}
