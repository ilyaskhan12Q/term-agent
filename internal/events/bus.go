package events

import (
	"context"
)

// Handler represents an event consumer callback.
type Handler func(event Event)

// EventBus defines the decoupled publish-subscribe contract.
type EventBus interface {
	Publish(event Event)
	Subscribe(eventType EventType, handler Handler) Subscription
	Shutdown(ctx context.Context) error
}

// Subscription allows unsubscribing from event notifications.
type Subscription interface {
	Unsubscribe()
}
