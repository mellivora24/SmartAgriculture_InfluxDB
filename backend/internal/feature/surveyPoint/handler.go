package surveyPoint

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateSurveyPointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	surveyPoint, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": surveyPoint})
}

func (h *Handler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid survey point id"})
		return
	}

	surveyPoint, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": surveyPoint})
}

func (h *Handler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid survey point id"})
		return
	}

	var req UpdateSurveyPointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	surveyPoint, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": surveyPoint})
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid survey point id"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=connecting connected disconnected"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid survey point id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "survey point deleted successfully"})
}

func (h *Handler) GetMCUSurveyPoints(c *gin.Context) {
	mcuIDStr := c.Param("mcu_id")
	mcuID, err := uuid.Parse(mcuIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mcu id"})
		return
	}

	surveyPoints, err := h.service.GetMCUSurveyPoints(c.Request.Context(), mcuID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": surveyPoints})
}

func (h *Handler) ListByMCU(c *gin.Context) {
	mcuIDStr := c.Query("mcu_id")
	if mcuIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mcu_id is required"})
		return
	}

	mcuID, err := uuid.Parse(mcuIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mcu id"})
		return
	}

	surveyPoints, err := h.service.ListByMCU(c.Request.Context(), mcuID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": surveyPoints})
}

func (h *Handler) ListByStatus(c *gin.Context) {
	status := c.Query("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	surveyPoints, err := h.service.ListByStatus(c.Request.Context(), status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": surveyPoints})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	surveyPoints := r.Group("/survey-points")
	{
		surveyPoints.POST("", h.Create)
		surveyPoints.GET("/:id", h.GetByID)
		surveyPoints.PUT("/:id", h.Update)
		surveyPoints.PUT("/:id/status", h.UpdateStatus)
		surveyPoints.DELETE("/:id", h.Delete)
		surveyPoints.GET("/mcu/:mcu_id", h.GetMCUSurveyPoints)
		surveyPoints.GET("/list", h.ListByMCU)
		surveyPoints.GET("/status", h.ListByStatus)
	}
}
