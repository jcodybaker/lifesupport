package httpapi

import (
	"encoding/json"
	"net/http"

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

// Device handlers

func (h *Handler) CreateDevice(w http.ResponseWriter, r *http.Request) {
	var dev api.Device
	if err := json.NewDecoder(r.Body).Decode(&dev); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.Store.CreateDevice(ctx, &dev); err != nil {
		http.Error(w, "Failed to create device: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dev)
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

func (h *Handler) ListDevices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	devices, err := h.Store.ListDevices(ctx)
	if err != nil {
		http.Error(w, "Failed to list devices: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

// Sensor handlers

func (h *Handler) CreateSensor(w http.ResponseWriter, r *http.Request) {
	var sensor api.Sensor
	if err := json.NewDecoder(r.Body).Decode(&sensor); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.Store.CreateSensor(ctx, &sensor); err != nil {
		http.Error(w, "Failed to create sensor: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sensor)
}

func (h *Handler) GetSensor(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	deviceID := params["device_id"]
	sensorID := params["sensor_id"]

	ctx := r.Context()
	sensor, err := h.Store.GetSensor(ctx, deviceID, sensorID)
	if err != nil {
		http.Error(w, "Sensor not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sensor)
}

func (h *Handler) UpdateSensor(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	deviceID := params["device_id"]
	sensorID := params["sensor_id"]

	var sensor api.Sensor
	if err := json.NewDecoder(r.Body).Decode(&sensor); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	sensor.DeviceID = deviceID
	sensor.ID = sensorID

	ctx := r.Context()
	if err := h.Store.UpdateSensor(ctx, &sensor); err != nil {
		http.Error(w, "Failed to update sensor: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sensor)
}

func (h *Handler) DeleteSensor(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	deviceID := params["device_id"]
	sensorID := params["sensor_id"]

	ctx := r.Context()
	if err := h.Store.DeleteSensor(ctx, deviceID, sensorID); err != nil {
		http.Error(w, "Failed to delete sensor: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListSensors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check if device_id query parameter is provided
	deviceID := r.URL.Query().Get("device_id")

	var sensors []*api.Sensor
	var err error

	if deviceID != "" {
		sensors, err = h.Store.ListSensorsByDeviceID(ctx, deviceID)
	} else {
		sensors, err = h.Store.ListSensors(ctx)
	}

	if err != nil {
		http.Error(w, "Failed to list sensors: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sensors)
}

func (h *Handler) GetSensorByTag(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	tag := params["tag"]

	ctx := r.Context()
	sensor, err := h.Store.GetSensorByTag(ctx, tag)
	if err != nil {
		http.Error(w, "Sensor not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sensor)
}

// Actuator handlers

func (h *Handler) CreateActuator(w http.ResponseWriter, r *http.Request) {
	var actuator api.Actuator
	if err := json.NewDecoder(r.Body).Decode(&actuator); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.Store.CreateActuator(ctx, &actuator); err != nil {
		http.Error(w, "Failed to create actuator: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(actuator)
}

func (h *Handler) GetActuator(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	deviceID := params["device_id"]
	actuatorID := params["actuator_id"]

	ctx := r.Context()
	actuator, err := h.Store.GetActuator(ctx, deviceID, actuatorID)
	if err != nil {
		http.Error(w, "Actuator not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actuator)
}

func (h *Handler) UpdateActuator(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	deviceID := params["device_id"]
	actuatorID := params["actuator_id"]

	var actuator api.Actuator
	if err := json.NewDecoder(r.Body).Decode(&actuator); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	actuator.DeviceID = deviceID
	actuator.ID = actuatorID

	ctx := r.Context()
	if err := h.Store.UpdateActuator(ctx, &actuator); err != nil {
		http.Error(w, "Failed to update actuator: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actuator)
}

func (h *Handler) DeleteActuator(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	deviceID := params["device_id"]
	actuatorID := params["actuator_id"]

	ctx := r.Context()
	if err := h.Store.DeleteActuator(ctx, deviceID, actuatorID); err != nil {
		http.Error(w, "Failed to delete actuator: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListActuators(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check if device_id query parameter is provided
	deviceID := r.URL.Query().Get("device_id")

	var actuators []*api.Actuator
	var err error

	if deviceID != "" {
		actuators, err = h.Store.ListActuatorsByDeviceID(ctx, deviceID)
	} else {
		actuators, err = h.Store.ListActuators(ctx)
	}

	if err != nil {
		http.Error(w, "Failed to list actuators: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actuators)
}

func (h *Handler) GetActuatorByTag(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	tag := params["tag"]

	ctx := r.Context()
	actuator, err := h.Store.GetActuatorByTag(ctx, tag)
	if err != nil {
		http.Error(w, "Actuator not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actuator)
}
