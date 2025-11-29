<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { fade, fly } from 'svelte/transition';
  import { Search, FileText, CheckSquare, User, Settings, Calendar, Plus } from 'lucide-svelte';
  import { api, type SearchResult } from '$lib/utils/api';

  export let open = false;

  let query = '';
  let selectedIndex = 0;
  let inputElement: HTMLInputElement;
  let searchResults: SearchResult[] = [];
  let searching = false;
  let searchTimeout: ReturnType<typeof setTimeout>;

  // Command groups
  const staticGroups = [
    {
      name: 'Navigation',
      items: [
        { icon: FileText, label: 'Go to Notes', action: () => goto('/notes') },
        { icon: CheckSquare, label: 'Go to Todos', action: () => goto('/todos') },
        { icon: User, label: 'Go to Accounts', action: () => goto('/accounts') },
        { icon: Calendar, label: 'Go to Calendar', action: () => goto('/calendar') },
        { icon: Settings, label: 'Go to Settings', action: () => goto('/settings') }
      ]
    },
    {
      name: 'Actions',
      items: [
        { icon: Plus, label: 'New Note', action: () => dispatch('quickCapture', 'note') },
        { icon: Plus, label: 'New Todo', action: () => dispatch('quickCapture', 'todo') }
      ]
    }
  ];

  $: if (open && inputElement) {
    setTimeout(() => {
      if (inputElement) inputElement.focus();
    }, 10);
  }
  
  $: if (!open) {
    query = '';
    selectedIndex = 0;
    searchResults = [];
  }

  async function performSearch(q: string) {
    if (q.length < 2) {
      searchResults = [];
      return;
    }
    
    searching = true;
    try {
      const results = await api.search(q);
      searchResults = results;
    } catch (e) {
      console.error('Search failed:', e);
      searchResults = [];
    } finally {
      searching = false;
    }
  }

  function handleInput(e: Event) {
    const value = (e.target as HTMLInputElement).value;
    query = value;
    selectedIndex = 0;
    
    clearTimeout(searchTimeout);
    if (value.length >= 2) {
      searchTimeout = setTimeout(() => performSearch(value), 300);
    } else {
      searchResults = [];
    }
  }

  // Combined items (static + search results)
  $: filteredStaticGroups = !query || query.length < 2
    ? staticGroups 
    : staticGroups.map(g => ({
        ...g,
        items: g.items.filter(i => i.label.toLowerCase().includes(query.toLowerCase()))
      })).filter(g => g.items.length > 0);

  $: searchResultGroup = searchResults.length > 0 ? [{
    name: 'Search Results',
    items: searchResults.map(r => ({
      icon: getResultIcon(r.type),
      label: r.title,
      action: () => goto(getResultLink(r)),
      description: r.snippet
    }))
  }] : [];

  $: allGroups = [...filteredStaticGroups, ...searchResultGroup];
  $: flatItems = allGroups.flatMap(g => g.items);

  function getResultIcon(type: string) {
    switch (type) {
      case 'note': return FileText;
      case 'account': return User;
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

  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  function handleKeydown(e: KeyboardEvent) {
    if (!open) return;

    if (e.key === 'Escape') {
      close();
      return;
    }

    if (e.key === 'ArrowDown') {
      e.preventDefault();
      selectedIndex = (selectedIndex + 1) % flatItems.length;
      // scrollIntoView(selectedIndex);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      selectedIndex = (selectedIndex - 1 + flatItems.length) % flatItems.length;
      // scrollIntoView(selectedIndex);
    } else if (e.key === 'Enter') {
      e.preventDefault();
      executeAction(flatItems[selectedIndex]);
    }
  }

  function executeAction(item: any) {
    if (item) {
      item.action();
      close();
    }
  }

  function close() {
    open = false;
  }

  onMount(() => {
    window.addEventListener('keydown', handleKeydown);
    return () => window.removeEventListener('keydown', handleKeydown);
  });
</script>

{#if open}
  <div class="fixed inset-0 z-50 flex items-start justify-center pt-[20vh] bg-black/50 backdrop-blur-sm" 
       transition:fade={{ duration: 150 }}
       on:click={close}
  >
    <div class="w-full max-w-xl bg-white dark:bg-gray-800 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-700 overflow-hidden flex flex-col max-h-[60vh]" 
         transition:fly={{ y: -20, duration: 200 }}
         on:click|stopPropagation
    >
      <!-- Search Input -->
      <div class="flex items-center px-4 py-3 border-b border-gray-200 dark:border-gray-700">
        <Search class="w-5 h-5 text-gray-400 mr-3" />
        <input 
          bind:this={inputElement}
          value={query}
          on:input={handleInput}
          type="text" 
          placeholder="Type a command or search..." 
          class="flex-1 bg-transparent border-none outline-none text-lg text-gray-900 dark:text-white placeholder-gray-400"
        />
        <div class="text-xs text-gray-400 border border-gray-200 dark:border-gray-700 rounded px-1.5 py-0.5">ESC</div>
      </div>

      <!-- Results List -->
      <div class="overflow-y-auto py-2">
        {#if flatItems.length === 0}
          <div class="px-4 py-8 text-center text-gray-500">
            {#if searching}
              Searching...
            {:else}
              No results found.
            {/if}
          </div>
        {:else}
          {#each allGroups as group}
            <div class="px-2 mb-2">
              <div class="px-2 py-1 text-xs font-medium text-gray-400 uppercase tracking-wider">
                {group.name}
              </div>
              {#each group.items as item}
                {@const index = flatItems.indexOf(item)}
                <button
                  class="w-full flex items-center px-3 py-2 rounded-lg text-left transition-colors
                         {index === selectedIndex ? 'bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300' : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700/50'}"
                  on:click={() => executeAction(item)}
                  on:mouseenter={() => selectedIndex = index}
                >
                  <svelte:component this={item.icon} class="w-4 h-4 mr-3 opacity-70 shrink-0" />
                  <div class="min-w-0 flex-1">
                    <div class="truncate">{item.label}</div>
                    {#if item.description}
                      <div class="text-xs text-gray-500 truncate opacity-70">{@html item.description}</div>
                    {/if}
                  </div>
                  {#if index === selectedIndex}
                    <span class="ml-auto text-xs opacity-50 shrink-0">↵</span>
                  {/if}
                </button>
              {/each}
            </div>
          {/each}
        {/if}
      </div>

      <!-- Footer -->
      <div class="bg-gray-50 dark:bg-gray-800/50 px-4 py-2 border-t border-gray-200 dark:border-gray-700 text-xs text-gray-500 flex justify-between items-center">
        <div class="flex gap-3">
          <span><kbd class="font-sans bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded px-1">↑↓</kbd> to navigate</span>
          <span><kbd class="font-sans bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded px-1">↵</kbd> to select</span>
        </div>
        <div class="flex items-center gap-1">
          <span class="w-2 h-2 rounded-full bg-green-500"></span>
          <span>Noted Command</span>
        </div>
      </div>
    </div>
  </div>
{/if}
