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
	r.HandleFunc("/api/devices/{id}", h.GetDevice).Methods("GET")
	r.HandleFunc("/api/devices/{id}", h.UpdateDevice).Methods("PUT")
	r.HandleFunc("/api/devices/{id}", h.DeleteDevice).Methods("DELETE")

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
