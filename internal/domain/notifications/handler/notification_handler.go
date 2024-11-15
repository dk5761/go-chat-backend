package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	models "github.com/dk5761/go-serv/internal/domain/notifications/model"
	"github.com/dk5761/go-serv/internal/domain/notifications/service"
)

type NotificationHandler struct {
	notificationService service.NotificationService
}

func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// SendNotification handles sending a single notification to a user
func (h *NotificationHandler) SendNotification(c *gin.Context) {
	var req struct {
		UserID      string            `json:"user_id" binding:"required"`
		Title       string            `json:"title" binding:"required"`
		Body        string            `json:"body" binding:"required"`
		DeviceToken string            `json:"device_token" binding:"required"`
		Data        map[string]string `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}

	notification := &models.Notification{

		UserID:      req.UserID,
		Title:       req.Title,
		Body:        req.Body,
		DeviceToken: req.DeviceToken,
		Data:        req.Data,
		CreatedAt:   time.Now(),
	}

	err := h.notificationService.SendNotification(c.Request.Context(), notification, req.UserID)
	if err != nil {
		zap.L().Error("Failed to send notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Notification sent successfully"})
}

// SendMulticastNotification handles sending a notification to multiple devices
func (h *NotificationHandler) SendMulticastNotification(c *gin.Context) {
	var req struct {
		Tokens []string          `json:"tokens" binding:"required"`
		Data   map[string]string `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}

	failedTokens, err := h.notificationService.SendMulticastNotification(c.Request.Context(), &models.Notification{
		Data: req.Data,
	}, req.Tokens)

	if err != nil {
		zap.L().Error("Failed to send multicast notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send multicast notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           "Multicast notification sent successfully",
		"failed_tokens":    failedTokens,
		"failed_token_num": len(failedTokens),
	})
}

// GetNotifications retrieves notifications for a user with pagination
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20 // Default value
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0 // Default value
	}

	notifications, err := h.notificationService.GetNotifications(c.Request.Context(), userID, limit, offset)
	if err != nil {
		zap.L().Error("Failed to retrieve notifications", zap.String("user_id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"metadata": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}
