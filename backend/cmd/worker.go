package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"lifesupport/backend/pkg/storer"

	"github.com/spf13/cobra"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start the Temporal worker",
	Long:  `Start the Temporal worker that processes workflows and activities.`,
	Run:   runWorker,
}

var (
	temporalHost  string
	taskQueue     string
	dbConnString  string
	workerOptions WorkerOptions
)

type WorkerOptions struct {
	MaxConcurrentActivityExecutionSize     int
	MaxConcurrentWorkflowTaskExecutionSize int
}

func init() {
	rootCmd.AddCommand(workerCmd)
	workerCmd.Flags().StringVar(&temporalHost, "temporal-host", "localhost:7233", "Temporal server host:port")
	workerCmd.Flags().StringVar(&taskQueue, "task-queue", "lifesupport-tasks", "Task queue name")
	workerCmd.Flags().StringVar(&dbConnString, "db", "postgres://postgres:postgres@localhost:5432/lifesupport?sslmode=disable", "Database connection string")
	workerCmd.Flags().IntVar(&workerOptions.MaxConcurrentActivityExecutionSize, "max-concurrent-activities", 10, "Maximum concurrent activity executions")
	workerCmd.Flags().IntVar(&workerOptions.MaxConcurrentWorkflowTaskExecutionSize, "max-concurrent-workflows", 10, "Maximum concurrent workflow task executions")
}

func runWorker(cmd *cobra.Command, args []string) {

	// Override with environment variable if set
	if envHost := os.Getenv("TEMPORAL_HOST"); envHost != "" {
		temporalHost = envHost
	}
	if envQueue := os.Getenv("TEMPORAL_TASK_QUEUE"); envQueue != "" {
		taskQueue = envQueue
	}
	if envDB := os.Getenv("DATABASE_URL"); envDB != "" {
		dbConnString = envDB
	}

	// Create storer
	store, err := storer.New(dbConnString)
	if err != nil {
		log.Fatalln("Unable to create storer", err)
	}
	defer store.Close()

	// Create Temporal client
	// c, err := client.Dial(client.Options{
	// 	HostPort: temporalHost,
	// })
	// if err != nil {
	// 	log.Fatalln("Unable to create Temporal client", err)
	// }
	// defer c.Close()

	// _ = worker.NewWorker(store)

	// Create worker
	// w := temporalWorker.New(c, taskQueue, temporalWorker.Options{
	// 	MaxConcurrentActivityExecutionSize:     workerOptions.MaxConcurrentActivityExecutionSize,
	// 	MaxConcurrentWorkflowTaskExecutionSize: workerOptions.MaxConcurrentWorkflowTaskExecutionSize,
	// })

	// w.RegisterWorkflowWithOptions(wrk.DeviceDiscoveryWorkflow, workflow.RegisterOptions{Name: "DeviceDiscoveryWorkflow"})

	// log.Printf("Starting Temporal worker on task queue: %s", taskQueue)
	// log.Printf("Connected to Temporal server at: %s", temporalHost)

	// // Start worker in a goroutine
	// go func() {
	// 	err := w.Run(temporalWorker.InterruptCh())
	// 	if err != nil {
	// 		log.Fatalln("Unable to start worker", err)
	// 	}
	// }()

	// Wait for interrupt signal to gracefully shutdown the worker
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Temporal worker...")
	// w.Stop()
	log.Println("Temporal worker stopped")
}
