package cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"lifesupport/backend/pkg/drivers/shelly"
	"lifesupport/backend/pkg/storer"
	"lifesupport/backend/pkg/temporallog"
	"lifesupport/backend/pkg/workflows"

	temporalWorker "go.temporal.io/sdk/worker"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.temporal.io/sdk/client"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start the Temporal worker",
	Long:  `Start the Temporal worker that processes workflows and activities.`,
	Run:   runWorker,
}

var (
	dbConnString    string
	temporalOptions TemporalOptions
	mqttOptions     MQTTOptions
	workerOptions   WorkerOptions
)

type TemporalOptions struct {
	Host              string
	Namespace         string
	TaskQueue         string
	Identity          string
	ConnectionTimeout time.Duration
}

type MQTTOptions struct {
	Broker                string
	ClientID              string
	Username              string
	Password              string
	KeepAlive             time.Duration
	CleanSession          bool
	AutoReconnect         bool
	ConnectTimeout        time.Duration
	TLSCACert             string
	TLSClientCert         string
	TLSClientKey          string
	TLSInsecureSkipVerify bool
}

type WorkerOptions struct {
	MaxConcurrentActivityExecutionSize     int
	MaxConcurrentWorkflowTaskExecutionSize int
}

func init() {
	rootCmd.AddCommand(workerCmd)

	// Configure Viper for automatic environment variable binding
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Database flags
	workerCmd.Flags().StringVar(&dbConnString, "db", "postgres://lifesupport:lifesupport@localhost:5432/lifesupport?sslmode=disable", "Database connection string")
	viper.BindPFlag("db", workerCmd.Flags().Lookup("db"))

	// Temporal flags
	workerCmd.Flags().StringVar(&temporalOptions.Host, "temporal-host", "localhost:7233", "Temporal server host:port")
	workerCmd.Flags().StringVar(&temporalOptions.Namespace, "temporal-namespace", "default", "Temporal namespace")
	workerCmd.Flags().StringVar(&temporalOptions.TaskQueue, "task-queue", "lifesupport-tasks", "Task queue name")
	workerCmd.Flags().StringVar(&temporalOptions.Identity, "temporal-identity", "", "Temporal worker identity (defaults to hostname)")
	workerCmd.Flags().DurationVar(&temporalOptions.ConnectionTimeout, "temporal-timeout", 10*time.Second, "Temporal connection timeout")
	viper.BindPFlag("temporal-host", workerCmd.Flags().Lookup("temporal-host"))
	viper.BindPFlag("temporal-namespace", workerCmd.Flags().Lookup("temporal-namespace"))
	viper.BindPFlag("task-queue", workerCmd.Flags().Lookup("task-queue"))
	viper.BindPFlag("temporal-identity", workerCmd.Flags().Lookup("temporal-identity"))
	viper.BindPFlag("temporal-timeout", workerCmd.Flags().Lookup("temporal-timeout"))

	// MQTT flags
	workerCmd.Flags().StringVar(&mqttOptions.Broker, "mqtt-broker", "tcp://localhost:1883", "MQTT broker URL")
	workerCmd.Flags().StringVar(&mqttOptions.ClientID, "mqtt-client-id", "lifesupport-worker", "MQTT client ID")
	workerCmd.Flags().StringVar(&mqttOptions.Username, "mqtt-username", "", "MQTT username")
	workerCmd.Flags().StringVar(&mqttOptions.Password, "mqtt-password", "", "MQTT password")
	workerCmd.Flags().DurationVar(&mqttOptions.KeepAlive, "mqtt-keepalive", 60*time.Second, "MQTT keep alive interval")
	workerCmd.Flags().BoolVar(&mqttOptions.CleanSession, "mqtt-clean-session", true, "MQTT clean session")
	workerCmd.Flags().BoolVar(&mqttOptions.AutoReconnect, "mqtt-auto-reconnect", true, "MQTT auto reconnect")
	workerCmd.Flags().DurationVar(&mqttOptions.ConnectTimeout, "mqtt-connect-timeout", 30*time.Second, "MQTT connection timeout")
	workerCmd.Flags().StringVar(&mqttOptions.TLSCACert, "mqtt-tls-ca-cert", "", "MQTT TLS CA certificate file path")
	workerCmd.Flags().StringVar(&mqttOptions.TLSClientCert, "mqtt-tls-client-cert", "", "MQTT TLS client certificate file path")
	workerCmd.Flags().StringVar(&mqttOptions.TLSClientKey, "mqtt-tls-client-key", "", "MQTT TLS client key file path")
	workerCmd.Flags().BoolVar(&mqttOptions.TLSInsecureSkipVerify, "mqtt-tls-insecure-skip-verify", false, "MQTT TLS skip certificate verification")
	viper.BindPFlag("mqtt-broker", workerCmd.Flags().Lookup("mqtt-broker"))
	viper.BindPFlag("mqtt-client-id", workerCmd.Flags().Lookup("mqtt-client-id"))
	viper.BindPFlag("mqtt-username", workerCmd.Flags().Lookup("mqtt-username"))
	viper.BindPFlag("mqtt-password", workerCmd.Flags().Lookup("mqtt-password"))
	viper.BindPFlag("mqtt-keepalive", workerCmd.Flags().Lookup("mqtt-keepalive"))
	viper.BindPFlag("mqtt-clean-session", workerCmd.Flags().Lookup("mqtt-clean-session"))
	viper.BindPFlag("mqtt-auto-reconnect", workerCmd.Flags().Lookup("mqtt-auto-reconnect"))
	viper.BindPFlag("mqtt-connect-timeout", workerCmd.Flags().Lookup("mqtt-connect-timeout"))
	viper.BindPFlag("mqtt-tls-ca-cert", workerCmd.Flags().Lookup("mqtt-tls-ca-cert"))
	viper.BindPFlag("mqtt-tls-client-cert", workerCmd.Flags().Lookup("mqtt-tls-client-cert"))
	viper.BindPFlag("mqtt-tls-client-key", workerCmd.Flags().Lookup("mqtt-tls-client-key"))
	viper.BindPFlag("mqtt-tls-insecure-skip-verify", workerCmd.Flags().Lookup("mqtt-tls-insecure-skip-verify"))

	// Worker flags
	workerCmd.Flags().IntVar(&workerOptions.MaxConcurrentActivityExecutionSize, "max-concurrent-activities", 10, "Maximum concurrent activity executions")
	workerCmd.Flags().IntVar(&workerOptions.MaxConcurrentWorkflowTaskExecutionSize, "max-concurrent-workflows", 10, "Maximum concurrent workflow task executions")
	viper.BindPFlag("max-concurrent-activities", workerCmd.Flags().Lookup("max-concurrent-activities"))
	viper.BindPFlag("max-concurrent-workflows", workerCmd.Flags().Lookup("max-concurrent-workflows"))
}

func createTLSConfig(opts MQTTOptions) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: opts.TLSInsecureSkipVerify,
	}

	// Load CA certificate if provided
	if opts.TLSCACert != "" {
		caCert, err := os.ReadFile(opts.TLSCACert)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, err
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Load client certificate and key if provided
	if opts.TLSClientCert != "" && opts.TLSClientKey != "" {
		cert, err := tls.LoadX509KeyPair(opts.TLSClientCert, opts.TLSClientKey)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

func runWorker(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	ctx = log.Logger.WithContext(ctx)
	var wg sync.WaitGroup

	// Get values from Viper (which handles env vars automatically)
	dbConnString = viper.GetString("db")
	temporalOptions.Host = viper.GetString("temporal-host")
	temporalOptions.Namespace = viper.GetString("temporal-namespace")
	temporalOptions.TaskQueue = viper.GetString("task-queue")
	temporalOptions.Identity = viper.GetString("temporal-identity")
	temporalOptions.ConnectionTimeout = viper.GetDuration("temporal-timeout")
	mqttOptions.Broker = viper.GetString("mqtt-broker")
	mqttOptions.ClientID = viper.GetString("mqtt-client-id")
	mqttOptions.Username = viper.GetString("mqtt-username")
	mqttOptions.Password = viper.GetString("mqtt-password")
	mqttOptions.KeepAlive = viper.GetDuration("mqtt-keepalive")
	mqttOptions.CleanSession = viper.GetBool("mqtt-clean-session")
	mqttOptions.AutoReconnect = viper.GetBool("mqtt-auto-reconnect")
	mqttOptions.ConnectTimeout = viper.GetDuration("mqtt-connect-timeout")
	mqttOptions.TLSCACert = viper.GetString("mqtt-tls-ca-cert")
	mqttOptions.TLSClientCert = viper.GetString("mqtt-tls-client-cert")
	mqttOptions.TLSClientKey = viper.GetString("mqtt-tls-client-key")
	mqttOptions.TLSInsecureSkipVerify = viper.GetBool("mqtt-tls-insecure-skip-verify")
	workerOptions.MaxConcurrentActivityExecutionSize = viper.GetInt("max-concurrent-activities")
	workerOptions.MaxConcurrentWorkflowTaskExecutionSize = viper.GetInt("max-concurrent-workflows")

	// Set default identity to hostname if not specified
	if temporalOptions.Identity == "" {
		hostname, err := os.Hostname()
		if err != nil {
			temporalOptions.Identity = "lifesupport-worker"
		} else {
			temporalOptions.Identity = hostname
		}
	}

	// Create storer
	store, err := storer.New(dbConnString)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to create storer")
	}
	defer store.Close()
	if err := store.InitSchema(ctx); err != nil {
		log.Fatal().Err(err).Msg("Unable to initialize database schema")
	}

	// Create Temporal client
	c, err := client.DialContext(ctx, client.Options{
		HostPort:  temporalOptions.Host,
		Namespace: temporalOptions.Namespace,
		Identity:  temporalOptions.Identity,
		Logger:    temporallog.NewTemporalLogger(log.Logger),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to create Temporal client")
	}
	defer c.Close()

	// Configure MQTT client
	mqttClientOptions := mqtt.NewClientOptions().
		AddBroker(mqttOptions.Broker).
		SetClientID(mqttOptions.ClientID).
		SetKeepAlive(mqttOptions.KeepAlive).
		SetCleanSession(mqttOptions.CleanSession).
		SetAutoReconnect(mqttOptions.AutoReconnect).
		SetConnectTimeout(mqttOptions.ConnectTimeout)

	if mqttOptions.Username != "" {
		mqttClientOptions.SetUsername(mqttOptions.Username)
	}
	if mqttOptions.Password != "" {
		mqttClientOptions.SetPassword(mqttOptions.Password)
	}

	// Configure TLS if certificates are provided
	if mqttOptions.TLSCACert != "" || mqttOptions.TLSClientCert != "" || mqttOptions.TLSInsecureSkipVerify {
		tlsConfig, err := createTLSConfig(mqttOptions)
		if err != nil {
			log.Fatal().Err(err).Msg("Unable to create TLS config for MQTT")
		}
		mqttClientOptions.SetTLSConfig(tlsConfig)
	}

	mqttClient := mqtt.NewClient(mqttClientOptions)
	token := mqttClient.Connect()
	token.WaitTimeout(mqttOptions.ConnectTimeout)
	if err := token.Error(); err != nil {
		log.Fatal().Err(err).Msg("Unable to connect to MQTT broker")
	}
	shellyDriver := shelly.New(mqttClient)
	if err := shellyDriver.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("Unable to start Shelly driver")
	}

	workflowCtx := workflows.New(log.Logger, store, shellyDriver)

	// Create worker
	w := temporalWorker.New(c, temporalOptions.TaskQueue, temporalWorker.Options{
		MaxConcurrentActivityExecutionSize:     workerOptions.MaxConcurrentActivityExecutionSize,
		MaxConcurrentWorkflowTaskExecutionSize: workerOptions.MaxConcurrentWorkflowTaskExecutionSize,
		Identity:                               temporalOptions.Identity,
	})

	workflowCtx.Register(w)

	log.Info().
		Str("task_queue", temporalOptions.TaskQueue).
		Str("namespace", temporalOptions.Namespace).
		Str("identity", temporalOptions.Identity).
		Msg("Starting Temporal worker")
	log.Info().
		Str("temporal_host", temporalOptions.Host).
		Str("mqtt_broker", mqttOptions.Broker).
		Str("mqtt_client_id", mqttOptions.ClientID).
		Msg("Connected to services")

	// // Start worker in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := w.Run(temporalWorker.InterruptCh())
		if err != nil {
			log.Fatal().Err(err).Msg("Unable to start worker")
		}
		log.Info().Msg("Temporal worker stopped")
	}()

	// Wait for interrupt signal to gracefully shutdown the worker
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Info().Msg("Shutting down MQTT client...")

	log.Info().Msg("Shutting down Temporal worker...")
	w.Stop()
	if err := shellyDriver.Stop(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Error stopping Shelly driver")
	}
	mqttClient.Disconnect(250)
	wg.Wait()

}
