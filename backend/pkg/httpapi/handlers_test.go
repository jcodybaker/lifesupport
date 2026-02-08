package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"lifesupport/backend/pkg/api"
	"lifesupport/backend/pkg/storer"
)

// Test helpers

func setupTestDB(t *testing.T) *storer.Storer {
	t.Helper()

	// Use a test database connection
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=lifesupport_test sslmode=disable"
	store, err := storer.New(connStr)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
		return nil
	}

	// Initialize schema
	ctx := context.Background()
	if err := store.InitSchema(ctx); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}

	// Clean up existing test data
	cleanupTestData(t, store)

	return store
}

func cleanupTestData(t *testing.T, store *storer.Storer) {
	t.Helper()
	ctx := context.Background()

	// Delete test systems (cascades to everything else)
	_ = store.DeleteSystem(ctx, "test-sys-001")
	_ = store.DeleteSystem(ctx, "test-sys-002")
}

func teardownTestDB(t *testing.T, store *storer.Storer) {
	t.Helper()
	if store != nil {
		cleanupTestData(t, store)
		store.Close()
	}
}

// System Tests

func TestCreateSystem(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}
	defer teardownTestDB(t, store)

	handler := NewHandler(store)

	system := api.System{
		ID:          "test-sys-001",
		Name:        "Test System",
		Description: "Test system description",
		Subsystems:  []*api.Subsystem{},
	}

	body, _ := json.Marshal(system)
	req := httptest.NewRequest("POST", "/api/systems", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateSystem(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var result api.System
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.ID != system.ID {
		t.Errorf("Expected ID %s, got %s", system.ID, result.ID)
	}

	if result.Name != system.Name {
		t.Errorf("Expected Name %s, got %s", system.Name, result.Name)
	}
}

func TestGetSystem(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}
	defer teardownTestDB(t, store)

	// Create a test system first
	ctx := context.Background()
	system := &api.System{
		ID:          "test-sys-001",
		Name:        "Test System",
		Description: "Test description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := store.CreateSystem(ctx, system); err != nil {
		t.Fatalf("Failed to create test system: %v", err)
	}

	handler := NewHandler(store)

	req := httptest.NewRequest("GET", "/api/systems/test-sys-001", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "test-sys-001"})
	w := httptest.NewRecorder()

	handler.GetSystem(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var result api.System
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.ID != system.ID {
		t.Errorf("Expected ID %s, got %s", system.ID, result.ID)
	}
}

func TestGetSystemNotFound(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}
	defer teardownTestDB(t, store)

	handler := NewHandler(store)

	req := httptest.NewRequest("GET", "/api/systems/nonexistent", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "nonexistent"})
	w := httptest.NewRecorder()

	handler.GetSystem(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUpdateSystem(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}
	defer teardownTestDB(t, store)

	// Create a test system first
	ctx := context.Background()
	system := &api.System{
		ID:          "test-sys-001",
		Name:        "Original Name",
		Description: "Original description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := store.CreateSystem(ctx, system); err != nil {
		t.Fatalf("Failed to create test system: %v", err)
	}

	handler := NewHandler(store)

	update := map[string]string{
		"name":        "Updated Name",
		"description": "Updated description",
	}

	body, _ := json.Marshal(update)
	req := httptest.NewRequest("PUT", "/api/systems/test-sys-001", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "test-sys-001"})
	w := httptest.NewRecorder()

	handler.UpdateSystem(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var result api.System
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.Name != "Updated Name" {
		t.Errorf("Expected Name 'Updated Name', got %s", result.Name)
	}
}

func TestDeleteSystem(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}
	defer teardownTestDB(t, store)

	// Create a test system first
	ctx := context.Background()
	system := &api.System{
		ID:          "test-sys-001",
		Name:        "Test System",
		Description: "Test description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := store.CreateSystem(ctx, system); err != nil {
		t.Fatalf("Failed to create test system: %v", err)
	}

	handler := NewHandler(store)

	req := httptest.NewRequest("DELETE", "/api/systems/test-sys-001", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "test-sys-001"})
	w := httptest.NewRecorder()

	handler.DeleteSystem(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	// Verify it's deleted
	_, err := store.GetSystem(ctx, "test-sys-001")
	if err == nil {
		t.Error("Expected system to be deleted, but it still exists")
	}
}

// Sensor Reading Tests

func TestStoreSensorReading(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}
	defer teardownTestDB(t, store)

	// Create system, subsystem, and device first
	ctx := context.Background()
	system := &api.System{
		ID:        "test-sys-001",
		Name:      "Test System",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := store.CreateSystem(ctx, system); err != nil {
		t.Fatalf("Failed to create test system: %v", err)
	}

	subsystem := &api.Subsystem{
		ID:   "test-sub-001",
		Name: "Test Subsystem",
		Type: api.SubsystemTypeAquarium,
	}
	if err := store.CreateSubsystem(ctx, subsystem, "test-sys-001"); err != nil {
		t.Fatalf("Failed to create test subsystem: %v", err)
	}

	dev := &api.Device{
		ID:     "test-dev-001",
		Driver: api.DriverShelly,
		Name:   "Test Device",
	}
	if err := store.CreateDevice(ctx, dev, "test-sub-001"); err != nil {
		t.Fatalf("Failed to create test device: %v", err)
	}

	handler := NewHandler(store)

	request := StoreSensorReadingRequest{
		DeviceID:   "test-dev-001",
		SensorID:   "test-sensor-001",
		SensorName: "Temperature Sensor",
		SensorType: api.SensorTypeTemperature,
		Reading: api.SensorReading{
			Value:     25.5,
			Unit:      api.UnitCelsius,
			Timestamp: time.Now(),
			Valid:     true,
		},
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/sensor-readings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.StoreSensorReading(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestGetSensorReadings(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}
	defer teardownTestDB(t, store)

	// Setup test data
	ctx := context.Background()
	system := &api.System{
		ID:        "test-sys-001",
		Name:      "Test System",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.CreateSystem(ctx, system)

	subsystem := &api.Subsystem{
		ID:   "test-sub-001",
		Name: "Test Subsystem",
		Type: api.SubsystemTypeAquarium,
	}
	store.CreateSubsystem(ctx, subsystem, "test-sys-001")

	dev := &api.Device{
		ID:     "test-dev-001",
		Driver: api.DriverShelly,
		Name:   "Test Device",
	}
	store.CreateDevice(ctx, dev, "test-sub-001")

	// Store a reading
	reading := &api.SensorReading{
		Value:     25.5,
		Unit:      api.UnitCelsius,
		Timestamp: time.Now(),
		Valid:     true,
	}
	store.StoreSensorReading(ctx, "test-dev-001", "test-sensor-001", "Temperature Sensor", api.SensorTypeTemperature, reading)

	handler := NewHandler(store)

	req := httptest.NewRequest("GET", "/api/sensor-readings?device_id=test-dev-001", nil)
	w := httptest.NewRecorder()

	handler.GetSensorReadings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var readings []*api.SensorReading
	if err := json.Unmarshal(w.Body.Bytes(), &readings); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(readings) == 0 {
		t.Error("Expected at least one reading, got none")
	}
}

func TestGetLatestSensorReading(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}
	defer teardownTestDB(t, store)

	// Setup test data
	ctx := context.Background()
	system := &api.System{
		ID:        "test-sys-001",
		Name:      "Test System",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.CreateSystem(ctx, system)

	subsystem := &api.Subsystem{
		ID:   "test-sub-001",
		Name: "Test Subsystem",
		Type: api.SubsystemTypeAquarium,
	}
	store.CreateSubsystem(ctx, subsystem, "test-sys-001")

	dev := &api.Device{
		ID:     "test-dev-001",
		Driver: api.DriverShelly,
		Name:   "Test Device",
	}
	store.CreateDevice(ctx, dev, "test-sub-001")

	// Store readings
	for i := 0; i < 3; i++ {
		reading := &api.SensorReading{
			Value:     float64(25 + i),
			Unit:      api.UnitCelsius,
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			Valid:     true,
		}
		store.StoreSensorReading(ctx, "test-dev-001", "test-sensor-001", "Temperature Sensor", api.SensorTypeTemperature, reading)
	}

	handler := NewHandler(store)

	req := httptest.NewRequest("GET", "/api/sensor-readings/test-sensor-001/latest", nil)
	req = mux.SetURLVars(req, map[string]string{"sensorId": "test-sensor-001"})
	w := httptest.NewRecorder()

	handler.GetLatestSensorReading(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var reading api.SensorReading
	if err := json.Unmarshal(w.Body.Bytes(), &reading); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Should get the latest reading (27)
	if reading.Value != 27 {
		t.Errorf("Expected latest value 27, got %f", reading.Value)
	}
}

// Actuator State Tests

func TestStoreActuatorState(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}
	defer teardownTestDB(t, store)

	// Setup test data
	ctx := context.Background()
	system := &api.System{
		ID:        "test-sys-001",
		Name:      "Test System",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.CreateSystem(ctx, system)

	subsystem := &api.Subsystem{
		ID:   "test-sub-001",
		Name: "Test Subsystem",
		Type: api.SubsystemTypeAquarium,
	}
	store.CreateSubsystem(ctx, subsystem, "test-sys-001")

	dev := &api.Device{
		ID:     "test-dev-001",
		Driver: api.DriverShelly,
		Name:   "Test Device",
	}
	store.CreateDevice(ctx, dev, "test-sub-001")

	handler := NewHandler(store)

	request := StoreActuatorStateRequest{
		DeviceID:     "test-dev-001",
		ActuatorID:   "test-actuator-001",
		ActuatorName: "Test Pump",
		ActuatorType: api.ActuatorTypeRelay,
		State: api.ActuatorState{
			Active:    true,
			Timestamp: time.Now(),
		},
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/actuator-states", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.StoreActuatorState(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

// Invalid Input Tests

func TestCreateSystemInvalidJSON(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}
	defer teardownTestDB(t, store)

	handler := NewHandler(store)

	req := httptest.NewRequest("POST", "/api/systems", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateSystem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetSensorReadingsInvalidTimeFormat(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}
	defer teardownTestDB(t, store)

	handler := NewHandler(store)

	req := httptest.NewRequest("GET", "/api/sensor-readings?start_time=invalid", nil)
	w := httptest.NewRecorder()

	handler.GetSensorReadings(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
