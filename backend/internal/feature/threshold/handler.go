package threshold

import (
	"backend/internal/shared"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	thresholds := router.Group("/thresholds")
	{
		thresholds.GET("/survey-point/:surveyPointId", h.GetBySurveyPoint)
		thresholds.PUT("/survey-point/:surveyPointId", h.UpdateThresholds)

		thresholds.GET("/survey-point/:surveyPointId/alerts", h.GetAlertHistory)
		thresholds.POST("/alerts/:alertId/acknowledge", h.AcknowledgeAlert)

		thresholds.GET("/survey-point/:surveyPointId/auto-pump-history", h.GetAutoPumpHistory)
	}
}

// GetBySurveyPoint lấy cấu hình ngưỡng theo survey point
func (h *Handler) GetBySurveyPoint(c *gin.Context) {
	surveyPointID, err := shared.ParseUUID(c.Param("surveyPointId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid survey point ID"})
		return
	}

	settings, err := h.service.GetBySurveyPoint(c.Request.Context(), surveyPointID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if settings == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Threshold settings not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": settings})
}

// UpdateThresholds cập nhật ngưỡng
func (h *Handler) UpdateThresholds(c *gin.Context) {
	surveyPointID, err := shared.ParseUUID(c.Param("surveyPointId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid survey point ID"})
		return
	}

	var req UpdateThresholdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateThresholds(c.Request.Context(), surveyPointID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get updated settings
	settings, err := h.service.GetBySurveyPoint(c.Request.Context(), surveyPointID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": settings})
}

// GetAlertHistory lấy lịch sử cảnh báo
func (h *Handler) GetAlertHistory(c *gin.Context) {
	surveyPointID, err := shared.ParseUUID(c.Param("surveyPointId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid survey point ID"})
		return
	}
	limit := 50
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil {
			limit = parsed
		}
	}

	alerts, err := h.service.GetAlertHistory(c.Request.Context(), surveyPointID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": alerts})
	c.JSON(http.StatusOK, gin.H{"data": alerts})
}

// AcknowledgeAlert xác nhận đã xem cảnh báo
func (h *Handler) AcknowledgeAlert(c *gin.Context) {
	alertID, err := shared.ParseUUID(c.Param("alertId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	claimsMap, ok := claims.(map[string]interface{})
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr, ok := claimsMap["user_id"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, err := shared.ParseUUID(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.service.AcknowledgeAlert(c.Request.Context(), alertID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"acknowledged": true}})
}

// GetAutoPumpHistory lấy lịch sử tự động bơm
func (h *Handler) GetAutoPumpHistory(c *gin.Context) {
	surveyPointID, err := shared.ParseUUID(c.Param("surveyPointId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid survey point ID"})
		return
	}
	limit := 50
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil {
			limit = parsed
		}
	}

	history, err := h.service.GetAutoPumpHistory(c.Request.Context(), surveyPointID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": history})
	c.JSON(http.StatusOK, gin.H{"data": history})
}
