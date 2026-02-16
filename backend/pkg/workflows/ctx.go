package workflows

import (
	"lifesupport/backend/pkg/drivers/shelly"
	"lifesupport/backend/pkg/storer"

	"github.com/rs/zerolog"
	temporalWorker "go.temporal.io/sdk/worker"
)

type WorkflowCtx struct {
	logger zerolog.Logger
	storer *storer.Storer

	// drivers
	shellyDriver *shelly.Driver
}

func New(logger zerolog.Logger, storer *storer.Storer, shellyDriver *shelly.Driver) *WorkflowCtx {
	return &WorkflowCtx{
		logger:       logger,
		storer:       storer,
		shellyDriver: shellyDriver,
	}
}

func (w *WorkflowCtx) Register(worker temporalWorker.Worker) {
	w.registerDiscoveryWorkflow(worker)
}
