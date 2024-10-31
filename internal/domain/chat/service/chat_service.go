package service

import (
	"context"
	"errors"
	"mime/multipart"

	"github.com/dk5761/go-serv/internal/domain/chat/models"
	"github.com/dk5761/go-serv/internal/domain/chat/repository"
	"github.com/dk5761/go-serv/internal/infrastructure/storage"
	"github.com/google/uuid"
)

type chatService struct {
	msgRepo        repository.MessageRepository
	storageService storage.StorageService
}

func NewChatService(msgRepo repository.MessageRepository, storageService storage.StorageService) ChatService {
	return &chatService{msgRepo, storageService}
}

// Implement the ChatService methods here

func (s *chatService) UploadFile(ctx context.Context, file multipart.File, fileName string) (string, error) {
	return s.storageService.UploadFile(ctx, file, fileName)
}

func (s *chatService) SendMessage(ctx context.Context, msg *models.Message) error {
	// Validate the message content
	if msg.Content == "" {
		return errors.New("message content cannot be empty")
	}

	// Save the message using the repository
	return s.msgRepo.SaveMessage(ctx, msg)
}

func (s *chatService) GetChatHistory(ctx context.Context, userID1, userID2 uuid.UUID) ([]*models.Message, error) {
	// Retrieve messages from the repository
	return s.msgRepo.GetMessages(ctx, userID1, userID2)
}
