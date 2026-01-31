<script>
  export let status = null;
</script>

<div class="status-panel">
  {#if status}
    <div class="status-item" class:healthy={status.healthy} class:unhealthy={!status.healthy}>
      <div class="status-icon">{status.healthy ? '✅' : '⚠️'}</div>
      <div class="status-content">
        <div class="status-label">System Status</div>
        <div class="status-value">{status.healthy ? 'Healthy' : 'Issues Detected'}</div>
      </div>
    </div>

    <div class="status-item">
      <div class="status-icon">🔔</div>
      <div class="status-content">
        <div class="status-label">Active Alerts</div>
        <div class="status-value">{status.active_alerts}</div>
      </div>
    </div>

    <div class="status-item">
      <div class="status-icon">⚙️</div>
      <div class="status-content">
        <div class="status-label">Devices Online</div>
        <div class="status-value">{status.devices_online}</div>
      </div>
    </div>

    <div class="status-item">
      <div class="status-icon">📊</div>
      <div class="status-content">
        <div class="status-label">Sensors Online</div>
        <div class="status-value">{status.sensors_online}</div>
      </div>
    </div>
  {/if}
</div>

<style>
  .status-panel {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 20px;
    margin-bottom: 30px;
  }

  .status-item {
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(10px);
    padding: 24px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    gap: 16px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    transition: transform 0.3s ease, box-shadow 0.3s ease;
  }

  .status-item:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 12px rgba(0, 0, 0, 0.15);
  }

  .status-item.healthy {
    border-left: 4px solid #48bb78;
  }

  .status-item.unhealthy {
    border-left: 4px solid #f56565;
  }

  .status-icon {
    font-size: 32px;
  }

  .status-content {
    flex: 1;
  }

  .status-label {
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: #718096;
    margin-bottom: 4px;
    font-weight: 600;
  }

  .status-value {
    font-size: 24px;
    font-weight: 700;
    color: #2d3748;
  }
</style>
