package handlers

import (
	"net/http"
	"smarthome/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type IntegrationHandler struct {
	deviceClient    *services.DeviceServiceClient
	telemetryClient *services.TelemetryServiceClient
}

func NewIntegrationHandler() *IntegrationHandler {
	return &IntegrationHandler{
		deviceClient:    services.NewDeviceServiceClient(),
		telemetryClient: services.NewTelemetryServiceClient(),
	}
}

func (h *IntegrationHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Device Service integration
	router.GET("/v2/devices", h.GetDevices)
	router.GET("/v2/devices/:id", h.GetDevice)
	router.POST("/v2/devices", h.CreateDevice)

	// Telemetry Service integration
	router.GET("/v2/telemetry", h.GetTelemetry)
	router.POST("/v2/telemetry", h.PostTelemetry)
	router.GET("/v2/telemetry/aggregated", h.GetAggregatedTelemetry)
}

// Device endpoints
func (h *IntegrationHandler) GetDevices(c *gin.Context) {
	location := c.Query("location")
	status := c.Query("status")

	devices, err := h.deviceClient.GetDevices(location, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, devices)
}

func (h *IntegrationHandler) GetDevice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	device, err := h.deviceClient.GetDevice(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, device)
}

func (h *IntegrationHandler) CreateDevice(c *gin.Context) {
	var req struct {
		Name       string `json:"name" binding:"required"`
		DeviceType string `json:"device_type" binding:"required"`
		Location   string `json:"location" binding:"required"`
		Room       string `json:"room" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device, err := h.deviceClient.CreateDevice(req.Name, req.DeviceType, req.Location, req.Room)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, device)
}

// Telemetry endpoints
func (h *IntegrationHandler) GetTelemetry(c *gin.Context) {
	deviceIDStr := c.Query("device_id")
	metricName := c.Query("metric_name")
	limitStr := c.DefaultQuery("limit", "100")

	deviceID, _ := strconv.Atoi(deviceIDStr)
	limit, _ := strconv.Atoi(limitStr)

	telemetry, err := h.telemetryClient.GetTelemetry(deviceID, metricName, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, telemetry)
}

func (h *IntegrationHandler) PostTelemetry(c *gin.Context) {
	var req struct {
		DeviceID   int     `json:"device_id" binding:"required"`
		MetricName string  `json:"metric_name" binding:"required"`
		Value      float64 `json:"value" binding:"required"`
		Unit       string  `json:"unit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	telemetry, err := h.telemetryClient.PostTelemetry(req.DeviceID, req.MetricName, req.Value, req.Unit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, telemetry)
}

func (h *IntegrationHandler) GetAggregatedTelemetry(c *gin.Context) {
	deviceIDStr := c.Query("device_id")
	metricName := c.Query("metric_name")
	interval := c.DefaultQuery("interval", "1 hour")

	if deviceIDStr == "" || metricName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id and metric_name are required"})
		return
	}

	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device_id"})
		return
	}

	aggregated, err := h.telemetryClient.GetAggregatedTelemetry(deviceID, metricName, interval)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, aggregated)
}
