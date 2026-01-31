<script>
  import { onMount } from 'svelte';

  let items = [];
  let newItem = { name: '', description: '' };
  let editingItem = null;

  const API_URL = '/api/items';

  async function loadItems() {
    try {
      const response = await fetch(API_URL);
      items = await response.json() || [];
    } catch (error) {
      console.error('Error loading items:', error);
    }
  }

  async function createItem() {
    if (!newItem.name.trim()) return;

    try {
      const response = await fetch(API_URL, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newItem)
      });

      if (response.ok) {
        newItem = { name: '', description: '' };
        await loadItems();
      }
    } catch (error) {
      console.error('Error creating item:', error);
    }
  }

  async function updateItem() {
    if (!editingItem || !editingItem.name.trim()) return;

    try {
      const response = await fetch(`${API_URL}/${editingItem.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(editingItem)
      });

      if (response.ok) {
        editingItem = null;
        await loadItems();
      }
    } catch (error) {
      console.error('Error updating item:', error);
    }
  }

  async function deleteItem(id) {
    try {
      const response = await fetch(`${API_URL}/${id}`, {
        method: 'DELETE'
      });

      if (response.ok) {
        await loadItems();
      }
    } catch (error) {
      console.error('Error deleting item:', error);
    }
  }

  function startEdit(item) {
    editingItem = { ...item };
  }

  function cancelEdit() {
    editingItem = null;
  }

  onMount(loadItems);
</script>

<main>
  <h1>LifeSupport App</h1>

  <section class="add-section">
    <h2>Add New Item</h2>
    <form on:submit|preventDefault={createItem}>
      <input
        type="text"
        placeholder="Name"
        bind:value={newItem.name}
        required
      />
      <input
        type="text"
        placeholder="Description"
        bind:value={newItem.description}
      />
      <button type="submit">Add Item</button>
    </form>
  </section>

  <section class="items-section">
    <h2>Items</h2>
    {#if items.length === 0}
      <p>No items yet. Add one above!</p>
    {:else}
      <ul>
        {#each items as item (item.id)}
          <li>
            {#if editingItem && editingItem.id === item.id}
              <form on:submit|preventDefault={updateItem} class="edit-form">
                <input
                  type="text"
                  bind:value={editingItem.name}
                  required
                />
                <input
                  type="text"
                  bind:value={editingItem.description}
                />
                <button type="submit">Save</button>
                <button type="button" on:click={cancelEdit}>Cancel</button>
              </form>
            {:else}
              <div class="item-content">
                <div>
                  <strong>{item.name}</strong>
                  {#if item.description}
                    <p>{item.description}</p>
                  {/if}
                </div>
                <div class="item-actions">
                  <button on:click={() => startEdit(item)}>Edit</button>
                  <button class="delete" on:click={() => deleteItem(item.id)}>Delete</button>
                </div>
              </div>
            {/if}
          </li>
        {/each}
      </ul>
    {/if}
  </section>
</main>

<style>
  main {
    max-width: 800px;
    margin: 0 auto;
    padding: 2rem;
    font-family: Arial, sans-serif;
  }

  h1 {
    text-align: center;
    color: #333;
  }

  section {
    margin: 2rem 0;
  }

  h2 {
    color: #555;
    border-bottom: 2px solid #eee;
    padding-bottom: 0.5rem;
  }

  form {
    display: flex;
    gap: 0.5rem;
    margin-top: 1rem;
  }

  input {
    flex: 1;
    padding: 0.5rem;
    border: 1px solid #ddd;
    border-radius: 4px;
  }

  button {
    padding: 0.5rem 1rem;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    background-color: #4CAF50;
    color: white;
  }

  button:hover {
    background-color: #45a049;
  }

  button.delete {
    background-color: #f44336;
  }

  button.delete:hover {
    background-color: #da190b;
  }

  ul {
    list-style: none;
    padding: 0;
  }

  li {
    background: #f9f9f9;
    margin: 0.5rem 0;
    padding: 1rem;
    border-radius: 4px;
    border: 1px solid #eee;
  }

  .item-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .item-actions {
    display: flex;
    gap: 0.5rem;
  }

  .edit-form {
    display: flex;
    gap: 0.5rem;
  }

  p {
    margin: 0.25rem 0 0 0;
    color: #666;
    font-size: 0.9rem;
  }
</style>
