package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/dk5761/go-serv/internal/domain/chat/models"
	"github.com/google/uuid"
)

type MessageRepository interface {
	SaveMessage(ctx context.Context, msg *models.Message) (primitive.ObjectID, error)
	GetMessages(ctx context.Context, userID1, userID2 uuid.UUID, limit, offset int) ([]*models.Message, error)
}
