<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { api, type Account } from '$lib/utils/api';
  import { addToast } from '$lib/stores';
  import { FileText, CheckSquare, X, Plus } from 'lucide-svelte';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';

  export let open = false;

  const dispatch = createEventDispatcher();

  let captureType: 'note' | 'todo' = 'note';
  let title = '';
  let content = '';
  let accountId = '';
  let priority = 'medium';
  let accounts: Account[] = [];
  let loading = false;

  onMount(async () => {
    try {
      accounts = await api.getAccounts();
    } catch (e) {
      console.error('Failed to load accounts:', e);
    }
  });

  function close() {
    open = false;
    title = '';
    content = '';
    accountId = '';
    priority = 'medium';
    dispatch('close');
  }

  async function handleSubmit() {
    if (!title.trim()) return;

    loading = true;
    try {
      const result = await api.quickCapture({
        type: captureType,
        title: title.trim(),
        content: content.trim() || undefined,
        account_id: accountId || undefined,
        priority: captureType === 'todo' ? priority : undefined,
        description: captureType === 'todo' ? content.trim() : undefined
      });

      addToast('success', `${captureType === 'note' ? 'Note' : 'Todo'} created`);
      close();

      if (captureType === 'note') {
        goto(`/notes/${result.id}`);
      } else {
        goto('/todos');
      }
    } catch (e) {
      addToast('error', `Failed to create ${captureType}`);
    } finally {
      loading = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      close();
    }
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      handleSubmit();
    }
  }
</script>

<svelte:window on:keydown={open ? handleKeydown : undefined} />

{#if open}
  <div class="fixed inset-0 z-[100] flex items-start justify-center pt-[15vh]">
    <button class="modal-backdrop" on:click={close} aria-label="Close modal"></button>
    
    <div class="relative z-[101] w-full max-w-lg mx-4 bg-[var(--color-card)] border border-[var(--color-border)] shadow-editorial-lg animate-scale-in" style="border-radius: 2px;">
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-[var(--color-border)]">
        <div class="flex items-center gap-3">
          <div class="p-2 bg-[var(--color-accent)]/10 border border-[var(--color-accent)]/20" style="border-radius: 2px;">
            <Plus class="w-4 h-4 text-[var(--color-accent)]" strokeWidth={1.5} />
          </div>
          <h2 class="font-serif text-lg">Quick Create</h2>
        </div>
        <button 
          class="p-2 hover:bg-[var(--color-card-hover)] transition-colors"
          style="border-radius: 2px;"
          on:click={close}
        >
          <X class="w-4 h-4" strokeWidth={1.5} />
        </button>
      </div>

      <!-- Type Toggle -->
      <div class="px-6 py-4 border-b border-[var(--color-border)]">
        <div class="flex gap-2">
          <button
            class="flex-1 flex items-center justify-center gap-2 py-3 border transition-all {captureType === 'note' ? 'border-[var(--color-accent)] bg-[var(--color-accent)]/5 text-[var(--color-accent)]' : 'border-[var(--color-border)] hover:border-[var(--color-accent)]/40'}"
            style="border-radius: 2px;"
            on:click={() => captureType = 'note'}
          >
            <FileText class="w-4 h-4" strokeWidth={1.5} />
            <span class="text-sm font-medium">Note</span>
          </button>
          <button
            class="flex-1 flex items-center justify-center gap-2 py-3 border transition-all {captureType === 'todo' ? 'border-[var(--color-accent)] bg-[var(--color-accent)]/5 text-[var(--color-accent)]' : 'border-[var(--color-border)] hover:border-[var(--color-accent)]/40'}"
            style="border-radius: 2px;"
            on:click={() => captureType = 'todo'}
          >
            <CheckSquare class="w-4 h-4" strokeWidth={1.5} />
            <span class="text-sm font-medium">Todo</span>
          </button>
        </div>
      </div>

      <!-- Form -->
      <form on:submit|preventDefault={handleSubmit} class="p-6">
        <div class="space-y-5">
          <div>
            <label class="label" for="quick-capture-title">Title</label>
            <!-- svelte-ignore a11y-autofocus -->
            <input
              id="quick-capture-title"
              type="text"
              class="input"
              placeholder={captureType === 'note' ? 'Note title...' : 'What needs to be done?'}
              bind:value={title}
              autofocus
            />
          </div>

          <div>
            <label class="label" for="quick-capture-content">{captureType === 'note' ? 'Quick notes' : 'Description'} (optional)</label>
            <textarea
              id="quick-capture-content"
              class="input"
              rows="3"
              placeholder={captureType === 'note' ? 'Jot down quick thoughts...' : 'Add more details...'}
              bind:value={content}
            ></textarea>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="label" for="quick-capture-account">Account (optional)</label>
              <select id="quick-capture-account" class="input" bind:value={accountId}>
                <option value="">No account</option>
                {#each accounts as account}
                  <option value={account.id}>{account.name}</option>
                {/each}
              </select>
            </div>

            {#if captureType === 'todo'}
              <div>
                <label class="label" for="quick-capture-priority">Priority</label>
                <select id="quick-capture-priority" class="input" bind:value={priority}>
                  <option value="low">Low</option>
                  <option value="medium">Medium</option>
                  <option value="high">High</option>
                </select>
              </div>
            {/if}
          </div>
        </div>

        <div class="flex items-center justify-between mt-8 pt-6 border-t border-[var(--color-border)]">
          <span class="text-xs text-[var(--color-muted)]">
            <kbd class="px-1.5 py-0.5 bg-[var(--color-bg)] border border-[var(--color-border)]" style="border-radius: 2px;">⌘</kbd>
            +
            <kbd class="px-1.5 py-0.5 bg-[var(--color-bg)] border border-[var(--color-border)]" style="border-radius: 2px;">Enter</kbd>
            to save
          </span>
          <div class="flex gap-3">
            <button type="button" class="btn-secondary" on:click={close}>
              Cancel
            </button>
            <button type="submit" class="btn-primary" disabled={!title.trim() || loading}>
              {loading ? 'Creating...' : 'Create'}
            </button>
          </div>
        </div>
      </form>
    </div>
  </div>
{/if}
