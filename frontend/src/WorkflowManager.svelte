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
    padding: 0;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.75rem;
    flex-wrap: wrap;
    gap: 0.5rem;
    padding: 0.6rem 1rem;
    background: #40E0D0;
    border-radius: 15px;
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
    font-size: 0.75rem;
    color: #0A1929;
    cursor: pointer;
    user-select: none;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    background: #AFEEEE;
    padding: 0.35rem 0.75rem;
    border-radius: 12px;
  }

  .auto-refresh-toggle input {
    cursor: pointer;
    width: 14px;
    height: 14px;
  }

  h2 {
    margin: 0;
    color: #0A1929;
    text-transform: uppercase;
    letter-spacing: 0.12em;
    font-size: 1.3rem;
    font-weight: 700;
  }

  h3 {
    margin: 0 0 0.5rem 0;
    color: #E0FFFF;
    font-size: 1rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  h4 {
    margin: 0 0 0.4rem 0;
    color: #40E0D0;
    font-size: 0.85rem;
    border-bottom: 2px solid #66CDAA;
    padding-bottom: 0.3rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .btn {
    padding: 0.4rem 1rem;
    border: none;
    border-radius: 15px;
    cursor: pointer;
    font-size: 0.8rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    transition: all 0.2s;
    white-space: nowrap;
  }

  .btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .btn-primary {
    background-color: #AFEEEE;
    color: #0A1929;
  }

  .btn-primary:hover:not(:disabled) {
    background-color: #E0FFFF;
    transform: translateY(-1px);
    box-shadow: 0 0 10px rgba(175, 238, 238, 0.5);
  }

  .btn-secondary {
    background-color: #66CDAA;
    color: #0A1929;
  }

  .btn-secondary:hover:not(:disabled) {
    background-color: #7FFFD4;
    transform: translateY(-1px);
  }

  .error-message {
    background-color: #FF7F50;
    color: #0A1929;
    padding: 0.6rem 1rem;
    border-radius: 15px;
    margin-bottom: 0.75rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    font-size: 0.8rem;
  }

  .content {
    display: grid;
    grid-template-columns: 1fr 2fr;
    gap: 0.75rem;
  }

  .workflow-list {
    background-color: #0F2030;
    border-radius: 15px;
    padding: 0.75rem;
    border: 2px solid #40E0D0;
  }

  .workflow-items {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }

  .workflow-item {
    padding: 0.6rem 1rem;
    background: linear-gradient(to right, #AFEEEE 0%, #AFEEEE 95%, #E0FFFF 95%, #E0FFFF 100%);
    border-radius: 15px;
    cursor: pointer;
    transition: all 0.2s;
    border: 2px solid transparent;
  }

  .workflow-item:hover {
    background: linear-gradient(to right, #E0FFFF 0%, #E0FFFF 95%, #FFFFFF 95%, #FFFFFF 100%);
    transform: translateX(3px);
  }

  .workflow-item.selected {
    background: linear-gradient(to right, #20B2AA 0%, #20B2AA 95%, #00CED1 95%, #00CED1 100%);
    box-shadow: 0 0 15px rgba(32, 178, 170, 0.5);
  }

  .workflow-header {
    display: flex;
    align-items: flex-start;
    gap: 0.6rem;
    margin-bottom: 0.3rem;
  }

  .workflow-icon {
    font-size: 1.2rem;
  }

  .workflow-info {
    flex: 1;
  }

  .workflow-id {
    font-family: monospace;
    font-size: 0.7rem;
    color: #0A1929;
    word-break: break-all;
    opacity: 0.7;
  }

  .workflow-time {
    font-size: 0.7rem;
    color: #0A1929;
    margin-top: 0.15rem;
    font-weight: 600;
    opacity: 0.7;
  }

  .workflow-status {
    font-weight: 700;
    font-size: 0.85rem;
    margin-bottom: 0.15rem;
    color: #0A1929;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .workflow-duration {
    font-size: 0.75rem;
    color: #0A1929;
    font-weight: 600;
    opacity: 0.8;
  }

  .workflow-details {
    background-color: #0F2030;
    border-radius: 15px;
    padding: 0.75rem;
    border: 2px solid #66CDAA;
  }

  .details-content {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .detail-card {
    background-color: #1A2F3F;
    padding: 0.75rem;
    border-radius: 12px;
    border-left: 4px solid #40E0D0;
  }

  .error-card {
    background-color: #1F1A1A;
    border-left-color: #FF7F50;
  }

  .success-card {
    background-color: #1A2F28;
    border-left-color: #AFEEEE;
  }

  .info-card {
    background-color: #1A252F;
    border-left-color: #40E0D0;
  }

  .status-display {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 0.75rem 1rem;
    background: linear-gradient(135deg, #40E0D0 0%, #66CDAA 100%);
    border-radius: 15px;
    border: 2px solid #AFEEEE;
  }

  .status-icon {
    font-size: 1.8rem;
  }

  .status-text {
    font-size: 1.3rem;
    font-weight: 700;
    color: #0A1929;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  dl {
    margin: 0;
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0.4rem 1rem;
  }

  dt {
    font-weight: 700;
    color: #40E0D0;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    font-size: 0.75rem;
  }

  dd {
    margin: 0;
    color: #E0FFFF;
    font-weight: 500;
    font-size: 0.8rem;
  }

  .monospace {
    font-family: monospace;
    font-size: 0.75rem;
    word-break: break-all;
    color: #7FDBFF;
  }

  .error-content {
    background-color: #0A1929;
    padding: 0.75rem;
    border-radius: 10px;
    font-family: monospace;
    font-size: 0.75rem;
    color: #FF7F50;
    white-space: pre-wrap;
    word-break: break-word;
    border: 2px solid #FF7F50;
  }

  .result-summary {
    background: linear-gradient(to right, #AFEEEE 0%, #E0FFFF 100%);
    padding: 0.6rem 1rem;
    border-radius: 12px;
    font-weight: 700;
    color: #0A1929;
    margin-bottom: 0.5rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    font-size: 0.8rem;
  }

  .discovered-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
  }

  .tag-item {
    background: linear-gradient(135deg, #40E0D0 0%, #7FDBFF 100%);
    color: #0A1929;
    padding: 0.4rem 0.75rem;
    border-radius: 12px;
    font-family: monospace;
    font-size: 0.75rem;
    border: 2px solid #AFEEEE;
    font-weight: 700;
  }

  .no-results {
    color: #66CDAA;
    font-style: italic;
    text-align: center;
    padding: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-size: 0.75rem;
  }

  .loading,
  .empty-state {
    text-align: center;
    color: #E0FFFF;
    padding: 1.5rem;
    font-size: 0.9rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .info-card p {
    margin: 0 0 0.5rem 0;
    color: #E0FFFF;
    line-height: 1.5;
    font-size: 0.8rem;
  }
</style>
