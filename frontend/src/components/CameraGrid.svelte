<script>
  export let cameras = [];

  function formatTimestamp(timestamp) {
    if (!timestamp) return 'Never';
    const date = new Date(timestamp);
    return date.toLocaleString();
  }
</script>

<div class="camera-grid">
  {#each cameras as camera (camera.id)}
    <div class="camera-card" class:disabled={!camera.enabled}>
      <div class="camera-header">
        <span class="camera-icon">📹</span>
        <div class="camera-info">
          <h3>{camera.name}</h3>
          {#if camera.location}
            <span class="camera-location">📍 {camera.location}</span>
          {/if}
        </div>
        {#if !camera.enabled}
          <span class="disabled-badge">Offline</span>
        {/if}
      </div>
      
      <div class="camera-body">
        {#if camera.enabled}
          <div class="camera-feed">
            <img src={camera.url} alt={camera.name} on:error={(e) => e.target.src = 'https://via.placeholder.com/320x180?text=Camera+Offline'} />
          </div>
        {:else}
          <div class="camera-placeholder">
            <p>Camera offline</p>
          </div>
        {/if}
        
        <div class="camera-updated">
          Last updated: {formatTimestamp(camera.last_updated)}
        </div>
      </div>
    </div>
  {/each}

  {#if cameras.length === 0}
    <div class="empty-state">
      <p>No cameras configured</p>
    </div>
  {/if}
</div>

<style>
  .camera-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 20px;
  }

  .camera-card {
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(10px);
    border-radius: 12px;
    padding: 20px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    transition: all 0.3s ease;
  }

  .camera-card:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 16px rgba(0, 0, 0, 0.15);
  }

  .camera-card.disabled {
    opacity: 0.7;
  }

  .camera-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 16px;
  }

  .camera-icon {
    font-size: 32px;
  }

  .camera-info {
    flex: 1;
  }

  .camera-info h3 {
    margin: 0 0 4px 0;
    font-size: 16px;
    color: #2d3748;
  }

  .camera-location {
    font-size: 12px;
    color: #718096;
  }

  .disabled-badge {
    background: #e2e8f0;
    color: #4a5568;
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 600;
  }

  .camera-body {
    border-top: 1px solid #e2e8f0;
    padding-top: 16px;
  }

  .camera-feed {
    border-radius: 8px;
    overflow: hidden;
    margin-bottom: 12px;
    background: #000;
  }

  .camera-feed img {
    width: 100%;
    height: auto;
    display: block;
  }

  .camera-placeholder {
    background: #e2e8f0;
    border-radius: 8px;
    padding: 60px 20px;
    text-align: center;
    color: #718096;
    margin-bottom: 12px;
  }

  .camera-updated {
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
