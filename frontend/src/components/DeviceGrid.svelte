<script>
  import { createEventDispatcher } from 'svelte';
  import { controlDevice } from '../api.js';

  export let devices = [];
  export let authenticated = false;

  const dispatch = createEventDispatcher();

  function getDeviceIcon(type) {
    switch (type) {
      case 'pump': return '⛽';
      case 'light': return '💡';
      case 'valve': return '🚰';
      default: return '⚙️';
    }
  }

  function getStatusColor(status) {
    switch (status) {
      case 'on': return '#48bb78';
      case 'off': return '#718096';
      case 'error': return '#f56565';
      default: return '#a0aec0';
    }
  }

  async function toggleDevice(device) {
    if (!authenticated) return;
    
    try {
      await controlDevice(device.id, 'toggle');
      dispatch('refresh');
    } catch (err) {
      alert('Failed to control device: ' + err.message);
    }
  }

  function formatTimestamp(timestamp) {
    if (!timestamp) return 'Never';
    const date = new Date(timestamp);
    return date.toLocaleString();
  }
</script>

<div class="device-grid">
  {#each devices as device (device.id)}
    <div class="device-card" class:disabled={!device.enabled}>
      <div class="device-header">
        <span class="device-icon">{getDeviceIcon(device.type)}</span>
        <div class="device-info">
          <h3>{device.name}</h3>
          <span class="device-type">{device.type}</span>
        </div>
        <div 
          class="status-indicator" 
          style="background-color: {getStatusColor(device.status)}"
          title="{device.status}"
        ></div>
      </div>
      
      <div class="device-body">
        <div class="device-status">
          Status: <strong style="color: {getStatusColor(device.status)}">{device.status}</strong>
        </div>
        
        <div class="device-updated">
          Last updated: {formatTimestamp(device.last_updated)}
        </div>

        {#if authenticated && device.enabled}
          <button 
            class="control-btn"
            class:on={device.status === 'on'}
            on:click={() => toggleDevice(device)}
          >
            {device.status === 'on' ? 'Turn Off' : 'Turn On'}
          </button>
        {:else if !device.enabled}
          <div class="disabled-notice">Device disabled</div>
        {/if}
      </div>
    </div>
  {/each}

  {#if devices.length === 0}
    <div class="empty-state">
      <p>No devices configured</p>
    </div>
  {/if}
</div>

<style>
  .device-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 20px;
  }

  .device-card {
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(10px);
    border-radius: 12px;
    padding: 20px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    transition: all 0.3s ease;
  }

  .device-card:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 16px rgba(0, 0, 0, 0.15);
  }

  .device-card.disabled {
    opacity: 0.6;
  }

  .device-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 16px;
  }

  .device-icon {
    font-size: 32px;
  }

  .device-info {
    flex: 1;
  }

  .device-info h3 {
    margin: 0;
    font-size: 16px;
    color: #2d3748;
  }

  .device-type {
    font-size: 12px;
    color: #718096;
    text-transform: capitalize;
  }

  .status-indicator {
    width: 16px;
    height: 16px;
    border-radius: 50%;
    box-shadow: 0 0 8px rgba(0, 0, 0, 0.2);
  }

  .device-body {
    border-top: 1px solid #e2e8f0;
    padding-top: 16px;
  }

  .device-status {
    font-size: 14px;
    color: #4a5568;
    margin-bottom: 8px;
  }

  .device-updated {
    font-size: 11px;
    color: #a0aec0;
    margin-bottom: 12px;
  }

  .control-btn {
    width: 100%;
    padding: 10px;
    border: none;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.3s ease;
    background: #48bb78;
    color: white;
  }

  .control-btn.on {
    background: #f56565;
  }

  .control-btn:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
  }

  .disabled-notice {
    text-align: center;
    padding: 8px;
    background: #e2e8f0;
    border-radius: 8px;
    font-size: 13px;
    color: #4a5568;
  }

  .empty-state {
    grid-column: 1 / -1;
    text-align: center;
    padding: 40px;
    background: rgba(255, 255, 255, 0.8);
    border-radius: 12px;
    color: #718096;
  }
</style>
