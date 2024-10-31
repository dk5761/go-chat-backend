package service

import (
	"context"
	"mime/multipart"

	"github.com/dk5761/go-serv/internal/domain/chat/models"
	"github.com/google/uuid"
)

type ChatService interface {
	SendMessage(ctx context.Context, msg *models.Message) error
	GetChatHistory(ctx context.Context, userID1, userID2 uuid.UUID) ([]*models.Message, error)
	UploadFile(ctx context.Context, file multipart.File, fileName string) (string, error)
}