package main

import (
	"log"

	"github.com/dk5761/go-serv/configs"
	"github.com/dk5761/go-serv/internal/infrastructure/cache"
	"github.com/dk5761/go-serv/internal/infrastructure/database"
	"github.com/dk5761/go-serv/internal/infrastructure/logging"
	"github.com/dk5761/go-serv/internal/infrastructure/middlewares"
	"github.com/dk5761/go-serv/internal/infrastructure/storage"
	"github.com/dk5761/go-serv/internal/infrastructure/tracing"
	"github.com/dk5761/go-serv/internal/routes"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Load configurations
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logging.InitLogger()

	// Initialize tracing
	if err := tracing.InitTracer(); err != nil {
		logging.Logger.Fatal("Failed to initialize tracer", zap.Error(err))
	}

	// Initialize databases
	db, err := database.InitPostgresDB(config.Postgres)
	if err != nil {
		logging.Logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}
	mongoDB, err := database.InitMongoDB(config.MongoDB)
	if err != nil {
		logging.Logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}

	// Initialize cache
	cacheClient := cache.InitRedisClient(config.Redis)

	// Initialize storage service
	var storageService storage.StorageService
	if config.Storage.Provider == "s3" {
		storageService = storage.NewS3StorageService(config.Storage.S3Config)
	} else if config.Storage.Provider == "gdrive" {
		var err error
		storageService, err = storage.NewGDriveStorageService(config.Storage.GDriveConfig)
		if err != nil {
			logging.Logger.Fatal("Failed to initialize Google Drive storage", zap.Error(err))
		}
	} else {
		logging.Logger.Fatal("Invalid storage provider specified")
	}

	// Set up Gin router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middlewares.RequestLogger())

	// Initialize routes
	routes.InitRoutes(router, db, mongoDB, cacheClient, storageService, config)

	// Start the server
	if err := router.Run(config.Server.Address); err != nil {
		logging.Logger.Fatal("Failed to run server", zap.Error(err))
	}
}
