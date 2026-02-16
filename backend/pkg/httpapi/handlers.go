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
