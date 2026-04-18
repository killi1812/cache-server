package service

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Hub manages active WebSocket connections for agents.
type Hub struct {
	mu     sync.RWMutex
	agents map[string]*websocket.Conn
}

func NewHub() *Hub {
	return &Hub{
		agents: make(map[string]*websocket.Conn),
	}
}

// Register adds an agent connection to the hub.
func (h *Hub) Register(name string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if old, ok := h.agents[name]; ok {
		old.Close()
	}

	h.agents[name] = conn
	zap.S().Infof("Agent '%s' registered in Hub", name)
}

// Unregister removes an agent connection from the hub.
func (h *Hub) Unregister(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.agents[name]; ok {
		delete(h.agents, name)
		zap.S().Infof("Agent '%s' unregistered from Hub", name)
	}
}

// NotifyAgent sends a JSON message to a specific agent.
func (h *Hub) NotifyAgent(name string, msg any) error {
	h.mu.RLock()
	conn, ok := h.agents[name]
	h.mu.RUnlock()

	if !ok {
		zap.S().Errorf("agent '%s' is not connected", name)
		return ErrAgentNotConnected
	}

	return conn.WriteJSON(msg)
}

var ErrAgentNotConnected = errors.New("agent is not connected")
