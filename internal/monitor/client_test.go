package monitor

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

func TestNewWsClient(t *testing.T) {
	conn := &websocket.Conn{}
	hub := &WsHub{}

	client := NewWsClient(conn, hub, zap.NewNop())

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
	// ReadPump is an infinite loop that requires a WebSocket connection
	// We can only test that the function exists and the client is properly initialized
	hub := NewWsHub(zap.NewNop())
	client := NewWsClient(nil, hub, zap.NewNop())

	assert.NotNil(t, client)
	assert.NotNil(t, client.ReadPump)
	// Note: Full testing would require WebSocket server setup
}

func TestWsClient_WritePump(t *testing.T) {
	// WritePump is an infinite loop that requires a WebSocket connection
	hub := NewWsHub(zap.NewNop())
	client := NewWsClient(nil, hub, zap.NewNop())

	assert.NotNil(t, client)
	assert.NotNil(t, client.WritePump)
	// Note: Full testing would require WebSocket server setup
}

func TestWsClient_Integration(t *testing.T) {
	// Setup Hub
	hub := NewWsHub(zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	// Setup Server
	server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		client := NewWsClient(ws, hub, zap.NewNop())
		hub.register <- client
		go client.WritePump()
		client.ReadPump()
	}))
	defer server.Close()

	// Connect Client
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, err := websocket.Dial(url, "", server.URL)
	assert.NoError(t, err)
	defer ws.Close()

	// Send message (client -> server)
	_, err = ws.Write([]byte("hello"))
	assert.NoError(t, err)

	// Wait a bit for processing
	time.Sleep(50 * time.Millisecond)

	// Verify client is registered
	assert.NotEmpty(t, hub.clients)

	// Broadcast message (server -> client)
	hub.broadcast <- []byte("broadcast")

	// Read message
	msg := make([]byte, 512)
	n, err := ws.Read(msg)
	assert.NoError(t, err)
	assert.Equal(t, "broadcast", string(msg[:n]))
}
