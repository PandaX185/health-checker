package monitor

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

type EventHandler func(ctx context.Context, event Event)

type EventBus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(eventType string, handler EventHandler)
}

type InMemoryEventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]EventHandler
	log         *zap.Logger
}

func NewInMemoryEventBus(log *zap.Logger) EventBus {
	return &InMemoryEventBus{
		subscribers: make(map[string][]EventHandler),
		log:         log,
	}
}

func (bus *InMemoryEventBus) Publish(ctx context.Context, event Event) error {
	bus.mu.RLock()
	handlers, exists := bus.subscribers[event.Type()]
	bus.mu.RUnlock()

	if !exists {
		return nil
	}

	for _, handler := range handlers {
		go safeHandle(ctx, handler, event, bus.log)
	}
	return nil
}

func (bus *InMemoryEventBus) Subscribe(eventType string, handler EventHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.subscribers[eventType] = append(bus.subscribers[eventType], handler)
}

func safeHandle(ctx context.Context, handler EventHandler, event Event, log *zap.Logger) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("panic recovered in event handler",
				zap.String("event_type", event.Type()),
				zap.Any("error", r),
			)
		}
	}()
	handler(ctx, event)
}
