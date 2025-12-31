package monitor

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"
)

type WsHub struct {
	clients    map[*WsClient]bool
	broadcast  chan []byte
	register   chan *WsClient
	unregister chan *WsClient
	log        *zap.Logger
}

func NewWsHub(log *zap.Logger) *WsHub {
	return &WsHub{
		clients:    make(map[*WsClient]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *WsClient),
		unregister: make(chan *WsClient),
		log:        log,
	}
}

func (h *WsHub) Run(ctx context.Context) {
	h.log.Info("WebSocket hub is running")
	for {
		select {
		case <-ctx.Done():
			h.shutdown()
			return
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *WsHub) shutdown() {
	for client := range h.clients {
		close(client.send)
		delete(h.clients, client)
	}
}

func (h *WsHub) BroadcastStatusChange(event StatusChangeEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	h.broadcast <- data
	return nil
}
