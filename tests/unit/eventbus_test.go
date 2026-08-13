package unit

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ilyaskhan/term-agent/internal/events"
)

func TestInMemoryEventBusPubSub(t *testing.T) {
	bus := events.NewInMemoryEventBus()
	defer bus.Shutdown(context.Background())

	var wg sync.WaitGroup
	wg.Add(2)

	received := make([]events.Event, 0)
	var mu sync.Mutex

	sub1 := bus.Subscribe(events.EventToolCallProposed, func(e events.Event) {
		mu.Lock()
		received = append(received, e)
		mu.Unlock()
		wg.Done()
	})
	defer sub1.Unsubscribe()

	sub2 := bus.Subscribe(events.EventToolCallProposed, func(e events.Event) {
		mu.Lock()
		received = append(received, e)
		mu.Unlock()
		wg.Done()
	})
	defer sub2.Unsubscribe()

	event := events.Event{
		ID:        "evt-1",
		Type:      events.EventToolCallProposed,
		SessionID: "sess-1",
		Timestamp: time.Now(),
	}

	bus.Publish(event)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		mu.Lock()
		if len(received) != 2 {
			t.Errorf("expected 2 events received, got %d", len(received))
		}
		mu.Unlock()
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for event handlers")
	}
}

func TestInMemoryEventBusUnsubscribe(t *testing.T) {
	bus := events.NewInMemoryEventBus()
	defer bus.Shutdown(context.Background())

	called := false
	sub := bus.Subscribe(events.EventMutationCommitted, func(e events.Event) {
		called = true
	})

	sub.Unsubscribe()

	bus.Publish(events.Event{
		ID:   "evt-2",
		Type: events.EventMutationCommitted,
	})

	if called {
		t.Error("handler was called after unsubscribe")
	}
}

func TestInMemoryEventBusPanicRecovery(t *testing.T) {
	bus := events.NewInMemoryEventBus()
	defer bus.Shutdown(context.Background())

	sub := bus.Subscribe(events.EventErrorOccurred, func(e events.Event) {
		panic("simulated subscriber failure")
	})
	defer sub.Unsubscribe()

	// Publishing should not crash application despite handler panic
	bus.Publish(events.Event{
		ID:   "evt-panic",
		Type: events.EventErrorOccurred,
	})
}
