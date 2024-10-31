package repository

import (
	"context"

	"github.com/dk5761/go-serv/internal/domain/chat/models"
	"github.com/google/uuid"
)

type MessageRepository interface {
	SaveMessage(ctx context.Context, msg *models.Message) error
	GetMessages(ctx context.Context, userID1, userID2 uuid.UUID) ([]*models.Message, error)
}
