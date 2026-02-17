package api

// SensorType identifies the type of sensor measurement
type SensorType string

const (
	SensorTypeTemperature     SensorType = "temperature"
	SensorTypePH              SensorType = "ph"
	SensorTypeFlowRate        SensorType = "flow_rate"
	SensorTypePower           SensorType = "power"
	SensorTypeWaterDepth      SensorType = "water_depth"
	SensorTypeHumidity        SensorType = "humidity"
	SensorTypeLightLevel      SensorType = "light_level"
	SensorTypeConductivity    SensorType = "conductivity"
	SensorTypeDissolvedOxygen SensorType = "dissolved_oxygen"
	SensorTypeBoolean         SensorType = "boolean"
	SensorTypeVolume          SensorType = "volume"
)

// Sensor provides a base implementation for sensors with tag support
type Sensor struct {
	ID         string            `json:"id"`
	DeviceID   string            `json:"device_id"`
	Name       string            `json:"name"`
	SensorType SensorType        `json:"sensor_type"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
}

func (s *Sensor) GetID() string {
	return s.ID
}

func (s *Sensor) GetName() string {
	return s.Name
}

func (s *Sensor) GetType() SensorType {
	return s.SensorType
}

func (s *Sensor) GetTags() []string {
	return s.Tags
}

func (s *Sensor) DefaultTag(deviceID string) string {
	return "device." + deviceID + ".sensor." + s.ID
}
