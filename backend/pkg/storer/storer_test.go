package storer

import (
	"context"
	"fmt"
	"os"
	"testing"

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

	// Clean up devices
	_, _ = store.db.ExecContext(ctx, "DELETE FROM devices")

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
	tables := []string{"devices"}

	for _, table := range tables {
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		var count int
		err := store.db.QueryRowContext(ctx, query).Scan(&count)
		if err != nil {
			t.Errorf("Table %s does not exist or is not accessible: %v", table, err)
		}
	}
}

func TestCreateAndGetDevice(t *testing.T) {
	store := setupTestDB(t)
	defer cleanupTestDB(t, store)

	ctx := context.Background()

	// Create device
	dev := &api.Device{
		ID:          "test-device-1",
		Driver:      api.DriverShelly,
		Name:        "Test Device",
		Description: "A test device",
		Metadata:    map[string]string{"ip": "192.168.1.100"},
	}

	if err := store.CreateDevice(ctx, dev); err != nil {
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

	dev := &api.Device{
		ID:          "test-device-update",
		Driver:      api.DriverShelly,
		Name:        "Original Device",
		Description: "Original Description",
		Metadata:    map[string]string{"key": "value"},
	}

	if err := store.CreateDevice(ctx, dev); err != nil {
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

	dev := &api.Device{
		ID:     "test-device-delete",
		Driver: api.DriverShelly,
		Name:   "Test Device",
	}

	if err := store.CreateDevice(ctx, dev); err != nil {
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

func TestCreateDeviceWithNestedSensorsAndActuators(t *testing.T) {
	store := setupTestDB(t)
	defer cleanupTestDB(t, store)

	ctx := context.Background()

	// Create device with nested sensors and actuators
	dev := &api.Device{
		ID:          "test-device-nested",
		Driver:      api.DriverShelly,
		Name:        "Test Device With Sensors",
		Description: "A test device with sensors and actuators",
		Metadata:    map[string]string{"location": "greenhouse"},
		Sensors: []*api.Sensor{
			{
				ID:         "temp-sensor-1",
				Name:       "Temperature Sensor",
				SensorType: api.SensorTypeTemperature,
				Metadata:   map[string]string{"unit": "celsius"},
				Tags:       []string{"device.test-device-nested.sensor.temp-sensor-1", "greenhouse.temp"},
			},
			{
				ID:         "ph-sensor-1",
				Name:       "pH Sensor",
				SensorType: api.SensorTypePH,
				Metadata:   map[string]string{"calibrated": "true"},
			},
		},
		Actuators: []*api.Actuator{
			{
				ID:           "pump-1",
				Name:         "Water Pump",
				ActuatorType: api.ActuatorTypePeristalticPump,
				Metadata:     map[string]string{"flow_rate": "100"},
				Tags:         []string{"device.test-device-nested.actuator.pump-1", "greenhouse.pump"},
			},
			{
				ID:           "light-1",
				Name:         "Grow Light",
				ActuatorType: api.ActuatorTypeDimmableLight,
			},
		},
	}

	// Create device with nested sensors and actuators
	if err := store.CreateDevice(ctx, dev); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	// Verify device was created
	retrievedDev, err := store.GetDevice(ctx, dev.ID)
	if err != nil {
		t.Fatalf("GetDevice() error = %v", err)
	}
	if retrievedDev.ID != dev.ID {
		t.Errorf("GetDevice() ID = %v, want %v", retrievedDev.ID, dev.ID)
	}

	// Verify sensors were created
	sensors, err := store.ListSensorsByDeviceID(ctx, dev.ID)
	if err != nil {
		t.Fatalf("ListSensorsByDeviceID() error = %v", err)
	}
	if len(sensors) != 2 {
		t.Fatalf("ListSensorsByDeviceID() returned %d sensors, want 2", len(sensors))
	}

	// Verify first sensor
	sensor1, err := store.GetSensor(ctx, dev.ID, "temp-sensor-1")
	if err != nil {
		t.Fatalf("GetSensor() error = %v", err)
	}
	if sensor1.Name != "Temperature Sensor" {
		t.Errorf("GetSensor() Name = %v, want Temperature Sensor", sensor1.Name)
	}
	if sensor1.SensorType != api.SensorTypeTemperature {
		t.Errorf("GetSensor() SensorType = %v, want %v", sensor1.SensorType, api.SensorTypeTemperature)
	}
	if sensor1.DeviceID != dev.ID {
		t.Errorf("GetSensor() DeviceID = %v, want %v", sensor1.DeviceID, dev.ID)
	}
	if len(sensor1.Tags) != 2 {
		t.Errorf("GetSensor() Tags length = %d, want 2", len(sensor1.Tags))
	}

	// Verify second sensor (should have default tag generated)
	sensor2, err := store.GetSensor(ctx, dev.ID, "ph-sensor-1")
	if err != nil {
		t.Fatalf("GetSensor() error = %v", err)
	}
	if sensor2.Name != "pH Sensor" {
		t.Errorf("GetSensor() Name = %v, want pH Sensor", sensor2.Name)
	}
	if len(sensor2.Tags) == 0 {
		t.Error("GetSensor() should have generated default tag")
	}

	// Verify actuators were created
	actuators, err := store.ListActuatorsByDeviceID(ctx, dev.ID)
	if err != nil {
		t.Fatalf("ListActuatorsByDeviceID() error = %v", err)
	}
	if len(actuators) != 2 {
		t.Fatalf("ListActuatorsByDeviceID() returned %d actuators, want 2", len(actuators))
	}

	// Verify first actuator
	actuator1, err := store.GetActuator(ctx, dev.ID, "pump-1")
	if err != nil {
		t.Fatalf("GetActuator() error = %v", err)
	}
	if actuator1.Name != "Water Pump" {
		t.Errorf("GetActuator() Name = %v, want Water Pump", actuator1.Name)
	}
	if actuator1.ActuatorType != api.ActuatorTypePeristalticPump {
		t.Errorf("GetActuator() ActuatorType = %v, want %v", actuator1.ActuatorType, api.ActuatorTypePeristalticPump)
	}
	if actuator1.DeviceID != dev.ID {
		t.Errorf("GetActuator() DeviceID = %v, want %v", actuator1.DeviceID, dev.ID)
	}
	if len(actuator1.Tags) != 2 {
		t.Errorf("GetActuator() Tags length = %d, want 2", len(actuator1.Tags))
	}

	// Verify second actuator (should have default tag generated)
	actuator2, err := store.GetActuator(ctx, dev.ID, "light-1")
	if err != nil {
		t.Fatalf("GetActuator() error = %v", err)
	}
	if actuator2.Name != "Grow Light" {
		t.Errorf("GetActuator() Name = %v, want Grow Light", actuator2.Name)
	}
	if len(actuator2.Tags) == 0 {
		t.Error("GetActuator() should have generated default tag")
	}

	// Test that we can look up sensors and actuators by tag
	sensorByTag, err := store.GetSensorByTag(ctx, "greenhouse.temp")
	if err != nil {
		t.Fatalf("GetSensorByTag() error = %v", err)
	}
	if sensorByTag.ID != "temp-sensor-1" {
		t.Errorf("GetSensorByTag() ID = %v, want temp-sensor-1", sensorByTag.ID)
	}

	actuatorByTag, err := store.GetActuatorByTag(ctx, "greenhouse.pump")
	if err != nil {
		t.Fatalf("GetActuatorByTag() error = %v", err)
	}
	if actuatorByTag.ID != "pump-1" {
		t.Errorf("GetActuatorByTag() ID = %v, want pump-1", actuatorByTag.ID)
	}

	// Verify cascading delete - deleting device should delete sensors and actuators
	if err := store.DeleteDevice(ctx, dev.ID); err != nil {
		t.Fatalf("DeleteDevice() error = %v", err)
	}

	// Verify sensors are deleted
	sensors, err = store.ListSensorsByDeviceID(ctx, dev.ID)
	if err != nil {
		t.Fatalf("ListSensorsByDeviceID() error = %v", err)
	}
	if len(sensors) != 0 {
		t.Errorf("ListSensorsByDeviceID() after device delete returned %d sensors, want 0", len(sensors))
	}

	// Verify actuators are deleted
	actuators, err = store.ListActuatorsByDeviceID(ctx, dev.ID)
	if err != nil {
		t.Fatalf("ListActuatorsByDeviceID() error = %v", err)
	}
	if len(actuators) != 0 {
		t.Errorf("ListActuatorsByDeviceID() after device delete returned %d actuators, want 0", len(actuators))
	}
}
