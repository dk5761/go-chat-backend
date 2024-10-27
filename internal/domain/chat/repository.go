package chat

import (
	"context"

	"github.com/google/uuid"
)

type MessageRepository interface {
	SaveMessage(ctx context.Context, msg *Message) error
	GetMessages(ctx context.Context, userID1, userID2 uuid.UUID) ([]*Message, error)
}
