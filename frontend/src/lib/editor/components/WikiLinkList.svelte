<script lang="ts">
  import { onMount } from 'svelte';
  import { api, type SearchResult } from '$lib/utils/api';
  import { FileText, Building2, CheckSquare } from 'lucide-svelte';

  export let items: any[] = [];
  export let command: any;
  
  let selectedIndex = 0;
  let element: HTMLElement;

  $: if (items) {
    selectedIndex = 0;
  }

  export function onKeyDown({ event }: { event: KeyboardEvent }) {
    if (event.key === 'ArrowUp') {
      upHandler();
      return true;
    }

    if (event.key === 'ArrowDown') {
      downHandler();
      return true;
    }

    if (event.key === 'Enter') {
      enterHandler();
      return true;
    }

    return false;
  }

  function upHandler() {
    selectedIndex = (selectedIndex + items.length - 1) % items.length;
  }

  function downHandler() {
    selectedIndex = (selectedIndex + 1) % items.length;
  }

  function enterHandler() {
    selectItem(selectedIndex);
  }

  function selectItem(index: number) {
    const item = items[index];
    if (item) {
      command(item);
    }
  }
  
  function getIcon(type: string) {
    switch (type) {
      case 'note': return FileText;
      case 'account': return Building2;
      case 'todo': return CheckSquare;
      default: return FileText;
    }
  }
</script>

{#if items.length > 0}
  <div 
    bind:this={element}
    class="bg-[var(--color-card)] border border-[var(--color-border)] shadow-xl rounded-lg overflow-hidden min-w-[200px] max-h-[300px] overflow-y-auto flex flex-col"
  >
    {#each items as item, index}
      <button
        class="flex items-center gap-2 px-3 py-2 text-left text-sm hover:bg-[var(--color-bg-elevated)] transition-colors {index === selectedIndex ? 'bg-[var(--color-bg-elevated)] text-[var(--color-text)]' : 'text-[var(--color-muted)]'}"
        on:click={() => selectItem(index)}
      >
        <svelte:component this={getIcon(item.type)} class="w-3.5 h-3.5 opacity-70" />
        <span class="truncate flex-1">{item.title}</span>
        {#if item.account_name}
          <span class="text-xs opacity-50 truncate max-w-[80px]">{item.account_name}</span>
        {/if}
      </button>
    {/each}
  </div>
{/if}
