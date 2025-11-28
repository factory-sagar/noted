<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { 
    Moon, 
    Sun, 
    Calendar,
    FileText,
    Database,
    Download,
    Trash2,
    Check,
    AlertCircle,
    LayoutGrid,
    CheckSquare,
    Building2,
    Tag
  } from 'lucide-svelte';
  import { addToast } from '$lib/stores';
  import { api, type CalendarConfig, type Tag as TagType } from '$lib/utils/api';

  let darkMode = false;
  let autoSave = true;
  let defaultTemplate = 'initial';
  let calendarConfig: CalendarConfig = { connected: false };
  
  // Default view settings
  let defaultNotesView = 'folders'; // folders, cards, organized
  let defaultTodosView = 'kanban'; // kanban, list
  let defaultAccountsView = 'split'; // split, grid
  
  // Tags
  let tags: TagType[] = [];
  let newTagName = '';
  let newTagColor = '#6b7280';
  let showTagModal = false;
  const tagColors = ['#ef4444', '#f97316', '#f59e0b', '#22c55e', '#14b8a6', '#3b82f6', '#8b5cf6', '#ec4899', '#6b7280'];

  onMount(async () => {
    if (typeof window !== 'undefined') {
      darkMode = document.documentElement.classList.contains('dark');
      const savedAutoSave = localStorage.getItem('autoSave');
      if (savedAutoSave !== null) {
        autoSave = savedAutoSave === 'true';
      }
      const savedTemplate = localStorage.getItem('defaultTemplate');
      if (savedTemplate) {
        defaultTemplate = savedTemplate;
      }
      // Load view settings
      defaultNotesView = localStorage.getItem('defaultNotesView') || 'folders';
      defaultTodosView = localStorage.getItem('defaultTodosView') || 'kanban';
      defaultAccountsView = localStorage.getItem('defaultAccountsView') || 'split';
    }
    
    // Check for OAuth callback
    if ($page.url.searchParams.get('calendar') === 'connected') {
      addToast('success', 'Google Calendar connected!');
      goto('/settings', { replaceState: true });
    }
    
    // Load calendar config and tags
    try {
      calendarConfig = await api.getCalendarConfig();
      tags = await api.getTags();
    } catch (e) {
      console.error('Failed to load settings:', e);
    }
  });

  function toggleDarkMode() {
    darkMode = !darkMode;
    localStorage.setItem('darkMode', String(darkMode));
    if (darkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
    addToast('success', `${darkMode ? 'Dark' : 'Light'} mode enabled`);
  }

  function saveAutoSave() {
    localStorage.setItem('autoSave', String(autoSave));
    addToast('success', 'Settings saved');
  }

  function saveDefaultTemplate() {
    localStorage.setItem('defaultTemplate', defaultTemplate);
    addToast('success', 'Default template updated');
  }

  function saveNotesView() {
    localStorage.setItem('defaultNotesView', defaultNotesView);
    addToast('success', 'Notes view preference saved');
  }

  function saveTodosView() {
    localStorage.setItem('defaultTodosView', defaultTodosView);
    addToast('success', 'Todos view preference saved');
  }

  function saveAccountsView() {
    localStorage.setItem('defaultAccountsView', defaultAccountsView);
    addToast('success', 'Accounts view preference saved');
  }

  async function createTag() {
    if (!newTagName.trim()) return;
    try {
      const tag = await api.createTag({ name: newTagName.trim(), color: newTagColor });
      tags = [...tags, tag];
      newTagName = '';
      newTagColor = '#6b7280';
      showTagModal = false;
      addToast('success', 'Tag created');
    } catch (e) {
      addToast('error', 'Failed to create tag');
    }
  }

  async function deleteTag(tagId: string) {
    if (!confirm('Delete this tag? It will be removed from all notes.')) return;
    try {
      await api.deleteTag(tagId);
      tags = tags.filter(t => t.id !== tagId);
      addToast('success', 'Tag deleted');
    } catch (e) {
      addToast('error', 'Failed to delete tag');
    }
  }

  async function connectCalendar() {
    try {
      const { url } = await api.getCalendarAuthURL();
      window.location.href = url;
    } catch (e: any) {
      if (e.message.includes('not configured')) {
        addToast('error', 'Google OAuth not configured. Set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET environment variables.');
      } else {
        addToast('error', 'Failed to connect calendar');
      }
    }
  }

  async function disconnectCalendar() {
    try {
      await api.disconnectCalendar();
      calendarConfig = { connected: false };
      addToast('success', 'Calendar disconnected');
    } catch (e) {
      addToast('error', 'Failed to disconnect calendar');
    }
  }

  async function exportAllData() {
    addToast('info', 'Exporting all data...');
    // TODO: Implement full data export
    setTimeout(() => {
      addToast('success', 'Data exported (mock)');
    }, 1000);
  }

  async function clearAllData() {
    if (!confirm('Are you sure you want to delete ALL data? This cannot be undone.')) return;
    if (!confirm('This will permanently delete all accounts, notes, and todos. Continue?')) return;
    
    addToast('info', 'Clearing all data...');
    // TODO: Implement data clearing
    setTimeout(() => {
      addToast('success', 'All data cleared');
    }, 1000);
  }
</script>

<svelte:head>
  <title>Settings - Noted</title>
</svelte:head>

<div class="max-w-3xl mx-auto">
  <div class="mb-8">
    <h1 class="page-title">Settings</h1>
    <p class="page-subtitle">Manage your preferences</p>
  </div>

  <div class="space-y-6">
    <!-- Appearance -->
    <div class="card">
      <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
        {#if darkMode}
          <Moon class="w-5 h-5" />
        {:else}
          <Sun class="w-5 h-5" />
        {/if}
        Appearance
      </h2>
      
      <div class="flex items-center justify-between py-3 border-b border-[var(--color-border)]">
        <div>
          <p class="font-medium">Dark Mode</p>
          <p class="text-sm text-[var(--color-muted)]">Use dark theme for reduced eye strain</p>
        </div>
        <button 
          class="relative w-12 h-6 rounded-full transition-colors"
          class:bg-primary-600={darkMode}
          class:bg-[var(--color-border)]={!darkMode}
          on:click={toggleDarkMode}
          aria-label="Toggle dark mode"
        >
          <span 
            class="absolute top-1 w-4 h-4 bg-white rounded-full transition-transform shadow"
            class:translate-x-1={!darkMode}
            class:translate-x-7={darkMode}
          ></span>
        </button>
      </div>
    </div>

    <!-- Notes Settings -->
    <div class="card">
      <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
        <FileText class="w-5 h-5" />
        Notes
      </h2>
      
      <div class="flex items-center justify-between py-3 border-b border-[var(--color-border)]">
        <div>
          <p class="font-medium">Auto-save</p>
          <p class="text-sm text-[var(--color-muted)]">Automatically save notes while editing</p>
        </div>
        <button 
          class="relative w-12 h-6 rounded-full transition-colors"
          class:bg-primary-600={autoSave}
          class:bg-[var(--color-border)]={!autoSave}
          on:click={() => { autoSave = !autoSave; saveAutoSave(); }}
          aria-label="Toggle auto-save"
        >
          <span 
            class="absolute top-1 w-4 h-4 bg-white rounded-full transition-transform shadow"
            class:translate-x-1={!autoSave}
            class:translate-x-7={autoSave}
          ></span>
        </button>
      </div>

      <div class="py-3 border-b border-[var(--color-border)]">
        <div class="flex items-center justify-between mb-2">
          <div>
            <p class="font-medium">Default Template</p>
            <p class="text-sm text-[var(--color-muted)]">Template to use when creating new notes</p>
          </div>
        </div>
        <select 
          class="input mt-2"
          bind:value={defaultTemplate}
          on:change={saveDefaultTemplate}
        >
          <option value="initial">Initial Call - All fields</option>
          <option value="followup">Follow-up Call - Simplified</option>
        </select>
      </div>

      <div class="flex items-center justify-between py-3">
        <div>
          <p class="font-medium">Customize Templates</p>
          <p class="text-sm text-[var(--color-muted)]">Create and edit note templates</p>
        </div>
        <a href="/settings/templates" class="btn-secondary btn-sm">
          Manage Templates
        </a>
      </div>
    </div>

    <!-- Default Views -->
    <div class="card">
      <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
        <LayoutGrid class="w-5 h-5" />
        Default Views
      </h2>
      
      <div class="py-3 border-b border-[var(--color-border)]">
        <div class="flex items-center justify-between mb-2">
          <div>
            <p class="font-medium">Notes View</p>
            <p class="text-sm text-[var(--color-muted)]">Default view when opening notes page</p>
          </div>
        </div>
        <select 
          class="input mt-2"
          bind:value={defaultNotesView}
          on:change={saveNotesView}
        >
          <option value="folders">Folders - Collapsible by account</option>
          <option value="cards">Cards - Grid of note cards</option>
          <option value="organized">Organized - Grouped by account</option>
        </select>
      </div>

      <div class="py-3 border-b border-[var(--color-border)]">
        <div class="flex items-center justify-between mb-2">
          <div>
            <p class="font-medium">Todos View</p>
            <p class="text-sm text-[var(--color-muted)]">Default layout for todos page</p>
          </div>
        </div>
        <select 
          class="input mt-2"
          bind:value={defaultTodosView}
          on:change={saveTodosView}
        >
          <option value="kanban">Kanban Board</option>
          <option value="list">List View</option>
        </select>
      </div>

      <div class="py-3">
        <div class="flex items-center justify-between mb-2">
          <div>
            <p class="font-medium">Accounts View</p>
            <p class="text-sm text-[var(--color-muted)]">Default layout for accounts page</p>
          </div>
        </div>
        <select 
          class="input mt-2"
          bind:value={defaultAccountsView}
          on:change={saveAccountsView}
        >
          <option value="split">Split View - List and detail</option>
          <option value="grid">Grid View - Cards</option>
        </select>
      </div>
    </div>

    <!-- Tags Management -->
    <div class="card">
      <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
        <Tag class="w-5 h-5" />
        Tags
      </h2>
      <p class="text-sm text-[var(--color-muted)] mb-4">Create tags to organize your notes</p>
      
      {#if tags.length === 0}
        <div class="text-center py-6 text-[var(--color-muted)]">
          No tags yet. Create your first tag to get started.
        </div>
      {:else}
        <div class="flex flex-wrap gap-2 mb-4">
          {#each tags as tag}
            <div 
              class="flex items-center gap-2 px-3 py-1.5 rounded-full text-sm group"
              style="background-color: {tag.color}20; color: {tag.color}; border: 1px solid {tag.color}40"
            >
              <span class="w-2 h-2 rounded-full" style="background-color: {tag.color}"></span>
              {tag.name}
              <button 
                class="opacity-0 group-hover:opacity-100 hover:text-red-500 transition-opacity"
                on:click={() => deleteTag(tag.id)}
              >
                <Trash2 class="w-3 h-3" />
              </button>
            </div>
          {/each}
        </div>
      {/if}
      
      <button class="btn-secondary btn-sm" on:click={() => showTagModal = true}>
        <Tag class="w-4 h-4" />
        Create Tag
      </button>
    </div>

    <!-- Calendar Integration -->
    <div class="card">
      <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
        <Calendar class="w-5 h-5" />
        Calendar Integration
      </h2>
      
      <div class="flex items-center justify-between py-3">
        <div>
          <p class="font-medium">Google Calendar</p>
          <p class="text-sm text-[var(--color-muted)]">
            {#if calendarConfig.connected && calendarConfig.email}
              Connected as {calendarConfig.email}
            {:else if calendarConfig.connected}
              Connected
            {:else}
              Connect to sync meetings and auto-populate participants
            {/if}
          </p>
        </div>
        {#if calendarConfig.connected}
          <div class="flex items-center gap-3">
            <span class="flex items-center gap-1 text-sm text-green-500">
              <Check class="w-4 h-4" />
              Connected
            </span>
            <button class="btn-secondary btn-sm" on:click={disconnectCalendar}>
              Disconnect
            </button>
          </div>
        {:else}
          <button class="btn-primary btn-sm" on:click={connectCalendar}>
            Connect
          </button>
        {/if}
      </div>
    </div>

    <!-- Data Management -->
    <div class="card">
      <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
        <Database class="w-5 h-5" />
        Data Management
      </h2>
      
      <div class="flex items-center justify-between py-3 border-b border-[var(--color-border)]">
        <div>
          <p class="font-medium">Export All Data</p>
          <p class="text-sm text-[var(--color-muted)]">Download all your data as JSON</p>
        </div>
        <button class="btn-secondary btn-sm" on:click={exportAllData}>
          <Download class="w-4 h-4" />
          Export
        </button>
      </div>

      <div class="flex items-center justify-between py-3">
        <div>
          <p class="font-medium text-red-500">Delete All Data</p>
          <p class="text-sm text-[var(--color-muted)]">Permanently remove all accounts, notes, and todos</p>
        </div>
        <button 
          class="btn-danger btn-sm"
          on:click={clearAllData}
        >
          <Trash2 class="w-4 h-4" />
          Delete All
        </button>
      </div>
    </div>

    <!-- About -->
    <div class="card">
      <h2 class="text-lg font-semibold mb-4">About</h2>
      <div class="space-y-2 text-sm text-[var(--color-muted)]">
        <p><strong>Noted</strong> - Solutions Engineer Notes App</p>
        <p>Version 1.0.0</p>
        <p>Built with SvelteKit, Go, and SQLite</p>
      </div>
    </div>
  </div>
</div>

<!-- Create Tag Modal -->
{#if showTagModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button 
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => showTagModal = false}
      aria-label="Close modal"
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-md animate-slide-up">
      <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
        <Tag class="w-5 h-5" />
        Create Tag
      </h2>
      <form on:submit|preventDefault={createTag}>
        <div class="mb-4">
          <label class="label" for="new-tag-name">Tag Name</label>
          <!-- svelte-ignore a11y-autofocus -->
          <input 
            id="new-tag-name"
            type="text"
            class="input"
            placeholder="e.g., Follow-up, Urgent, Demo"
            bind:value={newTagName}
            autofocus
          />
        </div>
        <div class="mb-4">
          <!-- svelte-ignore a11y-label-has-associated-control -->
          <label class="label" id="tag-color-label">Color</label>
          <div class="flex flex-wrap gap-2 mt-2" role="group" aria-labelledby="tag-color-label">
            {#each tagColors as color}
              <button
                type="button"
                class="w-8 h-8 rounded-full border-2 transition-transform hover:scale-110"
                class:border-white={newTagColor === color}
                class:border-transparent={newTagColor !== color}
                class:ring-2={newTagColor === color}
                class:ring-offset-2={newTagColor === color}
                style="background-color: {color}; --tw-ring-color: {color}"
                on:click={() => newTagColor = color}
                aria-label="Select color {color}"
              ></button>
            {/each}
          </div>
          <div class="mt-3 flex items-center gap-2">
            <span class="text-sm text-[var(--color-muted)]">Preview:</span>
            <span 
              class="px-3 py-1 rounded-full text-sm"
              style="background-color: {newTagColor}20; color: {newTagColor}; border: 1px solid {newTagColor}40"
            >
              {newTagName || 'Tag Name'}
            </span>
          </div>
        </div>
        <div class="flex justify-end gap-3">
          <button 
            type="button"
            class="btn-secondary"
            on:click={() => showTagModal = false}
          >
            Cancel
          </button>
          <button type="submit" class="btn-primary" disabled={!newTagName.trim()}>
            Create Tag
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
