package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageStatus string

const (
	Stored   MessageStatus = "stored"
	Sent     MessageStatus = "sent"
	Received MessageStatus = "received"
	Pending  MessageStatus = "pending"
	Read     MessageStatus = "read"
)

type Message struct {
	EventType   string             `bson:"event_type" json:"event_type"`
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TempID      string             `bson:"temp_id,omitempty" json:"temp_id,omitempty"`
	SenderID    string             `bson:"sender_id" json:"sender_id"`
	ReceiverID  string             `bson:"receiver_id" json:"receiver_id"`
	Content     string             `bson:"content" json:"content"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	FileURL     string             `bson:"file_url,omitempty" json:"file_url"`
	Delivered   bool               `bson:"delivered" json:"delivered"`
	DeliveredAt time.Time          `bson:"delivered_at,omitempty" json:"delivered_at,omitempty"`
	Status      MessageStatus      `bson:"status" json:"status"`
}
