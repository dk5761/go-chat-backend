package repository

import (
	"context"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/dk5761/go-serv/internal/domain/chat/models"
)

type MessageRepository interface {
	SaveMessage(ctx context.Context, msg *models.Message) (primitive.ObjectID, error)
	GetMessages(ctx context.Context, userID1, userID2 uuid.UUID, limit, offset int) ([]*models.Message, error)
	GetUndeliveredMessages(ctx context.Context, receiverID string) ([]*models.Message, error)
	MarkMessageAsDelivered(ctx context.Context, messageID primitive.ObjectID) error
	StoreUndeliveredMessage(ctx context.Context, msg *models.Message) (primitive.ObjectID, error)
	UpdateMessageStatus(ctx context.Context, messageID primitive.ObjectID, status models.MessageStatus) error
	GetMessage(ctx context.Context, messageID primitive.ObjectID) (*models.Message, error)
	MarkAcknowledgmentPending(ctx context.Context, messageID primitive.ObjectID) error
	GetPendingAcknowledgments(ctx context.Context, receiverID string) ([]*models.Message, error)
}
