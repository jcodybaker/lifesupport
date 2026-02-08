package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"lifesupport/backend/pkg/api"
	"lifesupport/backend/pkg/storer"
)

// Handler holds the dependencies for HTTP handlers
type Handler struct {
	Store *storer.Storer
}

// NewHandler creates a new Handler instance
func NewHandler(store *storer.Storer) *Handler {
	return &Handler{Store: store}
}

// System handlers

func (h *Handler) CreateSystem(w http.ResponseWriter, r *http.Request) {
	var sys api.System
	if err := json.NewDecoder(r.Body).Decode(&sys); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	sys.CreatedAt = time.Now()
	sys.UpdatedAt = time.Now()

	ctx := r.Context()
	if err := h.Store.CreateSystem(ctx, &sys); err != nil {
		http.Error(w, "Failed to create system: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sys)
}

func (h *Handler) GetSystem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	ctx := r.Context()
	sys, err := h.Store.GetSystem(ctx, id)
	if err != nil {
		http.Error(w, "System not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sys)
}

func (h *Handler) UpdateSystem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var sys api.System
	if err := json.NewDecoder(r.Body).Decode(&sys); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	sys.ID = id
	sys.UpdatedAt = time.Now()

	ctx := r.Context()
	if err := h.Store.UpdateSystem(ctx, &sys); err != nil {
		http.Error(w, "Failed to update system: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sys)
}

func (h *Handler) DeleteSystem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	ctx := r.Context()
	if err := h.Store.DeleteSystem(ctx, id); err != nil {
		http.Error(w, "Failed to delete system: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Subsystem handlers

type CreateSubsystemRequest struct {
	Subsystem api.Subsystem `json:"subsystem"`
	SystemID  string        `json:"system_id"`
}

func (h *Handler) CreateSubsystem(w http.ResponseWriter, r *http.Request) {
	var req CreateSubsystemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.Store.CreateSubsystem(ctx, &req.Subsystem, req.SystemID); err != nil {
		http.Error(w, "Failed to create subsystem: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req.Subsystem)
}

func (h *Handler) GetSubsystem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	ctx := r.Context()
	sub, err := h.Store.GetSubsystem(ctx, id)
	if err != nil {
		http.Error(w, "Subsystem not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sub)
}

func (h *Handler) UpdateSubsystem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var sub api.Subsystem
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	sub.ID = id

	ctx := r.Context()
	if err := h.Store.UpdateSubsystem(ctx, &sub); err != nil {
		http.Error(w, "Failed to update subsystem: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sub)
}

func (h *Handler) DeleteSubsystem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	ctx := r.Context()
	if err := h.Store.DeleteSubsystem(ctx, id); err != nil {
		http.Error(w, "Failed to delete subsystem: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Device handlers

type CreateDeviceRequest struct {
	Device      api.Device `json:"device"`
	SubsystemID string     `json:"subsystem_id"`
}

func (h *Handler) CreateDevice(w http.ResponseWriter, r *http.Request) {
	var req CreateDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.Store.CreateDevice(ctx, &req.Device, req.SubsystemID); err != nil {
		http.Error(w, "Failed to create device: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req.Device)
}

func (h *Handler) GetDevice(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	ctx := r.Context()
	dev, err := h.Store.GetDevice(ctx, id)
	if err != nil {
		http.Error(w, "Device not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dev)
}

func (h *Handler) UpdateDevice(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var dev api.Device
	if err := json.NewDecoder(r.Body).Decode(&dev); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	dev.ID = id

	ctx := r.Context()
	if err := h.Store.UpdateDevice(ctx, &dev); err != nil {
		http.Error(w, "Failed to update device: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dev)
}

func (h *Handler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	ctx := r.Context()
	if err := h.Store.DeleteDevice(ctx, id); err != nil {
		http.Error(w, "Failed to delete device: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Sensor reading handlers

type StoreSensorReadingRequest struct {
	DeviceID   string            `json:"device_id"`
	SensorID   string            `json:"sensor_id"`
	SensorName string            `json:"sensor_name"`
	SensorType api.SensorType    `json:"sensor_type"`
	Reading    api.SensorReading `json:"reading"`
}

func (h *Handler) StoreSensorReading(w http.ResponseWriter, r *http.Request) {
	var req StoreSensorReadingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.Store.StoreSensorReading(ctx, req.DeviceID, req.SensorID, req.SensorName, req.SensorType, &req.Reading); err != nil {
		http.Error(w, "Failed to store sensor reading: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) GetSensorReadings(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	filters := storer.SensorReadingFilters{}

	if deviceID := query.Get("device_id"); deviceID != "" {
		filters.DeviceID = &deviceID
	}

	if sensorID := query.Get("sensor_id"); sensorID != "" {
		filters.SensorID = &sensorID
	}

	if sensorType := query.Get("sensor_type"); sensorType != "" {
		st := api.SensorType(sensorType)
		filters.SensorType = &st
	}

	if startTime := query.Get("start_time"); startTime != "" {
		t, err := time.Parse(time.RFC3339, startTime)
		if err != nil {
			http.Error(w, "Invalid start_time format: "+err.Error(), http.StatusBadRequest)
			return
		}
		filters.StartTime = &t
	}

	if endTime := query.Get("end_time"); endTime != "" {
		t, err := time.Parse(time.RFC3339, endTime)
		if err != nil {
			http.Error(w, "Invalid end_time format: "+err.Error(), http.StatusBadRequest)
			return
		}
		filters.EndTime = &t
	}

	if limit := query.Get("limit"); limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			http.Error(w, "Invalid limit format: "+err.Error(), http.StatusBadRequest)
			return
		}
		filters.Limit = l
	}

	ctx := r.Context()
	readings, err := h.Store.GetSensorReadings(ctx, filters)
	if err != nil {
		http.Error(w, "Failed to get sensor readings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(readings)
}

func (h *Handler) GetLatestSensorReading(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sensorID := params["sensorId"]

	ctx := r.Context()
	reading, err := h.Store.GetLatestSensorReading(ctx, sensorID)
	if err != nil {
		http.Error(w, "Reading not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reading)
}

// Actuator state handlers

type StoreActuatorStateRequest struct {
	DeviceID     string            `json:"device_id"`
	ActuatorID   string            `json:"actuator_id"`
	ActuatorName string            `json:"actuator_name"`
	ActuatorType api.ActuatorType  `json:"actuator_type"`
	State        api.ActuatorState `json:"state"`
}

func (h *Handler) StoreActuatorState(w http.ResponseWriter, r *http.Request) {
	var req StoreActuatorStateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.Store.StoreActuatorState(ctx, req.DeviceID, req.ActuatorID, req.ActuatorName, req.ActuatorType, &req.State); err != nil {
		http.Error(w, "Failed to store actuator state: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) GetActuatorStates(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	filters := storer.ActuatorStateFilters{}

	if deviceID := query.Get("device_id"); deviceID != "" {
		filters.DeviceID = &deviceID
	}

	if actuatorID := query.Get("actuator_id"); actuatorID != "" {
		filters.ActuatorID = &actuatorID
	}

	if actuatorType := query.Get("actuator_type"); actuatorType != "" {
		at := api.ActuatorType(actuatorType)
		filters.ActuatorType = &at
	}

	if startTime := query.Get("start_time"); startTime != "" {
		t, err := time.Parse(time.RFC3339, startTime)
		if err != nil {
			http.Error(w, "Invalid start_time format: "+err.Error(), http.StatusBadRequest)
			return
		}
		filters.StartTime = &t
	}

	if endTime := query.Get("end_time"); endTime != "" {
		t, err := time.Parse(time.RFC3339, endTime)
		if err != nil {
			http.Error(w, "Invalid end_time format: "+err.Error(), http.StatusBadRequest)
			return
		}
		filters.EndTime = &t
	}

	if limit := query.Get("limit"); limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			http.Error(w, "Invalid limit format: "+err.Error(), http.StatusBadRequest)
			return
		}
		filters.Limit = l
	}

	ctx := r.Context()
	states, err := h.Store.GetActuatorStates(ctx, filters)
	if err != nil {
		http.Error(w, "Failed to get actuator states: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(states)
}

func (h *Handler) GetLatestActuatorState(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	actuatorID := params["actuatorId"]

	ctx := r.Context()
	state, err := h.Store.GetLatestActuatorState(ctx, actuatorID)
	if err != nil {
		http.Error(w, "State not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

// Maintenance handlers

type CleanupRequest struct {
	DaysOld int `json:"days_old"`
}
