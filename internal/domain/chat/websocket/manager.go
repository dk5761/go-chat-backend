package websocket

import (
	"errors"
	"sync"

	"github.com/dk5761/go-serv/internal/domain/chat/models"
)

type WebSocketManager struct {
	clients map[string]*models.Client // Map userID to WebSocket client
	mu      sync.RWMutex
}

func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{clients: make(map[string]*models.Client)}
}

// AddClient adds a new client to the manager
func (m *WebSocketManager) AddClient(client *models.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[client.ID] = client
	go client.Listen()
}

// RemoveClient removes a client from the manager
func (m *WebSocketManager) RemoveClient(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, userID)
}

// SendToClient sends a message to the specified user if they are connected
func (m *WebSocketManager) SendToClient(receiverID string, message *models.Message) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[receiverID]
	if !exists {
		return errors.New("receiver not connected")
	}

	client.SendCh <- message
	return nil
}
