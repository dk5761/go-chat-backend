package routes

import (
	"github.com/dk5761/go-serv/internal/infrastructure/container"
	"github.com/dk5761/go-serv/internal/infrastructure/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(router *gin.Engine, container *container.Container) {
	public := router.Group("/auth")
	{
		public.POST("/signup", container.AuthHandler.SignUp)
		public.POST("/login", container.AuthHandler.Login)
	}

	protected := router.Group("/api")
	protected.Use(middlewares.JWTAuthMiddleware(container.AuthHandler.JwtService, container.AuthHandler.UserRepo))

	{
		protected.GET("/profile", container.AuthHandler.Profile)
		protected.GET("/find", container.AuthHandler.GetUsers)

	}
}
