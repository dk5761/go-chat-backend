package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dk5761/go-serv/configs"
	"github.com/dk5761/go-serv/internal/infrastructure/cache"
	"github.com/dk5761/go-serv/internal/infrastructure/container"
	"github.com/dk5761/go-serv/internal/infrastructure/database"
	"github.com/dk5761/go-serv/internal/infrastructure/logging"
	"github.com/dk5761/go-serv/internal/infrastructure/storage"
	"github.com/dk5761/go-serv/internal/infrastructure/tracing"
	"github.com/dk5761/go-serv/internal/routes"
	"github.com/dk5761/go-serv/migrations"
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
	startServerWithGracefulShutdown(router, config, db, mongoDB)
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
	if err := migrations.RunMigrations(mongoDB); err != nil {
		logging.Logger.Fatal("Failed to run MongoDB migrations", zap.Error(err))
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
func startServerWithGracefulShutdown(router *gin.Engine, config *configs.Config, db *pgxpool.Pool, mongoDB *mongo.Database) {
	// Create an http.Server with the Gin router
	server := &http.Server{
		Addr:    config.Server.Address,
		Handler: router,
	}

	// Run the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()
	logging.Logger.Info("Server started on " + config.Server.Address)

	// Listen for OS signals for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal
	<-stop
	logging.Logger.Info("Shutting down server...")

	// Set a timeout context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt a graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logging.Logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	// Close database connections
	db.Close()
	if err := mongoDB.Client().Disconnect(ctx); err != nil {
		logging.Logger.Fatal("Failed to disconnect MongoDB client", zap.Error(err))
	}

	logging.Logger.Info("Server shutdown complete.")
}
