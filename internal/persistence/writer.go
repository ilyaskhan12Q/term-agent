package persistence

import (
	"context"
	"errors"
	"fmt"

	"log/slog"
	"sync"
	"time"
)

var (
	ErrWriterClosed  = errors.New("async db writer is closed")
	ErrQueueOverflow = errors.New("async db writer queue overflow")
)

// WriteTask represents an asynchronous database write operation.
type WriteTask func(ctx context.Context, db *DB) error

// AsyncWriter defines the contract for controlled async database writes to avoid blocking TUI rendering.
type AsyncWriter interface {
	Enqueue(task WriteTask) error
	Close(ctx context.Context) error
}

// ChannelAsyncWriter implements AsyncWriter using a buffered channel and a dedicated single background worker goroutine.
type ChannelAsyncWriter struct {
	db        *DB
	queue     chan WriteTask
	closeOnce sync.Once
	done      chan struct{}
	closed    chan struct{}
	logger    *slog.Logger
}

// NewChannelAsyncWriter initializes a ChannelAsyncWriter with a specified channel queue buffer size.
func NewChannelAsyncWriter(db *DB, queueSize int, logger *slog.Logger) *ChannelAsyncWriter {
	if queueSize <= 0 {
		queueSize = 500
	}
	if logger == nil {
		logger = slog.Default()
	}

	w := &ChannelAsyncWriter{
		db:     db,
		queue:  make(chan WriteTask, queueSize),
		done:   make(chan struct{}),
		closed: make(chan struct{}),
		logger: logger,
	}

	go w.processLoop()
	return w
}

// Enqueue submits a write operation to the async channel.
func (w *ChannelAsyncWriter) Enqueue(task WriteTask) error {
	if task == nil {
		return nil
	}

	select {
	case <-w.done:
		return ErrWriterClosed
	default:
	}

	select {
	case w.queue <- task:
		return nil
	case <-w.done:
		return ErrWriterClosed
	default:
		return ErrQueueOverflow
	}
}

func (w *ChannelAsyncWriter) processLoop() {
	defer close(w.closed)

	for {
		select {
		case task, ok := <-w.queue:
			if !ok {
				// Channel drained and closed
				return
			}
			w.executeTask(task)
		case <-w.done:
			// Drain remaining tasks in queue
			for {
				select {
				case task, ok := <-w.queue:
					if !ok {
						return
					}
					w.executeTask(task)
				default:
					return
				}
			}
		}
	}
}

func (w *ChannelAsyncWriter) executeTask(task WriteTask) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := task(ctx, w.db); err != nil {
		w.logger.Error("Async database write task failed", "error", err)
	}
}

// Close gracefully stops accepting new tasks, flushes queued tasks, and shuts down the background writer.
func (w *ChannelAsyncWriter) Close(ctx context.Context) error {
	w.closeOnce.Do(func() {
		close(w.done)
		close(w.queue)
	})

	select {
	case <-w.closed:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("async db writer close timed out: %w", ctx.Err())
	}
}
