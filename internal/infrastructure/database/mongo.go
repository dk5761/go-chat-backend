package database

import (
	"context"
	"time"

	"github.com/dk5761/go-serv/configs"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongoDB(cfg configs.MongoDBConfig) (*mongo.Database, error) {
	clientOpts := options.Client().ApplyURI(cfg.URI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}
	return client.Database(cfg.Database), nil
}
