package api

import "time"

// WorkflowStatus represents the status of a workflow execution
type WorkflowStatus string

const (
	WorkflowStatusPending    WorkflowStatus = "pending"
	WorkflowStatusInProgress WorkflowStatus = "in-progress"
	WorkflowStatusSuccess    WorkflowStatus = "success"
	WorkflowStatusError      WorkflowStatus = "error"
)

// WorkflowInfo contains information about a workflow execution
type WorkflowInfo struct {
	WorkflowID string         `json:"workflow_id"`
	RunID      string         `json:"run_id"`
	Status     WorkflowStatus `json:"status"`
	StartTime  time.Time      `json:"start_time"`
	CloseTime  *time.Time     `json:"close_time,omitempty"`
	Error      string         `json:"error,omitempty"`
}

// DiscoveryWorkflowInfo contains information about a discovery workflow
type DiscoveryWorkflowInfo struct {
	WorkflowInfo
	Result *DiscoveryResult `json:"result,omitempty"`
}

// StartWorkflowRequest is the request body for starting a workflow
type StartWorkflowRequest struct {
	Options *DiscoveryOptions `json:"options,omitempty"`
}

// StartWorkflowResponse is the response body for starting a workflow
type StartWorkflowResponse struct {
	WorkflowID string `json:"workflow_id"`
	RunID      string `json:"run_id"`
}
