package sensorData

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) WriteSensorData(c *gin.Context) {
	var data SensorData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.WriteSensorData(c.Request.Context(), &data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "sensor data written successfully"})
}

func (h *Handler) QuerySensorData(c *gin.Context) {
	var req QuerySensorDataRequest

	if surveyPointIDStr := c.Query("survey_point_id"); surveyPointIDStr != "" {
		surveyPointID, err := uuid.Parse(surveyPointIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid survey_point_id"})
			return
		}
		req.SurveyPointID = &surveyPointID
	}

	if mcuCode := c.Query("mcu_code"); mcuCode != "" {
		req.MCUCode = &mcuCode
	}

	if farmIDStr := c.Query("farm_id"); farmIDStr != "" {
		farmID, err := uuid.Parse(farmIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid farm_id"})
			return
		}
		req.FarmID = &farmID
	}

	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format"})
			return
		}
		req.StartTime = &startTime
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format"})
			return
		}
		req.EndTime = &endTime
	}

	req.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "100"))

	records, err := h.service.QuerySensorData(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": records})
}

func (h *Handler) QueryLatestData(c *gin.Context) {
	surveyPointIDStr := c.Param("survey_point_id")
	surveyPointID, err := uuid.Parse(surveyPointIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid survey_point_id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	records, err := h.service.QueryLatestData(c.Request.Context(), surveyPointID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": records})
}

func (h *Handler) QueryAggregation(c *gin.Context) {
	var req AggregationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	records, err := h.service.QueryAggregation(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": records})
}

func (h *Handler) CreateCommand(c *gin.Context) {
	var req CreateCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.CreateCommand(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": result})
}

func (h *Handler) UpdateCommandStatus(c *gin.Context) {
	commandIDStr := c.Param("command_id")
	commandID, err := uuid.Parse(commandIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid command_id"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=pending sent success failed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.UpdateCommandStatus(c.Request.Context(), commandID, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *Handler) GetPendingCommands(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	commands, err := h.service.GetPendingCommands(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": commands})
}

func (h *Handler) GetCommandHistory(c *gin.Context) {
	var surveyPointID *uuid.UUID
	if surveyPointIDStr := c.Query("survey_point_id"); surveyPointIDStr != "" {
		id, err := uuid.Parse(surveyPointIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid survey_point_id"})
			return
		}
		surveyPointID = &id
	}

	var deviceName *string
	if name := c.Query("device_name"); name != "" {
		deviceName = &name
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	commands, err := h.service.GetCommandHistory(c.Request.Context(), surveyPointID, deviceName, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": commands})
}

func (h *Handler) GetCommandByID(c *gin.Context) {
	commandIDStr := c.Param("command_id")
	commandID, err := uuid.Parse(commandIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid command_id"})
		return
	}

	command, err := h.service.GetCommandByID(c.Request.Context(), commandID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": command})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	sensor := r.Group("/sensor-data")
	{
		sensor.POST("", h.WriteSensorData)
		sensor.GET("", h.QuerySensorData)
		sensor.GET("/latest/:survey_point_id", h.QueryLatestData)
		sensor.POST("/aggregation", h.QueryAggregation)
	}

	commands := r.Group("/commands")
	{
		commands.POST("", h.CreateCommand)
		commands.PUT("/:command_id/status", h.UpdateCommandStatus)
		commands.GET("/pending", h.GetPendingCommands)
		commands.GET("/history", h.GetCommandHistory)
		commands.GET("/:command_id", h.GetCommandByID)
	}
}
