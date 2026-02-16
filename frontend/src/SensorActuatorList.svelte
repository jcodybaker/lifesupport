<script>
  import { onMount } from 'svelte';
  import { sensorAPI, actuatorAPI } from './api.js';

  let sensors = [];
  let actuators = [];
  let sensorReadings = {};
  let actuatorStates = {};
  let loading = false;
  let error = null;
  let showSensors = true;
  let showActuators = true;
  
  // Edit state
  let editingSensor = null;
  let editingActuator = null;
  let showEditSensorForm = false;
  let showEditActuatorForm = false;

  onMount(async () => {
    await loadAll();
    // Refresh data every 10 seconds
    const interval = setInterval(loadAll, 10000);
    return () => clearInterval(interval);
  });

  async function loadAll() {
    loading = true;
    error = null;
    try {
      await Promise.all([loadSensors(), loadActuators()]);
    } catch (err) {
      error = `Failed to load data: ${err.message}`;
      console.error('Error loading data:', err);
    } finally {
      loading = false;
    }
  }

  async function loadSensors() {
    try {
      const result = await sensorAPI.list();
      sensors = Array.isArray(result) ? result : [];
      
      // TODO: Enable when backend implements /sensor-readings/{id}/latest endpoint
      // const readingPromises = sensors.map(async (sensor) => {
      //   const reading = await sensorAPI.getLatest(sensor.id);
      //   if (reading) {
      //     sensorReadings[sensor.id] = reading;
      //   }
      // });
      // await Promise.all(readingPromises);
    } catch (err) {
      console.error('Error loading sensors:', err);
      sensors = [];
    }
  }

  async function loadActuators() {
    try {
      const result = await actuatorAPI.list();
      actuators = Array.isArray(result) ? result : [];
      
      // TODO: Enable when backend implements /actuator-states/{id}/latest endpoint
      // const statePromises = actuators.map(async (actuator) => {
      //   const state = await actuatorAPI.getLatest(actuator.id);
      //   if (state) {
      //     actuatorStates[actuator.id] = state;
      //   }
      // });
      // await Promise.all(statePromises);
    } catch (err) {
      console.error('Error loading actuators:', err);
      actuators = [];
    }
  }

  function formatTimestamp(timestamp) {
    if (!timestamp) return 'N/A';
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now - date;
    const diffSecs = Math.floor(diffMs / 1000);
    const diffMins = Math.floor(diffSecs / 60);
    const diffHours = Math.floor(diffMins / 60);
    
    if (diffSecs < 60) return `${diffSecs}s ago`;
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    return date.toLocaleDateString();
  }

  function getSensorTypeDisplay(type) {
    return type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
  }

  function getActuatorTypeDisplay(type) {
    return type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
  }

  function getSensorIcon(sensorType) {
    const iconMap = {
      'temperature': '🌡️',
      'humidity': '💧',
      'pressure': '🎚️',
      'power': '⚡',
      'energy': '🔋',
      'voltage': '⚡',
      'current': '🔌',
      'motion': '🏃',
      'light': '💡',
      'co2': '💨',
      'gas': '☁️',
      'smoke': '🔥',
      'water': '💧',
      'door': '🚪',
      'window': '🪟',
    };
    
    const type = sensorType?.toLowerCase() || '';
    for (const [key, icon] of Object.entries(iconMap)) {
      if (type.includes(key)) return icon;
    }
    return '📊'; // Default sensor icon
  }

  function getActuatorIcon(actuatorType) {
    const iconMap = {
      'switch': '🔌',
      'relay': '🔌',
      'light': '💡',
      'dimmer': '🔆',
      'motor': '⚙️',
      'valve': '🚰',
      'fan': '💨',
      'heater': '🔥',
      'cooler': '❄️',
      'pump': '⛽',
      'lock': '🔒',
      'door': '🚪',
      'shutter': '🪟',
      'blinds': '🪟',
    };
    
    const type = actuatorType?.toLowerCase() || '';
    for (const [key, icon] of Object.entries(iconMap)) {
      if (type.includes(key)) return icon;
    }
    return '⚡'; // Default actuator icon
  }

  // Sensor edit/delete functions
  function startEditSensor(sensor) {
    editingSensor = { ...sensor };
    showEditSensorForm = true;
  }

  function cancelEditSensor() {
    showEditSensorForm = false;
    editingSensor = null;
    error = null;
  }

  async function updateSensor() {
    if (!editingSensor.name?.trim()) {
      error = 'Sensor name is required';
      return;
    }

    try {
      await sensorAPI.update(editingSensor.device_id, editingSensor.id, {
        name: editingSensor.name,
        sensor_type: editingSensor.sensor_type,
        unit: editingSensor.unit,
        tags: editingSensor.tags || [],
      });
      await loadSensors();
      showEditSensorForm = false;
      editingSensor = null;
      error = null;
    } catch (err) {
      error = `Failed to update sensor: ${err.message}`;
      console.error('Error updating sensor:', err);
    }
  }

  async function deleteSensor(deviceId, sensorId) {
    if (!confirm('Are you sure you want to delete this sensor?')) {
      return;
    }

    try {
      await sensorAPI.delete(deviceId, sensorId);
      await loadSensors();
      error = null;
    } catch (err) {
      error = `Failed to delete sensor: ${err.message}`;
      console.error('Error deleting sensor:', err);
    }
  }

  // Actuator edit/delete functions
  function startEditActuator(actuator) {
    editingActuator = { ...actuator };
    showEditActuatorForm = true;
  }

  function cancelEditActuator() {
    showEditActuatorForm = false;
    editingActuator = null;
    error = null;
  }

  async function updateActuator() {
    if (!editingActuator.name?.trim()) {
      error = 'Actuator name is required';
      return;
    }

    try {
      await actuatorAPI.update(editingActuator.device_id, editingActuator.id, {
        name: editingActuator.name,
        actuator_type: editingActuator.actuator_type,
        tags: editingActuator.tags || [],
      });
      await loadActuators();
      showEditActuatorForm = false;
      editingActuator = null;
      error = null;
    } catch (err) {
      error = `Failed to update actuator: ${err.message}`;
      console.error('Error updating actuator:', err);
    }
  }

  async function deleteActuator(deviceId, actuatorId) {
    if (!confirm('Are you sure you want to delete this actuator?')) {
      return;
    }

    try {
      await actuatorAPI.delete(deviceId, actuatorId);
      await loadActuators();
      error = null;
    } catch (err) {
      error = `Failed to delete actuator: ${err.message}`;
      console.error('Error deleting actuator:', err);
    }
  }

  // Tag management functions
  let newSensorTag = '';
  let newActuatorTag = '';

  function addSensorTag() {
    const tag = newSensorTag.trim();
    if (tag && editingSensor) {
      if (!editingSensor.tags) {
        editingSensor.tags = [];
      }
      if (!editingSensor.tags.includes(tag)) {
        editingSensor.tags = [...editingSensor.tags, tag];
      }
      newSensorTag = '';
    }
  }

  function removeSensorTag(tag) {
    if (editingSensor && editingSensor.tags) {
      editingSensor.tags = editingSensor.tags.filter(t => t !== tag);
    }
  }

  function addActuatorTag() {
    const tag = newActuatorTag.trim();
    if (tag && editingActuator) {
      if (!editingActuator.tags) {
        editingActuator.tags = [];
      }
      if (!editingActuator.tags.includes(tag)) {
        editingActuator.tags = [...editingActuator.tags, tag];
      }
      newActuatorTag = '';
    }
  }

  function removeActuatorTag(tag) {
    if (editingActuator && editingActuator.tags) {
      editingActuator.tags = editingActuator.tags.filter(t => t !== tag);
    }
  }
</script>

<div class="sensor-actuator-list">
  <div class="header">
    <h2>Sensors & Actuators</h2>
    <div class="filter-buttons">
      <button 
        class="filter-btn" 
        class:active={showSensors}
        on:click={() => showSensors = !showSensors}
      >
        📊 Sensors ({sensors.length})
      </button>
      <button 
        class="filter-btn" 
        class:active={showActuators}
        on:click={() => showActuators = !showActuators}
      >
        ⚡ Actuators ({actuators.length})
      </button>
      <button class="btn-refresh" on:click={loadAll} disabled={loading}>
        {loading ? '⟳' : '↻'} Refresh
      </button>
    </div>
  </div>

  {#if error}
    <div class="error-message">{error}</div>
  {/if}

  {#if showEditSensorForm && editingSensor}
    <div class="edit-form-container">
      <h3>Edit Sensor: {editingSensor.id}</h3>
      <form on:submit|preventDefault={updateSensor}>
        <div class="form-row">
          <div class="form-group">
            <label for="edit-sensor-name">Name *</label>
            <input id="edit-sensor-name" bind:value={editingSensor.name} required />
          </div>
          <div class="form-group">
            <label for="edit-sensor-type">Type *</label>
            <input id="edit-sensor-type" bind:value={editingSensor.sensor_type} required />
          </div>
        </div>
        <div class="form-group">
          <label for="edit-sensor-unit">Unit</label>
          <input id="edit-sensor-unit" bind:value={editingSensor.unit} placeholder="°C, %, W, etc." />
        </div>
        <div class="form-group">
          <div class="form-label">Tags</div>
          <div class="tag-manager">
            <div class="tag-list">
              {#if editingSensor.tags && editingSensor.tags.length > 0}
                {#each editingSensor.tags as tag}
                  <span class="tag-edit">
                    {tag}
                    <button type="button" class="tag-remove" on:click={() => removeSensorTag(tag)}>×</button>
                  </span>
                {/each}
              {:else}
                <span class="no-tags">No tags</span>
              {/if}
            </div>
            <div class="tag-input-row">
              <input 
                type="text" 
                bind:value={newSensorTag} 
                placeholder="Add tag..." 
                on:keypress={(e) => e.key === 'Enter' && (e.preventDefault(), addSensorTag())}
              />
              <button type="button" class="btn-add-tag" on:click={addSensorTag}>+ Add</button>
            </div>
          </div>
        </div>
        <div class="form-actions">
          <button type="submit" class="btn btn-primary">Save Changes</button>
          <button type="button" class="btn btn-secondary" on:click={cancelEditSensor}>Cancel</button>
        </div>
      </form>
    </div>
  {/if}

  {#if showEditActuatorForm && editingActuator}
    <div class="edit-form-container">
      <h3>Edit Actuator: {editingActuator.id}</h3>
      <form on:submit|preventDefault={updateActuator}>
        <div class="form-row">
          <div class="form-group">
            <label for="edit-actuator-name">Name *</label>
            <input id="edit-actuator-name" bind:value={editingActuator.name} required />
          </div>
          <div class="form-group">
            <label for="edit-actuator-type">Type *</label>
            <input id="edit-actuator-type" bind:value={editingActuator.actuator_type} required />
          </div>
        </div>
        <div class="form-group">
          <div class="form-label">Tags</div>
          <div class="tag-manager">
            <div class="tag-list">
              {#if editingActuator.tags && editingActuator.tags.length > 0}
                {#each editingActuator.tags as tag}
                  <span class="tag-edit">
                    {tag}
                    <button type="button" class="tag-remove" on:click={() => removeActuatorTag(tag)}>×</button>
                  </span>
                {/each}
              {:else}
                <span class="no-tags">No tags</span>
              {/if}
            </div>
            <div class="tag-input-row">
              <input 
                type="text" 
                bind:value={newActuatorTag} 
                placeholder="Add tag..." 
                on:keypress={(e) => e.key === 'Enter' && (e.preventDefault(), addActuatorTag())}
              />
              <button type="button" class="btn-add-tag" on:click={addActuatorTag}>+ Add</button>
            </div>
          </div>
        </div>
        <div class="form-actions">
          <button type="submit" class="btn btn-primary">Save Changes</button>
          <button type="button" class="btn btn-secondary" on:click={cancelEditActuator}>Cancel</button>
        </div>
      </form>
    </div>
  {/if}

  <div class="items-grid">
    {#if loading && sensors.length === 0 && actuators.length === 0}
      <div class="loading">Loading...</div>
    {:else if sensors.length === 0 && actuators.length === 0}
      <div class="empty-state">No sensors or actuators found.</div>
    {:else}
      {#if showSensors}
        {#each sensors as sensor (`${sensor.device_id || 'unknown'}-${sensor.id}`)}
          <div class="item-card sensor-card">
            <div class="card-header">
              <span class="icon">{getSensorIcon(sensor.sensor_type)}</span>
              <div class="title-section">
                <h3 class="item-name">{sensor.name}</h3>
                <span class="item-type">{getSensorTypeDisplay(sensor.sensor_type)}</span>
              </div>
              <div class="card-actions">
                <button class="btn-icon" on:click={() => startEditSensor(sensor)} title="Edit">✏️</button>
                <button class="btn-icon" on:click={() => deleteSensor(sensor.device_id, sensor.id)} title="Delete">🗑️</button>
              </div>
            </div>
            <div class="card-body">
              <div class="item-id">{sensor.id}</div>
              {#if sensor.device_id}
                <div class="device-ref">Device: {sensor.device_id}</div>
              {/if}
              {#if sensorReadings[sensor.id]}
                <div class="reading-display">
                  <span class="reading-value">
                    {sensorReadings[sensor.id].value}
                    {#if sensorReadings[sensor.id].unit}
                      <span class="unit">{sensorReadings[sensor.id].unit}</span>
                    {/if}
                  </span>
                  <span class="reading-time">{formatTimestamp(sensorReadings[sensor.id].timestamp)}</span>
                  {#if !sensorReadings[sensor.id].valid}
                    <span class="warning">⚠️ Invalid</span>
                  {/if}
                </div>
              {:else}
                <div class="no-data">No recent readings</div>
              {/if}
              {#if sensor.tags && sensor.tags.length > 0}
                <div class="tags">
                  {#each sensor.tags as tag}
                    <span class="tag">{tag}</span>
                  {/each}
                </div>
              {/if}
            </div>
          </div>
        {/each}
      {/if}

      {#if showActuators}
        {#each actuators as actuator (`${actuator.device_id || 'unknown'}-${actuator.id}`)}
          <div class="item-card actuator-card">
            <div class="card-header">
              <span class="icon">{getActuatorIcon(actuator.actuator_type)}</span>
              <div class="title-section">
                <h3 class="item-name">{actuator.name}</h3>
                <span class="item-type">{getActuatorTypeDisplay(actuator.actuator_type)}</span>
              </div>
              <div class="card-actions">
                <button class="btn-icon" on:click={() => startEditActuator(actuator)} title="Edit">✏️</button>
                <button class="btn-icon" on:click={() => deleteActuator(actuator.device_id, actuator.id)} title="Delete">🗑️</button>
              </div>
            </div>
            <div class="card-body">
              <div class="item-id">{actuator.id}</div>
              {#if actuator.device_id}
                <div class="device-ref">Device: {actuator.device_id}</div>
              {/if}
              {#if actuatorStates[actuator.id]}
                <div class="state-display">
                  <span class="state-indicator" class:active={actuatorStates[actuator.id].active}>
                    {actuatorStates[actuator.id].active ? '🟢 Active' : '⚪ Inactive'}
                  </span>
                  <span class="state-time">{formatTimestamp(actuatorStates[actuator.id].timestamp)}</span>
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
              {#if actuator.tags && actuator.tags.length > 0}
                <div class="tags">
                  {#each actuator.tags as tag}
                    <span class="tag">{tag}</span>
                  {/each}
                </div>
              {/if}
            </div>
          </div>
        {/each}
      {/if}
    {/if}
  </div>
</div>

<style>
  .sensor-actuator-list {
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

  h2 {
    margin: 0;
    color: #20B2AA;
    font-size: 1.8rem;
    font-weight: 700;
  }

  .filter-buttons {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .filter-btn {
    padding: 0.5rem 1rem;
    border: 2px solid #20B2AA;
    background: transparent;
    color: #20B2AA;
    border-radius: 8px;
    cursor: pointer;
    font-size: 0.9rem;
    font-weight: 600;
    transition: all 0.2s;
  }

  .filter-btn:hover {
    background: rgba(32, 178, 170, 0.1);
  }

  .filter-btn.active {
    background: #20B2AA;
    color: #0A1929;
  }

  .btn-refresh {
    padding: 0.5rem 1rem;
    border: 1px solid #20B2AA;
    background: #0A1929;
    color: #20B2AA;
    border-radius: 8px;
    cursor: pointer;
    font-size: 1rem;
    transition: all 0.2s;
  }

  .btn-refresh:hover:not(:disabled) {
    background: #20B2AA;
    color: #0A1929;
  }

  .btn-refresh:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .error-message {
    background: rgba(255, 77, 77, 0.1);
    border-left: 4px solid #ff4d4d;
    color: #ff6b6b;
    padding: 1rem;
    border-radius: 4px;
    margin-bottom: 1rem;
  }

  .loading, .empty-state {
    text-align: center;
    padding: 3rem;
    color: #20B2AA;
    font-size: 1.2rem;
  }

  .items-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 1rem;
  }

  .item-card {
    background: linear-gradient(135deg, #1a2332 0%, #0d1520 100%);
    border: 2px solid #20B2AA;
    border-radius: 12px;
    padding: 1rem;
    transition: all 0.3s;
  }

  .item-card:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 16px rgba(32, 178, 170, 0.3);
  }

  .sensor-card {
    border-color: #20B2AA;
  }

  .actuator-card {
    border-color: #FFB347;
  }

  .card-header {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    margin-bottom: 1rem;
  }

  .icon {
    font-size: 2rem;
    line-height: 1;
    flex-shrink: 0;
  }

  .title-section {
    flex: 1;
    min-width: 0;
  }

  .item-name {
    margin: 0;
    color: #20B2AA;
    font-size: 1.1rem;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .actuator-card .item-name {
    color: #FFB347;
  }

  .item-type {
    display: block;
    color: #8BA3B8;
    font-size: 0.85rem;
    margin-top: 0.25rem;
  }

  .card-body {
    color: #D4E4F7;
  }

  .item-id {
    font-size: 0.8rem;
    color: #8BA3B8;
    font-family: monospace;
    margin-bottom: 0.5rem;
  }

  .device-ref {
    font-size: 0.8rem;
    color: #8BA3B8;
    margin-bottom: 0.75rem;
  }

  .reading-display, .state-display {
    background: rgba(32, 178, 170, 0.1);
    padding: 0.75rem;
    border-radius: 8px;
    margin-top: 0.75rem;
  }

  .actuator-card .state-display {
    background: rgba(255, 179, 71, 0.1);
  }

  .reading-value {
    display: block;
    font-size: 1.5rem;
    font-weight: 700;
    color: #20B2AA;
    margin-bottom: 0.25rem;
  }

  .unit {
    font-size: 1rem;
    color: #8BA3B8;
    margin-left: 0.25rem;
  }

  .reading-time, .state-time {
    display: block;
    font-size: 0.75rem;
    color: #8BA3B8;
  }

  .warning {
    display: inline-block;
    margin-left: 0.5rem;
    color: #ff6b6b;
    font-size: 0.85rem;
  }

  .state-indicator {
    display: block;
    font-size: 1.2rem;
    font-weight: 600;
    color: #8BA3B8;
    margin-bottom: 0.25rem;
  }

  .state-indicator.active {
    color: #4CAF50;
  }

  .state-params {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-top: 0.5rem;
  }

  .param {
    background: rgba(255, 255, 255, 0.1);
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.75rem;
    color: #D4E4F7;
  }

  .no-data {
    text-align: center;
    padding: 1rem;
    color: #8BA3B8;
    font-style: italic;
    font-size: 0.9rem;
  }

  .tags {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-top: 0.75rem;
  }

  .tag {
    background: rgba(32, 178, 170, 0.2);
    color: #20B2AA;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.75rem;
    border: 1px solid rgba(32, 178, 170, 0.3);
  }

  .tag-manager {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .tag-list {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    min-height: 2rem;
    align-items: center;
  }

  .tag-edit {
    background: rgba(32, 178, 170, 0.2);
    color: #20B2AA;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.85rem;
    border: 1px solid rgba(32, 178, 170, 0.3);
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
  }

  .tag-remove {
    background: transparent;
    border: none;
    color: #20B2AA;
    font-size: 1.2rem;
    line-height: 1;
    cursor: pointer;
    padding: 0;
    margin: 0;
    opacity: 0.7;
    transition: opacity 0.2s;
  }

  .tag-remove:hover {
    opacity: 1;
  }

  .no-tags {
    color: #8BA3B8;
    font-style: italic;
    font-size: 0.85rem;
  }

  .tag-input-row {
    display: flex;
    gap: 0.5rem;
  }

  .tag-input-row input {
    flex: 1;
  }

  .btn-add-tag {
    padding: 0.5rem 1rem;
    background: rgba(32, 178, 170, 0.2);
    border: 1px solid #20B2AA;
    color: #20B2AA;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.85rem;
    font-weight: 600;
    transition: all 0.2s;
    white-space: nowrap;
  }

  .btn-add-tag:hover {
    background: rgba(32, 178, 170, 0.3);
  }

  .card-actions {
    display: flex;
    gap: 0.25rem;
    margin-left: auto;
  }

  .btn-icon {
    background: transparent;
    border: none;
    font-size: 1.2rem;
    cursor: pointer;
    padding: 0.25rem;
    opacity: 0.7;
    transition: all 0.2s;
    line-height: 1;
  }

  .btn-icon:hover {
    opacity: 1;
    transform: scale(1.1);
  }

  .edit-form-container {
    background: linear-gradient(135deg, #1a2332 0%, #0d1520 100%);
    border: 2px solid #20B2AA;
    border-radius: 12px;
    padding: 1.5rem;
    margin-bottom: 1.5rem;
  }

  .edit-form-container h3 {
    margin: 0 0 1rem 0;
    color: #20B2AA;
    font-size: 1.2rem;
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }

  .form-group {
    margin-bottom: 1rem;
  }

  .form-group label {
    display: block;
    color: #20B2AA;
    font-size: 0.9rem;
    font-weight: 600;
    margin-bottom: 0.5rem;
  }

  .form-label {
    display: block;
    color: #20B2AA;
    font-size: 0.9rem;
    font-weight: 600;
    margin-bottom: 0.5rem;
  }

  .form-group input {
    width: 100%;
    padding: 0.5rem;
    background: rgba(32, 178, 170, 0.05);
    border: 2px solid rgba(32, 178, 170, 0.3);
    border-radius: 4px;
    color: #D4E4F7;
    font-size: 0.9rem;
  }

  .form-group input:focus {
    outline: none;
    border-color: #20B2AA;
  }

  .form-actions {
    display: flex;
    gap: 0.5rem;
    margin-top: 1rem;
  }

  .btn {
    padding: 0.5rem 1.5rem;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    font-size: 0.9rem;
    font-weight: 600;
    transition: all 0.2s;
  }

  .btn-primary {
    background: #20B2AA;
    color: #0A1929;
  }

  .btn-primary:hover {
    background: #1a9b94;
  }

  .btn-secondary {
    background: transparent;
    border: 2px solid #20B2AA;
    color: #20B2AA;
  }

  .btn-secondary:hover {
    background: rgba(32, 178, 170, 0.1);
  }

  @media (max-width: 768px) {
    .items-grid {
      grid-template-columns: 1fr;
    }

    .header {
      flex-direction: column;
      align-items: flex-start;
    }

    .filter-buttons {
      width: 100%;
      justify-content: space-between;
    }

    .form-row {
      grid-template-columns: 1fr;
    }
  }
</style>
