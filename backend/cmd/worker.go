package cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"lifesupport/backend/pkg/drivers/shelly"
	"lifesupport/backend/pkg/workflows"

	temporalWorker "go.temporal.io/sdk/worker"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start the Temporal worker",
	Long:  `Start the Temporal worker that processes workflows and activities.`,
	Run:   runWorker,
}

var (
	workerOptions WorkerOptions
	commonOptions CommonOptions
	mqttOptions   MQTTOptions
)

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

	// Add common database and temporal flags
	AddCommonFlags(workerCmd, &commonOptions)

	// Worker-specific temporal flags
	workerCmd.Flags().StringVar(&commonOptions.Temporal.TaskQueue, "task-queue", "lifesupport-tasks", "Task queue name")

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

	// Worker flags
	workerCmd.Flags().IntVar(&workerOptions.MaxConcurrentActivityExecutionSize, "max-concurrent-activities", 10, "Maximum concurrent activity executions")
	workerCmd.Flags().IntVar(&workerOptions.MaxConcurrentWorkflowTaskExecutionSize, "max-concurrent-workflows", 10, "Maximum concurrent workflow task executions")
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

	// Initialize options
	InitCommonOptions(&commonOptions)

	// Initialize database
	store, err := InitDatabase(ctx, commonOptions.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to create storer")
	}
	defer store.Close()
	log.Info().Str("db", commonOptions.DB).Msg("Connected to database")

	// Create Temporal client
	c, err := InitTemporalClient(ctx, commonOptions.Temporal)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to create Temporal client")
	}
	defer c.Close()

	// Initialize ClickHouse client
	clickhouseConn, err := InitClickHouse(ctx, commonOptions.ClickHouse)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to create ClickHouse client")
	}
	defer clickhouseConn.Close()

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
	shellyDriver := shelly.New(mqttClient, clickhouseConn)
	if err := shellyDriver.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("Unable to start Shelly driver")
	}

	workflowCtx := workflows.New(log.Logger, store, shellyDriver)

	// Create worker
	w := temporalWorker.New(c, commonOptions.Temporal.TaskQueue, temporalWorker.Options{
		MaxConcurrentActivityExecutionSize:     workerOptions.MaxConcurrentActivityExecutionSize,
		MaxConcurrentWorkflowTaskExecutionSize: workerOptions.MaxConcurrentWorkflowTaskExecutionSize,
		Identity:                               commonOptions.Temporal.Identity,
	})

	workflowCtx.Register(w)

	log.Info().
		Str("task_queue", commonOptions.Temporal.TaskQueue).
		Str("namespace", commonOptions.Temporal.Namespace).
		Str("identity", commonOptions.Temporal.Identity).
		Msg("Starting Temporal worker")
	log.Info().
		Str("temporal_host", commonOptions.Temporal.Host).
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
