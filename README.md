# Go Context Examples

A comprehensive collection of Go examples demonstrating different use cases and patterns for the `context` package in Go. This project showcases how to properly use contexts for cancellation, timeouts, and value propagation in concurrent and distributed systems.

## Project Overview

This repository contains several standalone examples that demonstrate various context patterns in real-world scenarios:

1. **Distributed Task Processor** - A concurrent worker pool with context-based cancellation and timeout management
2. **HTTP Server** - Demonstrates context propagation in HTTP handlers with database query timeouts
3. **Web Scraper** - Concurrent scraping with timeout and cancellation support
4. **Context WithValue** - Request metadata propagation through the call chain

## Contents

### 1. Distributed Task Processor (`distributed-tast-processor/`)

A sophisticated concurrent task processing system that demonstrates:

**Features:**

- **Worker Pool Pattern**: Manages a fixed number of worker goroutines to process tasks concurrently
- **Context-Based Cancellation**: All workers respect context cancellation for graceful shutdown
- **Task Timeouts**: Each task has its own timeout context (3 seconds) independent of the overall timeout
- **Retry Logic**: Tasks can be retried with exponential backoff on failure
- **Graceful Shutdown**: Properly closes channels and waits for workers to complete
- **Worker Monitoring**: Background goroutine that periodically logs worker status and queue length
- **Simulated Failures**: Random failures with decreasing probability on retries

**Key Components:**

- `Task`: Represents a unit of work with ID, data, and retry count
- `Result`: Contains task outcome, processing duration, and worker ID
- `WorkerPool`: Manages worker lifecycle, task distribution, and result collection
- `TaskGenerator`: Generates sample tasks with variable retry counts
- Result processor that tracks success/failure statistics

**Context Usage:**

- Overall timeout context (30 seconds) for the entire processing cycle
- Per-task timeout contexts (3 seconds) nested within the main context
- Context cancellation propagates to all workers and tasks

### 2. HTTP Server (`httpserver/`)

Demonstrates context usage in HTTP handlers with database operations:

**Features:**

- **HTTP Server Setup**: Configured with read, write, and idle timeouts
- **Context Propagation**: Request context flows from HTTP handler to database layer
- **Database Timeout**: Implements a 5-second timeout for database queries
- **Error Handling**: Properly handles timeout and cancellation scenarios

**Context Usage:**

- Uses the incoming request's context as the base context
- Creates a timeout context (5 seconds) for database operations
- Demonstrates `context.DeadlineExceeded` error handling

**API Endpoint:**

- `GET /data`: Fetches data from a simulated database with timeout protection

### 3. Web Scraper (`web-scrapper/`)

A concurrent web scraping example showing timeout management:

**Features:**

- **Concurrent Scraping**: Scrapes multiple websites simultaneously using goroutines
- **Context Timeout**: 3-second timeout ensures slow scrapes are cancelled
- **Mutex Protection**: Uses mutex to safely update shared map from concurrent goroutines
- **Selective Processing**: Only slow operations are cancelled; fast ones complete

**Context Usage:**

- Overall 3-second timeout context shared across all scraping operations
- Individual scrape operations respect the context cancellation
- Demonstrates partial success scenarios (some scrapes succeed, others are cancelled)

**Simulated Websites:**

- Fast website (5 seconds) - will be cancelled
- Slow website (10 seconds) - will be cancelled
- Medium website (2 seconds) - will complete successfully

### 4. Context WithValue (`withValue/`)

Demonstrates request metadata propagation using context values:

**Features:**

- **Request Tracing**: Propagates request ID through the entire call chain
- **User Context**: Carries user ID from request to downstream services
- **Type-Safe Keys**: Uses custom types for context keys to avoid collisions
- **Order Processing**: Simulates a multi-step order processing pipeline

**Context Usage:**

- Stores request metadata (request ID, user ID) in context
- Retrieves and uses context values in downstream functions
- Demonstrates proper logging with request context

**Processing Flow:**

1. Request handling with context creation
2. Order processing that retrieves context values
3. Payment validation with context
4. Inventory update with context

## Key Concepts Demonstrated

### Context Cancellation

All examples show how contexts can be used to gracefully cancel operations:

- Worker pool workers shut down cleanly when context is cancelled
- HTTP handlers cancel database operations on timeout
- Web scraper cancels slow operations
- All operations check `ctx.Done()` periodically

### Context Timeouts

Demonstrates different timeout patterns:

- Nested timeouts (overall + per-task timeouts)
- Request-level timeouts in HTTP handlers
- Timeout propagation in concurrent operations

### Context Values

Shows how to safely propagate request metadata:

- Using custom types for context keys
- Retrieving and using values in nested calls
- Logging with context metadata

### Concurrency Patterns

Combines context with Go concurrency primitives:

- Worker pools with context cancellation
- Goroutines with timeout protection
- Mutex usage for safe concurrent map access
- WaitGroup for goroutine coordination

## Use Cases

### When to Use Each Pattern

**Distributed Task Processor:**

- Background job processing systems
- Queue-based task execution
- Batch processing with retry logic
- Systems requiring worker monitoring

**HTTP Server:**

- API servers with database operations
- Request timeout enforcement
- Graceful degradation under load
- Network timeout management

**Web Scraper:**

- Concurrent data fetching
- Operations with variable completion times
- Cancelling slow/stuck operations
- Parallel API calls with timeout

**Context WithValue:**

- Request tracing and correlation
- Passing metadata through call chains
- Distributed logging
- User authentication context propagation

## Context Best Practices

These examples demonstrate Go context best practices:

1. **Always check `ctx.Done()`** in long-running operations
2. **Use timeouts** for operations that shouldn't run indefinitely
3. **Propagate contexts** to all function calls that might block
4. **Use custom types** for context keys to avoid collisions
5. **Create child contexts** for nested timeouts and cancellations
6. **Handle cancellation** gracefully in all goroutines
7. **Document expectations** when accepting context parameters

## Error Handling

All examples demonstrate proper error handling with contexts:

- Checking for `context.DeadlineExceeded` to distinguish timeouts
- Checking for `context.Cancelled` to handle cancellations
- Properly cleaning up resources on cancellation
- Logging context errors for debugging

## Architecture Highlights

### Distributed Task Processor Architecture

```
Main
├── WorkerPool (4 workers)
│   ├── Worker 1-4 (goroutines)
│   │   ├── Task Queue (channel)
│   │   ├── Context cancellation handling
│   │   └── Retry logic with timeout
│   └── Monitor (background goroutine)
├── Task Generator (goroutine)
└── Result Processor (goroutine)
```

### Web Scraper Architecture

```
Main
├── Context with 3s timeout
└── Scraper
    ├── Goroutine 1 (fast.com)
    ├── Goroutine 2 (slow.com) ✗ cancelled
    └── Goroutine 3 (medium.com) ✓
    └── Mutex-protected results map
```

## Observations

### Context Propagation

- Contexts flow naturally through function calls
- Child contexts inherit parent cancellation but can have their own timeouts
- Cancellation signal spreads to all operations using the context

### Concurrent Safety

- Worker pools safely distribute tasks across workers
- Web scraper uses mutex to protect shared state
- All goroutines respect context cancellation

### Resource Management

- Proper channel closing in worker pools
- WaitGroup usage for coordination
- Deferred cleanup in all goroutines

## Notes

- All examples use simulated operations (timers, random failures) to demonstrate patterns without external dependencies
- The code is structured for educational purposes with detailed logging
- Each example is self-contained and can be run independently
- Error handling is explicit to show how context errors should be managed

