<script>
  import { createEventDispatcher } from 'svelte';
  import { acknowledgeAlert, deleteAlert } from '../api.js';

  export let alerts = [];
  export let authenticated = false;

  const dispatch = createEventDispatcher();

  function getAlertIcon(type) {
    switch (type) {
      case 'critical': return '🚨';
      case 'error': return '❌';
      case 'warning': return '⚠️';
      default: return 'ℹ️';
    }
  }

  function getAlertColor(type) {
    switch (type) {
      case 'critical': return '#c53030';
      case 'error': return '#f56565';
      case 'warning': return '#ed8936';
      default: return '#4299e1';
    }
  }

  async function handleAcknowledge(alert) {
    if (!authenticated) return;
    
    try {
      await acknowledgeAlert(alert.id);
      dispatch('refresh');
    } catch (err) {
      console.error('Failed to acknowledge alert:', err);
    }
  }

  async function handleDelete(alert) {
    if (!authenticated) return;
    if (!confirm('Delete this alert?')) return;
    
    try {
      await deleteAlert(alert.id);
      dispatch('refresh');
    } catch (err) {
      console.error('Failed to delete alert:', err);
    }
  }

  function formatTimestamp(timestamp) {
    if (!timestamp) return '';
    const date = new Date(timestamp);
    return date.toLocaleString();
  }
</script>

<div class="alert-panel">
  <h2>🔔 Active Alerts</h2>
  <div class="alert-list">
    {#each alerts.filter(a => !a.acknowledged && !a.resolved_at) as alert (alert.id)}
      <div class="alert-item" style="border-left-color: {getAlertColor(alert.type)}">
        <div class="alert-header">
          <span class="alert-icon">{getAlertIcon(alert.type)}</span>
          <div class="alert-info">
            <span class="alert-type" style="color: {getAlertColor(alert.type)}">{alert.type.toUpperCase()}</span>
            {#if alert.source}
              <span class="alert-source">from {alert.source}</span>
            {/if}
          </div>
          <span class="alert-time">{formatTimestamp(alert.created_at)}</span>
        </div>
        
        <div class="alert-message">{alert.message}</div>
        
        {#if authenticated}
          <div class="alert-actions">
            <button class="btn-ack" on:click={() => handleAcknowledge(alert)}>
              Acknowledge
            </button>
            <button class="btn-delete" on:click={() => handleDelete(alert)}>
              Delete
            </button>
          </div>
        {/if}
      </div>
    {/each}

    {#if alerts.filter(a => !a.acknowledged && !a.resolved_at).length === 0}
      <div class="no-alerts">
        <p>✅ No active alerts</p>
      </div>
    {/if}
  </div>
</div>

<style>
  .alert-panel {
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(10px);
    border-radius: 12px;
    padding: 24px;
    margin-bottom: 30px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  }

  h2 {
    margin-top: 0;
    margin-bottom: 20px;
    color: #2d3748;
    font-size: 20px;
  }

  .alert-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .alert-item {
    background: white;
    border-left: 4px solid;
    border-radius: 8px;
    padding: 16px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  }

  .alert-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 12px;
  }

  .alert-icon {
    font-size: 24px;
  }

  .alert-info {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .alert-type {
    font-weight: 700;
    font-size: 13px;
  }

  .alert-source {
    font-size: 12px;
    color: #718096;
  }

  .alert-time {
    font-size: 11px;
    color: #a0aec0;
  }

  .alert-message {
    color: #4a5568;
    line-height: 1.6;
    margin-bottom: 12px;
  }

  .alert-actions {
    display: flex;
    gap: 8px;
  }

  .alert-actions button {
    padding: 6px 12px;
    border: none;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .btn-ack {
    background: #48bb78;
    color: white;
  }

  .btn-ack:hover {
    background: #38a169;
  }

  .btn-delete {
    background: #e2e8f0;
    color: #4a5568;
  }

  .btn-delete:hover {
    background: #cbd5e0;
  }

  .no-alerts {
    text-align: center;
    padding: 40px;
    color: #48bb78;
    font-weight: 600;
  }
</style>
