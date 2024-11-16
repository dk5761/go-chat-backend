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
	go m.deliverUndeliveredMessages(client)
	go m.sendPendingMessages(client)
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
			message.Status = models.Stored
			message.SenderID = client.ID
			message.EventType = "receive_message"

			messageID, err := m.msgRepo.SaveMessage(context.Background(), &message)
			if err != nil {
				logging.Logger.Error("Error saving message", zap.Error(err))
				continue
			}
			message.ID = messageID

			// Send acknowledgment back to sender client
			m.sendAcknowledgment(&message, models.Stored)

			// Try delivering to receiver if connected
			if err := m.SendToClient(message.ReceiverID, &message); err == nil {
				// Update message status to Delivered in DB
				m.msgRepo.UpdateMessageStatus(context.Background(), messageID, models.Sent)
			}

		case "ack_received":
			// Handle acknowledgment from receiver client
			messageID := message.ID // Assuming message ID is provided in acknowledgment
			// update the local message body with the Receieved status

			if err := m.msgRepo.UpdateMessageStatus(context.Background(), messageID, models.Received); err != nil {
				logging.Logger.Error("Error updating message status", zap.Error(err))
				continue
			}
			message.Status = models.Received

			if err := m.msgRepo.MarkMessageAsDelivered(context.Background(), messageID); err != nil {
				logging.Logger.Error("Error updating message status", zap.Error(err))
				continue
			}

			message.Delivered = true // Update the lmb with Deliver == true and time.
			message.DeliveredAt = time.Now()

			m.sendAcknowledgment(&message, models.Received)

		default:
			logging.Logger.Error("Unhandled event type",
				zap.String("client_id", client.ID),
				zap.String("event_type", message.EventType),
			)
		}
	}
}

func (m *WebSocketManager) sendAcknowledgment(message *models.Message, status models.MessageStatus) {
	m.mu.RLock()
	originalSenderClient, exists := m.clients[message.SenderID]
	m.mu.RUnlock()

	ackMessage := &models.Message{
		ID:          message.ID,
		SenderID:    message.SenderID,
		ReceiverID:  message.ReceiverID,
		EventType:   "acknowledgment",
		TempID:      message.TempID,
		Status:      status, // Send the status as acknowledgment type
		CreatedAt:   time.Now(),
		Delivered:   message.Delivered,
		DeliveredAt: message.DeliveredAt,
		Content:     message.Content,
		FileURL:     message.FileURL,
	}

	if exists {
		// If client is connected, send the acknowledgment over WebSocket
		select {
		case originalSenderClient.SendCh <- ackMessage:
			log.Printf("Acknowledgment sent to client %s", message.SenderID)
		default:
			log.Printf("SendCh is full; acknowledgment not sent to client %s", message.SenderID)
			// Optionally, mark acknowledgment as undelivered in the database for retry
			if err := m.msgRepo.MarkAcknowledgmentPending(context.Background(), message.ID); err != nil {
				logging.Logger.Error("Failed to mark acknowledgment as pending", zap.Error(err))
			}
		}
	} else {
		// If client is offline, store acknowledgment status as pending in the database
		if err := m.msgRepo.MarkAcknowledgmentPending(context.Background(), message.ID); err != nil {
			logging.Logger.Error("Failed to mark acknowledgment as pending", zap.Error(err))
		}
		log.Printf("Client %s is offline; acknowledgment stored as pending", message.SenderID)
	}
}

// sendPendingMessages retrieves and sends any pending messages to the reconnected client
func (m *WebSocketManager) sendPendingMessages(client *models.Client) {
	// Retrieve pending messages for this client
	messages, err := m.msgRepo.GetPendingAcknowledgments(context.Background(), client.ID)

	if err != nil {
		logging.Logger.Error("Failed to retrieve pending messages", zap.String("client_id", client.ID), zap.Error(err))
		return
	}

	for _, message := range messages {
		// Send each pending message to the client\

		ackMessage := &models.Message{
			ID:          message.ID,
			SenderID:    message.SenderID,
			ReceiverID:  message.ReceiverID,
			EventType:   "acknowledgment",
			TempID:      message.TempID,
			Status:      models.Received, // Send the status as acknowledgment type
			CreatedAt:   time.Now(),
			Delivered:   message.Delivered,
			DeliveredAt: message.DeliveredAt,
			Content:     message.Content,
			FileURL:     message.FileURL,
		}
		client.SendCh <- ackMessage
		messageID := message.ID // Assuming message ID is provided in acknowledgment

		fmt.Println("data", messageID)
		if err := m.msgRepo.UpdateMessageStatus(context.Background(), messageID, models.Received); err != nil {
			logging.Logger.Error("Error updating message status", zap.Error(err))
			continue
		}
		// Mark as delivered if sent successfully
		if err := m.msgRepo.MarkMessageAsDelivered(context.Background(), message.ID); err != nil {
			logging.Logger.Error("Error marking message as delivered", zap.String("client_id", client.ID), zap.Error(err))
		}
	}
}

// processMessage handles the received message and routes it as needed
// func (m *WebSocketManager) processMessage(client *models.Client, message *models.Message) error {

// 	message.SenderID = client.ID
// 	message.CreatedAt = time.Now()

// 	// Save message to the database
// 	id, err := m.msgRepo.SaveMessage(context.Background(), message)
// 	if err != nil {
// 		log.Printf("Failed to save message from client %s: %v\n", client.ID, err)
// 		return err
// 	}
// 	// You can add logic to handle different message types here
// 	// Example: Sending message to another client by ID

// 	message.ID = id
// 	if message.ReceiverID != "" {
// 		return m.SendToClient(message.ReceiverID, message)
// 	}
// 	return nil
// }
