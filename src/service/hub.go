package service

import (
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Hub struct {
	mu     sync.RWMutex
	agents map[string]*websocket.Conn
}

func NewHub() *Hub {
	return &Hub{
		agents: make(map[string]*websocket.Conn),
	}
}

func (h *Hub) Register(name string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// If agent already connected, close old connection
	if old, ok := h.agents[name]; ok {
		old.Close()
	}
	
	h.agents[name] = conn
	zap.S().Infof("Agent '%s' registered via WebSocket", name)
}

func (h *Hub) Unregister(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.agents, name)
	zap.S().Infof("Agent '%s' disconnected", name)
}

func (h *Hub) NotifyAgent(name string, msg any) error {
	h.mu.RLock()
	conn, ok := h.agents[name]
	h.mu.RUnlock()

	if !ok {
		zap.S().Warnf("Agent '%s' not connected, notification skipped", name)
		return nil
	}

	err := conn.WriteJSON(msg)
	if err != nil {
		zap.S().Errorf("Failed to send notification to agent '%s': %v", name, err)
		return err
	}

	return nil
}
