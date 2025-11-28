<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { api, type Account } from '$lib/utils/api';
  import { addToast } from '$lib/stores';
  import { 
    X, 
    FileText, 
    CheckSquare,
    Zap
  } from 'lucide-svelte';

  export let open = false;

  const dispatch = createEventDispatcher();

  let type: 'note' | 'todo' = 'note';
  let title = '';
  let content = '';
  let description = '';
  let accountId = '';
  let priority = 'medium';
  let accounts: Account[] = [];
  let loading = false;

  onMount(async () => {
    try {
      accounts = await api.getAccounts();
    } catch (err) {
      console.error('Failed to load accounts:', err);
    }
  });

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      close();
    }
  }

  function close() {
    open = false;
    resetForm();
    dispatch('close');
  }

  function resetForm() {
    title = '';
    content = '';
    description = '';
    accountId = '';
    priority = 'medium';
    type = 'note';
  }

  async function handleSubmit() {
    if (!title.trim()) {
      addToast('error', 'Title is required');
      return;
    }

    loading = true;
    try {
      const result = await api.quickCapture({
        type,
        title: title.trim(),
        content: type === 'note' ? content : undefined,
        description: type === 'todo' ? description : undefined,
        account_id: accountId || undefined,
        priority: type === 'todo' ? priority : undefined,
      });

      addToast('success', `${type === 'note' ? 'Note' : 'Todo'} created successfully`);
      close();
      dispatch('created', result);
    } catch (err: any) {
      addToast('error', err.message);
    } finally {
      loading = false;
    }
  }

  $: if (open) {
    setTimeout(() => {
      const input = document.querySelector('#quick-capture-title') as HTMLInputElement;
      input?.focus();
    }, 100);
  }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if open}
  <div class="fixed inset-0 z-50 flex items-start justify-center pt-[15vh]">
    <button 
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={close}
    ></button>
    
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl shadow-2xl w-full max-w-lg animate-slide-up">
      <!-- Header -->
      <div class="flex items-center justify-between p-4 border-b border-[var(--color-border)]">
        <div class="flex items-center gap-2">
          <Zap class="w-5 h-5 text-primary-500" />
          <h2 class="font-semibold">Quick Capture</h2>
        </div>
        <button 
          class="p-1.5 hover:bg-[var(--color-bg)] rounded-lg transition-colors"
          on:click={close}
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Type Toggle -->
      <div class="flex gap-2 p-4 pb-0">
        <button
          class="flex-1 flex items-center justify-center gap-2 py-2.5 rounded-lg border transition-colors"
          class:bg-primary-500={type === 'note'}
          class:text-white={type === 'note'}
          class:border-primary-500={type === 'note'}
          class:border-[var(--color-border)]={type !== 'note'}
          on:click={() => type = 'note'}
        >
          <FileText class="w-4 h-4" />
          Note
        </button>
        <button
          class="flex-1 flex items-center justify-center gap-2 py-2.5 rounded-lg border transition-colors"
          class:bg-primary-500={type === 'todo'}
          class:text-white={type === 'todo'}
          class:border-primary-500={type === 'todo'}
          class:border-[var(--color-border)]={type !== 'todo'}
          on:click={() => type = 'todo'}
        >
          <CheckSquare class="w-4 h-4" />
          Todo
        </button>
      </div>

      <!-- Form -->
      <form on:submit|preventDefault={handleSubmit} class="p-4 space-y-4">
        <div>
          <input
            id="quick-capture-title"
            type="text"
            class="input text-lg"
            placeholder={type === 'note' ? 'Note title...' : 'What needs to be done?'}
            bind:value={title}
            autofocus
          />
        </div>

        {#if type === 'note'}
          <div>
            <textarea
              class="input h-24 resize-none"
              placeholder="Quick notes (optional)..."
              bind:value={content}
            ></textarea>
          </div>
        {:else}
          <div>
            <textarea
              class="input h-20 resize-none"
              placeholder="Description (optional)..."
              bind:value={description}
            ></textarea>
          </div>
          
          <div>
            <label class="label">Priority</label>
            <div class="flex gap-2">
              {#each ['low', 'medium', 'high'] as p}
                <button
                  type="button"
                  class="flex-1 py-2 rounded-lg border text-sm font-medium transition-colors"
                  class:bg-green-500={priority === p && p === 'low'}
                  class:bg-amber-500={priority === p && p === 'medium'}
                  class:bg-red-500={priority === p && p === 'high'}
                  class:text-white={priority === p}
                  class:border-green-500={priority === p && p === 'low'}
                  class:border-amber-500={priority === p && p === 'medium'}
                  class:border-red-500={priority === p && p === 'high'}
                  class:border-[var(--color-border)]={priority !== p}
                  on:click={() => priority = p}
                >
                  {p.charAt(0).toUpperCase() + p.slice(1)}
                </button>
              {/each}
            </div>
          </div>
        {/if}

        <div>
          <label class="label">Account (optional)</label>
          <select class="input" bind:value={accountId}>
            <option value="">No account</option>
            {#each accounts as account}
              <option value={account.id}>{account.name}</option>
            {/each}
          </select>
        </div>

        <div class="flex gap-3 pt-2">
          <button 
            type="button" 
            class="btn-secondary flex-1"
            on:click={close}
          >
            Cancel
          </button>
          <button 
            type="submit" 
            class="btn-primary flex-1"
            disabled={loading || !title.trim()}
          >
            {loading ? 'Creating...' : 'Create'}
          </button>
        </div>
      </form>

      <!-- Keyboard hint -->
      <div class="px-4 pb-4">
        <p class="text-xs text-center text-[var(--color-muted)]">
          Press <kbd class="px-1.5 py-0.5 bg-[var(--color-bg)] rounded border border-[var(--color-border)] text-xs">Esc</kbd> to close
        </p>
      </div>
    </div>
  </div>
{/if}
