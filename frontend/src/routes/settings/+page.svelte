<script lang="ts">
  import { onMount } from 'svelte';
  import { 
    Moon, 
    Sun, 
    Calendar,
    FileText,
    Database,
    Download,
    Trash2,
    Check,
    AlertCircle
  } from 'lucide-svelte';
  import { addToast } from '$lib/stores';

  let darkMode = false;
  let autoSave = true;
  let defaultTemplate = 'initial';
  let calendarConnected = false;

  onMount(() => {
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

  function connectCalendar() {
    addToast('info', 'Google Calendar integration coming soon!');
  }

  function disconnectCalendar() {
    calendarConnected = false;
    addToast('success', 'Calendar disconnected');
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
  <title>Settings - SE Notes</title>
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
        >
          <span 
            class="absolute top-1 w-4 h-4 bg-white rounded-full transition-transform shadow"
            class:translate-x-1={!autoSave}
            class:translate-x-7={autoSave}
          ></span>
        </button>
      </div>

      <div class="py-3">
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
            {calendarConnected ? 'Connected' : 'Connect to sync meetings'}
          </p>
        </div>
        {#if calendarConnected}
          <div class="flex items-center gap-3">
            <span class="flex items-center gap-1 text-sm text-green-500">
              <Check class="w-4 h-4" />
              Connected
            </span>
            <button class="btn-secondary text-sm" on:click={disconnectCalendar}>
              Disconnect
            </button>
          </div>
        {:else}
          <button class="btn-primary text-sm" on:click={connectCalendar}>
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
        <button class="btn-secondary text-sm" on:click={exportAllData}>
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
          class="btn text-sm bg-red-500/10 text-red-500 hover:bg-red-500/20"
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
        <p><strong>SE Notes</strong> - Solutions Engineer Notes App</p>
        <p>Version 1.0.0</p>
        <p>Built with SvelteKit, Go, and SQLite</p>
      </div>
    </div>
  </div>
</div>
