package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/dk5761/go-serv/internal/domain/chat/models"
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
func (r *mongoMessageRepository) SaveMessage(ctx context.Context, msg *models.Message) (primitive.ObjectID, error) {
	// Set the creation timestamp
	msg.CreatedAt = time.Now()

	fmt.Println("save message", msg)

	// Insert the message into MongoDB
	result, err := r.collection.InsertOne(ctx, msg)
	if err != nil {
		fmt.Println("save message err", err)
		return primitive.NilObjectID, err
	}

	// Extract and return the ObjectID from the InsertedID field
	messageID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("failed to convert inserted ID to ObjectID")
	}

	return messageID, nil
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

// GetUndeliveredMessages retrieves undelivered messages for a given receiver ID.
func (r *mongoMessageRepository) GetUndeliveredMessages(ctx context.Context, receiverID string) ([]*models.Message, error) {
	fiveDaysAgo := time.Now().AddDate(0, 0, -5)

	// Define the filter to match undelivered messages for the receiver within the last 5 days
	filter := bson.M{
		"receiver_id": receiverID,
		"delivered":   false,
		"created_at": bson.M{
			"$gte": fiveDaysAgo, // Only include messages created within the last 5 days
		},
	}

	// Execute the query with the filter
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode the results into a slice of messages
	var messages []*models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// MarkMessageAsDelivered updates the delivery status of a message.
func (r *mongoMessageRepository) MarkMessageAsDelivered(ctx context.Context, messageID primitive.ObjectID) error {
	update := bson.M{
		"$set": bson.M{
			"delivered":    true,
			"delivered_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateByID(ctx, messageID, update, options.Update())
	return err
}

func (r *mongoMessageRepository) StoreUndeliveredMessage(ctx context.Context, msg *models.Message) (primitive.ObjectID, error) {
	// Set default fields for undelivered messages
	msg.CreatedAt = time.Now()
	msg.Delivered = false // Mark as undelivered

	// Insert the message into MongoDB
	result, err := r.collection.InsertOne(ctx, msg)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return result.InsertedID.(primitive.ObjectID), nil
}

func (r *mongoMessageRepository) UpdateMessageStatus(ctx context.Context, messageID primitive.ObjectID, status models.MessageStatus) error {
	filter := bson.M{"_id": messageID}
	update := bson.M{"$set": bson.M{"status": status}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *mongoMessageRepository) GetMessage(ctx context.Context, messageID primitive.ObjectID) (*models.Message, error) {
	// Define the filter for the message ID
	filter := bson.M{"_id": messageID}

	// Prepare a variable to hold the result
	var message models.Message

	// Execute the query
	err := r.collection.FindOne(ctx, filter).Decode(&message)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("message not found")
		}
		return nil, err
	}

	return &message, nil
}
