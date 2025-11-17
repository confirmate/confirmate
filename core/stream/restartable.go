// Copyright 2016-2025 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//                                 /$$$$$$  /$$                                     /$$
//                               /$$__  $$|__/                                    | $$
//   /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//  /$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/
//
// This file is part of Confirmate Core.

// Package stream provides automatic restart functionality for Connect bidirectional streams,
// ensuring continuous connections between components even when streams are closed or encounter errors.
//
// # Features
//
// - Automatic Restart: Streams automatically restart when they encounter errors or are closed
// - Exponential Backoff: Configurable retry logic with exponential backoff to prevent overwhelming the server
// - Thread-Safe: Safe for concurrent use across multiple goroutines
// - Customizable Callbacks: Monitor restart events with custom callback functions
// - Context-Aware: Respects context cancellation for graceful shutdown
//
// # Quick Start
//
// Create a factory function that creates new streams:
//
//	factory := func(ctx context.Context) *connect.BidiStreamForClient[Req, Res] {
//	    return client.StoreAssessmentResults(ctx)
//	}
//
// Configure auto-restart behavior:
//
//	config := stream.DefaultRestartConfig()
//	config.MaxRetries = 10 // Retry up to 10 times (0 = unlimited)
//	config.InitialBackoff = 100 * time.Millisecond
//	config.MaxBackoff = 30 * time.Second
//
// Create the restartable stream:
//
//	rs, err := stream.NewRestartableBidiStream(ctx, factory, config)
//	if err != nil {
//	    return err
//	}
//	defer rs.Close()
//
// Use the stream - it will automatically restart on errors:
//
//	err = rs.Send(request)
//	msg, err := rs.Receive()
//
// # Configuration
//
// The RestartConfig struct allows fine-grained control over restart behavior:
//
//   - MaxRetries: Maximum number of restart attempts (0 = unlimited)
//   - InitialBackoff: Initial delay before the first retry
//   - MaxBackoff: Maximum delay between retries
//   - BackoffMultiplier: Factor by which backoff increases after each retry
//   - OnRestart: Callback when a stream restart is attempted
//   - OnRestartSuccess: Callback when a stream restart succeeds
//   - OnRestartFailure: Callback when all restart attempts have failed
//
// # Monitoring
//
// Monitor stream health with built-in methods:
//
//	retryCount := rs.RetryCount()  // Number of times stream has restarted
//	lastError := rs.LastError()    // Last error that triggered a restart
//
// # Thread Safety
//
// All operations on RestartableBidiStream are thread-safe and can be called
// concurrently from multiple goroutines.
//
// # Best Practices
//
//   - Set MaxRetries in production to prevent infinite retry loops
//   - Use exponential backoff to avoid overwhelming servers
//   - Monitor RetryCount() to detect persistent connection issues
//   - Implement callbacks for logging and monitoring
//   - Always respect context cancellation for graceful shutdown
//   - Always call Close() when done with the stream
package stream

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"connectrpc.com/connect"
)

// RestartConfig contains configuration for automatic stream restart behavior.
type RestartConfig struct {
	// MaxRetries is the maximum number of restart attempts. 0 means unlimited retries.
	MaxRetries int

	// InitialBackoff is the initial delay before the first retry.
	InitialBackoff time.Duration

	// MaxBackoff is the maximum delay between retries.
	MaxBackoff time.Duration

	// BackoffMultiplier is the factor by which the backoff increases after each retry.
	BackoffMultiplier float64

	// OnRestart is called when a stream restart is attempted.
	OnRestart func(attempt int, err error)

	// OnRestartSuccess is called when a stream restart succeeds.
	OnRestartSuccess func(attempt int)

	// OnRestartFailure is called when all restart attempts have failed.
	OnRestartFailure func(err error)
}

// DefaultRestartConfig returns a RestartConfig with sensible defaults:
//   - MaxRetries: 0 (unlimited retries)
//   - InitialBackoff: 100ms
//   - MaxBackoff: 30s
//   - BackoffMultiplier: 2.0
//   - Logging callbacks for restart events
func DefaultRestartConfig() RestartConfig {
	return RestartConfig{
		MaxRetries:        0, // unlimited
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        30 * time.Second,
		BackoffMultiplier: 2.0,
		OnRestart: func(attempt int, err error) {
			slog.Info("Attempting to restart stream", "attempt", attempt, "error", err)
		},
		OnRestartSuccess: func(attempt int) {
			slog.Info("Stream restart successful", "attempt", attempt)
		},
		OnRestartFailure: func(err error) {
			slog.Error("Failed to restart stream after all attempts", "error", err)
		},
	}
}

// StreamFactory is a function that creates a new bidirectional stream.
type StreamFactory[Req, Res any] func(ctx context.Context) *connect.BidiStreamForClient[Req, Res]

// RestartableBidiStream wraps a Connect bidirectional stream with automatic restart functionality.
// It transparently handles connection failures by recreating the stream using the provided factory
// function, with exponential backoff between retry attempts. All operations are thread-safe.
type RestartableBidiStream[Req, Res any] struct {
	name    string
	factory StreamFactory[Req, Res]
	config  RestartConfig

	mu         sync.RWMutex
	stream     *connect.BidiStreamForClient[Req, Res]
	ctx        context.Context
	cancel     context.CancelFunc
	closed     bool
	lastError  error
	retryCount int
}

// NewRestartableBidiStream creates a new restartable bidirectional stream.
// The factory function is called to create a new stream initially and whenever
// a restart is needed due to errors. The returned stream must be closed with
// Close() when no longer needed to clean up resources and prevent further restarts.
// The name parameter is optional and used for logging; if empty, a default name is used.
func NewRestartableBidiStream[Req, Res any](
	ctx context.Context,
	factory StreamFactory[Req, Res],
	config RestartConfig,
	name ...string,
) (*RestartableBidiStream[Req, Res], error) {
	var (
		streamCtx context.Context
		cancel    context.CancelFunc
		streamName string
	)
	
	streamCtx, cancel = context.WithCancel(ctx)
	
	if len(name) > 0 && name[0] != "" {
		streamName = name[0]
	} else {
		streamName = "BidiStream"
	}

	rs := &RestartableBidiStream[Req, Res]{
		name:    streamName,
		factory: factory,
		config:  config,
		ctx:     streamCtx,
		cancel:  cancel,
	}

	// Create initial stream
	rs.stream = factory(streamCtx)
	if rs.stream == nil {
		cancel()
		return nil, fmt.Errorf("factory returned nil stream")
	}

	return rs, nil
}

// Send sends a message on the stream, automatically restarting if needed.
func (rs *RestartableBidiStream[Req, Res]) Send(msg *Req) (err error) {
	var (
		stream *connect.BidiStreamForClient[Req, Res]
	)
	
	rs.mu.RLock()
	if rs.closed {
		rs.mu.RUnlock()
		return io.EOF
	}
	stream = rs.stream
	rs.mu.RUnlock()

	err = stream.Send(msg)
	if err != nil {
		// Try to restart the stream
		if restartErr := rs.restart(err); restartErr != nil {
			return fmt.Errorf("failed to send and restart: %w", restartErr)
		}
		// Retry sending on the new stream
		rs.mu.RLock()
		stream = rs.stream
		rs.mu.RUnlock()
		return stream.Send(msg)
	}

	return nil
}

// Receive receives a message from the stream, automatically restarting if needed.
func (rs *RestartableBidiStream[Req, Res]) Receive() (msg *Res, err error) {
	var (
		stream *connect.BidiStreamForClient[Req, Res]
	)
	
	rs.mu.RLock()
	if rs.closed {
		rs.mu.RUnlock()
		return nil, io.EOF
	}
	stream = rs.stream
	rs.mu.RUnlock()

	msg, err = stream.Receive()
	if err != nil {
		// Try to restart the stream
		if restartErr := rs.restart(err); restartErr != nil {
			return nil, fmt.Errorf("failed to receive and restart: %w", restartErr)
		}
		// Retry receiving on the new stream
		rs.mu.RLock()
		stream = rs.stream
		rs.mu.RUnlock()
		return stream.Receive()
	}

	return msg, nil
}

// restart attempts to restart the stream with exponential backoff.
func (rs *RestartableBidiStream[Req, Res]) restart(originalErr error) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.closed {
		return fmt.Errorf("stream is closed")
	}

	// Store the last error
	rs.lastError = originalErr

	backoff := rs.config.InitialBackoff
	attempt := 0

	for {
		attempt++
		rs.retryCount++

		// Check if we've exceeded max retries
		if rs.config.MaxRetries > 0 && attempt > rs.config.MaxRetries {
			if rs.config.OnRestartFailure != nil {
				rs.config.OnRestartFailure(originalErr)
			}
			return fmt.Errorf("max retries exceeded: %w", originalErr)
		}

		// Call restart callback
		if rs.config.OnRestart != nil {
			rs.config.OnRestart(attempt, originalErr)
		}

		// Wait before retrying
		select {
		case <-rs.ctx.Done():
			return rs.ctx.Err()
		case <-time.After(backoff):
		}

		// Try to create a new stream
		newStream := rs.factory(rs.ctx)
		if newStream != nil {
			// Close old stream if it exists
			if rs.stream != nil {
				_ = rs.stream.CloseRequest()
				_ = rs.stream.CloseResponse()
			}

			rs.stream = newStream
			if rs.config.OnRestartSuccess != nil {
				rs.config.OnRestartSuccess(attempt)
			}
			return nil
		}

		// Calculate next backoff with exponential increase
		backoff = time.Duration(float64(backoff) * rs.config.BackoffMultiplier)
		if backoff > rs.config.MaxBackoff {
			backoff = rs.config.MaxBackoff
		}
	}
}

// CloseRequest closes the request side of the stream.
func (rs *RestartableBidiStream[Req, Res]) CloseRequest() error {
	rs.mu.RLock()
	stream := rs.stream
	rs.mu.RUnlock()

	if stream == nil {
		return nil
	}
	return stream.CloseRequest()
}

// CloseResponse closes the response side of the stream.
func (rs *RestartableBidiStream[Req, Res]) CloseResponse() error {
	rs.mu.RLock()
	stream := rs.stream
	rs.mu.RUnlock()

	if stream == nil {
		return nil
	}
	return stream.CloseResponse()
}

// Close closes the stream and cancels the context, preventing any further restarts.
func (rs *RestartableBidiStream[Req, Res]) Close() error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.closed {
		return nil
	}

	rs.closed = true
	rs.cancel()

	if rs.stream != nil {
		_ = rs.stream.CloseRequest()
		_ = rs.stream.CloseResponse()
	}

	return nil
}

// RetryCount returns the number of times the stream has been restarted.
func (rs *RestartableBidiStream[Req, Res]) RetryCount() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.retryCount
}

// LastError returns the last error that triggered a restart.
func (rs *RestartableBidiStream[Req, Res]) LastError() error {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.lastError
}
