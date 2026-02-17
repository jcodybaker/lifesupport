package api

import "time"

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
	UnitMilliliters  Unit = "mL"
)

// SensorReading represents a single sensor measurement
type SensorReading struct {
	Value     float64   `json:"value"`
	Unit      Unit      `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
	Valid     bool      `json:"valid"`
	Error     string    `json:"error,omitempty"`
}
