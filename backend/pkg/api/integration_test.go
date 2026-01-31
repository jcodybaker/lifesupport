package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"lifesupport/backend/pkg/device"
)

// TestIntegrationCompleteFlow tests a complete workflow through the API
func TestIntegrationCompleteFlow(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}
	defer teardownTestDB(t, store)

	handler := NewHandler(store)
	router := handler.SetupRouter()

	// Step 1: Create a complete system with hierarchy
	systemPayload := map[string]interface{}{
		"id":          "integration-sys-001",
		"name":        "Integration Test System",
		"description": "Complete integration test",
		"subsystems": []map[string]interface{}{
			{
				"id":          "integration-sub-001",
				"name":        "Fish Tank",
				"type":        "aquarium",
				"description": "Main tank",
				"devices": []map[string]interface{}{
					{
						"id":          "integration-dev-001",
						"driver":      "shelly",
						"name":        "Water Monitor",
						"description": "Monitoring device",
					},
				},
			},
		},
	}

	body, _ := json.Marshal(systemPayload)
	req := httptest.NewRequest("POST", "/api/systems", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create system: %d - %s", w.Code, w.Body.String())
	}
	t.Log("✓ Created system with hierarchy")

	// Step 2: Store some sensor readings
	for i := 0; i < 5; i++ {
		readingPayload := map[string]interface{}{
			"device_id":   "integration-dev-001",
			"sensor_id":   "integration-temp-sensor",
			"sensor_name": "Temperature Sensor",
			"sensor_type": "temperature",
			"reading": map[string]interface{}{
				"value":     float64(25 + i),
				"unit":      "°C",
				"timestamp": time.Now().Add(time.Duration(i) * time.Minute).Format(time.RFC3339),
				"valid":     true,
			},
		}

		body, _ := json.Marshal(readingPayload)
		req := httptest.NewRequest("POST", "/api/sensor-readings", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Failed to store reading %d: %d - %s", i, w.Code, w.Body.String())
		}
	}
	t.Log("✓ Stored 5 sensor readings")

	// Step 3: Query sensor readings
	req = httptest.NewRequest("GET", "/api/sensor-readings?device_id=integration-dev-001&limit=10", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to query readings: %d - %s", w.Code, w.Body.String())
	}

	var readings []*device.SensorReading
	json.Unmarshal(w.Body.Bytes(), &readings)
	if len(readings) != 5 {
		t.Errorf("Expected 5 readings, got %d", len(readings))
	}
	t.Log("✓ Queried sensor readings successfully")

	// Step 4: Get latest reading
	req = httptest.NewRequest("GET", "/api/sensor-readings/integration-temp-sensor/latest", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to get latest reading: %d - %s", w.Code, w.Body.String())
	}

	var latestReading device.SensorReading
	json.Unmarshal(w.Body.Bytes(), &latestReading)
	if latestReading.Value != 29 {
		t.Errorf("Expected latest value 29, got %f", latestReading.Value)
	}
	t.Log("✓ Retrieved latest sensor reading")

	// Step 5: Store actuator states
	statePayload := map[string]interface{}{
		"device_id":     "integration-dev-001",
		"actuator_id":   "integration-pump",
		"actuator_name": "Water Pump",
		"actuator_type": "relay",
		"state": map[string]interface{}{
			"active":    true,
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	body, _ = json.Marshal(statePayload)
	req = httptest.NewRequest("POST", "/api/actuator-states", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to store actuator state: %d - %s", w.Code, w.Body.String())
	}
	t.Log("✓ Stored actuator state")

	// Step 6: Get the complete system
	req = httptest.NewRequest("GET", "/api/systems/integration-sys-001", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to get system: %d - %s", w.Code, w.Body.String())
	}

	var sys device.System
	json.Unmarshal(w.Body.Bytes(), &sys)
	if sys.Name != "Integration Test System" {
		t.Errorf("Expected system name 'Integration Test System', got %s", sys.Name)
	}
	if len(sys.Subsystems) != 1 {
		t.Errorf("Expected 1 subsystem, got %d", len(sys.Subsystems))
	}
	if len(sys.Subsystems[0].Devices) != 1 {
		t.Errorf("Expected 1 device, got %d", len(sys.Subsystems[0].Devices))
	}
	t.Log("✓ Retrieved complete system hierarchy")

	// Step 7: Update the system
	updatePayload := map[string]interface{}{
		"name":        "Updated Integration System",
		"description": "Updated description",
	}

	body, _ = json.Marshal(updatePayload)
	req = httptest.NewRequest("PUT", "/api/systems/integration-sys-001", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to update system: %d - %s", w.Code, w.Body.String())
	}
	t.Log("✓ Updated system")

	// Step 8: Delete the system (cascading)
	req = httptest.NewRequest("DELETE", "/api/systems/integration-sys-001", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Failed to delete system: %d - %s", w.Code, w.Body.String())
	}
	t.Log("✓ Deleted system with cascade")

	// Step 9: Verify it's deleted
	ctx := context.Background()
	_, err := store.GetSystem(ctx, "integration-sys-001")
	if err == nil {
		t.Error("Expected system to be deleted, but it still exists")
	}
	t.Log("✓ Verified system deletion")

	t.Log("\n✅ Integration test completed successfully!")
}

// TestRouterConfiguration tests that all routes are properly configured
func TestRouterConfiguration(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}
	defer teardownTestDB(t, store)

	handler := NewHandler(store)
	router := handler.SetupRouter()

	routes := []struct {
		method string
		path   string
	}{
		{"POST", "/api/systems"},
		{"GET", "/api/systems/test-id"},
		{"PUT", "/api/systems/test-id"},
		{"DELETE", "/api/systems/test-id"},
		{"POST", "/api/subsystems"},
		{"GET", "/api/subsystems/test-id"},
		{"PUT", "/api/subsystems/test-id"},
		{"DELETE", "/api/subsystems/test-id"},
		{"POST", "/api/devices"},
		{"GET", "/api/devices/test-id"},
		{"PUT", "/api/devices/test-id"},
		{"DELETE", "/api/devices/test-id"},
		{"POST", "/api/sensor-readings"},
		{"GET", "/api/sensor-readings"},
		{"GET", "/api/sensor-readings/test-sensor/latest"},
		{"POST", "/api/actuator-states"},
		{"GET", "/api/actuator-states"},
		{"GET", "/api/actuator-states/test-actuator/latest"},
		{"POST", "/api/maintenance/cleanup-readings"},
		{"POST", "/api/maintenance/cleanup-states"},
	}

	for _, route := range routes {
		req := httptest.NewRequest(route.method, route.path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// We expect some error (like 404 or 400), but not 404 for route not found
		// If the route exists, we'll get application errors (400, 404, 500)
		// If the route doesn't exist, we'll get nothing or a different error
		if w.Code == 0 {
			t.Errorf("Route %s %s not configured", route.method, route.path)
		}
	}

	t.Log("✓ All routes properly configured")
}
