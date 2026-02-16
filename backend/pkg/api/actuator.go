package api

import "time"

// ActuatorType identifies the type of actuator
type ActuatorType string

const (
	ActuatorTypeRelay           ActuatorType = "relay"
	ActuatorTypePeristalticPump ActuatorType = "peristaltic_pump"
	ActuatorTypeDimmableLight   ActuatorType = "dimmable_light"
	ActuatorTypeServo           ActuatorType = "servo"
	ActuatorTypeValve           ActuatorType = "valve"
)

// ActuatorState represents the current state of an actuator
type ActuatorState struct {
	Active     bool               `json:"active"`
	Parameters map[string]float64 `json:"parameters,omitempty"`
	Timestamp  time.Time          `json:"timestamp"`
	Error      string             `json:"error,omitempty"`
}

// ActuatorCommand represents a command to send to an actuator
type ActuatorCommand struct {
	Action     string             `json:"action"`               // "on", "off", "set", "dispense", etc.
	Parameters map[string]float64 `json:"parameters,omitempty"` // e.g., "brightness": 75, "quantity": 100
}

// Actuator provides a base implementation for actuators with tag support
type Actuator struct {
	ID           string            `json:"id"`
	DeviceID     string            `json:"device_id"`
	Name         string            `json:"name"`
	ActuatorType ActuatorType      `json:"actuator_type"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Tags         []string          `json:"tags,omitempty"`
}

func (a *Actuator) GetID() string {
	return a.ID
}

func (a *Actuator) GetName() string {
	return a.Name
}

func (a *Actuator) GetType() ActuatorType {
	return a.ActuatorType
}

func (a *Actuator) GetTags() []string {
	return a.Tags
}

func (a *Actuator) DefaultTag(deviceID string) string {
	return "device." + deviceID + ".actuator." + a.ID
}
