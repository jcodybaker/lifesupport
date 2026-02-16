package workflows

import (
	"context"
	"lifesupport/backend/pkg/api"
	"time"

	"github.com/rs/zerolog"
	"go.temporal.io/sdk/activity"
	temporalWorker "go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func (w *WorkflowCtx) registerDiscoveryWorkflow(worker temporalWorker.Worker) {
	worker.RegisterWorkflow(w.DeviceDiscoveryWorkflow)
	worker.RegisterActivity(w.ShellyDiscovery)
}

type DiscoveryWorkflowResult struct {
	// Add any fields needed for the discovery workflow result
}

func (w *WorkflowCtx) DeviceDiscoveryWorkflow(ctx workflow.Context, params api.DiscoveryOptions) (*DiscoveryWorkflowResult, error) {
	// Get workflow logger - this is the deterministic way to log in workflows
	logger := workflow.GetLogger(ctx)
	info := workflow.GetInfo(ctx)

	logger.Info("Starting device discovery workflow",
		"WorkflowType", info.WorkflowType.Name,
		"WorkflowID", info.WorkflowExecution.ID,
		"RunID", info.WorkflowExecution.RunID,
		"TaskQueue", info.TaskQueueName,
	)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 30,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result *api.DiscoveryResult
	err := workflow.ExecuteActivity(ctx, w.ShellyDiscovery, params).Get(ctx, &result)
	if err != nil {
		logger.Error("Device discovery activity failed", "error", err)
		return nil, err
	}

	logger.Info("Device discovery workflow completed", "tagsFound", len(result.DiscoveredTags))
	return &DiscoveryWorkflowResult{}, nil
}

func (w *WorkflowCtx) ShellyDiscovery(ctx context.Context, params api.DiscoveryOptions) (*api.DiscoveryResult, error) {
	// Extract activity info and create structured logger
	info := activity.GetInfo(ctx)

	// Create a logger with temporal execution context
	logger := zerolog.Ctx(ctx)
	if logger.GetLevel() == zerolog.Disabled {
		// Fallback if no logger in context
		logger = &w.logger
	}

	// Enrich logger with temporal execution info
	activityLogger := logger.With().
		Str("WorkflowID", info.WorkflowExecution.ID).
		Str("RunID", info.WorkflowExecution.RunID).
		Str("ActivityID", info.ActivityID).
		Str("ActivityType", info.ActivityType.Name).
		Str("TaskQueue", info.TaskQueue).
		Int32("Attempt", info.Attempt).
		Logger()

	activityLogger.Info().Msg("Starting Shelly device discovery")

	result, err := w.shellyDriver.DiscoverDevices(activityLogger.WithContext(ctx), params, w.storer)
	if err != nil {
		activityLogger.Error().Err(err).Msg("Shelly device discovery failed")
		return nil, err
	}

	activityLogger.Info().
		Int("tagsFound", len(result.DiscoveredTags)).
		Msg("Shelly device discovery completed")

	return result, nil
}
