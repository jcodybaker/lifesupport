package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// SetupRouter creates and configures the API router
func (h *Handler) SetupRouter() *mux.Router {
	r := mux.NewRouter()

	// System endpoints
	r.HandleFunc("/api/systems", h.CreateSystem).Methods("POST")
	r.HandleFunc("/api/systems/{id}", h.GetSystem).Methods("GET")
	r.HandleFunc("/api/systems/{id}", h.UpdateSystem).Methods("PUT")
	r.HandleFunc("/api/systems/{id}", h.DeleteSystem).Methods("DELETE")

	// Subsystem endpoints
	r.HandleFunc("/api/subsystems", h.CreateSubsystem).Methods("POST")
	r.HandleFunc("/api/subsystems/{id}", h.GetSubsystem).Methods("GET")
	r.HandleFunc("/api/subsystems/{id}", h.UpdateSubsystem).Methods("PUT")
	r.HandleFunc("/api/subsystems/{id}", h.DeleteSubsystem).Methods("DELETE")

	// Device endpoints
	r.HandleFunc("/api/devices", h.CreateDevice).Methods("POST")
	r.HandleFunc("/api/devices/{id}", h.GetDevice).Methods("GET")
	r.HandleFunc("/api/devices/{id}", h.UpdateDevice).Methods("PUT")
	r.HandleFunc("/api/devices/{id}", h.DeleteDevice).Methods("DELETE")

	// Sensor reading endpoints
	r.HandleFunc("/api/sensor-readings", h.StoreSensorReading).Methods("POST")
	r.HandleFunc("/api/sensor-readings", h.GetSensorReadings).Methods("GET")
	r.HandleFunc("/api/sensor-readings/{sensorId}/latest", h.GetLatestSensorReading).Methods("GET")

	// Actuator state endpoints
	r.HandleFunc("/api/actuator-states", h.StoreActuatorState).Methods("POST")
	r.HandleFunc("/api/actuator-states", h.GetActuatorStates).Methods("GET")
	r.HandleFunc("/api/actuator-states/{actuatorId}/latest", h.GetLatestActuatorState).Methods("GET")

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
