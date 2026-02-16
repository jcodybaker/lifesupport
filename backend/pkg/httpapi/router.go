package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"
)

// SetupRouter creates and configures the API router
func (h *Handler) SetupRouter() *mux.Router {
	r := mux.NewRouter()

	// Device endpoints
	r.HandleFunc("/api/devices", h.CreateDevice).Methods("POST")
	r.HandleFunc("/api/devices", h.ListDevices).Methods("GET")
	r.HandleFunc("/api/devices/{id}", h.GetDevice).Methods("GET")
	r.HandleFunc("/api/devices/{id}", h.UpdateDevice).Methods("PUT")
	r.HandleFunc("/api/devices/{id}", h.DeleteDevice).Methods("DELETE")

	// Sensor endpoints
	r.HandleFunc("/api/sensors", h.CreateSensor).Methods("POST")
	r.HandleFunc("/api/sensors", h.ListSensors).Methods("GET")
	r.HandleFunc("/api/sensors/by-tag/{tag}", h.GetSensorByTag).Methods("GET")
	r.HandleFunc("/api/sensors/{device_id}/{sensor_id}", h.GetSensor).Methods("GET")
	r.HandleFunc("/api/sensors/{device_id}/{sensor_id}", h.UpdateSensor).Methods("PUT")
	r.HandleFunc("/api/sensors/{device_id}/{sensor_id}", h.DeleteSensor).Methods("DELETE")

	// Actuator endpoints
	r.HandleFunc("/api/actuators", h.CreateActuator).Methods("POST")
	r.HandleFunc("/api/actuators", h.ListActuators).Methods("GET")
	r.HandleFunc("/api/actuators/by-tag/{tag}", h.GetActuatorByTag).Methods("GET")
	r.HandleFunc("/api/actuators/{device_id}/{actuator_id}", h.GetActuator).Methods("GET")
	r.HandleFunc("/api/actuators/{device_id}/{actuator_id}", h.UpdateActuator).Methods("PUT")
	r.HandleFunc("/api/actuators/{device_id}/{actuator_id}", h.DeleteActuator).Methods("DELETE")

	// Workflow endpoints
	r.HandleFunc("/api/workflows/discovery", h.StartDiscoveryWorkflow).Methods("POST")
	r.HandleFunc("/api/workflows/{workflowId}", h.GetWorkflowStatus).Methods("GET")
	r.HandleFunc("/api/workflows", h.ListWorkflows).Methods("GET")

	// Enable CORS
	r.Use(CORSMiddleware)

	return r
}

// CORSMiddleware enables CORS for all routes
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
