package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/dk5761/go-serv/internal/domain/notifications/models"
)

type NotificationRepository interface {
	Save(notification *models.Notification) error
	FindPendingNotifications() ([]*models.Notification, error)
	DeleteOldNotifications(hours int) error
}

type MongoNotificationRepository struct {
	collection *mongo.Collection
}

func NewMongoMessageRepository(db *mongo.Database) NotificationRepository {
	return &MongoNotificationRepository{
		collection: db.Collection("messages"),
	}
}

func (r *MongoNotificationRepository) Save(notification *models.Notification) error {
	ctx := context.Background()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": notification.ID},
		bson.M{"$set": notification},
		options.Update().SetUpsert(true),
	)
	return err
}

func (r *MongoNotificationRepository) FindPendingNotifications() ([]*models.Notification, error) {
	ctx := context.Background()
	cursor, err := r.collection.Find(ctx, bson.M{"status": models.StatusPending})
	if err != nil {
		return nil, err
	}

	var notifications []*models.Notification
	err = cursor.All(ctx, &notifications)
	return notifications, err
}

func (r *MongoNotificationRepository) DeleteOldNotifications(hours int) error {
	ctx := context.Background()
	threshold := time.Now().Add(-time.Hour * time.Duration(hours))
	_, err := r.collection.DeleteMany(ctx, bson.M{
		"created_at": bson.M{"$lt": threshold},
	})
	return err
}
