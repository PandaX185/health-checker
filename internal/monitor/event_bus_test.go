package monitor

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewInMemoryEventBus(t *testing.T) {
	logger := zap.NewNop()
	bus := NewInMemoryEventBus(logger)

	assert.NotNil(t, bus)
	assert.IsType(t, &InMemoryEventBus{}, bus)
}

func TestEventBus_PublishAndSubscribe(t *testing.T) {
	logger := zap.NewNop()
	bus := NewInMemoryEventBus(logger)
	ctx := context.Background()

	// Channel to capture events
	eventChan := make(chan Event, 1)

	// Subscribe
	bus.Subscribe("StatusChange", func(ctx context.Context, event Event) {
		eventChan <- event
	})

	// Publish event
	event := StatusChangeEvent{
		ServiceID: 1,
		OldStatus: "UP",
		NewStatus: "DOWN",
		Timestamp: time.Now(),
	}

	err := bus.Publish(ctx, event)
	assert.NoError(t, err)

	// Wait for event
	select {
	case received := <-eventChan:
		sce, ok := received.(StatusChangeEvent)
		assert.True(t, ok)
		assert.Equal(t, 1, sce.ServiceID)
		assert.Equal(t, "UP", sce.OldStatus)
		assert.Equal(t, "DOWN", sce.NewStatus)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for event")
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	logger := zap.NewNop()
	bus := NewInMemoryEventBus(logger)
	ctx := context.Background()

	// Create multiple subscribers
	eventChan1 := make(chan Event, 1)
	eventChan2 := make(chan Event, 1)

	bus.Subscribe("StatusChange", func(ctx context.Context, event Event) {
		eventChan1 <- event
	})
	bus.Subscribe("StatusChange", func(ctx context.Context, event Event) {
		eventChan2 <- event
	})

	// Publish event
	event := StatusChangeEvent{
		ServiceID: 2,
		OldStatus: "DOWN",
		NewStatus: "UP",
		Timestamp: time.Now(),
	}

	err := bus.Publish(ctx, event)
	assert.NoError(t, err)

	// Both subscribers should receive the event
	select {
	case <-eventChan1:
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for event on subscriber 1")
	}

	select {
	case <-eventChan2:
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for event on subscriber 2")
	}
}

func TestEventBus_PublishWithoutSubscribers(t *testing.T) {
	logger := zap.NewNop()
	bus := NewInMemoryEventBus(logger)
	ctx := context.Background()

	// Publish event without any subscribers
	event := StatusChangeEvent{
		ServiceID: 3,
		OldStatus: "UP",
		NewStatus: "DOWN",
		Timestamp: time.Now(),
	}

	err := bus.Publish(ctx, event)
	assert.NoError(t, err) // Should not error
}

func TestEventBus_HandlerPanic(t *testing.T) {
	logger := zap.NewNop()
	bus := NewInMemoryEventBus(logger)
	ctx := context.Background()

	// Channel to verify second handler still executes
	eventChan := make(chan Event, 1)

	// Subscribe with panicking handler
	bus.Subscribe("StatusChange", func(ctx context.Context, event Event) {
		panic("handler panic")
	})

	// Subscribe with normal handler
	bus.Subscribe("StatusChange", func(ctx context.Context, event Event) {
		eventChan <- event
	})

	// Publish event
	event := StatusChangeEvent{
		ServiceID: 4,
		OldStatus: "UP",
		NewStatus: "DOWN",
		Timestamp: time.Now(),
	}

	err := bus.Publish(ctx, event)
	assert.NoError(t, err)

	// Second handler should still receive event despite first handler panicking
	select {
	case <-eventChan:
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout - second handler should still execute")
	}
}

func TestEventBus_ConcurrentPublish(t *testing.T) {
	logger := zap.NewNop()
	bus := NewInMemoryEventBus(logger)
	ctx := context.Background()

	var wg sync.WaitGroup
	eventCount := 0
	mu := sync.Mutex{}

	// Subscribe
	bus.Subscribe("StatusChange", func(ctx context.Context, event Event) {
		mu.Lock()
		eventCount++
		mu.Unlock()
	})

	// Publish multiple events concurrently
	numEvents := 10
	wg.Add(numEvents)

	for i := 0; i < numEvents; i++ {
		go func(id int) {
			defer wg.Done()
			event := StatusChangeEvent{
				ServiceID: id,
				OldStatus: "UP",
				NewStatus: "DOWN",
				Timestamp: time.Now(),
			}
			bus.Publish(ctx, event)
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond) // Wait for handlers to complete

	mu.Lock()
	assert.Equal(t, numEvents, eventCount)
	mu.Unlock()
}

func TestEventBus_DifferentEventTypes(t *testing.T) {
	logger := zap.NewNop()
	bus := NewInMemoryEventBus(logger)
	ctx := context.Background()

	statusChangeChan := make(chan Event, 1)
	otherEventChan := make(chan Event, 1)

	// Subscribe to StatusChange
	bus.Subscribe("StatusChange", func(ctx context.Context, event Event) {
		statusChangeChan <- event
	})

	// Subscribe to different event type
	bus.Subscribe("OtherEvent", func(ctx context.Context, event Event) {
		otherEventChan <- event
	})

	// Publish StatusChange event
	statusEvent := StatusChangeEvent{
		ServiceID: 5,
		OldStatus: "UP",
		NewStatus: "DOWN",
		Timestamp: time.Now(),
	}
	bus.Publish(ctx, statusEvent)

	// Only StatusChange subscriber should receive
	select {
	case <-statusChangeChan:
	case <-time.After(1 * time.Second):
		t.Fatal("StatusChange subscriber should receive event")
	}

	// OtherEvent subscriber should not receive
	select {
	case <-otherEventChan:
		t.Fatal("OtherEvent subscriber should not receive StatusChange event")
	case <-time.After(100 * time.Millisecond):
		// Expected - no event received
	}
}
