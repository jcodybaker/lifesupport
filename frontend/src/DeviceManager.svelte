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

      // Load latest states for each actuator using the new by-tag endpoint
      if (device.actuators && device.actuators.length > 0) {
        for (const actuator of device.actuators) {
          try {
            // Use the first tag (typically the default tag) to fetch status
            if (actuator.tags && actuator.tags.length > 0) {
              const status = await actuatorAPI.getStatusByTag(actuator.tags[0]);
              if (status) {
                actuatorStates[actuator.device_id+"/"+actuator.id] = status;
              }
            }
          } catch (err) {
            console.warn(`Failed to load status for actuator ${actuator.id}:`, err);
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
                      {#if actuatorStates[actuator.device_id+"/"+actuator.id]}
                        <div class="actuator-state">
                          <span class="state-status" class:active={actuatorStates[actuator.device_id+"/"+actuator.id].active}>
                            {actuatorStates[actuator.device_id+"/"+actuator.id].value > 0 ? '🟢 Active' : '⚪ Inactive'}
                          </span>
                          <span class="state-time">
                            {formatTimestamp(actuatorStates[actuator.device_id+"/"+actuator.id].timestamp)}
                          </span>
                          {#if actuatorStates[actuator.device_id+"/"+actuator.id].parameters && Object.keys(actuatorStates[actuator.device_id+"/"+actuator.id].parameters).length > 0}
                            <div class="state-params">
                              {#each Object.entries(actuatorStates[actuator.device_id+"/"+actuator.id].parameters) as [key, value]}
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
    padding: 0;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.75rem;
    padding: 0.6rem 1rem;
    background: #20B2AA;
    border-radius: 15px;
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
  }

  .btn-primary {
    background-color: #40E0D0;
    color: #0A1929;
  }

  .btn-primary:hover {
    background-color: #7FDBFF;
    transform: translateY(-1px);
    box-shadow: 0 0 10px rgba(64, 224, 208, 0.5);
  }

  .btn-secondary {
    background-color: #66CDAA;
    color: #0A1929;
  }

  .btn-secondary:hover {
    background-color: #7FFFD4;
    transform: translateY(-1px);
  }

  .btn-icon {
    background: #FF7F50;
    border: none;
    cursor: pointer;
    font-size: 1rem;
    padding: 0.35rem;
    border-radius: 50%;
    transition: all 0.2s;
    color: #0A1929;
  }

  .btn-icon:hover {
    transform: scale(1.1);
    background: #FFA07A;
    box-shadow: 0 0 8px rgba(255, 127, 80, 0.5);
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

  .form-container {
    background-color: #0F2030;
    padding: 1rem;
    border-radius: 15px;
    margin-bottom: 0.75rem;
    border: 2px solid #40E0D0;
  }

  .form-group {
    margin-bottom: 0.6rem;
  }

  .form-group label {
    display: block;
    margin-bottom: 0.3rem;
    font-weight: 700;
    color: #E0FFFF;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    font-size: 0.75rem;
  }

  .form-group input,
  .form-group select,
  .form-group textarea {
    width: 100%;
    padding: 0.4rem 0.6rem;
    border: 2px solid #66CDAA;
    border-radius: 10px;
    font-size: 0.85rem;
    background: #0A1929;
    color: #E0FFFF;
    font-weight: 500;
  }

  .form-group input:focus,
  .form-group select:focus,
  .form-group textarea:focus {
    outline: none;
    border-color: #40E0D0;
    box-shadow: 0 0 8px rgba(64, 224, 208, 0.3);
  }

  .form-group textarea {
    min-height: 60px;
    resize: vertical;
    font-family: monospace;
  }

  .form-actions {
    display: flex;
    gap: 0.5rem;
    margin-top: 0.75rem;
  }

  .content {
    display: grid;
    grid-template-columns: 1fr 2fr;
    gap: 0.75rem;
  }

  .device-list {
    background-color: #0F2030;
    border-radius: 15px;
    padding: 0.75rem;
    border: 2px solid #20B2AA;
  }

  .device-items {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }

  .device-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.6rem 1rem;
    background: linear-gradient(to right, #40E0D0 0%, #40E0D0 95%, #7FDBFF 95%, #7FDBFF 100%);
    border-radius: 15px;
    cursor: pointer;
    transition: all 0.2s;
    border: 2px solid transparent;
  }

  .device-item:hover {
    background: linear-gradient(to right, #7FDBFF 0%, #7FDBFF 95%, #AFEEEE 95%, #AFEEEE 100%);
    transform: translateX(3px);
  }

  .device-item.selected {
    background: linear-gradient(to right, #20B2AA 0%, #20B2AA 95%, #00CED1 95%, #00CED1 100%);
    box-shadow: 0 0 15px rgba(32, 178, 170, 0.5);
  }

  .device-info {
    flex: 1;
  }

  .device-name {
    font-weight: 700;
    color: #0A1929;
    margin-bottom: 0.15rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-size: 0.9rem;
  }

  .device-id {
    font-size: 0.7rem;
    color: #0A1929;
    font-family: monospace;
    opacity: 0.7;
  }

  .device-driver {
    display: inline-block;
    font-size: 0.65rem;
    background-color: #AFEEEE;
    color: #0A1929;
    padding: 0.15rem 0.5rem;
    border-radius: 10px;
    margin-top: 0.3rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .device-stats {
    margin-top: 0.3rem;
    font-size: 0.7rem;
    color: #0A1929;
    opacity: 0.8;
  }

  .stat {
    margin-right: 0.5rem;
    font-weight: 600;
  }

  .device-actions {
    display: flex;
    gap: 0.3rem;
  }

  .device-details {
    background-color: #0F2030;
    border-radius: 15px;
    padding: 0.75rem;
    border: 2px solid #66CDAA;
  }

  .details-content {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .detail-section {
    background-color: #1A2F3F;
    padding: 0.75rem;
    border-radius: 12px;
    border-left: 4px solid #40E0D0;
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

  .sensor-list,
  .actuator-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 0.5rem;
  }

  .sensor-card,
  .actuator-card {
    background: linear-gradient(135deg, #66CDAA 0%, #40E0D0 100%);
    padding: 0.75rem;
    border-radius: 12px;
    border: 2px solid #AFEEEE;
  }

  .sensor-header,
  .actuator-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.4rem;
  }

  .sensor-name,
  .actuator-name {
    font-weight: 700;
    color: #0A1929;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-size: 0.85rem;
  }

  .sensor-type,
  .actuator-type {
    font-size: 0.65rem;
    background-color: #AFEEEE;
    color: #0A1929;
    padding: 0.15rem 0.5rem;
    border-radius: 10px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .sensor-id,
  .actuator-id {
    font-size: 0.7rem;
    color: #0A1929;
    font-family: monospace;
    margin-bottom: 0.4rem;
    opacity: 0.7;
  }

  .sensor-reading,
  .actuator-state {
    border-top: 2px solid rgba(10, 25, 41, 0.3);
    padding-top: 0.5rem;
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
  }

  .reading-value {
    font-size: 1.2rem;
    font-weight: 700;
    color: #E0FFFF;
    text-shadow: 0 0 8px rgba(224, 255, 255, 0.5);
  }

  .reading-time,
  .state-time {
    font-size: 0.65rem;
    color: #0A1929;
    opacity: 0.7;
    font-weight: 600;
  }

  .reading-error {
    color: #FF7F50;
    font-size: 0.75rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .state-status {
    font-weight: 700;
    font-size: 0.95rem;
  }

  .state-status.active {
    color: #E0FFFF;
    text-shadow: 0 0 8px rgba(224, 255, 255, 0.5);
  }

  .state-params {
    display: flex;
    flex-wrap: wrap;
    gap: 0.3rem;
    margin-top: 0.3rem;
  }

  .param {
    font-size: 0.7rem;
    background-color: #20B2AA;
    color: #0A1929;
    padding: 0.15rem 0.5rem;
    border-radius: 8px;
    font-weight: 700;
  }

  .loading,
  .loading-small,
  .empty-state {
    text-align: center;
    color: #E0FFFF;
    padding: 1.5rem;
    font-size: 0.9rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .loading-small {
    padding: 0.75rem;
  }

  .no-data {
    color: #66CDAA;
    font-style: italic;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
</style>
