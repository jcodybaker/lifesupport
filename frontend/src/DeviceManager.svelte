<script>
  import { onMount } from 'svelte';
  import { deviceAPI, sensorAPI, actuatorAPI } from './api.js';

  let devices = [];
  let selectedDevice = null;
  let showCreateForm = false;
  let showEditForm = false;
  let loading = false;
  let error = null;

  let newDevice = {
    id: '',
    driver: 'shelly',
    name: '',
    description: '',
    metadata: {},
  };

  let editingDevice = null;

  // Sensor readings and actuator states for selected device
  let sensorReadings = {};
  let actuatorStates = {};
  let loadingDetails = false;

  onMount(async () => {
    await loadDevices();
  });

  async function loadDevices() {
    loading = true;
    error = null;
    try {
      devices = await deviceAPI.list();
    } catch (err) {
      error = `Failed to load devices: ${err.message}`;
      console.error('Error loading devices:', err);
      devices = [];
    } finally {
      loading = false;
    }
  }

  async function createDevice() {
    if (!newDevice.id.trim() || !newDevice.name.trim()) {
      error = 'Device ID and name are required';
      return;
    }

    try {
      // Creating device without subsystem_id for now
      await deviceAPI.create(newDevice, null);
      await loadDevices();
      showCreateForm = false;
      newDevice = {
        id: '',
        driver: 'shelly',
        name: '',
        description: '',
        metadata: {},
      };
      error = null;
    } catch (err) {
      error = `Failed to create device: ${err.message}`;
      console.error('Error creating device:', err);
    }
  }

  async function updateDevice() {
    if (!editingDevice.name.trim()) {
      error = 'Device name is required';
      return;
    }

    try {
      await deviceAPI.update(editingDevice.id, {
        driver: editingDevice.driver,
        name: editingDevice.name,
        description: editingDevice.description,
        metadata: editingDevice.metadata,
      });
      await loadDevices();
      showEditForm = false;
      editingDevice = null;
      error = null;
    } catch (err) {
      error = `Failed to update device: ${err.message}`;
      console.error('Error updating device:', err);
    }
  }

  async function deleteDevice(deviceId) {
    if (!confirm('Are you sure you want to delete this device?')) {
      return;
    }

    try {
      await deviceAPI.delete(deviceId);
      await loadDevices();
      if (selectedDevice?.id === deviceId) {
        selectedDevice = null;
      }
      error = null;
    } catch (err) {
      error = `Failed to delete device: ${err.message}`;
      console.error('Error deleting device:', err);
    }
  }

  function startEdit(device) {
    editingDevice = { ...device };
    showEditForm = true;
    showCreateForm = false;
  }

  function cancelEdit() {
    showEditForm = false;
    editingDevice = null;
  }

  function cancelCreate() {
    showCreateForm = false;
    newDevice = {
      id: '',
      driver: 'shelly',
      name: '',
      description: '',
      metadata: {},
    };
  }

  async function selectDevice(device) {
    selectedDevice = device;
    await loadDeviceDetails(device);
  }

  async function loadDeviceDetails(device) {
    loadingDetails = true;
    sensorReadings = {};
    actuatorStates = {};

    try {
      // Load latest readings for each sensor
      if (device.sensors && device.sensors.length > 0) {
        for (const sensor of device.sensors) {
          try {
            const reading = await sensorAPI.getLatest(sensor.id);
            sensorReadings[sensor.id] = reading;
          } catch (err) {
            console.warn(`Failed to load reading for sensor ${sensor.id}:`, err);
          }
        }
      }

      // Load latest states for each actuator
      if (device.actuators && device.actuators.length > 0) {
        for (const actuator of device.actuators) {
          try {
            const state = await actuatorAPI.getLatest(actuator.id);
            actuatorStates[actuator.id] = state;
          } catch (err) {
            console.warn(`Failed to load state for actuator ${actuator.id}:`, err);
          }
        }
      }
    } catch (err) {
      console.error('Error loading device details:', err);
    } finally {
      loadingDetails = false;
    }
  }

  function formatTimestamp(timestamp) {
    if (!timestamp) return 'N/A';
    return new Date(timestamp).toLocaleString();
  }

  function getSensorTypeDisplay(type) {
    return type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
  }

  function getActuatorTypeDisplay(type) {
    return type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
  }
</script>

<div class="device-manager">
  <div class="header">
    <h2>Device Management</h2>
    <button on:click={() => { showCreateForm = !showCreateForm; showEditForm = false; }} class="btn btn-primary">
      {showCreateForm ? 'Cancel' : '+ Add Device'}
    </button>
  </div>

  {#if error}
    <div class="error-message">{error}</div>
  {/if}

  {#if showCreateForm}
    <div class="form-container">
      <h3>Create New Device</h3>
      <form on:submit|preventDefault={createDevice}>
        <div class="form-group">
          <label for="device-id">Device ID *</label>
          <input id="device-id" bind:value={newDevice.id} placeholder="dev-001" required />
        </div>
        <div class="form-group">
          <label for="device-name">Name *</label>
          <input id="device-name" bind:value={newDevice.name} placeholder="Temperature Monitor" required />
        </div>
        <div class="form-group">
          <label for="device-driver">Driver</label>
          <select id="device-driver" bind:value={newDevice.driver}>
            <option value="shelly">Shelly</option>
            <option value="station">Station</option>
          </select>
        </div>
        <div class="form-group">
          <label for="device-description">Description</label>
          <textarea id="device-description" bind:value={newDevice.description} placeholder="Optional description"></textarea>
        </div>
        <div class="form-actions">
          <button type="submit" class="btn btn-primary">Create Device</button>
          <button type="button" class="btn btn-secondary" on:click={cancelCreate}>Cancel</button>
        </div>
      </form>
    </div>
  {/if}

  {#if showEditForm && editingDevice}
    <div class="form-container">
      <h3>Edit Device: {editingDevice.id}</h3>
      <form on:submit|preventDefault={updateDevice}>
        <div class="form-group">
          <label for="edit-device-name">Name *</label>
          <input id="edit-device-name" bind:value={editingDevice.name} required />
        </div>
        <div class="form-group">
          <label for="edit-device-driver">Driver</label>
          <select id="edit-device-driver" bind:value={editingDevice.driver}>
            <option value="shelly">Shelly</option>
            <option value="station">Station</option>
          </select>
        </div>
        <div class="form-group">
          <label for="edit-device-description">Description</label>
          <textarea id="edit-device-description" bind:value={editingDevice.description}></textarea>
        </div>
        <div class="form-actions">
          <button type="submit" class="btn btn-primary">Update Device</button>
          <button type="button" class="btn btn-secondary" on:click={cancelEdit}>Cancel</button>
        </div>
      </form>
    </div>
  {/if}

  <div class="content">
    <div class="device-list">
      <h3>Devices ({devices.length})</h3>
      {#if loading}
        <div class="loading">Loading devices...</div>
      {:else if devices.length === 0}
        <div class="empty-state">No devices found. Create one to get started.</div>
      {:else}
        <div class="device-items">
          {#each devices as device (device.id)}
            <div 
              class="device-item" 
              class:selected={selectedDevice?.id === device.id}
              on:click={() => selectDevice(device)}
            >
              <div class="device-info">
                <div class="device-name">{device.name}</div>
                <div class="device-id">{device.id}</div>
                <div class="device-driver">{device.driver}</div>
                {#if device.sensors || device.actuators}
                  <div class="device-stats">
                    {#if device.sensors}
                      <span class="stat">📊 {device.sensors.length} sensor{device.sensors.length !== 1 ? 's' : ''}</span>
                    {/if}
                    {#if device.actuators}
                      <span class="stat">⚡ {device.actuators.length} actuator{device.actuators.length !== 1 ? 's' : ''}</span>
                    {/if}
                  </div>
                {/if}
              </div>
              <div class="device-actions">
                <button class="btn-icon" on:click|stopPropagation={() => startEdit(device)} title="Edit">✏️</button>
                <button class="btn-icon" on:click|stopPropagation={() => deleteDevice(device.id)} title="Delete">🗑️</button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>

    <div class="device-details">
      {#if selectedDevice}
        <h3>Device Details: {selectedDevice.name}</h3>
        <div class="details-content">
          <div class="detail-section">
            <h4>Information</h4>
            <dl>
              <dt>ID:</dt><dd>{selectedDevice.id}</dd>
              <dt>Driver:</dt><dd>{selectedDevice.driver}</dd>
              {#if selectedDevice.description}
                <dt>Description:</dt><dd>{selectedDevice.description}</dd>
              {/if}
              {#if selectedDevice.tags && selectedDevice.tags.length > 0}
                <dt>Tags:</dt><dd>{selectedDevice.tags.join(', ')}</dd>
              {/if}
            </dl>
          </div>

          {#if selectedDevice.sensors && selectedDevice.sensors.length > 0}
            <div class="detail-section">
              <h4>Sensors ({selectedDevice.sensors.length})</h4>
              {#if loadingDetails}
                <div class="loading-small">Loading sensor data...</div>
              {:else}
                <div class="sensor-list">
                  {#each selectedDevice.sensors as sensor (sensor.id)}
                    <div class="sensor-card">
                      <div class="sensor-header">
                        <span class="sensor-name">{sensor.name}</span>
                        <span class="sensor-type">{getSensorTypeDisplay(sensor.sensor_type)}</span>
                      </div>
                      <div class="sensor-id">{sensor.id}</div>
                      {#if sensorReadings[sensor.id]}
                        <div class="sensor-reading">
                          <span class="reading-value">
                            {sensorReadings[sensor.id].value} {sensorReadings[sensor.id].unit}
                          </span>
                          <span class="reading-time">
                            {formatTimestamp(sensorReadings[sensor.id].timestamp)}
                          </span>
                          {#if !sensorReadings[sensor.id].valid}
                            <span class="reading-error">⚠️ Invalid</span>
                          {/if}
                        </div>
                      {:else}
                        <div class="no-data">No recent readings</div>
                      {/if}
                    </div>
                  {/each}
                </div>
              {/if}
            </div>
          {/if}

          {#if selectedDevice.actuators && selectedDevice.actuators.length > 0}
            <div class="detail-section">
              <h4>Actuators ({selectedDevice.actuators.length})</h4>
              {#if loadingDetails}
                <div class="loading-small">Loading actuator data...</div>
              {:else}
                <div class="actuator-list">
                  {#each selectedDevice.actuators as actuator (actuator.id)}
                    <div class="actuator-card">
                      <div class="actuator-header">
                        <span class="actuator-name">{actuator.name}</span>
                        <span class="actuator-type">{getActuatorTypeDisplay(actuator.actuator_type)}</span>
                      </div>
                      <div class="actuator-id">{actuator.id}</div>
                      {#if actuatorStates[actuator.id]}
                        <div class="actuator-state">
                          <span class="state-status" class:active={actuatorStates[actuator.id].active}>
                            {actuatorStates[actuator.id].active ? '🟢 Active' : '⚪ Inactive'}
                          </span>
                          <span class="state-time">
                            {formatTimestamp(actuatorStates[actuator.id].timestamp)}
                          </span>
                          {#if actuatorStates[actuator.id].parameters && Object.keys(actuatorStates[actuator.id].parameters).length > 0}
                            <div class="state-params">
                              {#each Object.entries(actuatorStates[actuator.id].parameters) as [key, value]}
                                <span class="param">{key}: {value}</span>
                              {/each}
                            </div>
                          {/if}
                        </div>
                      {:else}
                        <div class="no-data">No recent state</div>
                      {/if}
                    </div>
                  {/each}
                </div>
              {/if}
            </div>
          {/if}

          {#if (!selectedDevice.sensors || selectedDevice.sensors.length === 0) && (!selectedDevice.actuators || selectedDevice.actuators.length === 0)}
            <div class="empty-state">This device has no sensors or actuators configured.</div>
          {/if}
        </div>
      {:else}
        <div class="empty-state">Select a device to view details</div>
      {/if}
    </div>
  </div>
</div>

<style>
  .device-manager {
    padding: 1rem;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
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
    transition: background-color 0.2s;
  }

  .btn-primary {
    background-color: #4CAF50;
    color: white;
  }

  .btn-primary:hover {
    background-color: #45a049;
  }

  .btn-secondary {
    background-color: #757575;
    color: white;
  }

  .btn-secondary:hover {
    background-color: #616161;
  }

  .btn-icon {
    background: none;
    border: none;
    cursor: pointer;
    font-size: 1.2rem;
    padding: 0.25rem;
    transition: transform 0.2s;
  }

  .btn-icon:hover {
    transform: scale(1.2);
  }

  .error-message {
    background-color: #ffebee;
    color: #c62828;
    padding: 0.75rem;
    border-radius: 4px;
    margin-bottom: 1rem;
    border-left: 4px solid #c62828;
  }

  .form-container {
    background-color: #f5f5f5;
    padding: 1.5rem;
    border-radius: 8px;
    margin-bottom: 1.5rem;
  }

  .form-group {
    margin-bottom: 1rem;
  }

  .form-group label {
    display: block;
    margin-bottom: 0.25rem;
    font-weight: 500;
    color: #555;
  }

  .form-group input,
  .form-group select,
  .form-group textarea {
    width: 100%;
    padding: 0.5rem;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 0.9rem;
  }

  .form-group textarea {
    min-height: 80px;
    resize: vertical;
  }

  .form-actions {
    display: flex;
    gap: 0.5rem;
    margin-top: 1rem;
  }

  .content {
    display: grid;
    grid-template-columns: 1fr 2fr;
    gap: 1.5rem;
  }

  .device-list {
    background-color: white;
    border-radius: 8px;
    padding: 1rem;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  }

  .device-items {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .device-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    background-color: #f9f9f9;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.2s;
    border: 2px solid transparent;
  }

  .device-item:hover {
    background-color: #f0f0f0;
  }

  .device-item.selected {
    background-color: #e8f5e9;
    border-color: #4CAF50;
  }

  .device-info {
    flex: 1;
  }

  .device-name {
    font-weight: 600;
    color: #333;
    margin-bottom: 0.25rem;
  }

  .device-id {
    font-size: 0.85rem;
    color: #666;
    font-family: monospace;
  }

  .device-driver {
    display: inline-block;
    font-size: 0.75rem;
    background-color: #2196F3;
    color: white;
    padding: 0.15rem 0.5rem;
    border-radius: 12px;
    margin-top: 0.25rem;
  }

  .device-stats {
    margin-top: 0.5rem;
    font-size: 0.85rem;
    color: #777;
  }

  .stat {
    margin-right: 0.75rem;
  }

  .device-actions {
    display: flex;
    gap: 0.25rem;
  }

  .device-details {
    background-color: white;
    border-radius: 8px;
    padding: 1rem;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  }

  .details-content {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .detail-section {
    background-color: #fafafa;
    padding: 1rem;
    border-radius: 6px;
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

  .sensor-list,
  .actuator-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
    gap: 1rem;
  }

  .sensor-card,
  .actuator-card {
    background-color: white;
    padding: 1rem;
    border-radius: 6px;
    border: 1px solid #e0e0e0;
  }

  .sensor-header,
  .actuator-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
  }

  .sensor-name,
  .actuator-name {
    font-weight: 600;
    color: #333;
  }

  .sensor-type,
  .actuator-type {
    font-size: 0.75rem;
    background-color: #9C27B0;
    color: white;
    padding: 0.15rem 0.5rem;
    border-radius: 12px;
  }

  .sensor-id,
  .actuator-id {
    font-size: 0.8rem;
    color: #999;
    font-family: monospace;
    margin-bottom: 0.75rem;
  }

  .sensor-reading,
  .actuator-state {
    border-top: 1px solid #f0f0f0;
    padding-top: 0.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .reading-value {
    font-size: 1.2rem;
    font-weight: 600;
    color: #4CAF50;
  }

  .reading-time,
  .state-time {
    font-size: 0.75rem;
    color: #999;
  }

  .reading-error {
    color: #f44336;
    font-size: 0.8rem;
    font-weight: 600;
  }

  .state-status {
    font-weight: 600;
  }

  .state-status.active {
    color: #4CAF50;
  }

  .state-params {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-top: 0.5rem;
  }

  .param {
    font-size: 0.8rem;
    background-color: #e3f2fd;
    color: #1976D2;
    padding: 0.15rem 0.5rem;
    border-radius: 4px;
  }

  .loading,
  .loading-small,
  .empty-state {
    text-align: center;
    color: #999;
    padding: 2rem;
  }

  .loading-small {
    padding: 1rem;
  }

  .no-data {
    color: #999;
    font-style: italic;
    font-size: 0.85rem;
  }
</style>
