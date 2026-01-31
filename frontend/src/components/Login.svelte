<script>
  import { createEventDispatcher } from 'svelte';
  import { login } from '../api.js';

  const dispatch = createEventDispatcher();

  let username = '';
  let password = '';
  let error = '';
  let loading = false;

  async function handleSubmit() {
    error = '';
    loading = true;

    try {
      await login(username, password);
      dispatch('login');
    } catch (err) {
      error = err.message || 'Login failed';
    } finally {
      loading = false;
    }
  }
</script>

<div class="login-overlay">
  <div class="login-card">
    <h2>Admin Login</h2>
    <form on:submit|preventDefault={handleSubmit}>
      <div class="form-group">
        <label for="username">Username</label>
        <input
          id="username"
          type="text"
          bind:value={username}
          placeholder="Enter username"
          required
          disabled={loading}
        />
      </div>

      <div class="form-group">
        <label for="password">Password</label>
        <input
          id="password"
          type="password"
          bind:value={password}
          placeholder="Enter password"
          required
          disabled={loading}
        />
      </div>

      {#if error}
        <div class="error">{error}</div>
      {/if}

      <div class="button-group">
        <button type="submit" class="btn-submit" disabled={loading}>
          {loading ? 'Logging in...' : 'Login'}
        </button>
        <button type="button" class="btn-cancel" on:click={() => dispatch('cancel')} disabled={loading}>
          Cancel
        </button>
      </div>
    </form>
  </div>
</div>

<style>
  .login-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .login-card {
    background: white;
    padding: 40px;
    border-radius: 12px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
    max-width: 400px;
    width: 90%;
  }

  h2 {
    margin-top: 0;
    margin-bottom: 24px;
    color: #333;
    text-align: center;
  }

  .form-group {
    margin-bottom: 20px;
  }

  label {
    display: block;
    margin-bottom: 8px;
    font-weight: 600;
    color: #555;
  }

  input {
    width: 100%;
    padding: 12px;
    border: 2px solid #e2e8f0;
    border-radius: 8px;
    font-size: 14px;
    transition: border-color 0.3s;
    box-sizing: border-box;
  }

  input:focus {
    outline: none;
    border-color: #667eea;
  }

  input:disabled {
    background: #f7fafc;
    cursor: not-allowed;
  }

  .error {
    background: #fed7d7;
    color: #c53030;
    padding: 12px;
    border-radius: 8px;
    margin-bottom: 20px;
    font-size: 14px;
  }

  .button-group {
    display: flex;
    gap: 12px;
  }

  button {
    flex: 1;
    padding: 12px 24px;
    border: none;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.3s ease;
    font-size: 14px;
  }

  .btn-submit {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
  }

  .btn-submit:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
  }

  .btn-submit:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .btn-cancel {
    background: #e2e8f0;
    color: #4a5568;
  }

  .btn-cancel:hover:not(:disabled) {
    background: #cbd5e0;
  }
</style>
