<script>
  import { onMount } from 'svelte';
  import { workflowAPI } from './api.js';

  let workflows = [];
  let selectedWorkflow = null;
  let loading = false;
  let error = null;
  let starting = false;
  let autoRefresh = false;
  let refreshInterval = null;

  onMount(async () => {
    await loadWorkflows();
  });

  async function loadWorkflows() {
    loading = true;
    error = null;
    try {
      workflows = await workflowAPI.list();
      // Sort by start time, most recent first
      workflows.sort((a, b) => new Date(b.start_time) - new Date(a.start_time));
      
      // If we have a selected workflow, refresh its details
      if (selectedWorkflow) {
        const updated = workflows.find(w => w.workflow_id === selectedWorkflow.workflow_id);
        if (updated) {
          await selectWorkflow(updated);
        }
      }
    } catch (err) {
      error = `Failed to load workflows: ${err.message}`;
      console.error('Error loading workflows:', err);
      workflows = [];
    } finally {
      loading = false;
    }
  }

  async function startDiscoveryWorkflow() {
    starting = true;
    error = null;
    try {
      const response = await workflowAPI.startDiscovery({});
      await loadWorkflows();
      // Auto-select the newly started workflow
      const newWorkflow = workflows.find(w => w.workflow_id === response.workflow_id);
      if (newWorkflow) {
        await selectWorkflow(newWorkflow);
        // Enable auto-refresh for in-progress workflows
        enableAutoRefresh();
      }
    } catch (err) {
      error = `Failed to start discovery workflow: ${err.message}`;
      console.error('Error starting workflow:', err);
    } finally {
      starting = false;
    }
  }

  async function selectWorkflow(workflow) {
    try {
      // Fetch detailed status
      const details = await workflowAPI.getStatus(workflow.workflow_id);
      selectedWorkflow = details;
      
      // Enable auto-refresh if workflow is in-progress
      if (details.status === 'in-progress' || details.status === 'pending') {
        enableAutoRefresh();
      }
    } catch (err) {
      console.error('Error fetching workflow details:', err);
      selectedWorkflow = workflow;
    }
  }

  function enableAutoRefresh() {
    if (!autoRefresh) {
      autoRefresh = true;
      refreshInterval = setInterval(async () => {
        await loadWorkflows();
        // Stop auto-refresh if no in-progress workflows
        const hasInProgress = workflows.some(w => 
          w.status === 'in-progress' || w.status === 'pending'
        );
        if (!hasInProgress) {
          disableAutoRefresh();
        }
      }, 3000); // Refresh every 3 seconds
    }
  }

  function disableAutoRefresh() {
    autoRefresh = false;
    if (refreshInterval) {
      clearInterval(refreshInterval);
      refreshInterval = null;
    }
  }

  function toggleAutoRefresh() {
    if (autoRefresh) {
      disableAutoRefresh();
    } else {
      enableAutoRefresh();
    }
  }

  function formatTimestamp(timestamp) {
    if (!timestamp) return 'N/A';
    return new Date(timestamp).toLocaleString();
  }

  function getStatusColor(status) {
    switch (status) {
      case 'success':
        return '#4CAF50';
      case 'error':
        return '#f44336';
      case 'in-progress':
        return '#2196F3';
      case 'pending':
        return '#FF9800';
      default:
        return '#757575';
    }
  }

  function getStatusIcon(status) {
    switch (status) {
      case 'success':
        return '✅';
      case 'error':
        return '❌';
      case 'in-progress':
        return '🔄';
      case 'pending':
        return '⏳';
      default:
        return '❓';
    }
  }

  function getDuration(startTime, closeTime) {
    if (!startTime) return 'N/A';
    const start = new Date(startTime);
    const end = closeTime ? new Date(closeTime) : new Date();
    const durationMs = end - start;
    
    if (durationMs < 1000) {
      return `${durationMs}ms`;
    } else if (durationMs < 60000) {
      return `${(durationMs / 1000).toFixed(1)}s`;
    } else {
      return `${Math.floor(durationMs / 60000)}m ${Math.floor((durationMs % 60000) / 1000)}s`;
    }
  }

  // Cleanup on component destroy
  import { onDestroy } from 'svelte';
  onDestroy(() => {
    disableAutoRefresh();
  });
</script>

<div class="workflow-manager">
  <div class="header">
    <h2>Discovery Workflows</h2>
    <div class="header-actions">
      <label class="auto-refresh-toggle">
        <input type="checkbox" bind:checked={autoRefresh} on:change={toggleAutoRefresh} />
        Auto-refresh
      </label>
      <button on:click={loadWorkflows} class="btn btn-secondary" disabled={loading}>
        {loading ? '🔄 Refreshing...' : '🔄 Refresh'}
      </button>
      <button on:click={startDiscoveryWorkflow} class="btn btn-primary" disabled={starting}>
        {starting ? '⏳ Starting...' : '🔍 Start Discovery'}
      </button>
    </div>
  </div>

  {#if error}
    <div class="error-message">{error}</div>
  {/if}

  <div class="content">
    <div class="workflow-list">
      <h3>Recent Executions</h3>
      {#if loading && workflows.length === 0}
        <div class="loading">Loading workflows...</div>
      {:else if workflows.length === 0}
        <div class="empty-state">
          No workflows found. Click "Start Discovery" to begin device discovery.
        </div>
      {:else}
        <div class="workflow-items">
          {#each workflows as workflow (workflow.workflow_id)}
            <div 
              class="workflow-item" 
              class:selected={selectedWorkflow?.workflow_id === workflow.workflow_id}
              on:click={() => selectWorkflow(workflow)}
            >
              <div class="workflow-header">
                <span class="workflow-icon">{getStatusIcon(workflow.status)}</span>
                <div class="workflow-info">
                  <div class="workflow-id">{workflow.workflow_id}</div>
                  <div class="workflow-time">{formatTimestamp(workflow.start_time)}</div>
                </div>
              </div>
              <div class="workflow-status" style="color: {getStatusColor(workflow.status)}">
                {workflow.status.toUpperCase()}
              </div>
              <div class="workflow-duration">
                Duration: {getDuration(workflow.start_time, workflow.close_time)}
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>

    <div class="workflow-details">
      {#if selectedWorkflow}
        <h3>Workflow Details</h3>
        <div class="details-content">
          <div class="detail-card">
            <h4>Status</h4>
            <div class="status-display" style="border-color: {getStatusColor(selectedWorkflow.status)}">
              <span class="status-icon">{getStatusIcon(selectedWorkflow.status)}</span>
              <span class="status-text" style="color: {getStatusColor(selectedWorkflow.status)}">
                {selectedWorkflow.status.toUpperCase()}
              </span>
            </div>
          </div>

          <div class="detail-card">
            <h4>Workflow Information</h4>
            <dl>
              <dt>Workflow ID:</dt>
              <dd class="monospace">{selectedWorkflow.workflow_id}</dd>
              <dt>Run ID:</dt>
              <dd class="monospace">{selectedWorkflow.run_id}</dd>
              <dt>Start Time:</dt>
              <dd>{formatTimestamp(selectedWorkflow.start_time)}</dd>
              {#if selectedWorkflow.close_time}
                <dt>Close Time:</dt>
                <dd>{formatTimestamp(selectedWorkflow.close_time)}</dd>
              {/if}
              <dt>Duration:</dt>
              <dd>{getDuration(selectedWorkflow.start_time, selectedWorkflow.close_time)}</dd>
            </dl>
          </div>

          {#if selectedWorkflow.error}
            <div class="detail-card error-card">
              <h4>Error</h4>
              <div class="error-content">
                {selectedWorkflow.error}
              </div>
            </div>
          {/if}

          {#if selectedWorkflow.result}
            <div class="detail-card success-card">
              <h4>Discovery Results</h4>
              {#if selectedWorkflow.result.discovered_tags && selectedWorkflow.result.discovered_tags.length > 0}
                <div class="result-summary">
                  Found {selectedWorkflow.result.discovered_tags.length} device{selectedWorkflow.result.discovered_tags.length !== 1 ? 's' : ''}
                </div>
                <div class="discovered-tags">
                  {#each selectedWorkflow.result.discovered_tags as tag}
                    <div class="tag-item">{tag}</div>
                  {/each}
                </div>
              {:else}
                <div class="no-results">No devices discovered</div>
              {/if}
            </div>
          {/if}

          {#if selectedWorkflow.status === 'in-progress' || selectedWorkflow.status === 'pending'}
            <div class="detail-card info-card">
              <h4>⏳ Workflow in Progress</h4>
              <p>The discovery workflow is currently running. Results will appear here when complete.</p>
              {#if !autoRefresh}
                <button on:click={enableAutoRefresh} class="btn btn-secondary">
                  Enable Auto-Refresh
                </button>
              {/if}
            </div>
          {/if}
        </div>
      {:else}
        <div class="empty-state">
          Select a workflow execution to view details
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .workflow-manager {
    padding: 1rem;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
    flex-wrap: wrap;
    gap: 1rem;
  }

  .header-actions {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .auto-refresh-toggle {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.9rem;
    color: #666;
    cursor: pointer;
    user-select: none;
  }

  .auto-refresh-toggle input {
    cursor: pointer;
  }

  h2 {
    margin: 0;
    color: #333;
  }

  h3 {
    margin: 0 0 1rem 0;
    color: #555;
    font-size: 1.2rem;
  }

  h4 {
    margin: 0 0 0.75rem 0;
    color: #666;
    font-size: 1rem;
    border-bottom: 2px solid #e0e0e0;
    padding-bottom: 0.5rem;
  }

  .btn {
    padding: 0.5rem 1rem;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.9rem;
    transition: all 0.2s;
    white-space: nowrap;
  }

  .btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .btn-primary {
    background-color: #4CAF50;
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background-color: #45a049;
  }

  .btn-secondary {
    background-color: #757575;
    color: white;
  }

  .btn-secondary:hover:not(:disabled) {
    background-color: #616161;
  }

  .error-message {
    background-color: #ffebee;
    color: #c62828;
    padding: 0.75rem;
    border-radius: 4px;
    margin-bottom: 1rem;
    border-left: 4px solid #c62828;
  }

  .content {
    display: grid;
    grid-template-columns: 1fr 2fr;
    gap: 1.5rem;
  }

  .workflow-list {
    background-color: white;
    border-radius: 8px;
    padding: 1rem;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  }

  .workflow-items {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .workflow-item {
    padding: 1rem;
    background-color: #f9f9f9;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.2s;
    border: 2px solid transparent;
  }

  .workflow-item:hover {
    background-color: #f0f0f0;
  }

  .workflow-item.selected {
    background-color: #e3f2fd;
    border-color: #2196F3;
  }

  .workflow-header {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    margin-bottom: 0.5rem;
  }

  .workflow-icon {
    font-size: 1.5rem;
  }

  .workflow-info {
    flex: 1;
  }

  .workflow-id {
    font-family: monospace;
    font-size: 0.85rem;
    color: #666;
    word-break: break-all;
  }

  .workflow-time {
    font-size: 0.8rem;
    color: #999;
    margin-top: 0.25rem;
  }

  .workflow-status {
    font-weight: 600;
    font-size: 0.9rem;
    margin-bottom: 0.25rem;
  }

  .workflow-duration {
    font-size: 0.85rem;
    color: #777;
  }

  .workflow-details {
    background-color: white;
    border-radius: 8px;
    padding: 1rem;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  }

  .details-content {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .detail-card {
    background-color: #fafafa;
    padding: 1rem;
    border-radius: 6px;
    border-left: 4px solid #e0e0e0;
  }

  .error-card {
    background-color: #ffebee;
    border-left-color: #f44336;
  }

  .success-card {
    background-color: #e8f5e9;
    border-left-color: #4CAF50;
  }

  .info-card {
    background-color: #e3f2fd;
    border-left-color: #2196F3;
  }

  .status-display {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem;
    background-color: white;
    border-radius: 6px;
    border-left: 4px solid;
  }

  .status-icon {
    font-size: 2rem;
  }

  .status-text {
    font-size: 1.5rem;
    font-weight: 600;
  }

  dl {
    margin: 0;
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0.5rem 1rem;
  }

  dt {
    font-weight: 600;
    color: #666;
  }

  dd {
    margin: 0;
    color: #333;
  }

  .monospace {
    font-family: monospace;
    font-size: 0.9rem;
    word-break: break-all;
  }

  .error-content {
    background-color: white;
    padding: 1rem;
    border-radius: 4px;
    font-family: monospace;
    font-size: 0.85rem;
    color: #c62828;
    white-space: pre-wrap;
    word-break: break-word;
  }

  .result-summary {
    background-color: white;
    padding: 0.75rem;
    border-radius: 4px;
    font-weight: 600;
    color: #2e7d32;
    margin-bottom: 1rem;
  }

  .discovered-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .tag-item {
    background-color: white;
    color: #2e7d32;
    padding: 0.5rem 0.75rem;
    border-radius: 4px;
    font-family: monospace;
    font-size: 0.85rem;
    border: 1px solid #a5d6a7;
  }

  .no-results {
    color: #999;
    font-style: italic;
    text-align: center;
    padding: 1rem;
  }

  .loading,
  .empty-state {
    text-align: center;
    color: #999;
    padding: 2rem;
  }

  .info-card p {
    margin: 0 0 1rem 0;
    color: #666;
  }
</style>
