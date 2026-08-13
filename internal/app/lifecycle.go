package app

import (
	"context"
)

// Lifecycle defines the application startup and shutdown contracts.
type Lifecycle interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// State represents the current lifecycle state of the application.
type State string

const (
	StateInitializing State = "INITIALIZING"
	StateRunning      State = "RUNNING"
	StateShuttingDown State = "SHUTTING_DOWN"
	StateStopped      State = "STOPPED"
)
