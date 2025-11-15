package mcu

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
	var req CreateMCURequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mcu, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": mcu})
}

func (h *Handler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mcu id"})
		return
	}

	mcu, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mcu})
}

func (h *Handler) GetByCode(c *gin.Context) {
	mcuCode := c.Param("code")
	if mcuCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mcu code is required"})
		return
	}

	info, err := h.service.GetByCode(c.Request.Context(), mcuCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": info})
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	mcuCode := c.Param("code")
	if mcuCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mcu code is required"})
		return
	}

	var req UpdateMCUStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.UpdateStatus(c.Request.Context(), mcuCode, req.Status)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mcu id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "mcu deleted successfully"})
}

func (h *Handler) GetFarmMCUs(c *gin.Context) {
	farmIDStr := c.Param("farm_id")
	farmID, err := uuid.Parse(farmIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid farm id"})
		return
	}

	mcus, err := h.service.GetFarmMCUs(c.Request.Context(), farmID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mcus})
}

func (h *Handler) ListByFarm(c *gin.Context) {
	farmIDStr := c.Query("farm_id")
	if farmIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "farm_id is required"})
		return
	}

	farmID, err := uuid.Parse(farmIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid farm id"})
		return
	}

	mcus, err := h.service.ListByFarm(c.Request.Context(), farmID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mcus})
}

func (h *Handler) ListByStatus(c *gin.Context) {
	status := c.Query("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	mcus, err := h.service.ListByStatus(c.Request.Context(), status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mcus})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	mcus := r.Group("/mcus")
	{
		mcus.POST("", h.Create)
		mcus.GET("/:id", h.GetByID)
		mcus.GET("/code/:code", h.GetByCode)
		mcus.PUT("/code/:code/status", h.UpdateStatus)
		mcus.DELETE("/:id", h.Delete)
		mcus.GET("/farm/:farm_id", h.GetFarmMCUs)
		mcus.GET("/list", h.ListByFarm)
		mcus.GET("/status", h.ListByStatus)
	}
}
