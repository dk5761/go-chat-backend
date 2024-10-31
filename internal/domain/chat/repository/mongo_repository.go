package repository

import (
	"context"
	"time"

	"github.com/dk5761/go-serv/internal/domain/chat/models"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoMessageRepository struct {
	collection *mongo.Collection
}

func NewMongoMessageRepository(db *mongo.Database) MessageRepository {
	return &mongoMessageRepository{
		collection: db.Collection("messages"),
	}
}

// SaveMessage saves a new message to the MongoDB collection.
func (r *mongoMessageRepository) SaveMessage(ctx context.Context, msg *models.Message) error {
	// Set the creation time
	msg.CreatedAt = time.Now()

	// Insert the message into the collection
	_, err := r.collection.InsertOne(ctx, msg)
	if err != nil {
		return err
	}
	return nil
}

// GetMessages retrieves messages between two users, sorted by creation time.
func (r *mongoMessageRepository) GetMessages(ctx context.Context, userID1, userID2 uuid.UUID) ([]*models.Message, error) {
	// Create a filter to get messages where sender and receiver match
	filter := bson.M{
		"$or": []bson.M{
			{
				"sender_id":   userID1.String(),
				"receiver_id": userID2.String(),
			},
			{
				"sender_id":   userID2.String(),
				"receiver_id": userID1.String(),
			},
		},
	}

	// Define options to sort the messages by created_at
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"created_at", 1}}) // 1 for ascending order

	// Execute the query
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Iterate over the cursor and decode each message
	var messages []*models.Message
	for cursor.Next(ctx) {
		var msg models.Message
		if err := cursor.Decode(&msg); err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}
