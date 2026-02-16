package api

import "time"

// SensorType identifies the type of sensor measurement
type SensorType string

const (
	SensorTypeTemperature     SensorType = "temperature"
	SensorTypePH              SensorType = "ph"
	SensorTypeFlowRate        SensorType = "flow_rate"
	SensorTypePower           SensorType = "power"
	SensorTypeWaterDepth      SensorType = "water_depth"
	SensorTypeActuatorStatus  SensorType = "actuator_status"
	SensorTypeHumidity        SensorType = "humidity"
	SensorTypeLightLevel      SensorType = "light_level"
	SensorTypeConductivity    SensorType = "conductivity"
	SensorTypeDissolvedOxygen SensorType = "dissolved_oxygen"
)

// Unit represents the measurement unit
type Unit string

const (
	UnitCelsius      Unit = "°C"
	UnitFahrenheit   Unit = "°F"
	UnitPH           Unit = "pH"
	UnitLitersPerMin Unit = "L/min"
	UnitWatts        Unit = "W"
	UnitCentimeters  Unit = "cm"
	UnitPercent      Unit = "%"
	UnitLux          Unit = "lux"
	UnitMicroSiemens Unit = "µS/cm"
	UnitMgPerL       Unit = "mg/L"
)

// SensorReading represents a single sensor measurement
type SensorReading struct {
	Value     float64   `json:"value"`
	Unit      Unit      `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
	Valid     bool      `json:"valid"`
	Error     string    `json:"error,omitempty"`
}

// Sensor represents an abstract sensor interface
type Sensor interface {
	GetID() string
	GetName() string
	GetType() SensorType
	GetTags() []string
	DefaultTag(deviceID string) string
}

// BaseSensor provides a base implementation for sensors with tag support
type BaseSensor struct {
	ID         string            `json:"id"`
	DeviceID   string            `json:"device_id"`
	Name       string            `json:"name"`
	SensorType SensorType        `json:"sensor_type"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
}

func (s *BaseSensor) GetID() string {
	return s.ID
}

func (s *BaseSensor) GetName() string {
	return s.Name
}

func (s *BaseSensor) GetType() SensorType {
	return s.SensorType
}

func (s *BaseSensor) GetTags() []string {
	return s.Tags
}

func (s *BaseSensor) DefaultTag(deviceID string) string {
	return "device." + deviceID + ".sensor." + s.ID
}
