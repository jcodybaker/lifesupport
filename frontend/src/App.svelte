<script>
  import SensorActuatorList from './SensorActuatorList.svelte';
  import WorkflowManager from './WorkflowManager.svelte';

  let activeTab = 'sensors';

  function setActiveTab(tab) {
    activeTab = tab;
  }
</script>

<main>
  <header class="app-header">
    <div>
      <h1>LIFE SUPPORT</h1>
      <p class="subtitle">Sensor & Actuator Control</p>
    </div>
  </header>

  <nav class="tabs">
    <button 
      class="tab" 
      class:active={activeTab === 'sensors'}
      on:click={() => setActiveTab('sensors')}
    >
      📊⚡ SENSORS & ACTUATORS
    </button>
    <button 
      class="tab" 
      class:active={activeTab === 'workflows'}
      on:click={() => setActiveTab('workflows')}
    >
      🔍 DISCOVERY
    </button>
  </nav>

  <div class="tab-content">
    {#if activeTab === 'sensors'}
      <SensorActuatorList />
    {:else if activeTab === 'workflows'}
      <WorkflowManager />
    {/if}
  </div>
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    background: #0A1929;
    min-height: 100vh;
  }

  main {
    max-width: 1600px;
    margin: 0 auto;
    min-height: 100vh;
    padding: 0.5rem;
  }

  .app-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem 1.5rem;
    background: #20B2AA;
    border-radius: 0 0 25px 25px;
    margin-bottom: 0.5rem;
    position: relative;
  }

  .app-header::before {
    content: '';
    position: absolute;
    left: 0;
    top: 0;
    width: 20px;
    height: 100%;
    background: #008B8B;
    border-radius: 0 0 0 25px;
  }

  .app-header::after {
    content: '';
    position: absolute;
    right: 0;
    top: 0;
    width: 40px;
    height: 100%;
    background: #40E0D0;
    border-radius: 0 0 25px 0;
  }

  h1 {
    margin: 0;
    font-size: 1.8rem;
    font-weight: 700;
    color: #0A1929;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    z-index: 1;
  }

  .subtitle {
    margin: 0.2rem 0 0 0;
    font-size: 0.75rem;
    color: #0F2537;
    text-transform: uppercase;
    letter-spacing: 0.12em;
    font-weight: 600;
    z-index: 1;
  }

  .tabs {
    display: flex;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    background: #1A2F3F;
    border-radius: 20px;
    margin-bottom: 0.5rem;
  }

  .tab {
    padding: 0.5rem 1.5rem;
    border: none;
    background: #40E0D0;
    color: #0A1929;
    border-radius: 15px;
    cursor: pointer;
    font-size: 0.85rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    transition: all 0.2s;
    position: relative;
    overflow: hidden;
  }

  .tab:hover {
    background: #7FDBFF;
    transform: translateY(-1px);
  }

  .tab.active {
    background: #00CED1;
    color: #0A1929;
    box-shadow: 0 0 15px rgba(0, 206, 209, 0.5);
  }

  .tab-content {
    background: #1A2F3F;
    border-radius: 20px;
    padding: 1rem;
    min-height: calc(100vh - 200px);
    border: 2px solid #66CDAA;
  }
</style>
