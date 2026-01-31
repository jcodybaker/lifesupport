package storer_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"lifesupport/backend/pkg/device"
	"lifesupport/backend/pkg/storer"
)

// This example demonstrates how to use the storer package to persist device data
func Example() {
	// Connect to PostgreSQL database
	connString := "postgres://user:password@localhost:5432/lifesupport?sslmode=disable"
	store, err := storer.New(connString)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()

	// Initialize database schema
	if err := store.InitSchema(ctx); err != nil {
		log.Fatal(err)
	}

	// Create a system
	sys := &device.System{
		ID:          "sys-001",
		Name:        "Aquaponics System",
		Description: "Main aquaponics life support system",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create a subsystem
	subsystem := &device.Subsystem{
		ID:          "sub-001",
		Name:        "Fish Tank",
		Description: "Main fish tank subsystem",
		Type:        device.SubsystemTypeAquarium,
		Metadata: map[string]string{
			"capacity": "500L",
			"species":  "tilapia",
		},
	}

	// Create a device
	dev := &device.Device{
		ID:          "dev-001",
		Driver:      device.DriverShelly,
		Name:        "Water Monitor",
		Description: "Temperature and pH monitoring device",
		Metadata: map[string]string{
			"location": "tank-center",
			"version":  "1.0",
		},
	}

	// Add device to subsystem
	subsystem.Devices = []*device.Device{dev}

	// Add subsystem to system
	sys.Subsystems = []*device.Subsystem{subsystem}

	// Store the entire system hierarchy
	if err := store.CreateSystem(ctx, sys); err != nil {
		log.Fatal(err)
	}

	fmt.Println("System created successfully")

	// Store sensor readings
	reading := &device.SensorReading{
		Value:     25.5,
		Unit:      device.UnitCelsius,
		Timestamp: time.Now(),
		Valid:     true,
	}

	if err := store.StoreSensorReading(ctx, "dev-001", "sensor-temp-01", "Temperature Sensor", device.SensorTypeTemperature, reading); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Sensor reading stored")

	// Store actuator state
	state := &device.ActuatorState{
		Active: true,
		Parameters: map[string]float64{
			"brightness": 75.0,
		},
		Timestamp: time.Now(),
	}

	if err := store.StoreActuatorState(ctx, "dev-001", "actuator-light-01", "LED Light", device.ActuatorTypeDimmableLight, state); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Actuator state stored")

	// Retrieve the system
	retrievedSys, err := store.GetSystem(ctx, "sys-001")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Retrieved system: %s\n", retrievedSys.Name)

	// Query sensor readings
	filters := storer.SensorReadingFilters{
		DeviceID: stringPtr("dev-001"),
		Limit:    10,
	}

	readings, err := store.GetSensorReadings(ctx, filters)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d sensor readings\n", len(readings))

	// Get latest reading
	latest, err := store.GetLatestSensorReading(ctx, "sensor-temp-01")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Latest reading: %.2f%s\n", latest.Value, latest.Unit)

	// Output:
	// System created successfully
	// Sensor reading stored
	// Actuator state stored
	// Retrieved system: Aquaponics System
	// Found 1 sensor readings
	// Latest reading: 25.50°C
}

func stringPtr(s string) *string {
	return &s
}
