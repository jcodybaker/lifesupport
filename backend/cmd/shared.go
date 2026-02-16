package cmd

import (
	"context"
	"os"
	"time"

	"lifesupport/backend/pkg/storer"
	"lifesupport/backend/pkg/temporallog"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.temporal.io/sdk/client"
)

// CommonOptions holds configuration shared between commands
type CommonOptions struct {
	DB       string
	Temporal TemporalOptions
}

// TemporalOptions holds Temporal configuration
type TemporalOptions struct {
	Host              string
	Namespace         string
	TaskQueue         string
	Identity          string
	ConnectionTimeout time.Duration
}

// AddCommonFlags adds shared database and temporal flags to a command
func AddCommonFlags(cmd *cobra.Command, opts *CommonOptions) {
	// Database flags
	cmd.Flags().StringVar(&opts.DB, "db", "postgres://lifesupport:lifesupport@localhost:5432/lifesupport?sslmode=disable", "Database connection string")
	viper.BindPFlag("db", cmd.Flags().Lookup("db"))

	// Temporal flags
	cmd.Flags().StringVar(&opts.Temporal.Host, "temporal-host", "localhost:7233", "Temporal server host:port")
	cmd.Flags().StringVar(&opts.Temporal.Namespace, "temporal-namespace", "default", "Temporal namespace")
	cmd.Flags().StringVar(&opts.Temporal.Identity, "temporal-identity", "", "Temporal worker identity (defaults to hostname)")
	cmd.Flags().DurationVar(&opts.Temporal.ConnectionTimeout, "temporal-timeout", 10*time.Second, "Temporal connection timeout")
	viper.BindPFlag("temporal-host", cmd.Flags().Lookup("temporal-host"))
	viper.BindPFlag("temporal-namespace", cmd.Flags().Lookup("temporal-namespace"))
	viper.BindPFlag("temporal-identity", cmd.Flags().Lookup("temporal-identity"))
	viper.BindPFlag("temporal-timeout", cmd.Flags().Lookup("temporal-timeout"))
}

// LoadCommonOptions loads options from viper (which handles env vars and flags)
func LoadCommonOptions(opts *CommonOptions) {
	opts.DB = viper.GetString("db")
	opts.Temporal.Host = viper.GetString("temporal-host")
	opts.Temporal.Namespace = viper.GetString("temporal-namespace")
	opts.Temporal.Identity = viper.GetString("temporal-identity")
	opts.Temporal.ConnectionTimeout = viper.GetDuration("temporal-timeout")

	// Set default identity to hostname if not specified
	if opts.Temporal.Identity == "" {
		hostname, err := os.Hostname()
		if err != nil {
			opts.Temporal.Identity = "lifesupport"
		} else {
			opts.Temporal.Identity = hostname
		}
	}
}

// InitDatabase creates and initializes the database connection
func InitDatabase(ctx context.Context, connString string) (*storer.Storer, error) {
	store, err := storer.New(connString)
	if err != nil {
		return nil, err
	}

	if err := store.InitSchema(ctx); err != nil {
		store.Close()
		return nil, err
	}

	return store, nil
}

// InitTemporalClient creates a Temporal client with the given options
func InitTemporalClient(ctx context.Context, opts TemporalOptions) (client.Client, error) {
	c, err := client.DialContext(ctx, client.Options{
		HostPort:  opts.Host,
		Namespace: opts.Namespace,
		Identity:  opts.Identity,
		Logger:    temporallog.NewTemporalLogger(log.Logger),
	})
	if err != nil {
		return nil, err
	}

	log.Info().
		Str("host", opts.Host).
		Str("namespace", opts.Namespace).
		Str("identity", opts.Identity).
		Msg("Connected to Temporal")

	return c, nil
}
