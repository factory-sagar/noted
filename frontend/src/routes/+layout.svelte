<script lang="ts">
  import '../app.css';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
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
    X,
    Building2,
    Loader2,
    Plus,
    ArrowRight,
    Users,
    Trash2
  } from 'lucide-svelte';
  import { onMount } from 'svelte';
  import { api, type SearchResult } from '$lib/utils/api';
  import QuickCapture from '$lib/components/QuickCapture.svelte';
  import CommandPalette from '$lib/components/CommandPalette.svelte';
  import { theme } from '$lib/stores/theme';

  let darkMode = false;
  let sidebarOpen = true;
  let searchOpen = false; // Used for Command Palette now
  let quickCaptureOpen = false;
  let quickCaptureType: 'note' | 'todo' = 'note';
  
  // Removed legacy search state
  // let searchQuery = '';
  // let searchResults: SearchResult[] = [];
  // let searching = false;
  // let searchTimeout: ReturnType<typeof setTimeout>;

  const navItems = [
    { href: '/', label: 'Dashboard', icon: LayoutDashboard },
    { href: '/notes', label: 'Notes', icon: FileText },
    { href: '/todos', label: 'Todos', icon: CheckSquare },
    { href: '/accounts', label: 'Accounts', icon: Building2 },
    { href: '/contacts', label: 'Contacts', icon: Users },
    { href: '/calendar', label: 'Calendar', icon: Calendar },
    { href: '/trash', label: 'Trash', icon: Trash2 },
    { href: '/settings', label: 'Settings', icon: Settings },
  ];

  onMount(() => {
    if (typeof window !== 'undefined') {
      // Dark mode initialization
      darkMode = window.matchMedia('(prefers-color-scheme: dark)').matches;
      const savedDark = localStorage.getItem('darkMode');
      if (savedDark !== null) {
        darkMode = savedDark === 'true';
      }
      updateTheme();

      // Theme initialization
      theme.init();
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
      searchOpen = !searchOpen; // Toggles Command Palette
    }
    if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key === 'c') {
      e.preventDefault();
      quickCaptureOpen = !quickCaptureOpen;
      quickCaptureType = 'note';
    }
  }

  function isActive(href: string, currentPath: string): boolean {
    if (href === '/') return currentPath === '/';
    return currentPath.startsWith(href);
  }
  
  // Legacy search functions removed as they are handled in CommandPalette

</script>

<svelte:window on:keydown={handleKeydown} />

<div class="min-h-screen flex">
  <!-- Sidebar -->
  <aside 
    class="fixed inset-y-0 left-0 z-50 w-72 bg-[var(--color-card)] border-r border-[var(--color-border)] transform transition-transform duration-300 ease-out lg:translate-x-0"
    class:translate-x-0={sidebarOpen}
    class:-translate-x-full={!sidebarOpen}
  >
    <div class="flex flex-col h-full">
      <!-- Logo -->
      <div class="flex items-center justify-between h-20 px-6 border-b border-[var(--color-border)]">
        <a href="/" class="group">
          <h1 class="font-serif text-2xl tracking-tight text-[var(--color-text)] group-hover:text-[var(--color-accent)] transition-colors">
            Noted<span class="text-[var(--color-accent)]">.</span>
          </h1>
        </a>
        <button 
          class="btn-icon btn-icon-ghost lg:hidden"
          on:click={() => sidebarOpen = false}
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Navigation -->
      <nav class="flex-1 px-3 py-8 space-y-1 overflow-y-auto">
        <p class="px-4 mb-4 text-xs font-medium uppercase tracking-wider text-[var(--color-muted)]">
          Menu
        </p>
        {#each navItems as item}
          <a 
            href={item.href}
            class="nav-link {isActive(item.href, $page.url.pathname) ? 'nav-link-active' : 'nav-link-inactive'}"
          >
            <svelte:component this={item.icon} class="w-5 h-5" strokeWidth={1.5} />
            {item.label}
          </a>
        {/each}
      </nav>

      <!-- Theme Toggle -->
      <div class="px-3 py-6 border-t border-[var(--color-border)]">
        <button 
          class="nav-link nav-link-inactive w-full"
          on:click={toggleDarkMode}
        >
          {#if darkMode}
            <Sun class="w-5 h-5" strokeWidth={1.5} />
            <span>Light Mode</span>
          {:else}
            <Moon class="w-5 h-5" strokeWidth={1.5} />
            <span>Dark Mode</span>
          {/if}
        </button>
      </div>
    </div>
  </aside>

  <!-- Main Content -->
  <div class="flex-1 lg:ml-72">
    <!-- Top Bar -->
    <header class="sticky top-0 z-40 h-20 bg-[var(--color-bg)]/80 backdrop-blur-md border-b border-[var(--color-border)]">
      <div class="flex items-center justify-between h-full px-8">
        <button 
          class="btn-icon btn-icon-ghost lg:hidden"
          on:click={() => sidebarOpen = true}
        >
          <Menu class="w-5 h-5" />
        </button>

        <!-- Centered Search Bar -->
        <div class="flex-1 flex items-center justify-center px-4">
          <button 
            class="w-full max-w-xl flex items-center gap-3 px-4 py-2.5 bg-[var(--color-bg-elevated)] border border-[var(--color-border)] hover:border-[var(--color-accent)]/50 transition-colors text-left"
            style="border-radius: var(--radius);"
            on:click={() => searchOpen = true}
          >
            <Search class="w-4 h-4 text-[var(--color-muted)]" strokeWidth={1.5} />
            <span class="flex-1 text-[var(--color-muted)]">Type a command or search...</span>
            <kbd class="hidden sm:inline-flex items-center gap-1 px-1.5 py-0.5 text-xs bg-[var(--color-bg)] border border-[var(--color-border)]" style="border-radius: var(--radius);">
              ⌘K
            </kbd>
          </button>
        </div>

        <div class="flex items-center gap-4">
          <!-- Quick Create -->
          <button 
            class="btn-primary"
            on:click={() => quickCaptureOpen = true}
            title="Quick Create (⌘⇧C)"
          >
            <Plus class="w-4 h-4" strokeWidth={1.5} />
            <span class="hidden sm:inline">New</span>
          </button>
        </div>
      </div>
    </header>

    <!-- Page Content -->
    <main class="p-8 lg:p-12">
      <slot />
    </main>
  </div>
</div>

<!-- Command Palette -->
<CommandPalette 
  bind:open={searchOpen} 
  on:quickCapture={(e) => {
    quickCaptureType = e.detail;
    quickCaptureOpen = true;
  }}
/>

<!-- Quick Capture Modal -->
<QuickCapture 
  bind:open={quickCaptureOpen} 
  type={quickCaptureType}
  on:close={() => quickCaptureOpen = false}
/>
