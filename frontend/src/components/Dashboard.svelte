<script>
  import { onMount, onDestroy } from 'svelte';
  import { getSystemStatus, getSensors, getDevices, getCameras, getAlerts } from '../api.js';
  import SystemStatus from './SystemStatus.svelte';
  import SensorGrid from './SensorGrid.svelte';
  import DeviceGrid from './DeviceGrid.svelte';
  import CameraGrid from './CameraGrid.svelte';
  import AlertPanel from './AlertPanel.svelte';

  export let authenticated = false;

  let status = null;
  let sensors = [];
  let devices = [];
  let cameras = [];
  let alerts = [];
  let loading = true;
  let error = null;
  let refreshInterval;

  async function fetchData() {
    try {
      error = null;
      const [statusData, sensorsData, devicesData, camerasData, alertsData] = await Promise.all([
        getSystemStatus(),
        getSensors(),
        getDevices(),
        getCameras(),
        getAlerts(),
      ]);

      status = statusData;
      sensors = sensorsData || [];
      devices = devicesData || [];
      cameras = camerasData || [];
      alerts = alertsData || [];
    } catch (err) {
      error = err.message;
      console.error('Failed to fetch data:', err);
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    fetchData();
    // Refresh data every 10 seconds
    refreshInterval = setInterval(fetchData, 10000);
  });

  onDestroy(() => {
    if (refreshInterval) {
      clearInterval(refreshInterval);
    }
  });
</script>

<div class="dashboard">
  {#if loading}
    <div class="loading">
      <div class="spinner"></div>
      <p>Loading system data...</p>
    </div>
  {:else if error}
    <div class="error-message">
      <h3>⚠️ Error Loading Data</h3>
      <p>{error}</p>
      <button on:click={fetchData}>Retry</button>
    </div>
  {:else}
    <SystemStatus {status} />
    
    {#if alerts && alerts.length > 0}
      <AlertPanel {alerts} {authenticated} on:refresh={fetchData} />
    {/if}

    <div class="grid-section">
      <h2>📊 Sensors</h2>
      <SensorGrid {sensors} />
    </div>

    <div class="grid-section">
      <h2>⚙️ Devices</h2>
      <DeviceGrid {devices} {authenticated} on:refresh={fetchData} />
    </div>

    <div class="grid-section">
      <h2>📹 Cameras</h2>
      <CameraGrid {cameras} />
    </div>
  {/if}
</div>

<style>
  .dashboard {
    animation: fadeIn 0.5s ease-in;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .loading {
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(10px);
    padding: 60px;
    border-radius: 12px;
    text-align: center;
  }

  .spinner {
    width: 50px;
    height: 50px;
    margin: 0 auto 20px;
    border: 4px solid #e2e8f0;
    border-top-color: #667eea;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .error-message {
    background: #fed7d7;
    border: 2px solid #fc8181;
    color: #c53030;
    padding: 30px;
    border-radius: 12px;
    text-align: center;
  }

  .error-message h3 {
    margin-top: 0;
  }

  .error-message button {
    margin-top: 20px;
    padding: 10px 24px;
    background: #c53030;
    color: white;
    border: none;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
  }

  .error-message button:hover {
    background: #9b2c2c;
  }

  .grid-section {
    margin-top: 30px;
  }

  .grid-section h2 {
    color: white;
    margin-bottom: 16px;
    font-size: 24px;
  }
</style>
