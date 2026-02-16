<script>
  import DeviceManager from './DeviceManager.svelte';
  import WorkflowManager from './WorkflowManager.svelte';

  let activeTab = 'devices';

  function setActiveTab(tab) {
    activeTab = tab;
  }
</script>

<main>
  <header class="app-header">
    <h1>🌱 Life Support System</h1>
    <p class="subtitle">Device & Workflow Management</p>
  </header>

  <nav class="tabs">
    <button 
      class="tab" 
      class:active={activeTab === 'devices'}
      on:click={() => setActiveTab('devices')}
    >
      📱 Devices
    </button>
    <button 
      class="tab" 
      class:active={activeTab === 'workflows'}
      on:click={() => setActiveTab('workflows')}
    >
      🔍 Discovery
    </button>
  </nav>

  <div class="tab-content">
    {#if activeTab === 'devices'}
      <DeviceManager />
    {:else if activeTab === 'workflows'}
      <WorkflowManager />
    {/if}
  </div>
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    min-height: 100vh;
  }

  main {
    max-width: 1400px;
    margin: 0 auto;
    min-height: 100vh;
  }

  .app-header {
    text-align: center;
    padding: 2rem 1rem 1rem 1rem;
    color: white;
  }

  h1 {
    margin: 0;
    font-size: 2.5rem;
    font-weight: 700;
    text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.2);
  }

  .subtitle {
    margin: 0.5rem 0 0 0;
    font-size: 1.1rem;
    opacity: 0.9;
  }

  .tabs {
    display: flex;
    justify-content: center;
    gap: 0.5rem;
    padding: 1rem;
  }

  .tab {
    padding: 0.75rem 2rem;
    border: none;
    background-color: rgba(255, 255, 255, 0.2);
    color: white;
    border-radius: 8px 8px 0 0;
    cursor: pointer;
    font-size: 1rem;
    font-weight: 500;
    transition: all 0.3s;
    backdrop-filter: blur(10px);
  }

  .tab:hover {
    background-color: rgba(255, 255, 255, 0.3);
  }

  .tab.active {
    background-color: white;
    color: #667eea;
    box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.1);
  }

  .tab-content {
    background-color: white;
    border-radius: 12px;
    margin: 0 1rem 1rem 1rem;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
    min-height: calc(100vh - 250px);
  }
</style>
