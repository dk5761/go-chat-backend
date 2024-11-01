package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/dk5761/go-serv/internal/domain/chat/models"
	"github.com/dk5761/go-serv/internal/domain/chat/repository"
)

type WebSocketManager struct {
	clients map[string]*models.Client // Map userID to WebSocket client
	mu      sync.RWMutex
	msgRepo repository.MessageRepository
}

func NewWebSocketManager(msgRepo repository.MessageRepository) *WebSocketManager {
	return &WebSocketManager{
		clients: make(map[string]*models.Client),
		msgRepo: msgRepo,
	}
}

// AddClient adds a new client to the manager
func (m *WebSocketManager) AddClient(client *models.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[client.ID] = client
	go client.Listen()
	fmt.Println("inside ListenToClient")

	go m.listenToClient(client)
}

func (m *WebSocketManager) RemoveClient(clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if client, ok := m.clients[clientID]; ok {
		err := client.Conn.Close()
		if err != nil {
			return
		}
		delete(m.clients, clientID)
	}
}

// SendToClient sends a message to the specified user if they are connected
func (m *WebSocketManager) SendToClient(receiverID string, message *models.Message) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fmt.Println("inside SendToClient")

	client, exists := m.clients[receiverID]
	if !exists {
		return errors.New("receiver not connected")
	}

	client.SendCh <- message
	return nil
}

func (m *WebSocketManager) listenToClient(client *models.Client) {
	defer m.RemoveClient(client.ID)

	fmt.Println("inside ListenToClient")

	for {
		_, msgData, err := client.Conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message for client %s: %v\n", client.ID, err)
			break
		}

		var message models.Message
		if err := json.Unmarshal(msgData, &message); err != nil {
			log.Println("Failed to unmarshal message:", err)
			continue
		}

		if err := m.processMessage(client, &message); err != nil {
			log.Printf("Error processing message from client %s: %v\n", client.ID, err)
		}
	}
}

// processMessage handles the received message and routes it as needed
func (m *WebSocketManager) processMessage(client *models.Client, message *models.Message) error {

	message.SenderID = client.ID

	fmt.Println("inside ProcessMessage")

	// Save message to the database
	if err := m.msgRepo.SaveMessage(context.Background(), message); err != nil {
		log.Printf("Failed to save message from client %s: %v\n", client.ID, err)
		return err
	}
	// You can add logic to handle different message types here
	// Example: Sending message to another client by ID
	if message.ReceiverID != "" {
		return m.SendToClient(message.ReceiverID, message)
	}
	return nil
}
