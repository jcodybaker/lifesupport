<script>
  import { onMount } from 'svelte';
  import { login, isAuthenticated, logout } from './api.js';
  import Dashboard from './components/Dashboard.svelte';
  import Login from './components/Login.svelte';

  let authenticated = isAuthenticated();
  let showLogin = false;

  function handleLogin(event) {
    authenticated = true;
    showLogin = false;
  }

  function handleLogout() {
    logout();
    authenticated = false;
  }

  function toggleLogin() {
    showLogin = !showLogin;
  }
</script>

<main>
  <header>
    <h1>🐠 Life Support System</h1>
    <div class="auth-controls">
      {#if authenticated}
        <span class="admin-badge">Admin Mode</span>
        <button on:click={handleLogout} class="btn-logout">Logout</button>
      {:else}
        <button on:click={toggleLogin} class="btn-login">
          {showLogin ? 'Cancel' : 'Admin Login'}
        </button>
      {/if}
    </div>
  </header>

  {#if showLogin && !authenticated}
    <Login on:login={handleLogin} on:cancel={toggleLogin} />
  {/if}

  <Dashboard {authenticated} />
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: #333;
    min-height: 100vh;
  }

  main {
    max-width: 1400px;
    margin: 0 auto;
    padding: 20px;
  }

  header {
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(10px);
    padding: 20px 30px;
    border-radius: 12px;
    margin-bottom: 20px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  }

  h1 {
    margin: 0;
    font-size: 28px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .auth-controls {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .admin-badge {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    padding: 8px 16px;
    border-radius: 20px;
    font-size: 14px;
    font-weight: 600;
  }

  button {
    padding: 10px 20px;
    border: none;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.3s ease;
    font-size: 14px;
  }

  .btn-login {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
  }

  .btn-login:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
  }

  .btn-logout {
    background: #f56565;
    color: white;
  }

  .btn-logout:hover {
    background: #e53e3e;
    transform: translateY(-2px);
  }
</style>
