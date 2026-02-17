package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"lifesupport/backend/pkg/httpapi"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "Start the HTTP API server",
	Long:  `Start the HTTP API server that provides RESTful endpoints for the Life Support system.`,
	Run:   runHTTPServer,
}

var (
	httpPort    string
	httpOptions CommonOptions
)

func init() {
	rootCmd.AddCommand(httpCmd)

	// Configure Viper for automatic environment variable binding
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// HTTP-specific flags
	httpCmd.Flags().StringVarP(&httpPort, "port", "p", "8080", "Port to run the HTTP server on")
	viper.BindPFlag("port", httpCmd.Flags().Lookup("port"))

	// Add common database and temporal flags
	AddCommonFlags(httpCmd, &httpOptions)
}

func runHTTPServer(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	// Load configuration from flags and environment variables
	LoadCommonOptions(&httpOptions)
	httpPort = viper.GetString("port")

	// Initialize database
	store, err := InitDatabase(ctx, httpOptions.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer store.Close()
	log.Info().Str("db", httpOptions.DB).Msg("Connected to database")

	// Initialize ClickHouse client
	clickhouseConn, err := InitClickHouse(ctx, httpOptions.ClickHouse)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to connect to ClickHouse")
	} else {
		defer clickhouseConn.Close()
	}

	// Create Temporal client (optional - server will still work without it)
	temporalClient, err := InitTemporalClient(ctx, httpOptions.Temporal)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to connect to Temporal - workflow endpoints will not be available")
		temporalClient = nil
	}
	if temporalClient != nil {
		defer temporalClient.Close()
	}

	// Create API handler and setup router
	handler := httpapi.NewHandler(store, temporalClient)
	router := handler.SetupRouter()

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
