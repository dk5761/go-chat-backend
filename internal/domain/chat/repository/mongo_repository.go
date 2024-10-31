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

// NewMongoMessageRepository initializes a new instance of mongoMessageRepository
func NewMongoMessageRepository(db *mongo.Database) MessageRepository {
	return &mongoMessageRepository{
		collection: db.Collection("messages"),
	}
}

// SaveMessage saves a new message to the MongoDB collection
func (r *mongoMessageRepository) SaveMessage(ctx context.Context, msg *models.Message) error {
	// Set the creation timestamp
	msg.Timestamp = time.Now()

	// Insert the message into MongoDB
	_, err := r.collection.InsertOne(ctx, msg)
	return err
}

// GetMessages retrieves messages between two users, sorted by creation time with pagination support
func (r *mongoMessageRepository) GetMessages(ctx context.Context, userID1, userID2 uuid.UUID, limit, offset int) ([]*models.Message, error) {
	// Create a filter to match messages between userID1 and userID2
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

	// Define options to apply pagination and sorting by timestamp
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"timestamp", 1}}) // 1 for ascending order
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(offset))

	// Execute the query
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode each document in the cursor into a Message struct
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
