package service

import (
	"context"
	"mime/multipart"

	"github.com/dk5761/go-serv/internal/domain/chat/models"
	"github.com/google/uuid"
)

type ChatService interface {
	SendMessage(ctx context.Context, msg *models.Message, file multipart.File, fileName string) error
	GetChatHistory(ctx context.Context, userID1, userID2 uuid.UUID, limit, offset int) ([]*models.Message, error)
	UploadFile(ctx context.Context, file multipart.File, fileName string) (string, error)
	SendToClient(receiverID string, msg *models.Message) error
}
