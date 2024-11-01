package service

import (
	"context"
	"errors"
	"mime/multipart"
	"time"

	"github.com/dk5761/go-serv/internal/domain/chat/models"
	"github.com/dk5761/go-serv/internal/domain/chat/repository"
	"github.com/dk5761/go-serv/internal/domain/chat/websocket"
	"github.com/dk5761/go-serv/internal/infrastructure/storage"
	"github.com/google/uuid"
)

type chatService struct {
	msgRepo        repository.MessageRepository
	storageService storage.StorageService
	wsManager      *websocket.WebSocketManager
}

func NewChatService(msgRepo repository.MessageRepository, storageService storage.StorageService, wsManager *websocket.WebSocketManager) ChatService {
	return &chatService{msgRepo: msgRepo, storageService: storageService, wsManager: wsManager}
}

// UploadFile uploads a file and returns its URL.
func (s *chatService) UploadFile(ctx context.Context, file multipart.File, fileName string) (string, error) {
	return s.storageService.UploadFile(ctx, file, fileName)
}

// SendMessage validates and saves a message to the repository.
// If a file is attached, it uploads the file and saves the URL in the message.
func (s *chatService) SendMessage(ctx context.Context, msg *models.Message, file multipart.File, fileName string) error {

	// Set message timestamp
	msg.CreatedAt = time.Now()

	// Handle optional file upload
	if file != nil {
		fileURL, err := s.storageService.UploadFile(ctx, file, fileName)
		if err != nil {
			return errors.New("failed to upload file")
		}
		msg.FileURL = fileURL
	}

	// Save the message in the repository
	err := s.msgRepo.SaveMessage(ctx, msg)
	if err != nil {
		return err
	}

	// Send the message to the receiver over WebSocket
	// if msg.ReceiverID != "" {
	// 	err = s.wsManager.SendToClient(msg.ReceiverID, msg)
	// 	if err != nil {
	// 		fmt.Printf("Failed to send message over WebSocket: %v\n", err)
	// 		return errors.New("failed to send message over WebSocket")
	// 	}
	// }

	return nil
}

func (s *chatService) SendToClient(receiverID string, msg *models.Message) error {
	return s.wsManager.SendToClient(receiverID, msg)
}

// GetChatHistory retrieves messages between two users, supporting pagination for large histories.
func (s *chatService) GetChatHistory(ctx context.Context, userID1, userID2 uuid.UUID, limit, offset int) ([]*models.Message, error) {
	if limit <= 0 {
		limit = 20 // Default limit
	}

	// Retrieve messages from the repository with pagination
	return s.msgRepo.GetMessages(ctx, userID1, userID2, limit, offset)
}
