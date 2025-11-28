<script lang="ts">
  import '../app.css';
  import { page } from '$app/stores';
  import { 
    LayoutDashboard, 
    FileText, 
    CheckSquare, 
    Calendar, 
    Settings,
    Search,
    Moon,
    Sun,
    Menu,
    X
  } from 'lucide-svelte';
  import { onMount } from 'svelte';

  let darkMode = false;
  let sidebarOpen = true;
  let searchOpen = false;
  let searchQuery = '';

  const navItems = [
    { href: '/', label: 'Dashboard', icon: LayoutDashboard },
    { href: '/notes', label: 'Notes', icon: FileText },
    { href: '/todos', label: 'Todos', icon: CheckSquare },
    { href: '/calendar', label: 'Calendar', icon: Calendar },
    { href: '/settings', label: 'Settings', icon: Settings },
  ];

  onMount(() => {
    // Check system preference
    if (typeof window !== 'undefined') {
      darkMode = window.matchMedia('(prefers-color-scheme: dark)').matches;
      const saved = localStorage.getItem('darkMode');
      if (saved !== null) {
        darkMode = saved === 'true';
      }
      updateTheme();
    }
  });

  function toggleDarkMode() {
    darkMode = !darkMode;
    localStorage.setItem('darkMode', String(darkMode));
    updateTheme();
  }

  function updateTheme() {
    if (darkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault();
      searchOpen = !searchOpen;
    }
    if (e.key === 'Escape') {
      searchOpen = false;
    }
  }

  function isActive(href: string, currentPath: string): boolean {
    if (href === '/') return currentPath === '/';
    return currentPath.startsWith(href);
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="min-h-screen flex">
  <!-- Sidebar -->
  <aside 
    class="fixed inset-y-0 left-0 z-50 w-64 bg-[var(--color-card)] border-r border-[var(--color-border)] transform transition-transform duration-300 ease-in-out lg:translate-x-0"
    class:translate-x-0={sidebarOpen}
    class:-translate-x-full={!sidebarOpen}
  >
    <div class="flex flex-col h-full">
      <!-- Logo -->
      <div class="flex items-center justify-between h-16 px-6 border-b border-[var(--color-border)]">
        <a href="/" class="flex items-center gap-2">
          <div class="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center">
            <FileText class="w-5 h-5 text-white" />
          </div>
          <span class="text-lg font-semibold">SE Notes</span>
        </a>
        <button 
          class="lg:hidden p-1 hover:bg-[var(--color-border)] rounded"
          on:click={() => sidebarOpen = false}
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Navigation -->
      <nav class="flex-1 px-4 py-6 space-y-1 overflow-y-auto">
        {#each navItems as item}
          <a 
            href={item.href}
            class="sidebar-link {isActive(item.href, $page.url.pathname) ? 'sidebar-link-active' : 'sidebar-link-inactive'}"
          >
            <svelte:component this={item.icon} class="w-5 h-5" />
            {item.label}
          </a>
        {/each}
      </nav>

      <!-- Theme Toggle -->
      <div class="px-4 py-4 border-t border-[var(--color-border)]">
        <button 
          class="sidebar-link sidebar-link-inactive w-full"
          on:click={toggleDarkMode}
        >
          {#if darkMode}
            <Sun class="w-5 h-5" />
            Light Mode
          {:else}
            <Moon class="w-5 h-5" />
            Dark Mode
          {/if}
        </button>
      </div>
    </div>
  </aside>

  <!-- Main Content -->
  <div class="flex-1 lg:ml-64">
    <!-- Top Bar -->
    <header class="sticky top-0 z-40 h-16 bg-[var(--color-bg)]/80 backdrop-blur-sm border-b border-[var(--color-border)]">
      <div class="flex items-center justify-between h-full px-6">
        <div class="flex items-center gap-4">
          <button 
            class="lg:hidden p-2 hover:bg-[var(--color-card)] rounded-lg"
            on:click={() => sidebarOpen = true}
          >
            <Menu class="w-5 h-5" />
          </button>
        </div>

        <!-- Search -->
        <button 
          class="flex items-center gap-2 px-4 py-2 bg-[var(--color-card)] border border-[var(--color-border)] rounded-lg text-[var(--color-muted)] hover:border-primary-500 transition-colors"
          on:click={() => searchOpen = true}
        >
          <Search class="w-4 h-4" />
          <span class="text-sm">Search...</span>
          <kbd class="hidden sm:inline-flex items-center gap-1 px-2 py-0.5 text-xs bg-[var(--color-border)] rounded">
            ⌘K
          </kbd>
        </button>

        <div class="w-20"></div>
      </div>
    </header>

    <!-- Page Content -->
    <main class="p-6 animate-fade-in">
      <slot />
    </main>
  </div>
</div>

<!-- Search Modal -->
{#if searchOpen}
  <div class="fixed inset-0 z-[100] flex items-start justify-center pt-[20vh]">
    <!-- Backdrop -->
    <button 
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => searchOpen = false}
    ></button>
    
    <!-- Search Box -->
    <div class="relative w-full max-w-xl mx-4 bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl shadow-2xl animate-slide-up">
      <div class="flex items-center gap-3 px-4 py-3 border-b border-[var(--color-border)]">
        <Search class="w-5 h-5 text-[var(--color-muted)]" />
        <input 
          type="text"
          placeholder="Search notes, accounts, todos..."
          class="flex-1 bg-transparent outline-none text-[var(--color-text)] placeholder-[var(--color-muted)]"
          bind:value={searchQuery}
          autofocus
        />
        <kbd class="px-2 py-1 text-xs bg-[var(--color-border)] text-[var(--color-muted)] rounded">ESC</kbd>
      </div>
      
      <div class="max-h-80 overflow-y-auto p-2">
        {#if searchQuery.length === 0}
          <p class="px-4 py-8 text-center text-[var(--color-muted)]">
            Start typing to search...
          </p>
        {:else}
          <p class="px-4 py-8 text-center text-[var(--color-muted)]">
            No results found for "{searchQuery}"
          </p>
        {/if}
      </div>
    </div>
  </div>
{/if}
