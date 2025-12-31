package monitor

import (
	"io"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

type WsClient struct {
	conn   *websocket.Conn
	send   chan []byte
	hub    *WsHub
	logger *zap.Logger
}

func NewWsClient(conn *websocket.Conn, hub *WsHub, logger *zap.Logger) *WsClient {
	return &WsClient{
		conn:   conn,
		send:   make(chan []byte, 256),
		hub:    hub,
		logger: logger,
	}
}

func (c *WsClient) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		var msg string
		err := websocket.Message.Receive(c.conn, &msg)
		if err != nil {
			if err != io.EOF {
				c.logger.Error("websocket read error", zap.Error(err))
			}
			break
		}
	}
}

func (c *WsClient) WritePump() {
	ticker := time.NewTicker(50 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteClose(500)
				return
			}
			if err := websocket.Message.Send(c.conn, string(message)); err != nil {
				c.logger.Error("websocket write error", zap.Error(err))
				return
			}
		case <-ticker.C:
			if err := websocket.Message.Send(c.conn, "ping"); err != nil {
				c.logger.Error("websocket ping error", zap.Error(err))
				return
			}
		}
	}
}
