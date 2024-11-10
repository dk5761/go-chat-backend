package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/dk5761/go-serv/internal/domain/chat/models"
	"github.com/dk5761/go-serv/internal/domain/chat/repository"
	"github.com/dk5761/go-serv/internal/infrastructure/logging"
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

	go m.listenToClient(client)
	m.deliverUndeliveredMessages(client)
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

	client, exists := m.clients[receiverID]
	if !exists {
		return errors.New("receiver not connected")
	}

	if !exists {
		// Receiver is offline, store the message as undelivered
		_, err := m.msgRepo.StoreUndeliveredMessage(context.Background(), message)
		if err != nil {
			logging.Logger.Error("Failed to store undelivered message",
				zap.String("receiver_id", receiverID),
				zap.Error(err),
			)
			return err
		}
		log.Printf("Stored undelivered message for offline client %s", receiverID)
		return nil
	}

	// Receiver is online, send the message directly
	select {
	case client.SendCh <- message:
		m.markMessageAsDelivered(message.ID)
		return nil
	default:
		// If the SendCh is full, mark the message as undelivered
		_, err := m.msgRepo.StoreUndeliveredMessage(context.Background(), message)
		if err != nil {
			logging.Logger.Error("Failed to store undelivered message for full channel",
				zap.String("receiver_id", receiverID),
				zap.Error(err),
			)
		}
		return fmt.Errorf("client %s SendCh is full", receiverID)
	}
}

func (m *WebSocketManager) deliverUndeliveredMessages(client *models.Client) {
	undeliveredMessages, err := m.msgRepo.GetUndeliveredMessages(context.Background(), client.ID)
	if err != nil {
		logging.Logger.Error("Failed to fetch undelivered messages", zap.Error(err))
		return
	}

	for _, message := range undeliveredMessages {
		select {
		case client.SendCh <- message:
			// Mark message as delivered after sending
			m.markMessageAsDelivered(message.ID)
		default:
			log.Printf("Failed to send message to client %s; SendCh full", client.ID)
		}
	}
}

// markMessageAsDelivered updates a message's status to delivered in MongoDB
func (m *WebSocketManager) markMessageAsDelivered(messageID primitive.ObjectID) {
	if err := m.msgRepo.MarkMessageAsDelivered(context.Background(), messageID); err != nil {
		logging.Logger.Error("Failed to mark message as delivered", zap.Error(err))
	}
}

func (m *WebSocketManager) listenToClient(client *models.Client) {
	defer func() {
		m.RemoveClient(client.ID)
		_ = client.Conn.Close() // Ensure the connection is closed when done
	}()

	for {
		// Read incoming message from the WebSocket connection
		_, msgData, err := client.Conn.ReadMessage()
		if err != nil {
			logging.Logger.Error("Error reading message", zap.String("client_id", client.ID),
				zap.Error(err))
			break
		}

		// Unmarshal the message to determine the event type
		var message models.Message
		if err := json.Unmarshal(msgData, &message); err != nil {
			logging.Logger.Error("Failed to unmarshal message",
				zap.String("client_id", client.ID),
				zap.Error(err),
			)
			continue
		}

		// Process the message based on its EventType
		switch message.EventType {
		case "send_message":
			// Standard message event
			message.SenderID = client.ID
			fmt.Println("client_id", client.ID)
			message.EventType = "receive_message"
			if err := m.processMessage(client, &message); err != nil {
				logging.Logger.Error("Error processing send_message event",
					zap.String("client_id", client.ID),
					zap.Error(err),
				)
			} else {
				// Send back acknowledgment to sender with message details
				m.sendAcknowledgment(client, &message)
			}

		case "typing":
			// Notify the receiver that the sender is typing
			if err := m.notifyTypingEvent(message.ReceiverID, &message); err != nil {
				logging.Logger.Error("Error processing typing event",
					zap.String("client_id", client.ID),
					zap.Error(err),
				)
			}

		case "join":
			// Handle user joining the chat
			if err := m.handleUserJoin(client); err != nil {
				logging.Logger.Error("Error handling join event",
					zap.String("client_id", client.ID),
					zap.Error(err),
				)
			}

		case "leave":
			// Handle user leaving the chat
			if err := m.handleUserLeave(client); err != nil {

				logging.Logger.Error("Error handling leave event",
					zap.String("client_id", client.ID),
					zap.Error(err),
				)
			}
			return // Exit the loop if the client leaves

		default:
			logging.Logger.Error("Unhandled event type",
				zap.String("client_id", client.ID),
				zap.String("event_type", message.EventType),
			)
		}
	}
}

func (m *WebSocketManager) sendAcknowledgment(client *models.Client, message *models.Message) {
	ackMessage := &models.Message{
		ID:         message.ID,
		SenderID:   message.SenderID,
		ReceiverID: message.ReceiverID,
		Content:    message.Content,
		EventType:  "message_acknowledgment",
		CreatedAt:  message.CreatedAt,
	}

	select {
	case client.SendCh <- ackMessage:
		log.Printf("Acknowledgment sent to client %s", client.ID)
	default:
		log.Printf("Acknowledgment failed to send; SendCh full for client %s", client.ID)
	}
}

// processMessage handles the received message and routes it as needed
func (m *WebSocketManager) processMessage(client *models.Client, message *models.Message) error {

	message.SenderID = client.ID
	message.CreatedAt = time.Now()

	// Save message to the database
	id, err := m.msgRepo.SaveMessage(context.Background(), message)
	if err != nil {
		log.Printf("Failed to save message from client %s: %v\n", client.ID, err)
		return err
	}
	// You can add logic to handle different message types here
	// Example: Sending message to another client by ID

	message.ID = id
	if message.ReceiverID != "" {
		return m.SendToClient(message.ReceiverID, message)
	}
	return nil
}

func (m *WebSocketManager) notifyTypingEvent(receiverID string, msg *models.Message) error {
	m.mu.Lock()
	receiver, exists := m.clients[receiverID]
	m.mu.Unlock()

	if !exists {
		logging.Logger.Error("client not connected", zap.String("receiver_id", receiverID))
		return nil
	}

	select {
	case receiver.SendCh <- msg:
		return nil
	default:
		logging.Logger.Error("SendCh is full", zap.String("client ID", receiverID))
		return nil

	}
}

// handleUserJoin performs actions when a user joins
func (m *WebSocketManager) handleUserJoin(client *models.Client) error {
	log.Printf("User %s has joined", client.ID)
	// Additional join logic here
	return nil
}

// handleUserLeave performs cleanup when a user leaves
func (m *WebSocketManager) handleUserLeave(client *models.Client) error {
	log.Printf("User %s has left", client.ID)
	m.RemoveClient(client.ID)
	return nil
}
