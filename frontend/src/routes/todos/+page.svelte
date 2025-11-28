<script lang="ts">
  import { onMount } from 'svelte';
  import { dndzone } from 'svelte-dnd-action';
  import { flip } from 'svelte/animate';
  import { 
    Plus, 
    GripVertical, 
    FileText, 
    Calendar,
    Trash2,
    Link,
    CheckSquare,
    Square,
    X,
    ArrowRight,
    Building2,
    Filter,
    SortAsc,
    SortDesc,
    Check,
    AlertTriangle,
    ChevronDown
  } from 'lucide-svelte';
  import { api, type Todo, type Note, type Account } from '$lib/utils/api';
  import { addToast } from '$lib/stores';

  interface Column {
    id: string;
    title: string;
    items: Todo[];
  }

  // Main columns (3 across)
  let columns: Column[] = [
    { id: 'not_started', title: 'Not Started', items: [] },
    { id: 'in_progress', title: 'In Progress', items: [] },
    { id: 'stuck', title: 'Stuck', items: [] },
  ];
  
  // Completed is separate (full width)
  let completedColumn: Column = { id: 'completed', title: 'Completed', items: [] };

  let loading = true;
  let showNewTodoModal = false;
  let showLinkModal = false;
  let newTodoTitle = '';
  let newTodoDescription = '';
  let newTodoPriority = 'medium';
  let newTodoAccountId = '';
  let linkTodoId = '';
  let availableNotes: Note[] = [];
  let accounts: Account[] = [];
  
  // Bulk selection state
  let selectionMode = false;
  let selectedTodos: Set<string> = new Set();
  
  // Filter & Sort state
  let filterAccount = '';
  let filterPriority = '';
  let filterHasNotes = ''; // 'yes', 'no', or ''
  let sortBy = 'created_at'; // 'created_at', 'priority', 'title'
  let sortOrder: 'asc' | 'desc' = 'desc';
  let showFilters = false;

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
      
      // Distribute todos into columns (high priority first)
      const sortByPriority = (items: Todo[]) => {
        const priorityOrder = { high: 0, medium: 1, low: 2 };
        return items.sort((a, b) => (priorityOrder[a.priority] || 1) - (priorityOrder[b.priority] || 1));
      };
      
      columns = columns.map(col => ({
        ...col,
        items: sortByPriority(todos.filter(t => t.status === col.id))
      }));
      
      completedColumn = {
        ...completedColumn,
        items: sortByPriority(todos.filter(t => t.status === 'completed'))
      };
    } catch (e) {
      addToast('error', 'Failed to load todos');
    } finally {
      loading = false;
    }
  }

  function handleDndConsider(columnId: string, e: CustomEvent) {
    if (columnId === 'completed') {
      completedColumn.items = e.detail.items;
      completedColumn = completedColumn;
    } else {
      const column = columns.find(c => c.id === columnId);
      if (column) {
        column.items = e.detail.items;
        columns = columns;
      }
    }
  }

  async function handleDndFinalize(columnId: string, e: CustomEvent) {
    if (columnId === 'completed') {
      completedColumn.items = e.detail.items;
      completedColumn = completedColumn;
      
      for (const item of e.detail.items) {
        if (item.status !== 'completed') {
          try {
            await api.updateTodo(item.id, { status: 'completed' });
            item.status = 'completed';
          } catch (err) {
            addToast('error', 'Failed to update todo status');
          }
        }
      }
    } else {
      const column = columns.find(c => c.id === columnId);
      if (column) {
        column.items = e.detail.items;
        columns = columns;

        for (const item of e.detail.items) {
          if (item.status !== columnId) {
            try {
              await api.updateTodo(item.id, { status: columnId });
              item.status = columnId;
            } catch (err) {
              addToast('error', 'Failed to update todo status');
            }
          }
        }
      }
    }
  }

  async function createTodo() {
    if (!newTodoTitle.trim()) return;
    try {
      const todo = await api.createTodo({
        title: newTodoTitle.trim(),
        description: newTodoDescription.trim(),
        priority: newTodoPriority,
        status: 'not_started',
        account_id: newTodoAccountId || undefined
      });
      
      columns[0].items = [...columns[0].items, todo];
      columns = columns;
      
      newTodoTitle = '';
      newTodoDescription = '';
      newTodoPriority = 'medium';
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
        completedColumn.items = completedColumn.items.filter(t => t.id !== todoId);
        completedColumn = completedColumn;
      } else {
        const column = columns.find(c => c.id === columnId);
        if (column) {
          column.items = column.items.filter(t => t.id !== todoId);
          columns = columns;
        }
      }
      addToast('success', 'Todo deleted');
    } catch (e) {
      addToast('error', 'Failed to delete todo');
    }
  }

  // Quick actions for todos
  async function quickMarkComplete(todo: Todo, fromColumnId: string) {
    try {
      await api.updateTodo(todo.id, { status: 'completed' });
      
      // Remove from current column
      const fromColumn = columns.find(c => c.id === fromColumnId);
      if (fromColumn) {
        fromColumn.items = fromColumn.items.filter(t => t.id !== todo.id);
        columns = columns;
      }
      
      // Add to completed
      todo.status = 'completed';
      completedColumn.items = [todo, ...completedColumn.items];
      completedColumn = completedColumn;
      
      addToast('success', 'Marked as complete');
    } catch (e) {
      addToast('error', 'Failed to update todo');
    }
  }

  async function quickMarkStuck(todo: Todo, fromColumnId: string) {
    try {
      await api.updateTodo(todo.id, { status: 'stuck' });
      
      // Remove from current column
      const fromColumn = columns.find(c => c.id === fromColumnId);
      if (fromColumn) {
        fromColumn.items = fromColumn.items.filter(t => t.id !== todo.id);
        columns = columns;
      }
      
      // Add to stuck column
      const stuckColumn = columns.find(c => c.id === 'stuck');
      if (stuckColumn) {
        todo.status = 'stuck';
        stuckColumn.items = [todo, ...stuckColumn.items];
        columns = columns;
      }
      
      addToast('success', 'Moved to Stuck');
    } catch (e) {
      addToast('error', 'Failed to update todo');
    }
  }

  async function quickSetAccount(todo: Todo, accountId: string, columnId: string) {
    try {
      const account = accounts.find(a => a.id === accountId);
      await api.updateTodo(todo.id, { account_id: accountId || null });
      
      // Update local state
      const updateInColumn = (items: Todo[]) => {
        const item = items.find(t => t.id === todo.id);
        if (item) {
          item.account_id = accountId || undefined;
          item.account_name = account?.name || '';
        }
      };
      
      if (columnId === 'completed') {
        updateInColumn(completedColumn.items);
        completedColumn = completedColumn;
      } else {
        const column = columns.find(c => c.id === columnId);
        if (column) {
          updateInColumn(column.items);
          columns = columns;
        }
      }
      
      addToast('success', accountId ? `Tagged with ${account?.name}` : 'Account removed');
    } catch (e) {
      addToast('error', 'Failed to update account');
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
      await loadData(); // Reload to get updated linked_notes
      showLinkModal = false;
      addToast('success', 'Todo linked to note');
    } catch (e) {
      addToast('error', 'Failed to link todo');
    }
  }

  function getPriorityColor(priority: string): string {
    switch (priority) {
      case 'high': return 'border-l-red-500';
      case 'medium': return 'border-l-yellow-500';
      case 'low': return 'border-l-green-500';
      default: return 'border-l-gray-400';
    }
  }

  function getColumnColorClass(columnId: string): string {
    switch (columnId) {
      case 'not_started': return 'bg-gray-500';
      case 'in_progress': return 'bg-blue-500';
      case 'stuck': return 'bg-red-500';
      case 'completed': return 'bg-green-500';
      default: return 'bg-gray-400';
    }
  }

  function getColumnOutlineColor(columnId: string): string {
    switch (columnId) {
      case 'not_started': return '#6b7280'; // gray-500
      case 'in_progress': return '#3b82f6'; // blue-500
      case 'stuck': return '#ef4444'; // red-500
      case 'completed': return '#22c55e'; // green-500
      default: return '#9ca3af';
    }
  }

  function getPriorityIndicator(priority: string): { color: string; label: string } {
    switch (priority) {
      case 'high': return { color: '#ef4444', label: 'H' };
      case 'medium': return { color: '#f59e0b', label: 'M' };
      case 'low': return { color: '#22c55e', label: 'L' };
      default: return { color: '#6b7280', label: '-' };
    }
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric'
    });
  }

  // Bulk selection functions
  function toggleSelectionMode() {
    selectionMode = !selectionMode;
    if (!selectionMode) {
      selectedTodos = new Set();
    }
  }

  function toggleTodoSelection(todoId: string) {
    if (selectedTodos.has(todoId)) {
      selectedTodos.delete(todoId);
    } else {
      selectedTodos.add(todoId);
    }
    selectedTodos = selectedTodos; // Trigger reactivity
  }

  function selectAllInColumn(columnId: string) {
    const column = columns.find(c => c.id === columnId);
    if (column) {
      column.items.forEach(item => selectedTodos.add(item.id));
      selectedTodos = selectedTodos;
    }
  }

  function deselectAll() {
    selectedTodos = new Set();
  }

  async function bulkDelete() {
    if (selectedTodos.size === 0) return;
    if (!confirm(`Delete ${selectedTodos.size} todo(s)?`)) return;

    try {
      const promises = Array.from(selectedTodos).map(id => api.deleteTodo(id));
      await Promise.all(promises);
      
      // Remove from columns
      columns = columns.map(col => ({
        ...col,
        items: col.items.filter(t => !selectedTodos.has(t.id))
      }));
      
      addToast('success', `${selectedTodos.size} todo(s) deleted`);
      selectedTodos = new Set();
    } catch (e) {
      addToast('error', 'Failed to delete some todos');
    }
  }

  async function bulkMoveToStatus(newStatus: string) {
    if (selectedTodos.size === 0) return;

    try {
      const promises = Array.from(selectedTodos).map(id => 
        api.updateTodo(id, { status: newStatus })
      );
      await Promise.all(promises);
      
      // Move items between columns
      const movedItems: Todo[] = [];
      columns = columns.map(col => ({
        ...col,
        items: col.items.filter(t => {
          if (selectedTodos.has(t.id)) {
            t.status = newStatus;
            movedItems.push(t);
            return false;
          }
          return true;
        })
      }));
      
      // Add to target column
      const targetColumn = columns.find(c => c.id === newStatus);
      if (targetColumn) {
        targetColumn.items = [...targetColumn.items, ...movedItems];
        columns = columns;
      }
      
      addToast('success', `${selectedTodos.size} todo(s) moved`);
      selectedTodos = new Set();
    } catch (e) {
      addToast('error', 'Failed to move some todos');
    }
  }

  $: selectedCount = selectedTodos.size;

  // Filter and sort todos (high priority always at top)
  function filterAndSortTodos(todos: Todo[]): Todo[] {
    let filtered = todos;
    
    // Apply filters
    if (filterAccount) {
      filtered = filtered.filter(t => t.account_id === filterAccount);
    }
    if (filterPriority) {
      filtered = filtered.filter(t => t.priority === filterPriority);
    }
    if (filterHasNotes === 'yes') {
      filtered = filtered.filter(t => t.linked_notes && t.linked_notes.length > 0);
    } else if (filterHasNotes === 'no') {
      filtered = filtered.filter(t => !t.linked_notes || t.linked_notes.length === 0);
    }
    
    // Always sort high priority to top first, then apply secondary sort
    const priorityWeight = { high: 0, medium: 1, low: 2 };
    
    filtered.sort((a, b) => {
      // Primary sort: high priority always first
      const priorityDiff = (priorityWeight[a.priority] || 1) - (priorityWeight[b.priority] || 1);
      if (priorityDiff !== 0) return priorityDiff;
      
      // Secondary sort based on user selection
      let comparison = 0;
      switch (sortBy) {
        case 'title':
          comparison = a.title.localeCompare(b.title);
          break;
        case 'created_at':
        default:
          comparison = new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
          break;
      }
      return sortOrder === 'asc' ? -comparison : comparison;
    });
    
    return filtered;
  }

  function clearFilters() {
    filterAccount = '';
    filterPriority = '';
    filterHasNotes = '';
    sortBy = 'created_at';
    sortOrder = 'desc';
  }

  $: hasActiveFilters = filterAccount || filterPriority || filterHasNotes;
  
  // Apply filters to columns
  $: filteredColumns = columns.map(col => ({
    ...col,
    items: filterAndSortTodos(col.items)
  }));
  
  $: filteredCompletedColumn = {
    ...completedColumn,
    items: filterAndSortTodos(completedColumn.items)
  };

  // Get column color for drag outlines
  function getColumnColor(columnId: string): string {
    switch (columnId) {
      case 'not_started': return '#6b7280';
      case 'in_progress': return '#3b82f6';
      case 'stuck': return '#ef4444';
      case 'completed': return '#22c55e';
      default: return '#6b7280';
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
    <div class="flex items-center gap-3">
      <button 
        class="btn-secondary"
        class:bg-primary-500={showFilters}
        class:text-white={showFilters}
        on:click={() => showFilters = !showFilters}
      >
        <Filter class="w-4 h-4" />
        Filter
        {#if hasActiveFilters}
          <span class="ml-1 w-2 h-2 bg-orange-500 rounded-full"></span>
        {/if}
      </button>
      <button 
        class="btn-secondary"
        class:bg-primary-500={selectionMode}
        class:text-white={selectionMode}
        on:click={toggleSelectionMode}
      >
        <CheckSquare class="w-4 h-4" />
        {selectionMode ? 'Done Selecting' : 'Select'}
      </button>
      <button class="btn-primary" on:click={() => showNewTodoModal = true}>
        <Plus class="w-4 h-4" />
        New Todo
      </button>
    </div>
  </div>

  <!-- Filter Bar -->
  {#if showFilters}
    <div class="card mb-6 animate-slide-up">
      <div class="flex flex-wrap items-center gap-4">
        <div class="flex items-center gap-2">
          <label class="text-sm font-medium">Account:</label>
          <select class="input py-1.5 text-sm w-40" bind:value={filterAccount}>
            <option value="">All</option>
            {#each accounts as account}
              <option value={account.id}>{account.name}</option>
            {/each}
          </select>
        </div>
        <div class="flex items-center gap-2">
          <label class="text-sm font-medium">Priority:</label>
          <select class="input py-1.5 text-sm w-32" bind:value={filterPriority}>
            <option value="">All</option>
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
          </select>
        </div>
        <div class="flex items-center gap-2">
          <label class="text-sm font-medium">Linked Notes:</label>
          <select class="input py-1.5 text-sm w-32" bind:value={filterHasNotes}>
            <option value="">All</option>
            <option value="yes">Has Notes</option>
            <option value="no">No Notes</option>
          </select>
        </div>
        <div class="h-6 w-px bg-[var(--color-border)]"></div>
        <div class="flex items-center gap-2">
          <label class="text-sm font-medium">Sort:</label>
          <select class="input py-1.5 text-sm w-32" bind:value={sortBy}>
            <option value="created_at">Date Created</option>
            <option value="priority">Priority</option>
            <option value="title">Title</option>
          </select>
          <button 
            class="btn-ghost p-1.5"
            on:click={() => sortOrder = sortOrder === 'asc' ? 'desc' : 'asc'}
          >
            {#if sortOrder === 'asc'}
              <SortAsc class="w-4 h-4" />
            {:else}
              <SortDesc class="w-4 h-4" />
            {/if}
          </button>
        </div>
        {#if hasActiveFilters}
          <button class="btn-ghost text-sm text-red-500" on:click={clearFilters}>
            Clear Filters
          </button>
        {/if}
      </div>
    </div>
  {/if}

  <!-- Bulk Actions Bar -->
  {#if selectionMode && selectedCount > 0}
    <div class="fixed bottom-6 left-1/2 transform -translate-x-1/2 z-40 bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl shadow-2xl p-4 flex items-center gap-4 animate-slide-up">
      <span class="text-sm font-medium">{selectedCount} selected</span>
      <div class="h-6 w-px bg-[var(--color-border)]"></div>
      <div class="flex items-center gap-2">
        <button 
          class="btn-secondary text-sm py-1.5"
          on:click={() => bulkMoveToStatus('not_started')}
        >
          <ArrowRight class="w-4 h-4" />
          Not Started
        </button>
        <button 
          class="btn-secondary text-sm py-1.5"
          on:click={() => bulkMoveToStatus('in_progress')}
        >
          <ArrowRight class="w-4 h-4" />
          In Progress
        </button>
        <button 
          class="btn-secondary text-sm py-1.5"
          on:click={() => bulkMoveToStatus('completed')}
        >
          <ArrowRight class="w-4 h-4" />
          Completed
        </button>
      </div>
      <div class="h-6 w-px bg-[var(--color-border)]"></div>
      <button 
        class="btn text-sm py-1.5 bg-red-500/10 text-red-500 hover:bg-red-500/20"
        on:click={bulkDelete}
      >
        <Trash2 class="w-4 h-4" />
        Delete
      </button>
      <button 
        class="btn-ghost p-1.5"
        on:click={deselectAll}
      >
        <X class="w-4 h-4" />
      </button>
    </div>
  {/if}

  {#if loading}
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      {#each [1, 2, 3] as _}
        <div class="card animate-pulse">
          <div class="h-6 bg-[var(--color-border)] rounded w-32 mb-4"></div>
          <div class="space-y-3">
            {#each [1, 2] as __}
              <div class="h-24 bg-[var(--color-border)] rounded"></div>
            {/each}
          </div>
        </div>
      {/each}
    </div>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      {#each filteredColumns as column (column.id)}
        <div class="flex flex-col">
          <!-- Column Header -->
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-2">
              <div class="w-3 h-3 rounded-full {getColumnColorClass(column.id)}"></div>
              <h2 class="font-semibold">{column.title}</h2>
              <span class="text-sm text-[var(--color-muted)]">({column.items.length})</span>
            </div>
            {#if selectionMode && column.items.length > 0}
              <button 
                class="text-xs text-primary-500 hover:underline"
                on:click={() => selectAllInColumn(column.id)}
              >
                Select all
              </button>
            {/if}
          </div>

          <!-- Column Content -->
          <div 
            class="flex-1 min-h-[400px] p-3 bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl"
            use:dndzone={{
              items: column.items,
              flipDurationMs,
              dropTargetStyle: { outline: `2px dashed ${getColumnOutlineColor(column.id)}`, outlineOffset: '-2px' }
            }}
            on:consider={(e) => handleDndConsider(column.id, e)}
            on:finalize={(e) => handleDndFinalize(column.id, e)}
          >
            {#each column.items as todo (todo.id)}
              <div 
                class="bg-[var(--color-bg)] border border-[var(--color-border)] border-l-4 {getPriorityColor(todo.priority)} rounded-lg p-4 mb-3 group"
                class:cursor-grab={!selectionMode}
                class:cursor-pointer={selectionMode}
                class:ring-2={selectedTodos.has(todo.id)}
                class:ring-primary-500={selectedTodos.has(todo.id)}
                animate:flip={{ duration: flipDurationMs }}
                on:click={() => selectionMode && toggleTodoSelection(todo.id)}
              >
                <div class="flex items-start justify-between mb-2">
                  <div class="flex items-center gap-2">
                    {#if selectionMode}
                      <button 
                        class="p-0.5"
                        on:click|stopPropagation={() => toggleTodoSelection(todo.id)}
                      >
                        {#if selectedTodos.has(todo.id)}
                          <CheckSquare class="w-4 h-4 text-primary-500" />
                        {:else}
                          <Square class="w-4 h-4 text-[var(--color-muted)]" />
                        {/if}
                      </button>
                    {:else}
                      <GripVertical class="w-4 h-4 text-[var(--color-muted)] opacity-0 group-hover:opacity-100 transition-opacity" />
                    {/if}
                    <h3 class="font-medium text-sm">{todo.title}</h3>
                  </div>
                  <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    <!-- Quick complete button (show on in_progress and not_started) -->
                    {#if column.id === 'in_progress' || column.id === 'not_started'}
                      <button 
                        class="p-1 hover:bg-green-500/20 rounded"
                        title="Mark complete"
                        on:click|stopPropagation={() => quickMarkComplete(todo, column.id)}
                      >
                        <Check class="w-3.5 h-3.5 text-green-500" />
                      </button>
                    {/if}
                    <!-- Quick stuck button (show on in_progress) -->
                    {#if column.id === 'in_progress'}
                      <button 
                        class="p-1 hover:bg-red-500/20 rounded"
                        title="Mark as stuck"
                        on:click|stopPropagation={() => quickMarkStuck(todo, column.id)}
                      >
                        <AlertTriangle class="w-3.5 h-3.5 text-red-500" />
                      </button>
                    {/if}
                    <button 
                      class="p-1 hover:bg-[var(--color-border)] rounded"
                      title="Link to note"
                      on:click|stopPropagation={() => openLinkModal(todo.id)}
                    >
                      <Link class="w-3.5 h-3.5 text-[var(--color-muted)]" />
                    </button>
                    <button 
                      class="p-1 hover:bg-[var(--color-border)] rounded"
                      on:click|stopPropagation={() => deleteTodo(todo.id, column.id)}
                    >
                      <Trash2 class="w-3.5 h-3.5 text-red-500" />
                    </button>
                  </div>
                </div>
                
                {#if todo.description}
                  <p class="text-xs text-[var(--color-muted)] mb-3 line-clamp-2">{todo.description}</p>
                {/if}

                <!-- Account Tag / Quick Selector -->
                <div class="mb-2">
                  <div class="relative inline-block group/account">
                    <button
                      class="inline-flex items-center gap-1 px-2 py-0.5 text-xs rounded transition-colors {todo.account_name ? 'bg-orange-500/10 text-orange-600 dark:text-orange-400' : 'bg-[var(--color-border)] text-[var(--color-muted)] hover:bg-[var(--color-border)]/80'}"
                      on:click|stopPropagation
                    >
                      <Building2 class="w-3 h-3" />
                      {todo.account_name || 'Add account'}
                      <ChevronDown class="w-3 h-3" />
                    </button>
                    <div class="absolute left-0 top-full mt-1 z-10 hidden group-hover/account:block">
                      <div class="bg-[var(--color-card)] border border-[var(--color-border)] rounded-lg shadow-xl py-1 min-w-[140px]">
                        <button
                          class="w-full px-3 py-1.5 text-xs text-left hover:bg-[var(--color-bg)] transition-colors {!todo.account_id ? 'text-primary-500 font-medium' : ''}"
                          on:click|stopPropagation={() => quickSetAccount(todo, '', column.id)}
                        >
                          No account
                        </button>
                        {#each accounts as account}
                          <button
                            class="w-full px-3 py-1.5 text-xs text-left hover:bg-[var(--color-bg)] transition-colors {todo.account_id === account.id ? 'text-primary-500 font-medium' : ''}"
                            on:click|stopPropagation={() => quickSetAccount(todo, account.id, column.id)}
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
                  <div class="flex flex-wrap gap-1 mb-2">
                    {#each todo.linked_notes as note}
                      <a 
                        href="/notes/{note.id}"
                        class="inline-flex items-center gap-1 px-2 py-0.5 text-xs bg-primary-500/10 text-primary-500 rounded hover:bg-primary-500/20 transition-colors"
                        on:click|stopPropagation
                      >
                        <FileText class="w-3 h-3" />
                        {note.title}
                      </a>
                    {/each}
                  </div>
                {/if}

                <div class="flex items-center justify-between text-xs text-[var(--color-muted)]">
                  <span 
                    class="w-5 h-5 rounded flex items-center justify-center text-[10px] font-bold text-white"
                    style="background-color: {getPriorityIndicator(todo.priority).color}"
                    title="{todo.priority} priority"
                  >
                    {getPriorityIndicator(todo.priority).label}
                  </span>
                  {#if todo.due_date}
                    <span class="flex items-center gap-1">
                      <Calendar class="w-3 h-3" />
                      {formatDate(todo.due_date)}
                    </span>
                  {/if}
                </div>
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

    <!-- Completed Section (Full Width) -->
    <div class="mt-8">
      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center gap-2">
          <div class="w-3 h-3 rounded-full bg-green-500"></div>
          <h2 class="font-semibold">Completed</h2>
          <span class="text-sm text-[var(--color-muted)]">({filteredCompletedColumn.items.length})</span>
        </div>
        {#if selectionMode && filteredCompletedColumn.items.length > 0}
          <button 
            class="text-xs text-primary-500 hover:underline"
            on:click={() => selectAllInColumn('completed')}
          >
            Select all
          </button>
        {/if}
      </div>
      
      <div 
        class="min-h-[120px] p-3 bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl"
        use:dndzone={{
          items: filteredCompletedColumn.items,
          flipDurationMs,
          dropTargetStyle: { outline: '2px dashed #22c55e', outlineOffset: '-2px' }
        }}
        on:consider={(e) => handleDndConsider('completed', e)}
        on:finalize={(e) => handleDndFinalize('completed', e)}
      >
        {#if filteredCompletedColumn.items.length > 0}
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
            {#each filteredCompletedColumn.items as todo (todo.id)}
              <div 
                class="bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg p-3 group opacity-70 hover:opacity-100 transition-opacity"
                class:cursor-grab={!selectionMode}
                class:cursor-pointer={selectionMode}
                class:ring-2={selectedTodos.has(todo.id)}
                class:ring-primary-500={selectedTodos.has(todo.id)}
                animate:flip={{ duration: flipDurationMs }}
                on:click={() => selectionMode && toggleTodoSelection(todo.id)}
              >
                <div class="flex items-start justify-between">
                  <div class="flex items-center gap-2 min-w-0">
                    {#if selectionMode}
                      <button 
                        class="p-0.5 flex-shrink-0"
                        on:click|stopPropagation={() => toggleTodoSelection(todo.id)}
                      >
                        {#if selectedTodos.has(todo.id)}
                          <CheckSquare class="w-4 h-4 text-primary-500" />
                        {:else}
                          <Square class="w-4 h-4 text-[var(--color-muted)]" />
                        {/if}
                      </button>
                    {/if}
                    <span class="font-medium text-sm line-through truncate">{todo.title}</span>
                  </div>
                  <button 
                    class="p-1 hover:bg-[var(--color-border)] rounded opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0"
                    on:click|stopPropagation={() => deleteTodo(todo.id, 'completed')}
                  >
                    <Trash2 class="w-3.5 h-3.5 text-red-500" />
                  </button>
                </div>
                {#if todo.account_name}
                  <div class="mt-2">
                    <span class="inline-flex items-center gap-1 px-2 py-0.5 text-xs bg-orange-500/10 text-orange-600 dark:text-orange-400 rounded">
                      <Building2 class="w-3 h-3" />
                      {todo.account_name}
                    </span>
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        {:else}
          <div class="flex items-center justify-center h-20 text-[var(--color-muted)] text-sm">
            Completed items will appear here
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
          <label class="label">Priority</label>
          <select class="input" bind:value={newTodoPriority}>
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
          </select>
        </div>
        <div class="mb-4">
          <label class="label">Account Tag (optional)</label>
          <select class="input" bind:value={newTodoAccountId}>
            <option value="">No account</option>
            {#each accounts as account}
              <option value={account.id}>{account.name}</option>
            {/each}
          </select>
        </div>
        <div class="flex justify-end gap-3">
          <button 
            type="button"
            class="btn-secondary"
            on:click={() => showNewTodoModal = false}
          >
            Cancel
          </button>
          <button type="submit" class="btn-primary" disabled={!newTodoTitle.trim()}>
            Create Todo
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
        <p class="text-[var(--color-muted)]">No notes available. Create a note first.</p>
      {:else}
        <div class="max-h-80 overflow-y-auto space-y-2">
          {#each availableNotes as note}
            <button 
              class="w-full flex items-center gap-3 p-3 bg-[var(--color-bg)] rounded-lg hover:bg-[var(--color-border)] transition-colors text-left"
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
        <button 
          class="btn-secondary"
          on:click={() => showLinkModal = false}
        >
          Cancel
        </button>
      </div>
    </div>
  </div>
{/if}
