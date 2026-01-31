package models

import "time"

// Device represents a controllable device (pump, light, valve)
type Device struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // pump, light, valve
	ShellyID    string    `json:"shelly_id"`
	Status      string    `json:"status"` // on, off, error
	LastUpdated time.Time `json:"last_updated"`
	Enabled     bool      `json:"enabled"`
}

// Sensor represents a sensor in the system
type Sensor struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // temperature, ph, flow, weight, distance
	Unit        string    `json:"unit"`
	Location    string    `json:"location"`
	LastValue   float64   `json:"last_value"`
	LastUpdated time.Time `json:"last_updated"`
	Enabled     bool      `json:"enabled"`
}

// SensorReading represents a time-series sensor reading from ClickHouse
type SensorReading struct {
	SensorID  int       `json:"sensor_id"`
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// Camera represents a camera in the system
type Camera struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Location    string    `json:"location"`
	Enabled     bool      `json:"enabled"`
	LastUpdated time.Time `json:"last_updated"`
}

// SystemStatus represents the overall system status
type SystemStatus struct {
	Timestamp     time.Time `json:"timestamp"`
	Healthy       bool      `json:"healthy"`
	ActiveAlerts  int       `json:"active_alerts"`
	DevicesOnline int       `json:"devices_online"`
	SensorsOnline int       `json:"sensors_online"`
}

// DeviceCommand represents a command to control a device
type DeviceCommand struct {
	DeviceID int    `json:"device_id"`
	Action   string `json:"action"`          // on, off, toggle
	Value    *int   `json:"value,omitempty"` // for dimmer/variable control
}

// Alert represents a system alert
type Alert struct {
	ID           int        `json:"id"`
	Type         string     `json:"type"` // warning, error, critical
	Message      string     `json:"message"`
	Source       string     `json:"source"`
	Acknowledged bool       `json:"acknowledged"`
	CreatedAt    time.Time  `json:"created_at"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
}

// User represents an authenticated user (admin only)
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
}
