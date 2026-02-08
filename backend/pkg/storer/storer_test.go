package storer

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"lifesupport/backend/pkg/api"
)

// getTestConnString returns the connection string for the test database
func getTestConnString() string {
	connStr := os.Getenv("TEST_DB_CONN")
	if connStr == "" {
		return "postgres://lifesupport:lifesupport@localhost:5432/lifesupport?sslmode=disable"
	}
	return connStr
}

// setupTestDB creates a fresh database for testing
func setupTestDB(t *testing.T) *Storer {
	t.Helper()

	store, err := New(getTestConnString())
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	ctx := context.Background()
	if err := store.InitSchema(ctx); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}

	return store
}

// cleanupTestDB cleans up the test database
func cleanupTestDB(t *testing.T, store *Storer) {
	t.Helper()

	ctx := context.Background()

	// Clean up in correct order due to foreign keys
	_, _ = store.db.ExecContext(ctx, "DELETE FROM sensor_readings")
	_, _ = store.db.ExecContext(ctx, "DELETE FROM actuator_states")
	_, _ = store.db.ExecContext(ctx, "DELETE FROM devices")
	_, _ = store.db.ExecContext(ctx, "DELETE FROM subsystems")
	_, _ = store.db.ExecContext(ctx, "DELETE FROM systems")

	if err := store.Close(); err != nil {
		t.Errorf("Failed to close database: %v", err)
	}
}

func TestNew(t *testing.T) {
	store, err := New(getTestConnString())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer store.Close()

	if store.db == nil {
		t.Error("New() returned store with nil db")
	}
}

func TestInitSchema(t *testing.T) {
	store := setupTestDB(t)
	defer cleanupTestDB(t, store)

	// Verify tables exist by trying to query them
	ctx := context.Background()
	tables := []string{"systems", "subsystems", "devices", "sensor_readings", "actuator_states"}

	for _, table := range tables {
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		var count int
		err := store.db.QueryRowContext(ctx, query).Scan(&count)
		if err != nil {
			t.Errorf("Table %s does not exist or is not accessible: %v", table, err)
		}
	}
}

func TestCreateAndGetSystem(t *testing.T) {
	store := setupTestDB(t)
	defer cleanupTestDB(t, store)

	ctx := context.Background()
	now := time.Now()

	sys := &api.System{
		ID:          "test-system-1",
		Name:        "Test System",
		Description: "A test life support system",
		CreatedAt:   now,
		UpdatedAt:   now,
		Subsystems:  []*api.Subsystem{},
	}

	// Create system
	err := store.CreateSystem(ctx, sys)
	if err != nil {
		t.Fatalf("CreateSystem() error = %v", err)
	}

	// Get system
	retrieved, err := store.GetSystem(ctx, sys.ID)
	if err != nil {
		t.Fatalf("GetSystem() error = %v", err)
	}

	if retrieved.ID != sys.ID {
		t.Errorf("GetSystem() ID = %v, want %v", retrieved.ID, sys.ID)
	}
	if retrieved.Name != sys.Name {
		t.Errorf("GetSystem() Name = %v, want %v", retrieved.Name, sys.Name)
	}
	if retrieved.Description != sys.Description {
		t.Errorf("GetSystem() Description = %v, want %v", retrieved.Description, sys.Description)
	}
}

func TestUpdateSystem(t *testing.T) {
	store := setupTestDB(t)
	defer cleanupTestDB(t, store)

	ctx := context.Background()
	now := time.Now()

	sys := &api.System{
		ID:          "test-system-update",
		Name:        "Original Name",
		Description: "Original Description",
		CreatedAt:   now,
		UpdatedAt:   now,
		Subsystems:  []*api.Subsystem{},
	}

	// Create system
	if err := store.CreateSystem(ctx, sys); err != nil {
		t.Fatalf("CreateSystem() error = %v", err)
	}

	// Update system
	sys.Name = "Updated Name"
	sys.Description = "Updated Description"
	if err := store.UpdateSystem(ctx, sys); err != nil {
		t.Fatalf("UpdateSystem() error = %v", err)
	}

	// Get and verify
	retrieved, err := store.GetSystem(ctx, sys.ID)
	if err != nil {
		t.Fatalf("GetSystem() error = %v", err)
	}

	if retrieved.Name != "Updated Name" {
		t.Errorf("UpdateSystem() Name = %v, want Updated Name", retrieved.Name)
	}
	if retrieved.Description != "Updated Description" {
		t.Errorf("UpdateSystem() Description = %v, want Updated Description", retrieved.Description)
	}
}

func TestDeleteSystem(t *testing.T) {
	store := setupTestDB(t)
	defer cleanupTestDB(t, store)

	ctx := context.Background()
	now := time.Now()

	sys := &api.System{
		ID:          "test-system-delete",
		Name:        "Test System",
		Description: "To be deleted",
		CreatedAt:   now,
		UpdatedAt:   now,
		Subsystems:  []*api.Subsystem{},
	}

	// Create system
	if err := store.CreateSystem(ctx, sys); err != nil {
		t.Fatalf("CreateSystem() error = %v", err)
	}

	// Delete system
	if err := store.DeleteSystem(ctx, sys.ID); err != nil {
		t.Fatalf("DeleteSystem() error = %v", err)
	}

	// Verify it's gone
	_, err := store.GetSystem(ctx, sys.ID)
	if err == nil {
		t.Error("GetSystem() after delete should return error")
	}
}

func TestCreateAndGetSubsystem(t *testing.T) {
	store := setupTestDB(t)
	defer cleanupTestDB(t, store)

	ctx := context.Background()
	now := time.Now()

	// Create parent system first
	sys := &api.System{
		ID:          "test-system-sub",
		Name:        "Test System",
		Description: "System for subsystem test",
		CreatedAt:   now,
		UpdatedAt:   now,
		Subsystems:  []*api.Subsystem{},
	}

	if err := store.CreateSystem(ctx, sys); err != nil {
		t.Fatalf("CreateSystem() error = %v", err)
	}

	// Create subsystem
	sub := &api.Subsystem{
		ID:          "test-subsystem-1",
		Name:        "Test Subsystem",
		Description: "A test subsystem",
		Type:        api.SubsystemTypeAquarium,
		Metadata:    map[string]string{"key": "value"},
		Devices:     []*api.Device{},
		Subsystems:  []*api.Subsystem{},
	}

	if err := store.CreateSubsystem(ctx, sub, sys.ID); err != nil {
		t.Fatalf("CreateSubsystem() error = %v", err)
	}

	// Get subsystem
	retrieved, err := store.GetSubsystem(ctx, sub.ID)
	if err != nil {
		t.Fatalf("GetSubsystem() error = %v", err)
	}

	if retrieved.ID != sub.ID {
		t.Errorf("GetSubsystem() ID = %v, want %v", retrieved.ID, sub.ID)
	}
	if retrieved.Name != sub.Name {
		t.Errorf("GetSubsystem() Name = %v, want %v", retrieved.Name, sub.Name)
	}
	if retrieved.Type != sub.Type {
		t.Errorf("GetSubsystem() Type = %v, want %v", retrieved.Type, sub.Type)
	}
	if retrieved.Metadata["key"] != "value" {
		t.Errorf("GetSubsystem() Metadata = %v, want key=value", retrieved.Metadata)
	}
}

func TestUpdateSubsystem(t *testing.T) {
	store := setupTestDB(t)
	defer cleanupTestDB(t, store)

	ctx := context.Background()
	now := time.Now()

	// Create parent system first
	sys := &api.System{
		ID:        "test-system-sub-update",
		Name:      "Test System",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := store.CreateSystem(ctx, sys); err != nil {
		t.Fatalf("CreateSystem() error = %v", err)
	}

	// Create subsystem
	sub := &api.Subsystem{
		ID:          "test-subsystem-update",
		Name:        "Original Subsystem",
		Description: "Original Description",
		Type:        api.SubsystemTypeAquarium,
		Metadata:    map[string]string{"key": "value"},
	}

	if err := store.CreateSubsystem(ctx, sub, sys.ID); err != nil {
		t.Fatalf("CreateSubsystem() error = %v", err)
	}

	// Update subsystem
	sub.Name = "Updated Subsystem"
	sub.Description = "Updated Description"
	sub.Type = api.SubsystemTypeHydroponics
	sub.Metadata["key2"] = "value2"

	if err := store.UpdateSubsystem(ctx, sub); err != nil {
		t.Fatalf("UpdateSubsystem() error = %v", err)
	}

	// Get and verify
	retrieved, err := store.GetSubsystem(ctx, sub.ID)
	if err != nil {
		t.Fatalf("GetSubsystem() error = %v", err)
	}

	if retrieved.Name != "Updated Subsystem" {
		t.Errorf("UpdateSubsystem() Name = %v, want Updated Subsystem", retrieved.Name)
	}
	if retrieved.Type != api.SubsystemTypeHydroponics {
		t.Errorf("UpdateSubsystem() Type = %v, want hydroponics", retrieved.Type)
	}
}

func TestDeleteSubsystem(t *testing.T) {
	store := setupTestDB(t)
	defer cleanupTestDB(t, store)

	ctx := context.Background()
	now := time.Now()

	sys := &api.System{
		ID:        "test-system-sub-delete",
		Name:      "Test System",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := store.CreateSystem(ctx, sys); err != nil {
		t.Fatalf("CreateSystem() error = %v", err)
	}

	sub := &api.Subsystem{
		ID:   "test-subsystem-delete",
		Name: "Test Subsystem",
		Type: api.SubsystemTypeAquarium,
	}

	if err := store.CreateSubsystem(ctx, sub, sys.ID); err != nil {
		t.Fatalf("CreateSubsystem() error = %v", err)
	}

	// Delete subsystem
	if err := store.DeleteSubsystem(ctx, sub.ID); err != nil {
		t.Fatalf("DeleteSubsystem() error = %v", err)
	}

	// Verify it's gone
	_, err := store.GetSubsystem(ctx, sub.ID)
	if err == nil {
		t.Error("GetSubsystem() after delete should return error")
	}
}

func TestCreateAndGetDevice(t *testing.T) {
	store := setupTestDB(t)
	defer cleanupTestDB(t, store)

	ctx := context.Background()
	now := time.Now()

	// Create system and subsystem first
	sys := &api.System{
		ID:        "test-system-dev",
		Name:      "Test System",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateSystem(ctx, sys); err != nil {
		t.Fatalf("CreateSystem() error = %v", err)
	}

	sub := &api.Subsystem{
		ID:   "test-subsystem-dev",
		Name: "Test Subsystem",
		Type: api.SubsystemTypeAquarium,
	}
	if err := store.CreateSubsystem(ctx, sub, sys.ID); err != nil {
		t.Fatalf("CreateSubsystem() error = %v", err)
	}

	// Create device
	dev := &api.Device{
		ID:          "test-device-1",
		Driver:      api.DriverShelly,
		Name:        "Test Device",
		Description: "A test device",
		Metadata:    map[string]string{"ip": "192.168.1.100"},
	}

	if err := store.CreateDevice(ctx, dev, sub.ID); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	// Get device
	retrieved, err := store.GetDevice(ctx, dev.ID)
	if err != nil {
		t.Fatalf("GetDevice() error = %v", err)
	}

	if retrieved.ID != dev.ID {
		t.Errorf("GetDevice() ID = %v, want %v", retrieved.ID, dev.ID)
	}
	if retrieved.Driver != dev.Driver {
		t.Errorf("GetDevice() Driver = %v, want %v", retrieved.Driver, dev.Driver)
	}
	if retrieved.Name != dev.Name {
		t.Errorf("GetDevice() Name = %v, want %v", retrieved.Name, dev.Name)
	}
	if retrieved.Metadata["ip"] != "192.168.1.100" {
		t.Errorf("GetDevice() Metadata = %v, want ip=192.168.1.100", retrieved.Metadata)
	}
}

func TestUpdateDevice(t *testing.T) {
	store := setupTestDB(t)
	defer cleanupTestDB(t, store)

	ctx := context.Background()
	now := time.Now()

	sys := &api.System{
		ID:        "test-system-dev-update",
		Name:      "Test System",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateSystem(ctx, sys); err != nil {
		t.Fatalf("CreateSystem() error = %v", err)
	}

	sub := &api.Subsystem{
		ID:   "test-subsystem-dev-update",
		Name: "Test Subsystem",
		Type: api.SubsystemTypeAquarium,
	}
	if err := store.CreateSubsystem(ctx, sub, sys.ID); err != nil {
		t.Fatalf("CreateSubsystem() error = %v", err)
	}

	dev := &api.Device{
		ID:          "test-device-update",
		Driver:      api.DriverShelly,
		Name:        "Original Device",
		Description: "Original Description",
		Metadata:    map[string]string{"key": "value"},
	}

	if err := store.CreateDevice(ctx, dev, sub.ID); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	// Update device
	dev.Name = "Updated Device"
	dev.Description = "Updated Description"
	dev.Metadata["key2"] = "value2"

	if err := store.UpdateDevice(ctx, dev); err != nil {
		t.Fatalf("UpdateDevice() error = %v", err)
	}

	// Get and verify
	retrieved, err := store.GetDevice(ctx, dev.ID)
	if err != nil {
		t.Fatalf("GetDevice() error = %v", err)
	}

	if retrieved.Name != "Updated Device" {
		t.Errorf("UpdateDevice() Name = %v, want Updated Device", retrieved.Name)
	}
}

func TestDeleteDevice(t *testing.T) {
	store := setupTestDB(t)
	defer cleanupTestDB(t, store)

	ctx := context.Background()
	now := time.Now()

	sys := &api.System{
		ID:        "test-system-dev-delete",
		Name:      "Test System",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateSystem(ctx, sys); err != nil {
		t.Fatalf("CreateSystem() error = %v", err)
	}

	sub := &api.Subsystem{
		ID:   "test-subsystem-dev-delete",
		Name: "Test Subsystem",
		Type: api.SubsystemTypeAquarium,
	}
	if err := store.CreateSubsystem(ctx, sub, sys.ID); err != nil {
		t.Fatalf("CreateSubsystem() error = %v", err)
	}

	dev := &api.Device{
		ID:     "test-device-delete",
		Driver: api.DriverShelly,
		Name:   "Test Device",
	}

	if err := store.CreateDevice(ctx, dev, sub.ID); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	// Delete device
	if err := store.DeleteDevice(ctx, dev.ID); err != nil {
		t.Fatalf("DeleteDevice() error = %v", err)
	}

	// Verify it's gone
	_, err := store.GetDevice(ctx, dev.ID)
	if err == nil {
		t.Error("GetDevice() after delete should return error")
	}
}

func TestStoreSensorReading(t *testing.T) {
	store := setupTestDB(t)
	defer cleanupTestDB(t, store)

	ctx := context.Background()
	now := time.Now()

	// Create system, subsystem, and device
	sys := &api.System{
		ID:        "test-system-sensor",
		Name:      "Test System",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateSystem(ctx, sys); err != nil {
		t.Fatalf("CreateSystem() error = %v", err)
	}

	sub := &api.Subsystem{
		ID:   "test-subsystem-sensor",
		Name: "Test Subsystem",
		Type: api.SubsystemTypeAquarium,
	}
	if err := store.CreateSubsystem(ctx, sub, sys.ID); err != nil {
		t.Fatalf("CreateSubsystem() error = %v", err)
	}

	dev := &api.Device{
		ID:     "test-device-sensor",
		Driver: api.DriverShelly,
		Name:   "Test Device",
	}
	if err := store.CreateDevice(ctx, dev, sub.ID); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	// Store sensor reading
	reading := &api.SensorReading{
		Value:     25.5,
		Unit:      api.UnitCelsius,
		Timestamp: now,
		Valid:     true,
	}

	err := store.StoreSensorReading(ctx, dev.ID, "temp-sensor-1", "Temperature Sensor", api.SensorTypeTemperature, reading)
	if err != nil {
		t.Fatalf("StoreSensorReading() error = %v", err)
	}

	// Get sensor reading
	filters := SensorReadingFilters{
		DeviceID: &dev.ID,
		Limit:    1,
	}

	readings, err := store.GetSensorReadings(ctx, filters)
	if err != nil {
		t.Fatalf("GetSensorReadings() error = %v", err)
	}

	if len(readings) != 1 {
		t.Fatalf("GetSensorReadings() returned %d readings, want 1", len(readings))
	}

	if readings[0].Value != 25.5 {
		t.Errorf("GetSensorReadings() Value = %v, want 25.5", readings[0].Value)
	}
	if readings[0].Unit != api.UnitCelsius {
		t.Errorf("GetSensorReadings() Unit = %v, want °C", readings[0].Unit)
	}
	if !readings[0].Valid {
		t.Errorf("GetSensorReadings() Valid = false, want true")
	}
}

func TestGetLatestSensorReading(t *testing.T) {
	store := setupTestDB(t)
	defer cleanupTestDB(t, store)

	ctx := context.Background()
	now := time.Now()

	// Setup system, subsystem, device
	sys := &api.System{
		ID:        "test-system-sensor-latest",
		Name:      "Test System",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateSystem(ctx, sys); err != nil {
		t.Fatalf("CreateSystem() error = %v", err)
	}

	sub := &api.Subsystem{
		ID:   "test-subsystem-sensor-latest",
		Name: "Test Subsystem",
		Type: api.SubsystemTypeAquarium,
	}
	if err := store.CreateSubsystem(ctx, sub, sys.ID); err != nil {
		t.Fatalf("CreateSubsystem() error = %v", err)
	}

	dev := &api.Device{
		ID:     "test-device-sensor-latest",
		Driver: api.DriverShelly,
		Name:   "Test Device",
	}
	if err := store.CreateDevice(ctx, dev, sub.ID); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	sensorID := "temp-sensor-latest"

	// Store multiple readings
	for i := 0; i < 3; i++ {
		reading := &api.SensorReading{
			Value:     20.0 + float64(i),
			Unit:      api.UnitCelsius,
			Timestamp: now.Add(time.Duration(i) * time.Minute),
			Valid:     true,
		}

		err := store.StoreSensorReading(ctx, dev.ID, sensorID, "Temperature Sensor", api.SensorTypeTemperature, reading)
		if err != nil {
			t.Fatalf("StoreSensorReading() error = %v", err)
		}
	}

	// Get latest reading
	latest, err := store.GetLatestSensorReading(ctx, sensorID)
	if err != nil {
		t.Fatalf("GetLatestSensorReading() error = %v", err)
	}

	if latest.Value != 22.0 {
		t.Errorf("GetLatestSensorReading() Value = %v, want 22.0", latest.Value)
	}
}

func TestDeleteOldSensorReadings(t *testing.T) {
	store := setupTestDB(t)
	defer cleanupTestDB(t, store)

	ctx := context.Background()
	now := time.Now()

	// Setup system, subsystem, device
	sys := &api.System{
		ID:        "test-system-sensor-delete",
		Name:      "Test System",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateSystem(ctx, sys); err != nil {
		t.Fatalf("CreateSystem() error = %v", err)
	}

	sub := &api.Subsystem{
		ID:   "test-subsystem-sensor-delete",
		Name: "Test Subsystem",
		Type: api.SubsystemTypeAquarium,
	}
	if err := store.CreateSubsystem(ctx, sub, sys.ID); err != nil {
		t.Fatalf("CreateSubsystem() error = %v", err)
	}

	dev := &api.Device{
		ID:     "test-device-sensor-delete",
		Driver: api.DriverShelly,
		Name:   "Test Device",
	}
	if err := store.CreateDevice(ctx, dev, sub.ID); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	// Store old and new readings
	oldReading := &api.SensorReading{
		Value:     20.0,
		Unit:      api.UnitCelsius,
		Timestamp: now.Add(-48 * time.Hour),
		Valid:     true,
	}
	newReading := &api.SensorReading{
		Value:     25.0,
		Unit:      api.UnitCelsius,
		Timestamp: now,
		Valid:     true,
	}

	sensorID := "temp-sensor-delete"

	if err := store.StoreSensorReading(ctx, dev.ID, sensorID, "Temperature Sensor", api.SensorTypeTemperature, oldReading); err != nil {
		t.Fatalf("StoreSensorReading() error = %v", err)
	}
	if err := store.StoreSensorReading(ctx, dev.ID, sensorID, "Temperature Sensor", api.SensorTypeTemperature, newReading); err != nil {
		t.Fatalf("StoreSensorReading() error = %v", err)
	}

	// Delete old readings
	cutoff := now.Add(-24 * time.Hour)
	deleted, err := store.DeleteOldSensorReadings(ctx, cutoff)
	if err != nil {
		t.Fatalf("DeleteOldSensorReadings() error = %v", err)
	}

	if deleted != 1 {
		t.Errorf("DeleteOldSensorReadings() deleted %d rows, want 1", deleted)
	}

	// Verify only new reading remains
	latest, err := store.GetLatestSensorReading(ctx, sensorID)
	if err != nil {
		t.Fatalf("GetLatestSensorReading() error = %v", err)
	}

	if latest.Value != 25.0 {
		t.Errorf("GetLatestSensorReading() Value = %v, want 25.0", latest.Value)
	}
}

func TestStoreActuatorState(t *testing.T) {
	store := setupTestDB(t)
	defer cleanupTestDB(t, store)

	ctx := context.Background()
	now := time.Now()

	// Setup system, subsystem, device
	sys := &api.System{
		ID:        "test-system-actuator",
		Name:      "Test System",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateSystem(ctx, sys); err != nil {
		t.Fatalf("CreateSystem() error = %v", err)
	}

	sub := &api.Subsystem{
		ID:   "test-subsystem-actuator",
		Name: "Test Subsystem",
		Type: api.SubsystemTypeAquarium,
	}
	if err := store.CreateSubsystem(ctx, sub, sys.ID); err != nil {
		t.Fatalf("CreateSubsystem() error = %v", err)
	}

	dev := &api.Device{
		ID:     "test-device-actuator",
		Driver: api.DriverShelly,
		Name:   "Test Device",
	}
	if err := store.CreateDevice(ctx, dev, sub.ID); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	// Store actuator state
	state := &api.ActuatorState{
		Active:     true,
		Parameters: map[string]float64{"brightness": 75.0},
		Timestamp:  now,
	}

	err := store.StoreActuatorState(ctx, dev.ID, "light-1", "Light", api.ActuatorTypeDimmableLight, state)
	if err != nil {
		t.Fatalf("StoreActuatorState() error = %v", err)
	}

	// Get actuator state
	filters := ActuatorStateFilters{
		DeviceID: &dev.ID,
		Limit:    1,
	}

	states, err := store.GetActuatorStates(ctx, filters)
	if err != nil {
		t.Fatalf("GetActuatorStates() error = %v", err)
	}

	if len(states) != 1 {
		t.Fatalf("GetActuatorStates() returned %d states, want 1", len(states))
	}

	if !states[0].Active {
		t.Errorf("GetActuatorStates() Active = false, want true")
	}
	if states[0].Parameters["brightness"] != 75.0 {
		t.Errorf("GetActuatorStates() brightness = %v, want 75.0", states[0].Parameters["brightness"])
	}
}

func TestGetLatestActuatorState(t *testing.T) {
	store := setupTestDB(t)
	defer cleanupTestDB(t, store)

	ctx := context.Background()
	now := time.Now()

	// Setup system, subsystem, device
	sys := &api.System{
		ID:        "test-system-actuator-latest",
		Name:      "Test System",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateSystem(ctx, sys); err != nil {
		t.Fatalf("CreateSystem() error = %v", err)
	}

	sub := &api.Subsystem{
		ID:   "test-subsystem-actuator-latest",
		Name: "Test Subsystem",
		Type: api.SubsystemTypeAquarium,
	}
	if err := store.CreateSubsystem(ctx, sub, sys.ID); err != nil {
		t.Fatalf("CreateSubsystem() error = %v", err)
	}

	dev := &api.Device{
		ID:     "test-device-actuator-latest",
		Driver: api.DriverShelly,
		Name:   "Test Device",
	}
	if err := store.CreateDevice(ctx, dev, sub.ID); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	actuatorID := "light-latest"

	// Store multiple states
	for i := 0; i < 3; i++ {
		state := &api.ActuatorState{
			Active:     i%2 == 0,
			Parameters: map[string]float64{"brightness": float64(25 * i)},
			Timestamp:  now.Add(time.Duration(i) * time.Minute),
		}

		err := store.StoreActuatorState(ctx, dev.ID, actuatorID, "Light", api.ActuatorTypeDimmableLight, state)
		if err != nil {
			t.Fatalf("StoreActuatorState() error = %v", err)
		}
	}

	// Get latest state
	latest, err := store.GetLatestActuatorState(ctx, actuatorID)
	if err != nil {
		t.Fatalf("GetLatestActuatorState() error = %v", err)
	}

	if !latest.Active {
		t.Errorf("GetLatestActuatorState() Active = false, want true")
	}
	if latest.Parameters["brightness"] != 50.0 {
		t.Errorf("GetLatestActuatorState() brightness = %v, want 50.0", latest.Parameters["brightness"])
	}
}

func TestDeleteOldActuatorStates(t *testing.T) {
	store := setupTestDB(t)
	defer cleanupTestDB(t, store)

	ctx := context.Background()
	now := time.Now()

	// Setup system, subsystem, device
	sys := &api.System{
		ID:        "test-system-actuator-delete",
		Name:      "Test System",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateSystem(ctx, sys); err != nil {
		t.Fatalf("CreateSystem() error = %v", err)
	}

	sub := &api.Subsystem{
		ID:   "test-subsystem-actuator-delete",
		Name: "Test Subsystem",
		Type: api.SubsystemTypeAquarium,
	}
	if err := store.CreateSubsystem(ctx, sub, sys.ID); err != nil {
		t.Fatalf("CreateSubsystem() error = %v", err)
	}

	dev := &api.Device{
		ID:     "test-device-actuator-delete",
		Driver: api.DriverShelly,
		Name:   "Test Device",
	}
	if err := store.CreateDevice(ctx, dev, sub.ID); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	actuatorID := "light-delete"

	// Store old and new states
	oldState := &api.ActuatorState{
		Active:     true,
		Parameters: map[string]float64{},
		Timestamp:  now.Add(-48 * time.Hour),
	}
	newState := &api.ActuatorState{
		Active:     false,
		Parameters: map[string]float64{},
		Timestamp:  now,
	}

	if err := store.StoreActuatorState(ctx, dev.ID, actuatorID, "Light", api.ActuatorTypeRelay, oldState); err != nil {
		t.Fatalf("StoreActuatorState() error = %v", err)
	}
	if err := store.StoreActuatorState(ctx, dev.ID, actuatorID, "Light", api.ActuatorTypeRelay, newState); err != nil {
		t.Fatalf("StoreActuatorState() error = %v", err)
	}

	// Delete old states
	cutoff := now.Add(-24 * time.Hour)
	deleted, err := store.DeleteOldActuatorStates(ctx, cutoff)
	if err != nil {
		t.Fatalf("DeleteOldActuatorStates() error = %v", err)
	}

	if deleted != 1 {
		t.Errorf("DeleteOldActuatorStates() deleted %d rows, want 1", deleted)
	}

	// Verify only new state remains
	latest, err := store.GetLatestActuatorState(ctx, actuatorID)
	if err != nil {
		t.Fatalf("GetLatestActuatorState() error = %v", err)
	}

	if latest.Active {
		t.Errorf("GetLatestActuatorState() Active = true, want false")
	}
}

func TestHierarchicalSystemWithSubsystems(t *testing.T) {
	store := setupTestDB(t)
	defer cleanupTestDB(t, store)

	ctx := context.Background()
	now := time.Now()

	// Create system with nested subsystems
	sys := &api.System{
		ID:          "test-system-hierarchy",
		Name:        "Test System",
		Description: "System with nested subsystems",
		CreatedAt:   now,
		UpdatedAt:   now,
		Subsystems: []*api.Subsystem{
			{
				ID:          "parent-sub",
				Name:        "Parent Subsystem",
				Type:        api.SubsystemTypeAquarium,
				Description: "Parent subsystem",
				Metadata:    map[string]string{"level": "1"},
				Devices: []*api.Device{
					{
						ID:          "parent-device",
						Driver:      api.DriverShelly,
						Name:        "Parent Device",
						Description: "Device in parent",
					},
				},
				Subsystems: []*api.Subsystem{
					{
						ID:          "child-sub",
						Name:        "Child Subsystem",
						Type:        api.SubsystemTypeFiltration,
						Description: "Child subsystem",
						Metadata:    map[string]string{"level": "2"},
						Devices: []*api.Device{
							{
								ID:          "child-device",
								Driver:      api.DriverShelly,
								Name:        "Child Device",
								Description: "Device in child",
							},
						},
					},
				},
			},
		},
	}

	// Create system with all nested structures
	if err := store.CreateSystem(ctx, sys); err != nil {
		t.Fatalf("CreateSystem() error = %v", err)
	}

	// Retrieve system and verify hierarchy
	retrieved, err := store.GetSystem(ctx, sys.ID)
	if err != nil {
		t.Fatalf("GetSystem() error = %v", err)
	}

	if len(retrieved.Subsystems) != 1 {
		t.Fatalf("GetSystem() returned %d subsystems, want 1", len(retrieved.Subsystems))
	}

	parentSub := retrieved.Subsystems[0]
	if parentSub.ID != "parent-sub" {
		t.Errorf("Parent subsystem ID = %v, want parent-sub", parentSub.ID)
	}
	if len(parentSub.Devices) != 1 {
		t.Errorf("Parent subsystem has %d devices, want 1", len(parentSub.Devices))
	}
	if len(parentSub.Subsystems) != 1 {
		t.Errorf("Parent subsystem has %d child subsystems, want 1", len(parentSub.Subsystems))
	}

	childSub := parentSub.Subsystems[0]
	if childSub.ID != "child-sub" {
		t.Errorf("Child subsystem ID = %v, want child-sub", childSub.ID)
	}
	if len(childSub.Devices) != 1 {
		t.Errorf("Child subsystem has %d devices, want 1", len(childSub.Devices))
	}
}
