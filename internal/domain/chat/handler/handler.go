package handler

import (
	"net/http"

	"github.com/dk5761/go-serv/internal/domain/chat/models"
	"github.com/dk5761/go-serv/internal/domain/chat/service"
	"github.com/dk5761/go-serv/internal/infrastructure/logging"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type ChatHandler struct {
	chatService service.ChatService
}

func NewChatHandler(chatService service.ChatService) *ChatHandler {
	return &ChatHandler{chatService}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Handler for uploading files
func (h *ChatHandler) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file"})
		return
	}
	defer file.Close()

	// Generate a unique file name or use the original
	fileName := header.Filename

	// Call the service to upload the file
	fileURL, err := h.chatService.UploadFile(c.Request.Context(), file, fileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"file_url": fileURL})
}

func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logging.Logger.Error("Failed to upgrade to WebSocket", zap.Error(err))
		return
	}
	defer wsConn.Close()

	// Handle WebSocket communication here
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	var msg models.Message
	if err := c.BindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get the sender ID from context (set by JWT middleware)
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	senderID, ok := userIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}
	msg.SenderID = senderID.String()

	// Call the service to send the message
	if err := h.chatService.SendMessage(c.Request.Context(), &msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "message sent"})
}

func (h *ChatHandler) GetChatHistory(c *gin.Context) {
	// Extract user IDs from query parameters or request context
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID1, ok := userIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	receiverIDStr := c.Query("receiver_id")
	if receiverIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "receiver_id is required"})
		return
	}
	userID2, err := uuid.Parse(receiverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid receiver_id"})
		return
	}

	// Get chat history from the service
	messages, err := h.chatService.GetChatHistory(c.Request.Context(), userID1, userID2)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chat history"})
		return
	}

	c.JSON(http.StatusOK, messages)
}
