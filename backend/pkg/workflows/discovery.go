package workflows

import (
	"context"
	"lifesupport/backend/pkg/drivers"
	"time"

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

func (w *WorkflowCtx) DeviceDiscoveryWorkflow(ctx workflow.Context, params drivers.DiscoveryOptions) (*DiscoveryWorkflowResult, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 30,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result *drivers.DiscoveryResult
	err := workflow.ExecuteActivity(ctx, w.ShellyDiscovery, params).Get(ctx, &result)
	if err != nil {
		return nil, err
	}

	return &DiscoveryWorkflowResult{}, nil
}

func (w *WorkflowCtx) ShellyDiscovery(ctx context.Context, params drivers.DiscoveryOptions) (*drivers.DiscoveryResult, error) {
	result, err := w.shellyDriver.DiscoverDevices(ctx, params, w.storer)
	if err != nil {
		return nil, err
	}
	return result, nil
}
