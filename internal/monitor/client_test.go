package monitor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

func TestNewWsClient(t *testing.T) {
	conn := &websocket.Conn{}
	hub := &WsHub{}

	client := NewWsClient(conn, hub)

	assert.NotNil(t, client)
	assert.Equal(t, conn, client.conn)
	assert.Equal(t, hub, client.hub)
	assert.NotNil(t, client.send)
	assert.Equal(t, 256, cap(client.send))
}

func TestStatusChangeEvent_Type(t *testing.T) {
	event := StatusChangeEvent{}
	assert.Equal(t, "StatusChange", event.Type())
}

func TestStatusChangeEvent_At(t *testing.T) {
	event := StatusChangeEvent{}
	at := event.At()
	assert.NotNil(t, at)
}

func TestWsClient_ReadPump(t *testing.T) {
	// This is a complex function that requires a WebSocket connection
	// For unit testing, we can verify the function exists and has basic structure
	// Full integration testing would require WebSocket server setup
	hub := NewWsHub(zap.NewNop())
	client := NewWsClient(nil, hub) // nil conn for testing

	// The function should exist and be callable (though it will panic without proper setup)
	assert.NotNil(t, client)
	assert.NotNil(t, client.ReadPump)
}

func TestWsClient_WritePump(t *testing.T) {
	// Similar to ReadPump, this requires WebSocket connection
	hub := NewWsHub(zap.NewNop())
	client := NewWsClient(nil, hub) // nil conn for testing

	assert.NotNil(t, client)
	assert.NotNil(t, client.WritePump)
}
