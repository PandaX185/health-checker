package monitor

import (
	"io"
	"time"

	"golang.org/x/net/websocket"
)

type WsClient struct {
	conn *websocket.Conn
	send chan []byte
	hub  *WsHub
}

func NewWsClient(conn *websocket.Conn, hub *WsHub) *WsClient {
	return &WsClient{
		conn: conn,
		send: make(chan []byte, 256),
		hub:  hub,
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
				break
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
				return
			}
			if err := websocket.Message.Send(c.conn, string(message)); err != nil {
				return
			}
		case <-ticker.C:
			if err := websocket.Message.Send(c.conn, "ping"); err != nil {
				return
			}
		}
	}
}
