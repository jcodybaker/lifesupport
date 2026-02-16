package api

type DriverName string

const (
	DriverShelly  DriverName = "shelly"
	DriverStation DriverName = "station"
)

// Device represents a physical device that may contain multiple sensors and actuators
type Device struct {
	ID          string            `json:"id"`
	Driver      DriverName        `json:"driver"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Sensors     []*Sensor         `json:"sensors"`
	Actuators   []*Actuator       `json:"actuators"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
}

// DefaultTag returns the default hierarchical tag for this device
func (d *Device) DefaultTag() string {
	return "device." + d.ID
}

// EnsureDefaultTag ensures the device has its default tag
func (d *Device) EnsureDefaultTag() {
	defaultTag := d.DefaultTag()
	hasDefault := false
	for _, tag := range d.Tags {
		if tag == defaultTag {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		d.Tags = append([]string{defaultTag}, d.Tags...)
	}
}

// GetSensorByID finds a sensor by its ID within the device
func (d *Device) GetSensorByID(id string) *Sensor {
	for _, sensor := range d.Sensors {
		if sensor.GetID() == id {
			return sensor
		}
	}
	return nil
}

// GetActuatorByID finds an actuator by its ID within the device
func (d *Device) GetActuatorByID(id string) *Actuator {
	for _, actuator := range d.Actuators {
		if actuator.GetID() == id {
			return actuator
		}
	}
	return nil
}

// GetSensorsByType returns all sensors of a specific type
func (d *Device) GetSensorsByType(sensorType SensorType) []*Sensor {
	var result []*Sensor
	for _, sensor := range d.Sensors {
		if sensor.GetType() == sensorType {
			result = append(result, sensor)
		}
	}
	return result
}

// GetActuatorsByType returns all actuators of a specific type
func (d *Device) GetActuatorsByType(actuatorType ActuatorType) []*Actuator {
	var result []*Actuator
	for _, actuator := range d.Actuators {
		if actuator.GetType() == actuatorType {
			result = append(result, actuator)
		}
	}
	return result
}
