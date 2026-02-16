package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lifesupport/backend/pkg/httpapi"
	"lifesupport/backend/pkg/storer"

	"github.com/rs/zerolog/log"
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
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer store.Close()

	// Initialize database schema
	ctx := context.Background()
	if err := store.InitSchema(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize schema")
	}

	// Create API handler and setup router
	handler := httpapi.NewHandler(store)
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
		log.Info().Str("port", httpPort).Msg("Life Support HTTP API server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server error")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down HTTP server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("HTTP server forced to shutdown")
	}

	log.Info().Msg("HTTP server stopped")
}
