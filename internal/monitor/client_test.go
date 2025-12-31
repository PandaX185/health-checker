package monitor

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
