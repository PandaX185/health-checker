package monitor

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

type mockConn struct {
	writeData []byte
	mu        sync.Mutex
}

func (m *mockConn) Write(data []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeData = append(m.writeData, data...)
	return len(data), nil
}

func (m *mockConn) Read(data []byte) (int, error) {
	time.Sleep(100 * time.Millisecond)
	return 0, nil
}

func (m *mockConn) Close() error {
	return nil
}

func TestNewWsHub(t *testing.T) {
	logger := zap.NewNop()
	hub := NewWsHub(logger)

	assert.NotNil(t, hub)
	assert.NotNil(t, hub.clients)
	assert.NotNil(t, hub.broadcast)
	assert.NotNil(t, hub.register)
	assert.NotNil(t, hub.unregister)
}

func TestWsHub_Run(t *testing.T) {
	logger := zap.NewNop()
	hub := NewWsHub(logger)
	ctx := context.Background()

	// Start hub in background
	go hub.Run(ctx)

	// Create a mock client
	mockWsConn := &websocket.Conn{}
	client := &WsClient{
		conn: mockWsConn,
		send: make(chan []byte, 256),
		hub:  hub,
	}

	// Register client
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	// Check client is registered
	assert.Contains(t, hub.clients, client)

	// Unregister client
	hub.unregister <- client
	time.Sleep(50 * time.Millisecond)

	// Check client is unregistered
	assert.NotContains(t, hub.clients, client)
}

func TestWsHub_Broadcast(t *testing.T) {
	logger := zap.NewNop()
	hub := NewWsHub(logger)
	ctx := context.Background()

	// Start hub
	go hub.Run(ctx)

	// Create mock clients
	client1 := &WsClient{
		send: make(chan []byte, 256),
		hub:  hub,
	}
	client2 := &WsClient{
		send: make(chan []byte, 256),
		hub:  hub,
	}

	// Register clients
	hub.register <- client1
	hub.register <- client2
	time.Sleep(50 * time.Millisecond)

	// Broadcast message
	message := []byte("test message")
	hub.broadcast <- message
	time.Sleep(50 * time.Millisecond)

	// Both clients should receive message
	select {
	case msg := <-client1.send:
		assert.Equal(t, message, msg)
	case <-time.After(1 * time.Second):
		t.Fatal("Client 1 should receive broadcast")
	}

	select {
	case msg := <-client2.send:
		assert.Equal(t, message, msg)
	case <-time.After(1 * time.Second):
		t.Fatal("Client 2 should receive broadcast")
	}
}

func TestWsHub_BroadcastStatusChange(t *testing.T) {
	logger := zap.NewNop()
	hub := NewWsHub(logger)
	ctx := context.Background()

	// Start hub
	go hub.Run(ctx)

	// Create mock client
	client := &WsClient{
		send: make(chan []byte, 256),
		hub:  hub,
	}

	// Register client
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	// Broadcast status change
	event := StatusChangeEvent{
		ServiceID: 1,
		OldStatus: "UP",
		NewStatus: "DOWN",
		Timestamp: time.Now(),
	}

	err := hub.BroadcastStatusChange(event)
	assert.NoError(t, err)

	// Client should receive serialized event
	select {
	case msg := <-client.send:
		assert.Contains(t, string(msg), "ServiceID")
		assert.Contains(t, string(msg), "OldStatus")
		assert.Contains(t, string(msg), "NewStatus")
	case <-time.After(1 * time.Second):
		t.Fatal("Client should receive status change event")
	}
}

func TestWsHub_ClientDisconnect(t *testing.T) {
	logger := zap.NewNop()
	hub := NewWsHub(logger)
	ctx := context.Background()

	// Start hub
	go hub.Run(ctx)

	// Create mock client
	client := &WsClient{
		send: make(chan []byte, 256),
		hub:  hub,
	}

	// Register client
	hub.register <- client
	time.Sleep(50 * time.Millisecond)
	assert.Contains(t, hub.clients, client)

	// Unregister client properly
	hub.unregister <- client
	time.Sleep(50 * time.Millisecond)

	// Client should be removed
	assert.NotContains(t, hub.clients, client)
}

func TestWsHub_MultipleClientsOneDisconnects(t *testing.T) {
	logger := zap.NewNop()
	hub := NewWsHub(logger)
	ctx := context.Background()

	// Start hub
	go hub.Run(ctx)

	// Create two clients
	client1 := &WsClient{
		send: make(chan []byte, 256),
		hub:  hub,
	}
	client2 := &WsClient{
		send: make(chan []byte, 256),
		hub:  hub,
	}

	// Register both
	hub.register <- client1
	hub.register <- client2
	time.Sleep(50 * time.Millisecond)

	assert.Contains(t, hub.clients, client1)
	assert.Contains(t, hub.clients, client2)

	// Disconnect client1
	hub.unregister <- client1
	time.Sleep(50 * time.Millisecond)

	assert.NotContains(t, hub.clients, client1)
	assert.Contains(t, hub.clients, client2)

	// Broadcast should only reach client2
	hub.broadcast <- []byte("test")
	time.Sleep(50 * time.Millisecond)

	// Client1 channel should not receive (it might be closed or empty)
	select {
	case <-client2.send:
		// Expected - client2 receives
	case <-time.After(1 * time.Second):
		t.Fatal("Client2 should receive message")
	}
}
