package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"lifesupport/backend/pkg/api"
	"lifesupport/backend/pkg/storer"
)

func main() {
	var err error
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "host=localhost port=5432 user=postgres password=postgres dbname=lifesupport sslmode=disable"
	}

	store, err := storer.New(connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer store.Close()

	// Initialize database schema
	ctx := context.Background()
	if err := store.InitSchema(ctx); err != nil {
		log.Fatal("Failed to initialize schema:", err)
	}

	// Create API handler and setup router
	handler := api.NewHandler(store)
	router := handler.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Life Support API server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
