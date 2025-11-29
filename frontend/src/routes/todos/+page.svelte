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
    RotateCcw,
    Trash,
    RefreshCw,
    Pencil
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
  let deletedItems: Todo[] = [];
  let loading = true;
  let showTrash = false;
  let showNewTodoModal = false;
  let showLinkModal = false;
  let newTodoTitle = '';
  let newTodoDescription = '';
  let newTodoAccountId = '';
  let linkTodoId = '';
  let availableNotes: Note[] = [];
  let accounts: Account[] = [];
  let showEditModal = false;
  let editingTodo: Todo | null = null;
  let editTitle = '';
  let editDescription = '';

  const flipDurationMs = 200;

  onMount(async () => {
    await loadData();
  });

  async function loadData() {
    try {
      loading = true;
      const [todos, accountsData, deleted] = await Promise.all([
        api.getTodos(),
        api.getAccounts(),
        api.getDeletedTodos()
      ]);
      
      accounts = accountsData;
      deletedItems = deleted;
      
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
    try {
      let deletedTodo: Todo | undefined;
      
      // Optimistic update
      if (columnId === 'completed') {
        deletedTodo = completedItems.find(t => t.id === todoId);
        completedItems = completedItems.filter(t => t.id !== todoId);
      } else {
        const colIndex = columns.findIndex(c => c.id === columnId);
        if (colIndex !== -1) {
          deletedTodo = columns[colIndex].items.find(t => t.id === todoId);
          columns[colIndex].items = columns[colIndex].items.filter(t => t.id !== todoId);
          columns = columns;
        }
      }
      if (deletedTodo) {
        deletedItems = [deletedTodo, ...deletedItems];
      }
      addToast('success', 'Moved to trash');
      
      await api.deleteTodo(todoId);
    } catch (e) {
      addToast('error', 'Failed to delete todo');
      await loadData(); // Reload to revert
    }
  }

  function openEditModal(todo: Todo) {
    editingTodo = todo;
    editTitle = todo.title;
    editDescription = todo.description || '';
    showEditModal = true;
  }

  async function saveEditTodo() {
    if (!editingTodo || !editTitle.trim()) return;
    try {
      await api.updateTodo(editingTodo.id, {
        title: editTitle.trim(),
        description: editDescription.trim() || undefined
      });
      
      // Update in columns or completed
      for (const column of columns) {
        const idx = column.items.findIndex(t => t.id === editingTodo!.id);
        if (idx !== -1) {
          column.items[idx].title = editTitle.trim();
          column.items[idx].description = editDescription.trim();
          columns = columns;
          break;
        }
      }
      const compIdx = completedItems.findIndex(t => t.id === editingTodo!.id);
      if (compIdx !== -1) {
        completedItems[compIdx].title = editTitle.trim();
        completedItems[compIdx].description = editDescription.trim();
        completedItems = completedItems;
      }
      
      showEditModal = false;
      editingTodo = null;
      addToast('success', 'Todo updated');
    } catch (e) {
      addToast('error', 'Failed to update todo');
    }
  }

  async function restoreFromTrash(todoId: string) {
    try {
      await api.restoreTodo(todoId);
      const todo = deletedItems.find(t => t.id === todoId);
      if (todo) {
        deletedItems = deletedItems.filter(t => t.id !== todoId);
        const colIndex = columns.findIndex(c => c.id === todo.status);
        if (todo.status === 'completed') {
          completedItems = [todo, ...completedItems];
        } else if (colIndex !== -1) {
          columns[colIndex].items = [todo, ...columns[colIndex].items];
          columns = columns;
        } else {
          columns[0].items = [todo, ...columns[0].items];
          columns = columns;
        }
      }
      addToast('success', 'Todo restored');
    } catch (e) {
      addToast('error', 'Failed to restore');
    }
  }

  async function permanentDelete(todoId: string) {
    if (!confirm('Permanently delete this todo?')) return;
    try {
      await api.permanentDeleteTodo(todoId);
      deletedItems = deletedItems.filter(t => t.id !== todoId);
      addToast('success', 'Permanently deleted');
    } catch (e) {
      addToast('error', 'Failed to delete');
    }
  }

  async function advanceTodo(todo: Todo, fromColumnId: string) {
    const targetStatus = fromColumnId === 'not_started' ? 'in_progress' : 'completed';
    try {
      await api.updateTodo(todo.id, { status: targetStatus });
      const colIndex = columns.findIndex(c => c.id === fromColumnId);
      if (colIndex !== -1) {
        columns[colIndex].items = columns[colIndex].items.filter(t => t.id !== todo.id);
        columns = columns;
      }
      todo.status = targetStatus;
      if (targetStatus === 'completed') {
        completedItems = [todo, ...completedItems];
        addToast('success', 'Marked complete');
      } else {
        const targetIndex = columns.findIndex(c => c.id === targetStatus);
        columns[targetIndex].items = [todo, ...columns[targetIndex].items];
        columns = columns;
        addToast('success', 'Started');
      }
    } catch (e) {
      addToast('error', 'Failed to update');
    }
  }

  async function markStuck(todo: Todo, fromColumnId: string) {
    try {
      await api.updateTodo(todo.id, { status: 'stuck' });
      const colIndex = columns.findIndex(c => c.id === fromColumnId);
      if (colIndex !== -1) {
        columns[colIndex].items = columns[colIndex].items.filter(t => t.id !== todo.id);
      }
      const stuckIndex = columns.findIndex(c => c.id === 'stuck');
      todo.status = 'stuck';
      columns[stuckIndex].items = [todo, ...columns[stuckIndex].items];
      columns = columns;
      addToast('success', 'Moved to Stuck');
    } catch (e) {
      addToast('error', 'Failed to update');
    }
  }

  async function restoreTodo(todo: Todo, toStatus: Todo['status']) {
    try {
      await api.updateTodo(todo.id, { status: toStatus });
      completedItems = completedItems.filter(t => t.id !== todo.id);
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
      await api.updateTodo(todo.id, { account_id: accountId || undefined });
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
      case 'not_started': return 'bg-charcoal-400';
      case 'in_progress': return 'bg-blue-500';
      case 'stuck': return 'bg-red-500';
      case 'completed': return 'bg-emerald-500';
      default: return 'bg-charcoal-400';
    }
  }

  function getOutlineColor(columnId: string): string {
    switch (columnId) {
      case 'not_started': return 'var(--color-border)';
      case 'in_progress': return '#3b82f6';
      case 'stuck': return '#ef4444';
      case 'completed': return '#10b981';
      default: return 'var(--color-border)';
    }
  }
</script>

<svelte:head>
  <title>Todos - Noted</title>
</svelte:head>

<div class="max-w-full mx-auto">
  <!-- Header -->
  <div class="flex items-start justify-between mb-12">
    <div class="page-header mb-0">
      <div class="divider-accent mb-6"></div>
      <h1 class="page-title">Todos</h1>
      <p class="page-subtitle">Track your follow-up items</p>
    </div>
    <button class="btn-primary" on:click={() => showNewTodoModal = true}>
      <Plus class="w-4 h-4" strokeWidth={1.5} />
      New Todo
    </button>
  </div>

  {#if loading}
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      {#each [1, 2, 3] as _}
        <div class="card">
          <div class="skeleton h-6 w-32 mb-4"></div>
          <div class="space-y-3">
            <div class="skeleton h-20"></div>
            <div class="skeleton h-20"></div>
          </div>
        </div>
      {/each}
    </div>
  {:else}
    <!-- Kanban Columns -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-12">
      {#each columns as column (column.id)}
        <div class="flex flex-col">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-2 h-2 {getColumnColor(column.id)}" style="border-radius: 1px;"></div>
            <h2 class="font-serif text-lg">{column.title}</h2>
            <span class="text-sm text-[var(--color-muted)]">({column.items.length})</span>
          </div>

          <div 
            class="flex-1 min-h-[400px] p-4 bg-[var(--color-card)] border border-[var(--color-border)]"
            style="border-radius: 2px;"
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
                class="bg-[var(--color-bg)] border border-[var(--color-border)] p-4 mb-3 cursor-grab group hover:border-[var(--color-accent)]/40 transition-colors"
                style="border-radius: 2px;"
                animate:flip={{ duration: flipDurationMs }}
              >
                <div class="flex items-start justify-between gap-2 mb-2">
                  <div class="flex items-center gap-2 min-w-0">
                    <GripVertical class="w-4 h-4 text-[var(--color-muted)] opacity-0 group-hover:opacity-100 flex-shrink-0" strokeWidth={1.5} />
                    <h3 class="font-medium text-sm">{todo.title}</h3>
                  </div>
                  <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 flex-shrink-0">
                    <button 
                      class="btn-icon-sm btn-icon-success"
                      title={column.id === 'not_started' ? 'Start' : 'Complete'}
                      on:click|stopPropagation={() => advanceTodo(todo, column.id)}
                    >
                      <Check class="w-4 h-4" strokeWidth={1.5} />
                    </button>
                    {#if column.id === 'in_progress'}
                      <button 
                        class="btn-icon-sm btn-icon-danger"
                        title="Mark stuck"
                        on:click|stopPropagation={() => markStuck(todo, column.id)}
                      >
                        <AlertTriangle class="w-4 h-4" strokeWidth={1.5} />
                      </button>
                    {/if}
                    <button 
                      class="btn-icon-sm btn-icon-ghost"
                      title="Edit"
                      on:click|stopPropagation={() => openEditModal(todo)}
                    >
                      <Pencil class="w-4 h-4" strokeWidth={1.5} />
                    </button>
                    <button 
                      class="btn-icon-sm btn-icon-ghost"
                      title="Link to note"
                      on:click|stopPropagation={() => openLinkModal(todo.id)}
                    >
                      <Link class="w-4 h-4" strokeWidth={1.5} />
                    </button>
                    <button 
                      class="btn-icon-sm btn-icon-danger"
                      title="Delete"
                      on:click|stopPropagation={() => deleteTodo(todo.id, column.id)}
                    >
                      <Trash2 class="w-4 h-4" strokeWidth={1.5} />
                    </button>
                  </div>
                </div>

                {#if todo.description}
                  <p class="text-xs text-[var(--color-muted)] mb-3 ml-6 line-clamp-2">{todo.description}</p>
                {/if}

                <div class="ml-6">
                  <div class="relative inline-block group/dropdown">
                    <button
                      class="inline-flex items-center gap-1.5 px-2.5 py-1 text-xs transition-colors {todo.account_name ? 'tag-accent' : 'tag-default'}"
                      on:click|stopPropagation
                    >
                      <Building2 class="w-3 h-3" strokeWidth={1.5} />
                      {todo.account_name || 'Add account'}
                      <ChevronDown class="w-3 h-3" strokeWidth={1.5} />
                    </button>
                    <div class="absolute left-0 top-full mt-1 z-20 hidden group-hover/dropdown:block">
                      <div class="bg-[var(--color-card)] border border-[var(--color-border)] shadow-editorial py-1 min-w-[150px]" style="border-radius: 2px;">
                        <button
                          class="w-full px-3 py-2 text-xs text-left hover:bg-[var(--color-bg)] {!todo.account_id ? 'text-[var(--color-accent)]' : ''}"
                          on:click|stopPropagation={() => setAccount(todo, '', column.id)}
                        >
                          No account
                        </button>
                        {#each accounts as account}
                          <button
                            class="w-full px-3 py-2 text-xs text-left hover:bg-[var(--color-bg)] {todo.account_id === account.id ? 'text-[var(--color-accent)]' : ''}"
                            on:click|stopPropagation={() => setAccount(todo, account.id, column.id)}
                          >
                            {account.name}
                          </button>
                        {/each}
                      </div>
                    </div>
                  </div>
                </div>

                {#if todo.linked_notes && todo.linked_notes.length > 0}
                  <div class="flex flex-wrap gap-1.5 mt-3 ml-6">
                    {#each todo.linked_notes as note}
                      <a 
                        href="/notes/{note.id}"
                        class="inline-flex items-center gap-1 px-2 py-0.5 text-xs bg-[var(--color-accent)]/10 text-[var(--color-accent)] hover:bg-[var(--color-accent)]/20 transition-colors"
                        style="border-radius: 2px;"
                        on:click|stopPropagation
                      >
                        <FileText class="w-3 h-3" strokeWidth={1.5} />
                        {note.title}
                      </a>
                    {/each}
                  </div>
                {/if}
              </div>
            {:else}
              <div class="flex items-center justify-center h-32 text-[var(--color-muted)] text-sm border border-dashed border-[var(--color-border)]" style="border-radius: 2px;">
                Drop items here
              </div>
            {/each}
          </div>
        </div>
      {/each}
    </div>

    <!-- Completed Section -->
    <div class="mb-12">
      <div class="flex items-center gap-3 mb-4">
        <div class="w-2 h-2 bg-emerald-500" style="border-radius: 1px;"></div>
        <h2 class="font-serif text-lg">Completed</h2>
        <span class="text-sm text-[var(--color-muted)]">({completedItems.length})</span>
      </div>
      
      <div 
        class="min-h-[100px] p-4 bg-[var(--color-card)] border border-[var(--color-border)] flex flex-wrap gap-4"
        style="border-radius: 2px;"
        use:dndzone={{
          items: completedItems,
          flipDurationMs,
          dropTargetStyle: { outline: '2px dashed #10b981', outlineOffset: '-2px' }
        }}
        on:consider={(e) => handleDndConsider('completed', e)}
        on:finalize={(e) => handleDndFinalize('completed', e)}
      >
        {#each completedItems as todo (todo.id)}
          <div 
            class="bg-[var(--color-bg)] border border-[var(--color-border)] p-4 cursor-grab group opacity-60 hover:opacity-100 w-full md:w-[calc(50%-8px)] lg:w-[calc(25%-12px)] transition-opacity"
            style="border-radius: 2px;"
            animate:flip={{ duration: flipDurationMs }}
          >
            <div class="flex items-start justify-between gap-2">
              <span class="text-sm line-through">{todo.title}</span>
              <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 flex-shrink-0">
                <button 
                  class="btn-icon-sm btn-icon-ghost"
                  title="Edit"
                  on:click|stopPropagation={() => openEditModal(todo)}
                >
                  <Pencil class="w-3.5 h-3.5" strokeWidth={1.5} />
                </button>
                <button 
                  class="btn-icon-sm text-blue-500 hover:bg-blue-500/10"
                  title="Restore"
                  on:click|stopPropagation={() => restoreTodo(todo, 'in_progress')}
                >
                  <RotateCcw class="w-3.5 h-3.5" strokeWidth={1.5} />
                </button>
                <button 
                  class="btn-icon-sm btn-icon-danger"
                  title="Delete"
                  on:click|stopPropagation={() => deleteTodo(todo.id, 'completed')}
                >
                  <Trash2 class="w-3.5 h-3.5" strokeWidth={1.5} />
                </button>
              </div>
            </div>
            {#if todo.account_name}
              <div class="mt-2">
                <span class="tag-accent text-[10px]">
                  <Building2 class="w-3 h-3" strokeWidth={1.5} />
                  {todo.account_name}
                </span>
              </div>
            {/if}
          </div>
        {:else}
          <div class="flex items-center justify-center w-full h-16 text-[var(--color-muted)] text-sm">
            Completed items appear here
          </div>
        {/each}
      </div>
    </div>

    <!-- Trash Section -->
    {#if deletedItems.length > 0}
      <div>
        <button 
          class="flex items-center gap-2 mb-4 text-[var(--color-muted)] hover:text-[var(--color-text)] transition-colors"
          on:click={() => showTrash = !showTrash}
        >
          <Trash class="w-4 h-4" strokeWidth={1.5} />
          <span class="font-medium">Trash</span>
          <span class="text-sm">({deletedItems.length})</span>
          <ChevronDown class="w-4 h-4 transition-transform {showTrash ? 'rotate-180' : ''}" strokeWidth={1.5} />
        </button>
        
        {#if showTrash}
          <div class="p-6 bg-[var(--color-card)] border border-dashed border-[var(--color-border)]" style="border-radius: 2px;">
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              {#each deletedItems as todo (todo.id)}
                <div class="bg-[var(--color-bg)] border border-[var(--color-border)] p-4 opacity-50 hover:opacity-100 transition-opacity group" style="border-radius: 2px;">
                  <div class="flex items-start justify-between gap-2">
                    <span class="text-sm line-through">{todo.title}</span>
                    <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 flex-shrink-0">
                      <button 
                        class="btn-icon-sm btn-icon-success"
                        title="Restore"
                        on:click={() => restoreFromTrash(todo.id)}
                      >
                        <RefreshCw class="w-3.5 h-3.5" strokeWidth={1.5} />
                      </button>
                      <button 
                        class="btn-icon-sm btn-icon-danger"
                        title="Delete permanently"
                        on:click={() => permanentDelete(todo.id)}
                      >
                        <Trash2 class="w-3.5 h-3.5" strokeWidth={1.5} />
                      </button>
                    </div>
                  </div>
                  {#if todo.account_name}
                    <div class="mt-2">
                      <span class="tag-accent text-[10px]">
                        <Building2 class="w-3 h-3" strokeWidth={1.5} />
                        {todo.account_name}
                      </span>
                    </div>
                  {/if}
                </div>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {/if}
  {/if}
</div>

<!-- New Todo Modal -->
{#if showNewTodoModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button class="modal-backdrop" on:click={() => showNewTodoModal = false} aria-label="Close modal"></button>
    <div class="relative modal-content animate-scale-in">
      <h2 class="modal-title">New Todo</h2>
      <form on:submit|preventDefault={createTodo}>
        <div class="mb-6">
          <label class="label" for="todos-new-title">Title</label>
          <!-- svelte-ignore a11y-autofocus -->
          <input 
            id="todos-new-title"
            type="text"
            class="input"
            placeholder="What needs to be done?"
            bind:value={newTodoTitle}
            autofocus
          />
        </div>
        <div class="mb-6">
          <label class="label" for="todos-new-desc">Description (optional)</label>
          <textarea 
            id="todos-new-desc"
            class="input"
            rows="3"
            placeholder="Add more details..."
            bind:value={newTodoDescription}
          ></textarea>
        </div>
        <div class="mb-6">
          <label class="label" for="todos-new-account">Account (optional)</label>
          <select id="todos-new-account" class="input" bind:value={newTodoAccountId}>
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
    <button class="modal-backdrop" on:click={() => showLinkModal = false} aria-label="Close modal"></button>
    <div class="relative modal-content animate-scale-in">
      <h2 class="modal-title">Link to Note</h2>
      {#if availableNotes.length === 0}
        <p class="text-[var(--color-muted)]">No notes available.</p>
      {:else}
        <div class="max-h-80 overflow-y-auto space-y-2">
          {#each availableNotes as note}
            <button 
              class="w-full flex items-center gap-4 p-4 bg-[var(--color-bg)] border border-[var(--color-border)] hover:border-[var(--color-accent)]/40 transition-colors text-left"
              style="border-radius: 2px;"
              on:click={() => linkToNote(note.id)}
            >
              <FileText class="w-5 h-5 text-[var(--color-accent)]" strokeWidth={1.5} />
              <div>
                <p class="font-medium">{note.title}</p>
                <p class="text-sm text-[var(--color-muted)]">{note.account_name || 'No account'}</p>
              </div>
            </button>
          {/each}
        </div>
      {/if}
      <div class="flex justify-end mt-6">
        <button class="btn-secondary" on:click={() => showLinkModal = false}>Cancel</button>
      </div>
    </div>
  </div>
{/if}

<!-- Edit Todo Modal -->
{#if showEditModal && editingTodo}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button class="modal-backdrop" on:click={() => showEditModal = false} aria-label="Close modal"></button>
    <div class="relative modal-content animate-scale-in">
      <h2 class="modal-title">Edit Todo</h2>
      <form on:submit|preventDefault={saveEditTodo}>
        <div class="mb-4">
          <label class="label" for="todos-edit-title">Title</label>
          <!-- svelte-ignore a11y-autofocus -->
          <input 
            id="todos-edit-title"
            type="text"
            class="input"
            bind:value={editTitle}
            autofocus
          />
        </div>
        <div class="mb-6">
          <label class="label" for="todos-edit-desc">Description (optional)</label>
          <textarea
            id="todos-edit-desc"
            class="input"
            rows="3"
            placeholder="Add more details..."
            bind:value={editDescription}
          ></textarea>
        </div>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn-secondary" on:click={() => showEditModal = false}>
            Cancel
          </button>
          <button type="submit" class="btn-primary" disabled={!editTitle.trim()}>
            Save Changes
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
