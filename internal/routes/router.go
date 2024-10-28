package routes

import (
	"github.com/dk5761/go-serv/internal/infrastructure/container"
	"github.com/dk5761/go-serv/internal/infrastructure/middlewares"
	"github.com/gin-gonic/gin"
)

func InitRoutes(router *gin.Engine, container *container.Container) {
	// Apply global middlewares
	router.Use(middlewares.ErrorHandler())
	router.Use(middlewares.StartTracingMiddleware())

	router.Use(middlewares.TracingMiddleware())
	router.Use(middlewares.TraceIDResponseMiddleware())
	router.Use(middlewares.RequestLogger())
	// router.Use(middlewares.TraceIDResponseMiddleware())

	// Register feature routes
	RegisterAuthRoutes(router, container)
	RegisterChatRoutes(router, container)
}
