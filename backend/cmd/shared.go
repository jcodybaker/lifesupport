package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"lifesupport/backend/pkg/storer"
	"lifesupport/backend/pkg/temporallog"

	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
)

// CommonOptions holds configuration shared between commands
type CommonOptions struct {
	DB         string
	Temporal   TemporalOptions
	ClickHouse ClickHouseOptions
}

// TemporalOptions holds Temporal configuration
type TemporalOptions struct {
	Host              string
	Namespace         string
	TaskQueue         string
	Identity          string
	ConnectionTimeout time.Duration
}

// ClickHouseOptions holds ClickHouse configuration
type ClickHouseOptions struct {
	Addrs           []string
	Database        string
	Username        string
	Password        string
	DialTimeout     time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	TLS             bool
}

// AddCommonFlags adds shared database and temporal flags to a command
func AddCommonFlags(cmd *cobra.Command, opts *CommonOptions) {
	// Database flags
	cmd.Flags().StringVar(&opts.DB, "db", "postgres://lifesupport:lifesupport@localhost:5432/lifesupport?sslmode=disable", "Database connection string")

	// Temporal flags
	cmd.Flags().StringVar(&opts.Temporal.Host, "temporal-host", "localhost:7233", "Temporal server host:port")
	cmd.Flags().StringVar(&opts.Temporal.Namespace, "temporal-namespace", "default", "Temporal namespace")
	cmd.Flags().StringVar(&opts.Temporal.Identity, "temporal-identity", "", "Temporal worker identity (defaults to hostname)")
	cmd.Flags().DurationVar(&opts.Temporal.ConnectionTimeout, "temporal-timeout", 10*time.Second, "Temporal connection timeout")

	// ClickHouse flags
	cmd.Flags().StringSliceVar(&opts.ClickHouse.Addrs, "clickhouse-addrs", []string{"localhost:9000"}, "ClickHouse server addresses")
	cmd.Flags().StringVar(&opts.ClickHouse.Database, "clickhouse-database", "default", "ClickHouse database name")
	cmd.Flags().StringVar(&opts.ClickHouse.Username, "clickhouse-username", "default", "ClickHouse username")
	cmd.Flags().StringVar(&opts.ClickHouse.Password, "clickhouse-password", "", "ClickHouse password")
	cmd.Flags().DurationVar(&opts.ClickHouse.DialTimeout, "clickhouse-dial-timeout", 10*time.Second, "ClickHouse dial timeout")
	cmd.Flags().IntVar(&opts.ClickHouse.MaxOpenConns, "clickhouse-max-open-conns", 10, "ClickHouse max open connections")
	cmd.Flags().IntVar(&opts.ClickHouse.MaxIdleConns, "clickhouse-max-idle-conns", 5, "ClickHouse max idle connections")
	cmd.Flags().DurationVar(&opts.ClickHouse.ConnMaxLifetime, "clickhouse-conn-max-lifetime", time.Hour, "ClickHouse connection max lifetime")
	cmd.Flags().BoolVar(&opts.ClickHouse.TLS, "clickhouse-tls", false, "Enable TLS for ClickHouse connection")
}

// InitCommonOptions initializes default values that require runtime logic
func InitCommonOptions(opts *CommonOptions) {
	// Set default identity to hostname if not specified
	if opts.Temporal.Identity == "" {
		hostname, err := os.Hostname()
		if err != nil {
			opts.Temporal.Identity = "lifesupport"
		} else {
			opts.Temporal.Identity = hostname
		}
	}
	
	log.Debug().Str("db", opts.DB).Msg("Database config")
	log.Debug().Strs("addrs", opts.ClickHouse.Addrs).Str("database", opts.ClickHouse.Database).Str("username", opts.ClickHouse.Username).Bool("tls", opts.ClickHouse.TLS).Msg("ClickHouse config")
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

// InitClickHouse creates a ClickHouse client with the given options
func InitClickHouse(ctx context.Context, opts ClickHouseOptions) (driver.Conn, error) {
	connOptions := &clickhouse.Options{
		Addr: opts.Addrs,
		Auth: clickhouse.Auth{
			Database: opts.Database,
			Username: opts.Username,
			Password: opts.Password,
		},
		DialTimeout:     opts.DialTimeout,
		MaxOpenConns:    opts.MaxOpenConns,
		MaxIdleConns:    opts.MaxIdleConns,
		ConnMaxLifetime: opts.ConnMaxLifetime,
	}

	if opts.TLS {
		connOptions.TLS = &tls.Config{
			InsecureSkipVerify: false,
		}
	}

	conn, err := clickhouse.Open(connOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to open ClickHouse connection: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	log.Info().
		Strs("addrs", opts.Addrs).
		Str("database", opts.Database).
		Msg("Connected to ClickHouse")

	return conn, nil
}
