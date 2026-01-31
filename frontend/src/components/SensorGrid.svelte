<script>
  export let sensors = [];

  function getSensorIcon(type) {
    switch (type) {
      case 'temperature': return '🌡️';
      case 'ph': return '⚗️';
      case 'flow': return '🌊';
      case 'weight': return '⚖️';
      case 'distance': return '📏';
      default: return '📊';
    }
  }

  function formatValue(value, unit) {
    if (value === null || value === undefined) return 'N/A';
    return `${value.toFixed(2)} ${unit}`;
  }

  function formatTimestamp(timestamp) {
    if (!timestamp) return 'Never';
    const date = new Date(timestamp);
    return date.toLocaleString();
  }
</script>

<div class="sensor-grid">
  {#each sensors as sensor (sensor.id)}
    <div class="sensor-card" class:disabled={!sensor.enabled}>
      <div class="sensor-header">
        <span class="sensor-icon">{getSensorIcon(sensor.type)}</span>
        <div class="sensor-info">
          <h3>{sensor.name}</h3>
          <span class="sensor-type">{sensor.type}</span>
        </div>
        {#if !sensor.enabled}
          <span class="disabled-badge">Disabled</span>
        {/if}
      </div>
      
      <div class="sensor-body">
        <div class="sensor-value">
          {formatValue(sensor.last_value, sensor.unit)}
        </div>
        {#if sensor.location}
          <div class="sensor-location">📍 {sensor.location}</div>
        {/if}
        <div class="sensor-updated">
          Last updated: {formatTimestamp(sensor.last_updated)}
        </div>
      </div>
    </div>
  {/each}

  {#if sensors.length === 0}
    <div class="empty-state">
      <p>No sensors configured</p>
    </div>
  {/if}
</div>

<style>
  .sensor-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 20px;
  }

  .sensor-card {
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(10px);
    border-radius: 12px;
    padding: 20px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    transition: all 0.3s ease;
  }

  .sensor-card:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 16px rgba(0, 0, 0, 0.15);
  }

  .sensor-card.disabled {
    opacity: 0.6;
  }

  .sensor-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 16px;
  }

  .sensor-icon {
    font-size: 32px;
  }

  .sensor-info {
    flex: 1;
  }

  .sensor-info h3 {
    margin: 0;
    font-size: 16px;
    color: #2d3748;
  }

  .sensor-type {
    font-size: 12px;
    color: #718096;
    text-transform: capitalize;
  }

  .disabled-badge {
    background: #e2e8f0;
    color: #4a5568;
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 600;
  }

  .sensor-body {
    border-top: 1px solid #e2e8f0;
    padding-top: 16px;
  }

  .sensor-value {
    font-size: 28px;
    font-weight: 700;
    color: #2d3748;
    margin-bottom: 8px;
  }

  .sensor-location {
    font-size: 13px;
    color: #4a5568;
    margin-bottom: 4px;
  }

  .sensor-updated {
    font-size: 11px;
    color: #a0aec0;
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
