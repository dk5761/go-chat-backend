package chat

import (
	"context"
	"errors"
	"mime/multipart"

	"github.com/dk5761/go-serv/internal/infrastructure/storage"
	"github.com/google/uuid"
)

type ChatService interface {
	SendMessage(ctx context.Context, msg *Message) error
	GetChatHistory(ctx context.Context, userID1, userID2 uuid.UUID) ([]*Message, error)
	UploadFile(ctx context.Context, file multipart.File, fileName string) (string, error)
}

type chatService struct {
	msgRepo        MessageRepository
	storageService storage.StorageService
}

func NewChatService(msgRepo MessageRepository, storageService storage.StorageService) ChatService {
	return &chatService{msgRepo, storageService}
}

// Implement the ChatService methods here

func (s *chatService) UploadFile(ctx context.Context, file multipart.File, fileName string) (string, error) {
	return s.storageService.UploadFile(ctx, file, fileName)
}

func (s *chatService) SendMessage(ctx context.Context, msg *Message) error {
	// Validate the message content
	if msg.Content == "" {
		return errors.New("message content cannot be empty")
	}

	// Save the message using the repository
	return s.msgRepo.SaveMessage(ctx, msg)
}

func (s *chatService) GetChatHistory(ctx context.Context, userID1, userID2 uuid.UUID) ([]*Message, error) {
	// Retrieve messages from the repository
	return s.msgRepo.GetMessages(ctx, userID1, userID2)
}
