package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	logFormat string
	logLevel  string
)

var rootCmd = &cobra.Command{
	Use:   "lifesupport-backend",
	Short: "Life Support Backend Service",
	Long:  `Life Support Backend Service provides HTTP API and Temporal worker functionality.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initLogger()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	ctx := context.Background()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Configure Viper for automatic environment variable binding
	viper.AutomaticEnv()

	// Global flags for logging configuration
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "pretty", "Log output format (json or pretty)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
}

// initLogger initializes the global zerolog logger based on the provided flags
func initLogger() {
	// Set log level
	level := zerolog.InfoLevel
	switch strings.ToLower(logLevel) {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn", "warning":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	default:
		fmt.Fprintf(os.Stderr, "Invalid log level '%s', defaulting to 'info'\n", logLevel)
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Set log format
	switch strings.ToLower(logFormat) {
	case "json":
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	case "pretty":
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05",
		})
	default:
		fmt.Fprintf(os.Stderr, "Invalid log format '%s', defaulting to 'pretty'\n", logFormat)
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05",
		})
	}
}
