package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lifesupport/backend/pkg/api"
	"lifesupport/backend/pkg/storer"

	"github.com/spf13/cobra"
)

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "Start the HTTP API server",
	Long:  `Start the HTTP API server that provides RESTful endpoints for the Life Support system.`,
	Run:   runHTTPServer,
}

var (
	httpPort string
)

func init() {
	rootCmd.AddCommand(httpCmd)
	httpCmd.Flags().StringVarP(&httpPort, "port", "p", "8080", "Port to run the HTTP server on")
}

func runHTTPServer(cmd *cobra.Command, args []string) {
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

	// Override port with flag if provided
	if httpPort == "" {
		httpPort = os.Getenv("PORT")
		if httpPort == "" {
			httpPort = "8080"
		}
	}

	server := &http.Server{
		Addr:    ":" + httpPort,
		Handler: router,
	}

	// Setup graceful shutdown
	go func() {
		log.Printf("Life Support HTTP API server starting on port %s", httpPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down HTTP server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("HTTP server forced to shutdown:", err)
	}

	log.Println("HTTP server stopped")
}
