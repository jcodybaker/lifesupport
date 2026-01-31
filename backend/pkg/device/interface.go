package device

import (
	"time"
)

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
	GetReading() (*SensorReading, error)
	GetLastReading() *SensorReading
}

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

// Actuator represents an abstract actuator interface
type Actuator interface {
	GetID() string
	GetName() string
	GetType() ActuatorType
	GetState() (*ActuatorState, error)
	SendCommand(cmd ActuatorCommand) error
}

type DriverName string

const (
	DriverShelly  DriverName = "shelly"
	DriverStation DriverName = "station"
)

// Device represents a physical device that may contain multiple sensors and actuators
type Device struct {
	ID          string            `json:"id"`
	Driver      DriverName            `json:"driver"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Sensors     []Sensor          `json:"sensors"`
	Actuators   []Actuator        `json:"actuators"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// GetSensorByID finds a sensor by its ID within the device
func (d *Device) GetSensorByID(id string) Sensor {
	for _, sensor := range d.Sensors {
		if sensor.GetID() == id {
			return sensor
		}
	}
	return nil
}

// GetActuatorByID finds an actuator by its ID within the device
func (d *Device) GetActuatorByID(id string) Actuator {
	for _, actuator := range d.Actuators {
		if actuator.GetID() == id {
			return actuator
		}
	}
	return nil
}

// GetSensorsByType returns all sensors of a specific type
func (d *Device) GetSensorsByType(sensorType SensorType) []Sensor {
	var result []Sensor
	for _, sensor := range d.Sensors {
		if sensor.GetType() == sensorType {
			result = append(result, sensor)
		}
	}
	return result
}

// GetActuatorsByType returns all actuators of a specific type
func (d *Device) GetActuatorsByType(actuatorType ActuatorType) []Actuator {
	var result []Actuator
	for _, actuator := range d.Actuators {
		if actuator.GetType() == actuatorType {
			result = append(result, actuator)
		}
	}
	return result
}

// Subsystem represents a collection of devices organized hierarchically
type Subsystem struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Type        SubsystemType     `json:"type"`
	Devices     []*Device         `json:"devices"`
	Subsystems  []*Subsystem      `json:"subsystems,omitempty"` // Child subsystems
	Parent      *Subsystem        `json:"-"`                    // Parent subsystem (not serialized)
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// SubsystemType identifies the type of subsystem
type SubsystemType string

const (
	SubsystemTypeAquarium       SubsystemType = "aquarium"
	SubsystemTypeHydroponics    SubsystemType = "hydroponics"
	SubsystemTypeReservoir      SubsystemType = "reservoir"
	SubsystemTypeFiltration     SubsystemType = "filtration"
	SubsystemTypeLighting       SubsystemType = "lighting"
	SubsystemTypeNutrientDosing SubsystemType = "nutrient_dosing"
	SubsystemTypeWaterExchange  SubsystemType = "water_exchange"
	SubsystemTypeEnvironmental  SubsystemType = "environmental"
)

// GetAllDevices returns all devices in this subsystem and its children
func (s *Subsystem) GetAllDevices() []*Device {
	devices := make([]*Device, len(s.Devices))
	copy(devices, s.Devices)

	for _, child := range s.Subsystems {
		devices = append(devices, child.GetAllDevices()...)
	}

	return devices
}

// GetDeviceByID finds a device by ID in this subsystem and its children
func (s *Subsystem) GetDeviceByID(id string) *Device {
	for _, device := range s.Devices {
		if device.ID == id {
			return device
		}
	}

	for _, child := range s.Subsystems {
		if device := child.GetDeviceByID(id); device != nil {
			return device
		}
	}

	return nil
}

// GetSubsystemByID finds a child subsystem by ID recursively
func (s *Subsystem) GetSubsystemByID(id string) *Subsystem {
	if s.ID == id {
		return s
	}

	for _, child := range s.Subsystems {
		if found := child.GetSubsystemByID(id); found != nil {
			return found
		}
	}

	return nil
}

// AddDevice adds a device to the subsystem
func (s *Subsystem) AddDevice(device *Device) {
	s.Devices = append(s.Devices, device)
}

// AddSubsystem adds a child subsystem and sets the parent reference
func (s *Subsystem) AddSubsystem(child *Subsystem) {
	child.Parent = s
	s.Subsystems = append(s.Subsystems, child)
}

// GetAllSensors returns all sensors from all devices in this subsystem and children
func (s *Subsystem) GetAllSensors() []Sensor {
	var sensors []Sensor

	for _, device := range s.GetAllDevices() {
		sensors = append(sensors, device.Sensors...)
	}

	return sensors
}

// GetAllActuators returns all actuators from all devices in this subsystem and children
func (s *Subsystem) GetAllActuators() []Actuator {
	var actuators []Actuator

	for _, device := range s.GetAllDevices() {
		actuators = append(actuators, device.Actuators...)
	}

	return actuators
}

// System represents the entire life support system
type System struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Subsystems  []*Subsystem `json:"subsystems"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// GetSubsystemByID finds a subsystem by ID across the entire system
func (sys *System) GetSubsystemByID(id string) *Subsystem {
	for _, subsystem := range sys.Subsystems {
		if found := subsystem.GetSubsystemByID(id); found != nil {
			return found
		}
	}
	return nil
}

// GetDeviceByID finds a device by ID across the entire system
func (sys *System) GetDeviceByID(id string) *Device {
	for _, subsystem := range sys.Subsystems {
		if device := subsystem.GetDeviceByID(id); device != nil {
			return device
		}
	}
	return nil
}

// GetAllSensors returns all sensors in the entire system
func (sys *System) GetAllSensors() []Sensor {
	var sensors []Sensor

	for _, subsystem := range sys.Subsystems {
		sensors = append(sensors, subsystem.GetAllSensors()...)
	}

	return sensors
}

// GetAllActuators returns all actuators in the entire system
func (sys *System) GetAllActuators() []Actuator {
	var actuators []Actuator

	for _, subsystem := range sys.Subsystems {
		actuators = append(actuators, subsystem.GetAllActuators()...)
	}

	return actuators
}
