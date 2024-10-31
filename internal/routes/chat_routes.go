package routes

// import (
//

// 	"github.com/dk5761/go-serv/internal/infrastructure/container"
// 	"github.com/dk5761/go-serv/internal/infrastructure/middlewares"
// 	"github.com/gin-gonic/gin"

// )

// func RegisterChatRoutes(router *gin.Engine, container *container.Container) {
// 	protected := router.Group("/api/chat")
// 	protected.Use(middlewares.JWTAuthMiddleware(container.AuthHandler.JwtService, container.AuthHandler.UserRepo))
// 	{
// 		protected.GET("/ws", container.ChatHandler.HandleWebSocket)
// 		protected.POST("/send", container.ChatHandler.SendMessage)
// 	}
// }
