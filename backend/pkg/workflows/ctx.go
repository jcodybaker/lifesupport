package workflows

import (
	"lifesupport/backend/pkg/drivers/shelly"
	"lifesupport/backend/pkg/storer"

	temporalWorker "go.temporal.io/sdk/worker"
)

type WorkflowCtx struct {
	storer *storer.Storer

	// drivers
	shellyDriver *shelly.Driver
}

func New(storer *storer.Storer, shellyDriver *shelly.Driver) *WorkflowCtx {
	return &WorkflowCtx{
		storer:       storer,
		shellyDriver: shellyDriver,
	}
}

func (w *WorkflowCtx) Register(worker temporalWorker.Worker) {
	w.registerDiscoveryWorkflow(worker)
}
