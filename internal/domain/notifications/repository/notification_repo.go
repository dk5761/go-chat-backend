package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	models "github.com/dk5761/go-serv/internal/domain/notifications/model"
)

type mongoNotificationRepository struct {
	collection *mongo.Collection
}

func NewMongoNotificationRepository(db *mongo.Database) NotificationRepository {
	return &mongoNotificationRepository{
		collection: db.Collection("notifications"),
	}
}

func (r *mongoNotificationRepository) SaveNotification(ctx context.Context, notification *models.Notification) error {
	_, err := r.collection.InsertOne(ctx, notification)
	if err != nil {
		return err
	}
	return nil
}

func (r *mongoNotificationRepository) GetNotifications(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error) {
	filter := bson.M{"user_id": userID}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{"created_at", -1}}) // Sort by created_at descending

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []*models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *mongoNotificationRepository) UpdateNotificationStatus(ctx context.Context, notificationID string, status models.NotificationStatus) error {
	filter := bson.M{"_id": notificationID}
	update := bson.M{"$set": bson.M{"status": status}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}
