package main

import (
	"log"

	"github.com/dk5761/go-serv/configs"
	"github.com/dk5761/go-serv/internal/infrastructure/cache"
	"github.com/dk5761/go-serv/internal/infrastructure/container"
	"github.com/dk5761/go-serv/internal/infrastructure/database"
	"github.com/dk5761/go-serv/internal/infrastructure/logging"
	"github.com/dk5761/go-serv/internal/infrastructure/storage"
	"github.com/dk5761/go-serv/internal/infrastructure/tracing"
	"github.com/dk5761/go-serv/internal/routes"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func main() {
	// Load configurations
	config := initConfig()

	// Initialize Logger
	logging.InitLogger()

	// Initialize Tracing
	initTracer()

	// Initialize Databases
	db := initPostgres(config)
	mongoDB := initMongoDB(config)

	// Initialize Cache
	cacheClient := cache.InitRedisClient(config.Redis)

	// Initialize Storage Service
	storageService := initStorage(config)

	// Set up Dependency Container
	cont := container.NewContainer(db, mongoDB, cacheClient, storageService, config)

	// Set up Gin router
	router := setupRouter()

	// Initialize routes
	routes.InitRoutes(router, cont)

	// Start the server
	startServer(router, config)
}

// initConfig loads the application configuration
func initConfig() *configs.Config {
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return config
}

// initTracer initializes tracing
func initTracer() {
	if err := tracing.InitTracer(); err != nil {
		logging.Logger.Fatal("Failed to initialize tracer", zap.Error(err))
	}
}

// initPostgres initializes the PostgreSQL database
func initPostgres(config *configs.Config) *pgxpool.Pool {
	db, err := database.InitPostgresDB(config.Postgres)
	if err != nil {
		logging.Logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}
	return db
}

// initMongoDB initializes the MongoDB database
func initMongoDB(config *configs.Config) *mongo.Database {
	mongoDB, err := database.InitMongoDB(config.MongoDB)
	if err != nil {
		logging.Logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	return mongoDB
}

// initStorage initializes the storage service based on the provider
func initStorage(config *configs.Config) storage.StorageService {
	logging.Logger.Info("Initializing storage with provider", zap.String("provider", config.Storage.Provider))

	var storageService storage.StorageService
	var err error

	switch config.Storage.Provider {
	case "s3":
		storageService = storage.NewS3StorageService(config.Storage.S3Config)
	case "gdrive":
		storageService, err = storage.NewGDriveStorageService(config.Storage.GDriveConfig)
		if err != nil {
			logging.Logger.Fatal("Failed to initialize Google Drive storage", zap.Error(err))
		}
	default:
		logging.Logger.Fatal("Invalid storage provider specified", zap.String("provider", config.Storage.Provider))
	}

	return storageService
}

// setupRouter configures Gin router with middleware
func setupRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	return router
}

// startServer starts the HTTP server
func startServer(router *gin.Engine, config *configs.Config) {
	if err := router.Run(config.Server.Address); err != nil {
		logging.Logger.Fatal("Failed to run server", zap.Error(err))
	}
}
