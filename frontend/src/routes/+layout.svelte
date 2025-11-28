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
    ArrowRight
  } from 'lucide-svelte';
  import { onMount } from 'svelte';
  import { api, type SearchResult } from '$lib/utils/api';
  import QuickCapture from '$lib/components/QuickCapture.svelte';

  let darkMode = false;
  let sidebarOpen = true;
  let searchOpen = false;
  let quickCaptureOpen = false;
  let searchQuery = '';
  let searchResults: SearchResult[] = [];
  let searching = false;
  let searchTimeout: ReturnType<typeof setTimeout>;

  const navItems = [
    { href: '/', label: 'Dashboard', icon: LayoutDashboard },
    { href: '/notes', label: 'Notes', icon: FileText },
    { href: '/todos', label: 'Todos', icon: CheckSquare },
    { href: '/accounts', label: 'Accounts', icon: Building2 },
    { href: '/calendar', label: 'Calendar', icon: Calendar },
    { href: '/settings', label: 'Settings', icon: Settings },
  ];

  onMount(() => {
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
      if (searchOpen) {
        searchQuery = '';
        searchResults = [];
      }
    }
    if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key === 'c') {
      e.preventDefault();
      quickCaptureOpen = !quickCaptureOpen;
    }
    if (e.key === 'Escape') {
      searchOpen = false;
      quickCaptureOpen = false;
    }
  }

  function isActive(href: string, currentPath: string): boolean {
    if (href === '/') return currentPath === '/';
    return currentPath.startsWith(href);
  }

  async function performSearch(query: string) {
    if (query.length < 2) {
      searchResults = [];
      return;
    }
    
    searching = true;
    try {
      searchResults = await api.search(query);
    } catch (e) {
      console.error('Search failed:', e);
      searchResults = [];
    } finally {
      searching = false;
    }
  }

  function handleSearchInput(e: Event) {
    const value = (e.target as HTMLInputElement).value;
    searchQuery = value;
    
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => performSearch(value), 300);
  }

  function highlightMatch(text: string, query: string): string {
    if (!query || query.length < 2) return text;
    const regex = new RegExp(`(${query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
    return text.replace(regex, '<mark class="bg-primary-200 dark:bg-primary-900/50 px-0.5">$1</mark>');
  }

  function getResultIcon(type: string) {
    switch (type) {
      case 'note': return FileText;
      case 'account': return Building2;
      case 'todo': return CheckSquare;
      default: return FileText;
    }
  }

  function getResultLink(result: SearchResult): string {
    switch (result.type) {
      case 'note': return `/notes/${result.id}`;
      case 'account': return `/accounts`;
      case 'todo': return `/todos`;
      default: return '/';
    }
  }

  function selectResult(result: SearchResult) {
    searchOpen = false;
    goto(getResultLink(result));
  }
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

        <div class="flex-1 flex items-center justify-end gap-4">
          <!-- Search -->
          <button 
            class="btn-secondary"
            on:click={() => searchOpen = true}
          >
            <Search class="w-4 h-4" strokeWidth={1.5} />
            <span class="hidden sm:inline">Search</span>
            <kbd class="hidden sm:inline-flex items-center gap-1 px-1.5 py-0.5 text-xs bg-[var(--color-bg)] border border-[var(--color-border)]" style="border-radius: 2px;">
              ⌘K
            </kbd>
          </button>

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

<!-- Search Modal -->
{#if searchOpen}
  <div class="fixed inset-0 z-[100] flex items-start justify-center pt-[15vh]">
    <button 
      class="modal-backdrop"
      on:click={() => searchOpen = false}
    ></button>
    
    <div class="relative z-[101] w-full max-w-2xl mx-4 bg-[var(--color-card)] border border-[var(--color-border)] shadow-editorial-lg animate-scale-in" style="border-radius: 2px;">
      <div class="flex items-center gap-4 px-6 py-4 border-b border-[var(--color-border)]">
        {#if searching}
          <Loader2 class="w-5 h-5 text-[var(--color-muted)] animate-spin" strokeWidth={1.5} />
        {:else}
          <Search class="w-5 h-5 text-[var(--color-muted)]" strokeWidth={1.5} />
        {/if}
        <input 
          type="text"
          placeholder="Search notes, accounts, todos..."
          class="flex-1 bg-transparent outline-none text-[var(--color-text)] placeholder-[var(--color-muted)] text-lg"
          value={searchQuery}
          on:input={handleSearchInput}
          autofocus
        />
        <kbd class="px-2 py-1 text-xs bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-muted)]" style="border-radius: 2px;">ESC</kbd>
      </div>
      
      <div class="max-h-96 overflow-y-auto">
        {#if searchQuery.length < 2}
          <div class="px-6 py-12 text-center">
            <p class="text-[var(--color-muted)]">Type at least 2 characters to search</p>
          </div>
        {:else if searching}
          <div class="px-6 py-12 text-center">
            <p class="text-[var(--color-muted)]">Searching...</p>
          </div>
        {:else if searchResults.length === 0}
          <div class="px-6 py-12 text-center">
            <p class="text-[var(--color-muted)]">No results found for "{searchQuery}"</p>
          </div>
        {:else}
          <div class="p-2">
            {#each searchResults as result}
              <button
                class="w-full flex items-start gap-4 p-4 hover:bg-[var(--color-card-hover)] transition-colors text-left group"
                on:click={() => selectResult(result)}
              >
                <div class="p-2 bg-[var(--color-bg)] border border-[var(--color-border)]" style="border-radius: 2px;">
                  <svelte:component 
                    this={getResultIcon(result.type)} 
                    class="w-4 h-4 text-[var(--color-muted)]" 
                    strokeWidth={1.5}
                  />
                </div>
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-3 mb-1">
                    <span class="font-medium">
                      {@html highlightMatch(result.title, searchQuery)}
                    </span>
                    <span class="tag-default text-[10px]">
                      {result.type}
                    </span>
                  </div>
                  {#if result.snippet}
                    <p class="text-sm text-[var(--color-muted)] line-clamp-2">
                      {@html result.snippet}
                    </p>
                  {/if}
                </div>
                <ArrowRight class="w-4 h-4 text-[var(--color-muted)] opacity-0 group-hover:opacity-100 transition-opacity" strokeWidth={1.5} />
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<!-- Quick Capture Modal -->
<QuickCapture 
  bind:open={quickCaptureOpen} 
  on:close={() => quickCaptureOpen = false}
/>
