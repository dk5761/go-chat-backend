package handler

import (
	"fmt"
	"net/http"

	"github.com/dk5761/go-serv/internal/domain/chat/models"
	"github.com/dk5761/go-serv/internal/domain/chat/service"
	ws "github.com/dk5761/go-serv/internal/domain/chat/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ChatHandler struct {
	chatService service.ChatService
	wsManager   *ws.WebSocketManager
}

func NewChatHandler(chatService service.ChatService, wsManager *ws.WebSocketManager) *ChatHandler {
	return &ChatHandler{chatService, wsManager}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// UploadFile handles file uploads through the ChatService
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

// HandleWebSocket manages WebSocket connections for real-time chat
//func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
//	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade WebSocket"})
//		return
//	}
//
//	userID := c.Query("userID")
//	client := &models.Client{ID: userID, Conn: conn, SendCh: make(chan *models.Message)}
//	h.wsManager.AddClient(client)
//
//	go func() {
//		defer func(conn *websocket.Conn) {
//			err := conn.Close()
//			if err != nil {
//				log.Fatal("Error while closing websocket connection", err)
//			}
//		}(conn)
//		for {
//			_, msgData, err := conn.ReadMessage()
//			if err != nil {
//				h.wsManager.RemoveClient(userID)
//				break
//			}
//
//			var message models.Message
//			if err := json.Unmarshal(msgData, &message); err != nil {
//				continue
//			}
//
//			// Set the sender ID and call SendMessage
//			message.SenderID = userID
//			if err := h.chatService.SendMessage(context.Background(), &message, nil, ""); err != nil {
//				continue
//			}
//
//			fmt.Println("message.1")
//
//			// Send message to the receiver
//			if err := h.chatService.SendToClient(message.ReceiverID, &message); err != nil {
//				// If the receiver is not connected, handle if needed
//				continue
//			}
//
//			fmt.Println("message.2")
//
//		}
//	}()
//}

func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade WebSocket"})
		return
	}

	fmt.Println("inside HandleWebSocket")

	userID := c.Query("userID")
	client := &models.Client{
		ID:     userID,
		Conn:   conn,
		SendCh: make(chan *models.Message, 10),
	}

	h.wsManager.AddClient(client)
}

// SendMessage handles sending messages with optional file support
func (h *ChatHandler) SendMessage(c *gin.Context) {
	var msg models.Message
	if err := c.BindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get the sender ID from the context (set by JWT middleware)
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

	// Call the service to send the message (without file)
	if err := h.chatService.SendMessage(c.Request.Context(), &msg, nil, ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "message sent"})
}

// GetChatHistory retrieves chat history between two users with pagination support
func (h *ChatHandler) GetChatHistory(c *gin.Context) {
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

	// Pagination parameters
	limit, offset := 20, 0
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := c.Query("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	// Get chat history from the service
	messages, err := h.chatService.GetChatHistory(c.Request.Context(), userID1, userID2, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chat history"})
		return
	}

	c.JSON(http.StatusOK, messages)
}
