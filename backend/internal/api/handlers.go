package api

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/cody/lifesupport/internal/auth"
	"github.com/cody/lifesupport/internal/database"
	"github.com/cody/lifesupport/internal/models"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	postgres   *database.PostgresDB
	clickhouse *database.ClickHouseDB
}

func NewHandler(pg *database.PostgresDB, ch *database.ClickHouseDB) *Handler {
	return &Handler{
		postgres:   pg,
		clickhouse: ch,
	}
}

// SetupRoutes configures all API routes
func (h *Handler) SetupRoutes(r *gin.Engine) {
	// Public routes (read-only)
	public := r.Group("/api")
	public.Use(auth.OptionalAuthMiddleware())
	{
		public.GET("/status", h.GetSystemStatus)
		public.GET("/devices", h.GetDevices)
		public.GET("/sensors", h.GetSensors)
		public.GET("/sensors/:id/readings", h.GetSensorReadings)
		public.GET("/cameras", h.GetCameras)
		public.GET("/alerts", h.GetAlerts)
	}

	// Authentication
	r.POST("/api/login", h.Login)

	// Protected routes (admin only)
	admin := r.Group("/api/admin")
	admin.Use(auth.AuthMiddleware())
	{
		admin.POST("/devices/:id/control", h.ControlDevice)
		admin.PUT("/devices/:id", h.UpdateDevice)
		admin.POST("/devices", h.CreateDevice)
		admin.DELETE("/devices/:id", h.DeleteDevice)

		admin.PUT("/sensors/:id", h.UpdateSensor)
		admin.POST("/sensors", h.CreateSensor)
		admin.DELETE("/sensors/:id", h.DeleteSensor)

		admin.PUT("/cameras/:id", h.UpdateCamera)
		admin.POST("/cameras", h.CreateCamera)
		admin.DELETE("/cameras/:id", h.DeleteCamera)

		admin.PUT("/alerts/:id/acknowledge", h.AcknowledgeAlert)
		admin.DELETE("/alerts/:id", h.DeleteAlert)
	}
}

// GetSystemStatus returns overall system health
func (h *Handler) GetSystemStatus(c *gin.Context) {
	var devicesOnline, sensorsOnline, activeAlerts int

	h.postgres.DB.QueryRow("SELECT COUNT(*) FROM devices WHERE enabled = true AND status != 'error'").Scan(&devicesOnline)
	h.postgres.DB.QueryRow("SELECT COUNT(*) FROM sensors WHERE enabled = true").Scan(&sensorsOnline)
	h.postgres.DB.QueryRow("SELECT COUNT(*) FROM alerts WHERE acknowledged = false AND resolved_at IS NULL").Scan(&activeAlerts)

	status := models.SystemStatus{
		Timestamp:     time.Now(),
		Healthy:       activeAlerts == 0,
		ActiveAlerts:  activeAlerts,
		DevicesOnline: devicesOnline,
		SensorsOnline: sensorsOnline,
	}

	c.JSON(http.StatusOK, status)
}

// GetDevices returns all devices
func (h *Handler) GetDevices(c *gin.Context) {
	rows, err := h.postgres.DB.Query(`
		SELECT id, name, type, shelly_id, status, last_updated, enabled 
		FROM devices 
		ORDER BY name
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var d models.Device
		if err := rows.Scan(&d.ID, &d.Name, &d.Type, &d.ShellyID, &d.Status, &d.LastUpdated, &d.Enabled); err != nil {
			continue
		}
		devices = append(devices, d)
	}

	c.JSON(http.StatusOK, devices)
}

// GetSensors returns all sensors
func (h *Handler) GetSensors(c *gin.Context) {
	rows, err := h.postgres.DB.Query(`
		SELECT id, name, type, unit, location, last_value, last_updated, enabled 
		FROM sensors 
		ORDER BY name
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var sensors []models.Sensor
	for rows.Next() {
		var s models.Sensor
		var lastValue sql.NullFloat64
		if err := rows.Scan(&s.ID, &s.Name, &s.Type, &s.Unit, &s.Location, &lastValue, &s.LastUpdated, &s.Enabled); err != nil {
			continue
		}
		if lastValue.Valid {
			s.LastValue = lastValue.Float64
		}
		sensors = append(sensors, s)
	}

	c.JSON(http.StatusOK, sensors)
}

// GetSensorReadings returns time-series data for a sensor
func (h *Handler) GetSensorReadings(c *gin.Context) {
	sensorID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}

	// Default to last 24 hours
	hoursStr := c.DefaultQuery("hours", "24")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil {
		hours = 24
	}

	end := time.Now()
	start := end.Add(-time.Duration(hours) * time.Hour)

	readings, err := h.clickhouse.GetReadings(context.Background(), sensorID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, readings)
}

// GetCameras returns all cameras
func (h *Handler) GetCameras(c *gin.Context) {
	rows, err := h.postgres.DB.Query(`
		SELECT id, name, url, location, enabled, last_updated 
		FROM cameras 
		ORDER BY name
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var cameras []models.Camera
	for rows.Next() {
		var cam models.Camera
		if err := rows.Scan(&cam.ID, &cam.Name, &cam.URL, &cam.Location, &cam.Enabled, &cam.LastUpdated); err != nil {
			continue
		}
		cameras = append(cameras, cam)
	}

	c.JSON(http.StatusOK, cameras)
}

// GetAlerts returns system alerts
func (h *Handler) GetAlerts(c *gin.Context) {
	rows, err := h.postgres.DB.Query(`
		SELECT id, type, message, source, acknowledged, created_at, resolved_at 
		FROM alerts 
		ORDER BY created_at DESC 
		LIMIT 100
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var a models.Alert
		var resolvedAt sql.NullTime
		if err := rows.Scan(&a.ID, &a.Type, &a.Message, &a.Source, &a.Acknowledged, &a.CreatedAt, &resolvedAt); err != nil {
			continue
		}
		if resolvedAt.Valid {
			a.ResolvedAt = &resolvedAt.Time
		}
		alerts = append(alerts, a)
	}

	c.JSON(http.StatusOK, alerts)
}

// Login authenticates a user
func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := h.postgres.DB.QueryRow(
		"SELECT id, username, password_hash FROM users WHERE username = $1",
		req.Username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	var resp models.LoginResponse
	resp.Token = token
	resp.User.ID = user.ID
	resp.User.Username = user.Username

	c.JSON(http.StatusOK, resp)
}

// ControlDevice sends a command to a device
func (h *Handler) ControlDevice(c *gin.Context) {
	deviceID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	var cmd models.DeviceCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Integrate with Shelly API to actually control the device
	// For now, just update the database status

	newStatus := cmd.Action
	if cmd.Action == "toggle" {
		// Query current status and toggle it
		var currentStatus string
		h.postgres.DB.QueryRow("SELECT status FROM devices WHERE id = $1", deviceID).Scan(&currentStatus)
		if currentStatus == "on" {
			newStatus = "off"
		} else {
			newStatus = "on"
		}
	}

	_, err = h.postgres.DB.Exec(
		"UPDATE devices SET status = $1, last_updated = $2 WHERE id = $3",
		newStatus, time.Now(), deviceID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device command sent", "new_status": newStatus})
}

// UpdateDevice updates device configuration
func (h *Handler) UpdateDevice(c *gin.Context) {
	deviceID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	var device models.Device
	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = h.postgres.DB.Exec(
		"UPDATE devices SET name = $1, enabled = $2, last_updated = $3 WHERE id = $4",
		device.Name, device.Enabled, time.Now(), deviceID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device updated"})
}

// CreateDevice creates a new device
func (h *Handler) CreateDevice(c *gin.Context) {
	var device models.Device
	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id int
	err := h.postgres.DB.QueryRow(
		"INSERT INTO devices (name, type, shelly_id, status, enabled) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		device.Name, device.Type, device.ShellyID, "off", device.Enabled,
	).Scan(&id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Device created"})
}

// DeleteDevice removes a device
func (h *Handler) DeleteDevice(c *gin.Context) {
	deviceID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	_, err = h.postgres.DB.Exec("DELETE FROM devices WHERE id = $1", deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device deleted"})
}

// CreateSensor creates a new sensor
func (h *Handler) CreateSensor(c *gin.Context) {
	var sensor models.Sensor
	if err := c.ShouldBindJSON(&sensor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id int
	err := h.postgres.DB.QueryRow(
		"INSERT INTO sensors (name, type, unit, location, enabled) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		sensor.Name, sensor.Type, sensor.Unit, sensor.Location, sensor.Enabled,
	).Scan(&id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Sensor created"})
}

// UpdateSensor updates sensor configuration
func (h *Handler) UpdateSensor(c *gin.Context) {
	sensorID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}

	var sensor models.Sensor
	if err := c.ShouldBindJSON(&sensor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = h.postgres.DB.Exec(
		"UPDATE sensors SET name = $1, location = $2, enabled = $3, last_updated = $4 WHERE id = $5",
		sensor.Name, sensor.Location, sensor.Enabled, time.Now(), sensorID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sensor updated"})
}

// DeleteSensor removes a sensor
func (h *Handler) DeleteSensor(c *gin.Context) {
	sensorID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}

	_, err = h.postgres.DB.Exec("DELETE FROM sensors WHERE id = $1", sensorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sensor deleted"})
}

// CreateCamera creates a new camera
func (h *Handler) CreateCamera(c *gin.Context) {
	var camera models.Camera
	if err := c.ShouldBindJSON(&camera); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id int
	err := h.postgres.DB.QueryRow(
		"INSERT INTO cameras (name, url, location, enabled) VALUES ($1, $2, $3, $4) RETURNING id",
		camera.Name, camera.URL, camera.Location, camera.Enabled,
	).Scan(&id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Camera created"})
}

// UpdateCamera updates camera configuration
func (h *Handler) UpdateCamera(c *gin.Context) {
	cameraID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid camera ID"})
		return
	}

	var camera models.Camera
	if err := c.ShouldBindJSON(&camera); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = h.postgres.DB.Exec(
		"UPDATE cameras SET name = $1, url = $2, location = $3, enabled = $4, last_updated = $5 WHERE id = $6",
		camera.Name, camera.URL, camera.Location, camera.Enabled, time.Now(), cameraID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Camera updated"})
}

// DeleteCamera removes a camera
func (h *Handler) DeleteCamera(c *gin.Context) {
	cameraID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid camera ID"})
		return
	}

	_, err = h.postgres.DB.Exec("DELETE FROM cameras WHERE id = $1", cameraID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Camera deleted"})
}

// AcknowledgeAlert marks an alert as acknowledged
func (h *Handler) AcknowledgeAlert(c *gin.Context) {
	alertID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	_, err = h.postgres.DB.Exec("UPDATE alerts SET acknowledged = true WHERE id = $1", alertID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert acknowledged"})
}

// DeleteAlert removes an alert
func (h *Handler) DeleteAlert(c *gin.Context) {
	alertID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	_, err = h.postgres.DB.Exec("DELETE FROM alerts WHERE id = $1", alertID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert deleted"})
}
