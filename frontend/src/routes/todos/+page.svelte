<script lang="ts">
  import { onMount } from 'svelte';
  import { dndzone } from 'svelte-dnd-action';
  import { flip } from 'svelte/animate';
  import { 
    Plus, 
    GripVertical, 
    FileText, 
    Trash2,
    Link,
    Building2,
    Check,
    AlertTriangle,
    ChevronDown,
    RotateCcw
  } from 'lucide-svelte';
  import { api, type Todo, type Note, type Account } from '$lib/utils/api';
  import { addToast } from '$lib/stores';

  interface Column {
    id: string;
    title: string;
    items: Todo[];
  }

  let columns: Column[] = [
    { id: 'not_started', title: 'Not Started', items: [] },
    { id: 'in_progress', title: 'In Progress', items: [] },
    { id: 'stuck', title: 'Stuck', items: [] },
  ];
  
  let completedItems: Todo[] = [];
  let loading = true;
  let showNewTodoModal = false;
  let showLinkModal = false;
  let newTodoTitle = '';
  let newTodoDescription = '';
  let newTodoAccountId = '';
  let linkTodoId = '';
  let availableNotes: Note[] = [];
  let accounts: Account[] = [];

  const flipDurationMs = 200;

  onMount(async () => {
    await loadData();
  });

  async function loadData() {
    try {
      loading = true;
      const [todos, accountsData] = await Promise.all([
        api.getTodos(),
        api.getAccounts()
      ]);
      
      accounts = accountsData;
      
      columns = [
        { id: 'not_started', title: 'Not Started', items: todos.filter(t => t.status === 'not_started') },
        { id: 'in_progress', title: 'In Progress', items: todos.filter(t => t.status === 'in_progress') },
        { id: 'stuck', title: 'Stuck', items: todos.filter(t => t.status === 'stuck') },
      ];
      
      completedItems = todos.filter(t => t.status === 'completed');
    } catch (e) {
      addToast('error', 'Failed to load todos');
    } finally {
      loading = false;
    }
  }

  function handleDndConsider(columnId: string, e: CustomEvent) {
    if (columnId === 'completed') {
      completedItems = e.detail.items;
    } else {
      const colIndex = columns.findIndex(c => c.id === columnId);
      if (colIndex !== -1) {
        columns[colIndex].items = e.detail.items;
        columns = columns;
      }
    }
  }

  async function handleDndFinalize(columnId: string, e: CustomEvent) {
    const newItems = e.detail.items;
    
    if (columnId === 'completed') {
      completedItems = newItems;
      for (const item of newItems) {
        if (item.status !== 'completed') {
          await updateTodoStatus(item.id, 'completed');
          item.status = 'completed';
        }
      }
    } else {
      const colIndex = columns.findIndex(c => c.id === columnId);
      if (colIndex !== -1) {
        columns[colIndex].items = newItems;
        columns = columns;
        
        for (const item of newItems) {
          if (item.status !== columnId) {
            await updateTodoStatus(item.id, columnId);
            item.status = columnId;
          }
        }
      }
    }
  }

  async function updateTodoStatus(todoId: string, status: string) {
    try {
      await api.updateTodo(todoId, { status });
    } catch (err) {
      addToast('error', 'Failed to update todo');
    }
  }

  async function createTodo() {
    if (!newTodoTitle.trim()) return;
    try {
      const todo = await api.createTodo({
        title: newTodoTitle.trim(),
        description: newTodoDescription.trim(),
        status: 'not_started',
        account_id: newTodoAccountId || undefined
      });
      
      columns[0].items = [...columns[0].items, todo];
      columns = columns;
      
      newTodoTitle = '';
      newTodoDescription = '';
      newTodoAccountId = '';
      showNewTodoModal = false;
      addToast('success', 'Todo created');
    } catch (e) {
      addToast('error', 'Failed to create todo');
    }
  }

  async function deleteTodo(todoId: string, columnId: string) {
    if (!confirm('Delete this todo?')) return;
    try {
      await api.deleteTodo(todoId);
      if (columnId === 'completed') {
        completedItems = completedItems.filter(t => t.id !== todoId);
      } else {
        const colIndex = columns.findIndex(c => c.id === columnId);
        if (colIndex !== -1) {
          columns[colIndex].items = columns[colIndex].items.filter(t => t.id !== todoId);
          columns = columns;
        }
      }
      addToast('success', 'Todo deleted');
    } catch (e) {
      addToast('error', 'Failed to delete todo');
    }
  }

  async function markComplete(todo: Todo, fromColumnId: string) {
    try {
      await api.updateTodo(todo.id, { status: 'completed' });
      
      // Remove from source
      if (fromColumnId === 'completed') return;
      const colIndex = columns.findIndex(c => c.id === fromColumnId);
      if (colIndex !== -1) {
        columns[colIndex].items = columns[colIndex].items.filter(t => t.id !== todo.id);
        columns = columns;
      }
      
      // Add to completed
      todo.status = 'completed';
      completedItems = [todo, ...completedItems];
      
      addToast('success', 'Marked complete');
    } catch (e) {
      addToast('error', 'Failed to update');
    }
  }

  async function markStuck(todo: Todo, fromColumnId: string) {
    try {
      await api.updateTodo(todo.id, { status: 'stuck' });
      
      // Remove from source
      const colIndex = columns.findIndex(c => c.id === fromColumnId);
      if (colIndex !== -1) {
        columns[colIndex].items = columns[colIndex].items.filter(t => t.id !== todo.id);
      }
      
      // Add to stuck
      const stuckIndex = columns.findIndex(c => c.id === 'stuck');
      todo.status = 'stuck';
      columns[stuckIndex].items = [todo, ...columns[stuckIndex].items];
      columns = columns;
      
      addToast('success', 'Moved to Stuck');
    } catch (e) {
      addToast('error', 'Failed to update');
    }
  }

  async function restoreTodo(todo: Todo, toStatus: string) {
    try {
      await api.updateTodo(todo.id, { status: toStatus });
      
      // Remove from completed
      completedItems = completedItems.filter(t => t.id !== todo.id);
      
      // Add to target column
      const colIndex = columns.findIndex(c => c.id === toStatus);
      todo.status = toStatus;
      columns[colIndex].items = [todo, ...columns[colIndex].items];
      columns = columns;
      
      addToast('success', 'Todo restored');
    } catch (e) {
      addToast('error', 'Failed to restore');
    }
  }

  async function setAccount(todo: Todo, accountId: string, columnId: string) {
    try {
      const account = accounts.find(a => a.id === accountId);
      await api.updateTodo(todo.id, { account_id: accountId || null });
      
      todo.account_id = accountId || undefined;
      todo.account_name = account?.name || '';
      
      if (columnId === 'completed') {
        completedItems = completedItems;
      } else {
        columns = columns;
      }
      
      addToast('success', accountId ? `Tagged: ${account?.name}` : 'Account removed');
    } catch (e) {
      addToast('error', 'Failed to update');
    }
  }

  async function openLinkModal(todoId: string) {
    linkTodoId = todoId;
    try {
      availableNotes = await api.getNotes();
      showLinkModal = true;
    } catch (e) {
      addToast('error', 'Failed to load notes');
    }
  }

  async function linkToNote(noteId: string) {
    try {
      await api.linkTodoToNote(linkTodoId, noteId);
      await loadData();
      showLinkModal = false;
      addToast('success', 'Linked to note');
    } catch (e) {
      addToast('error', 'Failed to link');
    }
  }

  function getColumnColor(columnId: string): string {
    switch (columnId) {
      case 'not_started': return 'bg-gray-500';
      case 'in_progress': return 'bg-blue-500';
      case 'stuck': return 'bg-red-500';
      case 'completed': return 'bg-green-500';
      default: return 'bg-gray-400';
    }
  }

  function getOutlineColor(columnId: string): string {
    switch (columnId) {
      case 'not_started': return '#6b7280';
      case 'in_progress': return '#3b82f6';
      case 'stuck': return '#ef4444';
      case 'completed': return '#22c55e';
      default: return '#9ca3af';
    }
  }
</script>

<svelte:head>
  <title>Todos - Noted</title>
</svelte:head>

<div class="max-w-full mx-auto">
  <div class="flex items-center justify-between mb-8">
    <div>
      <h1 class="page-title">Todos</h1>
      <p class="page-subtitle">Track your follow-up items</p>
    </div>
    <button class="btn-primary" on:click={() => showNewTodoModal = true}>
      <Plus class="w-4 h-4" />
      New Todo
    </button>
  </div>

  {#if loading}
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      {#each [1, 2, 3] as _}
        <div class="card animate-pulse">
          <div class="h-6 bg-[var(--color-border)] rounded w-32 mb-4"></div>
          <div class="space-y-3">
            <div class="h-20 bg-[var(--color-border)] rounded"></div>
            <div class="h-20 bg-[var(--color-border)] rounded"></div>
          </div>
        </div>
      {/each}
    </div>
  {:else}
    <!-- Main Columns -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      {#each columns as column (column.id)}
        <div class="flex flex-col">
          <div class="flex items-center gap-2 mb-4">
            <div class="w-3 h-3 rounded-full {getColumnColor(column.id)}"></div>
            <h2 class="font-semibold">{column.title}</h2>
            <span class="text-sm text-[var(--color-muted)]">({column.items.length})</span>
          </div>

          <div 
            class="flex-1 min-h-[400px] p-3 bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl"
            use:dndzone={{
              items: column.items,
              flipDurationMs,
              dropTargetStyle: { outline: `2px dashed ${getOutlineColor(column.id)}`, outlineOffset: '-2px' }
            }}
            on:consider={(e) => handleDndConsider(column.id, e)}
            on:finalize={(e) => handleDndFinalize(column.id, e)}
          >
            {#each column.items as todo (todo.id)}
              <div 
                class="bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg p-4 mb-3 cursor-grab group"
                animate:flip={{ duration: flipDurationMs }}
              >
                <!-- Header -->
                <div class="flex items-start justify-between gap-2 mb-1">
                  <div class="flex items-center gap-2 min-w-0">
                    <GripVertical class="w-4 h-4 text-[var(--color-muted)] opacity-0 group-hover:opacity-100 flex-shrink-0" />
                    <h3 class="font-medium text-sm truncate">{todo.title}</h3>
                  </div>
                  <div class="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 flex-shrink-0">
                    <!-- Complete button -->
                    {#if column.id !== 'completed'}
                      <button 
                        class="p-1.5 hover:bg-green-500/20 rounded"
                        title="Mark complete"
                        on:click|stopPropagation={() => markComplete(todo, column.id)}
                      >
                        <Check class="w-4 h-4 text-green-500" />
                      </button>
                    {/if}
                    <!-- Stuck button (only on in_progress) -->
                    {#if column.id === 'in_progress'}
                      <button 
                        class="p-1.5 hover:bg-red-500/20 rounded"
                        title="Mark stuck"
                        on:click|stopPropagation={() => markStuck(todo, column.id)}
                      >
                        <AlertTriangle class="w-4 h-4 text-red-500" />
                      </button>
                    {/if}
                    <!-- Link button -->
                    <button 
                      class="p-1.5 hover:bg-[var(--color-border)] rounded"
                      title="Link to note"
                      on:click|stopPropagation={() => openLinkModal(todo.id)}
                    >
                      <Link class="w-4 h-4 text-[var(--color-muted)]" />
                    </button>
                    <!-- Delete -->
                    <button 
                      class="p-1.5 hover:bg-red-500/20 rounded"
                      title="Delete"
                      on:click|stopPropagation={() => deleteTodo(todo.id, column.id)}
                    >
                      <Trash2 class="w-4 h-4 text-red-500" />
                    </button>
                  </div>
                </div>

                <!-- Description -->
                {#if todo.description}
                  <p class="text-xs text-[var(--color-muted)] mb-2 ml-6 line-clamp-2">{todo.description}</p>
                {/if}

                <!-- Account Tag -->
                <div class="ml-6">
                  <div class="relative inline-block group/dropdown">
                    <button
                      class="inline-flex items-center gap-1 px-2 py-1 text-xs rounded transition-colors {todo.account_name ? 'bg-orange-500/10 text-orange-600 dark:text-orange-400' : 'bg-[var(--color-border)] text-[var(--color-muted)]'}"
                      on:click|stopPropagation
                    >
                      <Building2 class="w-3 h-3" />
                      {todo.account_name || 'Add account'}
                      <ChevronDown class="w-3 h-3" />
                    </button>
                    <div class="absolute left-0 top-full mt-1 z-20 hidden group-hover/dropdown:block">
                      <div class="bg-[var(--color-card)] border border-[var(--color-border)] rounded-lg shadow-xl py-1 min-w-[150px]">
                        <button
                          class="w-full px-3 py-1.5 text-xs text-left hover:bg-[var(--color-bg)] {!todo.account_id ? 'text-primary-500 font-medium' : ''}"
                          on:click|stopPropagation={() => setAccount(todo, '', column.id)}
                        >
                          No account
                        </button>
                        {#each accounts as account}
                          <button
                            class="w-full px-3 py-1.5 text-xs text-left hover:bg-[var(--color-bg)] {todo.account_id === account.id ? 'text-primary-500 font-medium' : ''}"
                            on:click|stopPropagation={() => setAccount(todo, account.id, column.id)}
                          >
                            {account.name}
                          </button>
                        {/each}
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Linked Notes -->
                {#if todo.linked_notes && todo.linked_notes.length > 0}
                  <div class="flex flex-wrap gap-1 mt-2 ml-6">
                    {#each todo.linked_notes as note}
                      <a 
                        href="/notes/{note.id}"
                        class="inline-flex items-center gap-1 px-2 py-0.5 text-xs bg-primary-500/10 text-primary-500 rounded hover:bg-primary-500/20"
                        on:click|stopPropagation
                      >
                        <FileText class="w-3 h-3" />
                        {note.title}
                      </a>
                    {/each}
                  </div>
                {/if}
              </div>
            {:else}
              <div class="flex items-center justify-center h-32 text-[var(--color-muted)] text-sm">
                Drop items here
              </div>
            {/each}
          </div>
        </div>
      {/each}
    </div>

    <!-- Completed Section -->
    <div class="mt-8">
      <div class="flex items-center gap-2 mb-4">
        <div class="w-3 h-3 rounded-full bg-green-500"></div>
        <h2 class="font-semibold">Completed</h2>
        <span class="text-sm text-[var(--color-muted)]">({completedItems.length})</span>
      </div>
      
      <div 
        class="min-h-[100px] p-3 bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl"
        use:dndzone={{
          items: completedItems,
          flipDurationMs,
          dropTargetStyle: { outline: '2px dashed #22c55e', outlineOffset: '-2px' }
        }}
        on:consider={(e) => handleDndConsider('completed', e)}
        on:finalize={(e) => handleDndFinalize('completed', e)}
      >
        {#if completedItems.length > 0}
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3">
            {#each completedItems as todo (todo.id)}
              <div 
                class="bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg p-3 cursor-grab group opacity-60 hover:opacity-100"
                animate:flip={{ duration: flipDurationMs }}
              >
                <div class="flex items-start justify-between gap-2">
                  <span class="text-sm line-through truncate">{todo.title}</span>
                  <div class="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 flex-shrink-0">
                    <button 
                      class="p-1 hover:bg-blue-500/20 rounded"
                      title="Restore to In Progress"
                      on:click|stopPropagation={() => restoreTodo(todo, 'in_progress')}
                    >
                      <RotateCcw class="w-3.5 h-3.5 text-blue-500" />
                    </button>
                    <button 
                      class="p-1 hover:bg-red-500/20 rounded"
                      title="Delete"
                      on:click|stopPropagation={() => deleteTodo(todo.id, 'completed')}
                    >
                      <Trash2 class="w-3.5 h-3.5 text-red-500" />
                    </button>
                  </div>
                </div>
                {#if todo.account_name}
                  <div class="mt-2">
                    <span class="inline-flex items-center gap-1 px-2 py-0.5 text-xs bg-orange-500/10 text-orange-500 rounded">
                      <Building2 class="w-3 h-3" />
                      {todo.account_name}
                    </span>
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        {:else}
          <div class="flex items-center justify-center h-16 text-[var(--color-muted)] text-sm">
            Completed items appear here
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<!-- New Todo Modal -->
{#if showNewTodoModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button 
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => showNewTodoModal = false}
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-md animate-slide-up">
      <h2 class="text-lg font-semibold mb-4">New Todo</h2>
      <form on:submit|preventDefault={createTodo}>
        <div class="mb-4">
          <label class="label">Title</label>
          <input 
            type="text"
            class="input"
            placeholder="What needs to be done?"
            bind:value={newTodoTitle}
            autofocus
          />
        </div>
        <div class="mb-4">
          <label class="label">Description (optional)</label>
          <textarea 
            class="input"
            rows="3"
            placeholder="Add more details..."
            bind:value={newTodoDescription}
          ></textarea>
        </div>
        <div class="mb-4">
          <label class="label">Account (optional)</label>
          <select class="input" bind:value={newTodoAccountId}>
            <option value="">No account</option>
            {#each accounts as account}
              <option value={account.id}>{account.name}</option>
            {/each}
          </select>
        </div>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn-secondary" on:click={() => showNewTodoModal = false}>
            Cancel
          </button>
          <button type="submit" class="btn-primary" disabled={!newTodoTitle.trim()}>
            Create
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- Link to Note Modal -->
{#if showLinkModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button 
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => showLinkModal = false}
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-md animate-slide-up">
      <h2 class="text-lg font-semibold mb-4">Link to Note</h2>
      {#if availableNotes.length === 0}
        <p class="text-[var(--color-muted)]">No notes available.</p>
      {:else}
        <div class="max-h-80 overflow-y-auto space-y-2">
          {#each availableNotes as note}
            <button 
              class="w-full flex items-center gap-3 p-3 bg-[var(--color-bg)] rounded-lg hover:bg-[var(--color-border)] text-left"
              on:click={() => linkToNote(note.id)}
            >
              <FileText class="w-5 h-5 text-primary-500" />
              <div>
                <p class="font-medium">{note.title}</p>
                <p class="text-sm text-[var(--color-muted)]">{note.account_name || 'No account'}</p>
              </div>
            </button>
          {/each}
        </div>
      {/if}
      <div class="flex justify-end mt-4">
        <button class="btn-secondary" on:click={() => showLinkModal = false}>Cancel</button>
      </div>
    </div>
  </div>
{/if}
