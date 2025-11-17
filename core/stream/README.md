# Auto-Restart Connect Streams

This package provides automatic restart functionality for Connect bidirectional streams, ensuring continuous connections between components even when streams are closed or encounter errors.

## Features

- **Automatic Restart**: Streams automatically restart when they encounter errors or are closed
- **Exponential Backoff**: Configurable retry logic with exponential backoff to prevent overwhelming the server
- **Thread-Safe**: Safe for concurrent use across multiple goroutines
- **Customizable Callbacks**: Monitor restart events with custom callback functions
- **Context-Aware**: Respects context cancellation for graceful shutdown

## Quick Start

```go
package main

import (
    "context"
    "log"
    "net/http"
    "time"

    "confirmate.io/core/api/orchestrator"
    "confirmate.io/core/api/orchestrator/orchestratorconnect"
    "confirmate.io/core/stream"
    "connectrpc.com/connect"
)

func main() {
    // Create your Connect client
    client := orchestratorconnect.NewOrchestratorClient(
        http.DefaultClient,
        "http://localhost:8080",
    )

    // Define a factory function that creates new streams
    factory := func(ctx context.Context) *connect.BidiStreamForClient[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse] {
        return client.StoreAssessmentResults(ctx)
    }

    // Configure auto-restart behavior
    config := stream.DefaultRestartConfig()
    config.MaxRetries = 10 // Retry up to 10 times (0 = unlimited)
    config.InitialBackoff = 100 * time.Millisecond
    config.MaxBackoff = 30 * time.Second

    // Create the restartable stream
    ctx := context.Background()
    rs, err := stream.NewRestartableBidiStream(ctx, factory, config)
    if err != nil {
        log.Fatalf("Failed to create stream: %v", err)
    }
    defer rs.Close()

    // Use the stream - it will automatically restart on errors
    for {
        msg, err := rs.Receive()
        if err != nil {
            log.Printf("Error receiving: %v", err)
            continue
        }
        // Process message...
        log.Printf("Received: %v", msg)
    }
}
```

## Configuration

### RestartConfig

Configure how the stream restarts on errors:

```go
type RestartConfig struct {
    // MaxRetries is the maximum number of restart attempts. 
    // 0 means unlimited retries.
    MaxRetries int

    // InitialBackoff is the initial delay before the first retry.
    InitialBackoff time.Duration

    // MaxBackoff is the maximum delay between retries.
    MaxBackoff time.Duration

    // BackoffMultiplier is the factor by which the backoff increases 
    // after each retry.
    BackoffMultiplier float64

    // OnRestart is called when a stream restart is attempted.
    OnRestart func(attempt int, err error)

    // OnRestartSuccess is called when a stream restart succeeds.
    OnRestartSuccess func(attempt int)

    // OnRestartFailure is called when all restart attempts have failed.
    OnRestartFailure func(err error)
}
```

### Default Configuration

```go
config := stream.DefaultRestartConfig()
// Returns:
// MaxRetries: 0 (unlimited)
// InitialBackoff: 100ms
// MaxBackoff: 30s
// BackoffMultiplier: 2.0
```

## Advanced Usage

### Custom Callbacks

Monitor stream health and restart events:

```go
config := stream.DefaultRestartConfig()
config.OnRestart = func(attempt int, err error) {
    log.Printf("Restarting stream (attempt %d) due to: %v", attempt, err)
}
config.OnRestartSuccess = func(attempt int) {
    log.Printf("Stream restarted successfully after %d attempts", attempt)
}
config.OnRestartFailure = func(err error) {
    log.Printf("Failed to restart stream after all attempts: %v", err)
}
```

### Monitoring Stream Health

```go
// Get the number of times the stream has been restarted
retryCount := rs.RetryCount()

// Get the last error that triggered a restart
lastError := rs.LastError()

log.Printf("Stream has restarted %d times", retryCount)
if lastError != nil {
    log.Printf("Last error: %v", lastError)
}
```

### Context Cancellation

The stream respects context cancellation for graceful shutdown:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

rs, err := stream.NewRestartableBidiStream(ctx, factory, config)
if err != nil {
    log.Fatal(err)
}
defer rs.Close()

// When context is cancelled, the stream will stop restarting
```

## How It Works

1. **Initial Connection**: The stream factory is called to create the initial connection
2. **Error Detection**: When `Send()` or `Receive()` encounters an error, the restart mechanism is triggered
3. **Exponential Backoff**: The system waits before attempting to restart, with increasing delays between attempts
4. **Automatic Retry**: A new stream is created using the factory function
5. **Resume Operation**: Once restarted, the operation is retried automatically

## Testing

The package includes comprehensive tests demonstrating:

- Basic stream operations
- Automatic restart after connection loss
- Exponential backoff behavior
- Thread safety with concurrent operations
- Proper cleanup and resource management
- Context cancellation handling

Run tests with:
```bash
go test ./stream/...
```

## Use Cases

### Assessment Tool Connections

Maintain continuous connections between assessment tools and the orchestrator:

```go
// Assessment tool client that needs continuous connection
factory := func(ctx context.Context) *connect.BidiStreamForClient[...] {
    return client.StoreAssessmentResults(ctx)
}

rs, _ := stream.NewRestartableBidiStream(ctx, factory, config)
defer rs.Close()

// Send assessment results continuously
for result := range assessmentResults {
    if err := rs.Send(result); err != nil {
        log.Printf("Error sending: %v", err)
        // Stream will automatically restart
    }
}
```

### Metric Subscriptions

Subscribe to metric change events with automatic reconnection:

```go
factory := func(ctx context.Context) *connect.ServerStreamForClient[...] {
    return client.SubscribeMetricChangeEvents(ctx)
}

// Receive events continuously
for {
    event, err := rs.Receive()
    if err != nil {
        log.Printf("Error: %v", err)
        continue // Stream restarts automatically
    }
    handleMetricChange(event)
}
```

## Best Practices

1. **Set Appropriate Retry Limits**: For production, set `MaxRetries` to prevent infinite retry loops
2. **Use Exponential Backoff**: Don't overwhelm the server with rapid retry attempts
3. **Monitor Retry Count**: Track `RetryCount()` to detect persistent connection issues
4. **Implement Callbacks**: Use callbacks for logging and monitoring
5. **Respect Context**: Always use context cancellation for graceful shutdown
6. **Close Streams**: Always call `Close()` when done with the stream

## Thread Safety

All operations on `RestartableBidiStream` are thread-safe and can be called concurrently from multiple goroutines. The internal state is protected with read-write locks.

## License

Copyright 2016-2025 Fraunhofer AISEC

SPDX-License-Identifier: Apache-2.0
