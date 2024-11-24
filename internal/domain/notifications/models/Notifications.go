package models

import (
	"time"
)

type NotificationStatus string

const (
	StatusPending NotificationStatus = "PENDING"
	StatusSent    NotificationStatus = "SENT"
	StatusFailed  NotificationStatus = "FAILED"
)

type Notification struct {
	ID         string             `bson:"_id"`
	Token      string             `bson:"token"`
	Title      string             `bson:"title"`
	Body       string             `bson:"body"`
	Data       map[string]string  `bson:"data"`
	CreatedAt  time.Time          `bson:"created_at"`
	RetryCount int                `bson:"retry_count"`
	Status     NotificationStatus `bson:"status"`
}
