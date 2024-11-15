package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationStatus string

const (
	StatusPending   NotificationStatus = "pending"
	StatusDelivered NotificationStatus = "delivered"
	StatusFailed    NotificationStatus = "failed"
)

type Notification struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	Title       string             `bson:"title" json:"title"`
	Body        string             `bson:"body" json:"body"`
	DeviceToken string             `bson:"device_token" json:"device_token"`
	Data        map[string]string  `bson:"data,omitempty" json:"data,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	Status      NotificationStatus `bson:"status" json:"status"`
}
