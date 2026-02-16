package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"

	"lifesupport/backend/pkg/api"
)

const (
	discoveryWorkflowName = "DeviceDiscoveryWorkflow"
	defaultTaskQueue      = "lifesupport-tasks"
)

// StartDiscoveryWorkflow handles POST /api/workflows/discovery
func (h *Handler) StartDiscoveryWorkflow(w http.ResponseWriter, r *http.Request) {
	if h.TemporalClient == nil {
		http.Error(w, "Temporal client not configured", http.StatusServiceUnavailable)
		return
	}

	var request api.StartWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Use default options if not provided
	options := api.DiscoveryOptions{}
	if request.Options != nil {
		options = *request.Options
	}

	// Generate a unique workflow ID
	workflowID := "discovery-" + uuid.New().String()

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: defaultTaskQueue,
	}

	ctx := r.Context()
	we, err := h.TemporalClient.ExecuteWorkflow(ctx, workflowOptions, discoveryWorkflowName, options)
	if err != nil {
		http.Error(w, "Failed to start workflow: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := api.StartWorkflowResponse{
		WorkflowID: we.GetID(),
		RunID:      we.GetRunID(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetWorkflowStatus handles GET /api/workflows/{workflowId}
func (h *Handler) GetWorkflowStatus(w http.ResponseWriter, r *http.Request) {
	if h.TemporalClient == nil {
		http.Error(w, "Temporal client not configured", http.StatusServiceUnavailable)
		return
	}

	params := mux.Vars(r)
	workflowID := params["workflowId"]

	ctx := r.Context()

	// Get workflow description to determine status
	desc, err := h.TemporalClient.DescribeWorkflowExecution(ctx, workflowID, "")
	if err != nil {
		http.Error(w, "Workflow not found: "+err.Error(), http.StatusNotFound)
		return
	}

	workflowInfo := api.WorkflowInfo{
		WorkflowID: desc.WorkflowExecutionInfo.Execution.WorkflowId,
		RunID:      desc.WorkflowExecutionInfo.Execution.RunId,
		StartTime:  desc.WorkflowExecutionInfo.StartTime.AsTime(),
	}

	// Determine status based on workflow state
	if desc.WorkflowExecutionInfo.CloseTime != nil {
		closeTime := desc.WorkflowExecutionInfo.CloseTime.AsTime()
		workflowInfo.CloseTime = &closeTime

		switch desc.WorkflowExecutionInfo.Status {
		case 1: // COMPLETED
			workflowInfo.Status = api.WorkflowStatusSuccess
		case 2, 3, 4, 5: // FAILED, CANCELED, TERMINATED, CONTINUED_AS_NEW
			workflowInfo.Status = api.WorkflowStatusError
			if desc.WorkflowExecutionInfo.Status == 2 {
				// Get failure message if available
				workflowInfo.Error = "Workflow failed"
			} else if desc.WorkflowExecutionInfo.Status == 3 {
				workflowInfo.Error = "Workflow canceled"
			} else if desc.WorkflowExecutionInfo.Status == 4 {
				workflowInfo.Error = "Workflow terminated"
			}
		case 6: // TIMED_OUT
			workflowInfo.Status = api.WorkflowStatusError
			workflowInfo.Error = "Workflow timed out"
		default:
			workflowInfo.Status = api.WorkflowStatusError
			workflowInfo.Error = "Unknown workflow status"
		}
	} else {
		// Workflow is still running
		workflowInfo.Status = api.WorkflowStatusInProgress
	}

	// If workflow is for discovery, try to get the result
	if workflowID[:9] == "discovery" && workflowInfo.Status == api.WorkflowStatusSuccess {
		discoveryInfo := api.DiscoveryWorkflowInfo{
			WorkflowInfo: workflowInfo,
		}

		// Try to get workflow result if completed
		var result api.DiscoveryResult
		err := h.TemporalClient.GetWorkflow(ctx, workflowID, "").Get(ctx, &result)
		if err == nil {
			discoveryInfo.Result = &result
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(discoveryInfo)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workflowInfo)
}

// ListWorkflows handles GET /api/workflows
func (h *Handler) ListWorkflows(w http.ResponseWriter, r *http.Request) {
	if h.TemporalClient == nil {
		http.Error(w, "Temporal client not configured", http.StatusServiceUnavailable)
		return
	}

	ctx := r.Context()

	// Query for discovery workflows - list recent ones
	query := "WorkflowType = 'DeviceDiscoveryWorkflow'"

	// Create a timeout context for the list operation
	listCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// ListWorkflow returns response and error
	resp, err := h.TemporalClient.ListWorkflow(listCtx, &workflowservice.ListWorkflowExecutionsRequest{
		Query:    query,
		PageSize: 100,
	})
	if err != nil {
		http.Error(w, "Failed to list workflows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	workflows := make([]api.WorkflowInfo, 0)
	if resp != nil && resp.Executions != nil {
		for _, exec := range resp.Executions {
			workflowInfo := api.WorkflowInfo{
				WorkflowID: exec.Execution.WorkflowId,
				RunID:      exec.Execution.RunId,
				StartTime:  exec.StartTime.AsTime(),
			}

			if exec.CloseTime != nil {
				closeTime := exec.CloseTime.AsTime()
				workflowInfo.CloseTime = &closeTime
			}

			// Determine status
			if exec.CloseTime != nil {
				switch exec.Status {
				case 1: // COMPLETED
					workflowInfo.Status = api.WorkflowStatusSuccess
				case 2: // FAILED
					workflowInfo.Status = api.WorkflowStatusError
					workflowInfo.Error = "Workflow failed"
				case 3: // CANCELED
					workflowInfo.Status = api.WorkflowStatusError
					workflowInfo.Error = "Workflow canceled"
				case 4: // TERMINATED
					workflowInfo.Status = api.WorkflowStatusError
					workflowInfo.Error = "Workflow terminated"
				case 6: // TIMED_OUT
					workflowInfo.Status = api.WorkflowStatusError
					workflowInfo.Error = "Workflow timed out"
				default:
					workflowInfo.Status = api.WorkflowStatusError
				}
			} else {
				workflowInfo.Status = api.WorkflowStatusInProgress
			}

			workflows = append(workflows, workflowInfo)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workflows)
}
