package events

import (
	"context"
	"fmt"
	"sync"
)

type subscriptionImpl struct {
	id        int64
	eventType EventType
	bus       *InMemoryEventBus
}

func (s *subscriptionImpl) Unsubscribe() {
	s.bus.unsubscribe(s.eventType, s.id)
}

// InMemoryEventBus implements a concurrency-safe pub-sub event bus.
type InMemoryEventBus struct {
	mu          sync.RWMutex
	subscribers map[EventType]map[int64]Handler
	nextSubID   int64
	closed      bool
}

// NewInMemoryEventBus creates a new EventBus instance.
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		subscribers: make(map[EventType]map[int64]Handler),
	}
}

// Subscribe registers a handler callback for a specific event type.
func (b *InMemoryEventBus) Subscribe(eventType EventType, handler Handler) Subscription {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return &subscriptionImpl{}
	}

	if _, exists := b.subscribers[eventType]; !exists {
		b.subscribers[eventType] = make(map[int64]Handler)
	}

	b.nextSubID++
	subID := b.nextSubID
	b.subscribers[eventType][subID] = handler

	return &subscriptionImpl{
		id:        subID,
		eventType: eventType,
		bus:       b,
	}
}

func (b *InMemoryEventBus) unsubscribe(eventType EventType, subID int64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subs, exists := b.subscribers[eventType]; exists {
		delete(subs, subID)
		if len(subs) == 0 {
			delete(b.subscribers, eventType)
		}
	}
}

// Publish broadcasts an event to all registered subscribers safely.
func (b *InMemoryEventBus) Publish(event Event) {
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return
	}

	subsCopy := make([]Handler, 0)
	if subs, exists := b.subscribers[event.Type]; exists {
		for _, h := range subs {
			subsCopy = append(subsCopy, h)
		}
	}
	b.mu.RUnlock()

	for _, handler := range subsCopy {
		func(h Handler) {
			defer func() {
				if r := recover(); r != nil {
					// Prevent a single panicking handler from crashing the event bus or application
					_ = fmt.Sprintf("event handler panic recovered: %v", r)
				}
			}()
			h(event)
		}(handler)
	}
}

// Shutdown safely closes the event bus.
func (b *InMemoryEventBus) Shutdown(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.closed = true
	b.subscribers = make(map[EventType]map[int64]Handler)
	return nil
}
