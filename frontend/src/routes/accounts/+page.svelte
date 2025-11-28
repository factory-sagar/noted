<script lang="ts">
  import { onMount } from 'svelte';
  import { 
    Building2, 
    FileText, 
    CheckSquare, 
    Plus,
    ChevronRight,
    Users,
    DollarSign,
    Trash2,
    Edit3,
    Search
  } from 'lucide-svelte';
  import { api, type Account, type Note, type Todo } from '$lib/utils/api';
  import { addToast } from '$lib/stores';

  let accounts: Account[] = [];
  let notes: Note[] = [];
  let todos: Todo[] = [];
  let loading = true;
  let searchQuery = '';
  let selectedAccountId: string | null = null;
  
  // Modal states
  let showNewAccountModal = false;
  let showEditAccountModal = false;
  let editingAccount: Account | null = null;
  let newAccountName = '';
  let newAccountOwner = '';

  onMount(async () => {
    await loadData();
  });

  async function loadData() {
    try {
      loading = true;
      const [accountsData, notesData, todosData] = await Promise.all([
        api.getAccounts(),
        api.getNotes(),
        api.getTodos()
      ]);
      accounts = accountsData;
      notes = notesData;
      todos = todosData;
    } catch (e) {
      addToast('error', 'Failed to load data');
    } finally {
      loading = false;
    }
  }

  function getNotesForAccount(accountId: string): Note[] {
    return notes.filter(n => n.account_id === accountId);
  }

  function getTodosForAccount(accountId: string): Todo[] {
    return todos.filter(t => t.account_id === accountId);
  }

  function getAccountStats(accountId: string) {
    const accountNotes = getNotesForAccount(accountId);
    const accountTodos = getTodosForAccount(accountId);
    const completedTodos = accountTodos.filter(t => t.status === 'completed').length;
    
    return {
      noteCount: accountNotes.length,
      todoCount: accountTodos.length,
      completedTodos,
      pendingTodos: accountTodos.length - completedTodos
    };
  }

  async function createAccount() {
    if (!newAccountName.trim()) return;
    try {
      const account = await api.createAccount({ 
        name: newAccountName.trim(),
        account_owner: newAccountOwner.trim() || undefined
      });
      accounts = [...accounts, account];
      newAccountName = '';
      newAccountOwner = '';
      showNewAccountModal = false;
      addToast('success', 'Account created');
    } catch (e) {
      addToast('error', 'Failed to create account');
    }
  }

  async function updateAccount() {
    if (!editingAccount || !editingAccount.name.trim()) return;
    try {
      await api.updateAccount(editingAccount.id, {
        name: editingAccount.name,
        account_owner: editingAccount.account_owner
      });
      accounts = accounts.map(a => a.id === editingAccount!.id ? editingAccount! : a);
      showEditAccountModal = false;
      editingAccount = null;
      addToast('success', 'Account updated');
    } catch (e) {
      addToast('error', 'Failed to update account');
    }
  }

  async function deleteAccount(accountId: string) {
    const stats = getAccountStats(accountId);
    if (!confirm(`Delete this account? This will also delete ${stats.noteCount} notes and ${stats.todoCount} todos.`)) return;
    try {
      await api.deleteAccount(accountId);
      accounts = accounts.filter(a => a.id !== accountId);
      notes = notes.filter(n => n.account_id !== accountId);
      todos = todos.filter(t => t.account_id !== accountId);
      if (selectedAccountId === accountId) {
        selectedAccountId = null;
      }
      addToast('success', 'Account deleted');
    } catch (e) {
      addToast('error', 'Failed to delete account');
    }
  }

  function openEditModal(account: Account) {
    editingAccount = { ...account };
    showEditAccountModal = true;
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric'
    });
  }

  function getStatusColor(status: string): string {
    switch (status) {
      case 'completed': return 'bg-green-500';
      case 'in_progress': return 'bg-blue-500';
      default: return 'bg-gray-400';
    }
  }

  $: filteredAccounts = searchQuery
    ? accounts.filter(a => 
        a.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        (a.account_owner && a.account_owner.toLowerCase().includes(searchQuery.toLowerCase()))
      )
    : accounts;

  $: selectedAccount = selectedAccountId ? accounts.find(a => a.id === selectedAccountId) : null;
  $: selectedNotes = selectedAccountId ? getNotesForAccount(selectedAccountId) : [];
  $: selectedTodos = selectedAccountId ? getTodosForAccount(selectedAccountId) : [];
</script>

<svelte:head>
  <title>Accounts - SE Notes</title>
</svelte:head>

<div class="max-w-7xl mx-auto">
  <div class="flex items-center justify-between mb-8">
    <div>
      <h1 class="page-title">Accounts</h1>
      <p class="page-subtitle">Manage customer accounts and their data</p>
    </div>
    <button class="btn-primary" on:click={() => showNewAccountModal = true}>
      <Plus class="w-4 h-4" />
      New Account
    </button>
  </div>

  <!-- Search -->
  <div class="mb-6">
    <div class="relative">
      <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-[var(--color-muted)]" />
      <input 
        type="text"
        placeholder="Search accounts..."
        class="input pl-10"
        bind:value={searchQuery}
      />
    </div>
  </div>

  {#if loading}
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div class="lg:col-span-1">
        <div class="card animate-pulse">
          <div class="space-y-4">
            {#each [1, 2, 3] as _}
              <div class="h-20 bg-[var(--color-border)] rounded"></div>
            {/each}
          </div>
        </div>
      </div>
      <div class="lg:col-span-2">
        <div class="card animate-pulse h-96"></div>
      </div>
    </div>
  {:else}
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Accounts List -->
      <div class="lg:col-span-1">
        <div class="card p-2">
          {#if filteredAccounts.length === 0}
            <div class="text-center py-8 text-[var(--color-muted)]">
              {searchQuery ? 'No accounts found' : 'No accounts yet'}
            </div>
          {:else}
            <div class="space-y-1">
              {#each filteredAccounts as account (account.id)}
                {@const stats = getAccountStats(account.id)}
                <button
                  class="w-full flex items-center justify-between p-3 rounded-lg transition-colors text-left group hover:bg-[var(--color-bg)] {selectedAccountId === account.id ? 'bg-primary-100 dark:bg-primary-900/20' : ''}"
                  on:click={() => selectedAccountId = account.id}
                >
                  <div class="flex items-center gap-3 min-w-0">
                    <div class="w-10 h-10 rounded-lg bg-primary-500/10 flex items-center justify-center flex-shrink-0">
                      <Building2 class="w-5 h-5 text-primary-500" />
                    </div>
                    <div class="min-w-0">
                      <p class="font-medium truncate">{account.name}</p>
                      <p class="text-xs text-[var(--color-muted)]">
                        {stats.noteCount} notes · {stats.todoCount} todos
                      </p>
                    </div>
                  </div>
                  <ChevronRight class="w-4 h-4 text-[var(--color-muted)] opacity-0 group-hover:opacity-100" />
                </button>
              {/each}
            </div>
          {/if}
        </div>
      </div>

      <!-- Account Detail -->
      <div class="lg:col-span-2">
        {#if selectedAccount}
          {@const stats = getAccountStats(selectedAccount.id)}
          <div class="space-y-6">
            <!-- Account Header -->
            <div class="card">
              <div class="flex items-start justify-between">
                <div class="flex items-center gap-4">
                  <div class="w-14 h-14 rounded-xl bg-primary-500/10 flex items-center justify-center">
                    <Building2 class="w-7 h-7 text-primary-500" />
                  </div>
                  <div>
                    <h2 class="text-xl font-semibold">{selectedAccount.name}</h2>
                    {#if selectedAccount.account_owner}
                      <p class="text-[var(--color-muted)] flex items-center gap-1">
                        <Users class="w-4 h-4" />
                        {selectedAccount.account_owner}
                      </p>
                    {/if}
                  </div>
                </div>
                <div class="flex items-center gap-2">
                  <button 
                    class="btn-ghost p-2"
                    on:click={() => openEditModal(selectedAccount)}
                  >
                    <Edit3 class="w-4 h-4" />
                  </button>
                  <button 
                    class="btn-ghost p-2 text-red-500"
                    on:click={() => deleteAccount(selectedAccount.id)}
                  >
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </div>

              <!-- Stats -->
              <div class="grid grid-cols-4 gap-4 mt-6">
                <div class="text-center p-3 bg-[var(--color-bg)] rounded-lg">
                  <p class="text-2xl font-bold">{stats.noteCount}</p>
                  <p class="text-xs text-[var(--color-muted)]">Notes</p>
                </div>
                <div class="text-center p-3 bg-[var(--color-bg)] rounded-lg">
                  <p class="text-2xl font-bold">{stats.todoCount}</p>
                  <p class="text-xs text-[var(--color-muted)]">Todos</p>
                </div>
                <div class="text-center p-3 bg-[var(--color-bg)] rounded-lg">
                  <p class="text-2xl font-bold text-green-500">{stats.completedTodos}</p>
                  <p class="text-xs text-[var(--color-muted)]">Completed</p>
                </div>
                <div class="text-center p-3 bg-[var(--color-bg)] rounded-lg">
                  <p class="text-2xl font-bold text-orange-500">{stats.pendingTodos}</p>
                  <p class="text-xs text-[var(--color-muted)]">Pending</p>
                </div>
              </div>

              {#if selectedAccount.budget || selectedAccount.est_engineers}
                <div class="flex gap-4 mt-4 pt-4 border-t border-[var(--color-border)]">
                  {#if selectedAccount.budget}
                    <div class="flex items-center gap-2 text-sm">
                      <DollarSign class="w-4 h-4 text-[var(--color-muted)]" />
                      <span>Budget: ${selectedAccount.budget.toLocaleString()}</span>
                    </div>
                  {/if}
                  {#if selectedAccount.est_engineers}
                    <div class="flex items-center gap-2 text-sm">
                      <Users class="w-4 h-4 text-[var(--color-muted)]" />
                      <span>Est. Engineers: {selectedAccount.est_engineers}</span>
                    </div>
                  {/if}
                </div>
              {/if}
            </div>

            <!-- Notes Section -->
            <div class="card">
              <div class="flex items-center justify-between mb-4">
                <h3 class="font-semibold flex items-center gap-2">
                  <FileText class="w-5 h-5" />
                  Notes ({selectedNotes.length})
                </h3>
                <a href="/notes" class="text-sm text-primary-500 hover:underline">
                  View All
                </a>
              </div>
              {#if selectedNotes.length === 0}
                <p class="text-[var(--color-muted)] text-sm py-4 text-center">No notes for this account yet.</p>
              {:else}
                <div class="space-y-2">
                  {#each selectedNotes.slice(0, 5) as note}
                    <a 
                      href="/notes/{note.id}"
                      class="flex items-center justify-between p-3 bg-[var(--color-bg)] rounded-lg hover:bg-[var(--color-border)] transition-colors"
                    >
                      <div class="flex items-center gap-3">
                        <FileText class="w-4 h-4 text-primary-500" />
                        <div>
                          <p class="font-medium text-sm">{note.title}</p>
                          <p class="text-xs text-[var(--color-muted)]">
                            {note.template_type} · {formatDate(note.created_at)}
                          </p>
                        </div>
                      </div>
                      <ChevronRight class="w-4 h-4 text-[var(--color-muted)]" />
                    </a>
                  {/each}
                  {#if selectedNotes.length > 5}
                    <p class="text-sm text-[var(--color-muted)] text-center py-2">
                      +{selectedNotes.length - 5} more notes
                    </p>
                  {/if}
                </div>
              {/if}
            </div>

            <!-- Todos Section -->
            <div class="card">
              <div class="flex items-center justify-between mb-4">
                <h3 class="font-semibold flex items-center gap-2">
                  <CheckSquare class="w-5 h-5" />
                  Todos ({selectedTodos.length})
                </h3>
                <a href="/todos" class="text-sm text-primary-500 hover:underline">
                  View All
                </a>
              </div>
              {#if selectedTodos.length === 0}
                <p class="text-[var(--color-muted)] text-sm py-4 text-center">No todos for this account yet.</p>
              {:else}
                <div class="space-y-2">
                  {#each selectedTodos.slice(0, 5) as todo}
                    <div class="flex items-center justify-between p-3 bg-[var(--color-bg)] rounded-lg">
                      <div class="flex items-center gap-3">
                        <div class="w-2 h-2 rounded-full {getStatusColor(todo.status)}"></div>
                        <div>
                          <p class="font-medium text-sm" class:line-through={todo.status === 'completed'}>
                            {todo.title}
                          </p>
                          <p class="text-xs text-[var(--color-muted)] capitalize">
                            {todo.status.replace('_', ' ')} · {todo.priority} priority
                          </p>
                        </div>
                      </div>
                    </div>
                  {/each}
                  {#if selectedTodos.length > 5}
                    <p class="text-sm text-[var(--color-muted)] text-center py-2">
                      +{selectedTodos.length - 5} more todos
                    </p>
                  {/if}
                </div>
              {/if}
            </div>
          </div>
        {:else}
          <div class="card h-96 flex items-center justify-center text-[var(--color-muted)]">
            <div class="text-center">
              <Building2 class="w-12 h-12 mx-auto mb-4 opacity-50" />
              <p>Select an account to view details</p>
            </div>
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<!-- New Account Modal -->
{#if showNewAccountModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button 
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => showNewAccountModal = false}
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-md animate-slide-up">
      <h2 class="text-lg font-semibold mb-4">New Account</h2>
      <form on:submit|preventDefault={createAccount}>
        <div class="mb-4">
          <label class="label">Account Name</label>
          <input 
            type="text"
            class="input"
            placeholder="e.g., Acme Corp"
            bind:value={newAccountName}
            autofocus
          />
        </div>
        <div class="mb-4">
          <label class="label">Account Owner (optional)</label>
          <input 
            type="text"
            class="input"
            placeholder="Sales rep name"
            bind:value={newAccountOwner}
          />
        </div>
        <div class="flex justify-end gap-3">
          <button 
            type="button"
            class="btn-secondary"
            on:click={() => showNewAccountModal = false}
          >
            Cancel
          </button>
          <button type="submit" class="btn-primary" disabled={!newAccountName.trim()}>
            Create Account
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- Edit Account Modal -->
{#if showEditAccountModal && editingAccount}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button 
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => showEditAccountModal = false}
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-md animate-slide-up">
      <h2 class="text-lg font-semibold mb-4">Edit Account</h2>
      <form on:submit|preventDefault={updateAccount}>
        <div class="mb-4">
          <label class="label">Account Name</label>
          <input 
            type="text"
            class="input"
            bind:value={editingAccount.name}
            autofocus
          />
        </div>
        <div class="mb-4">
          <label class="label">Account Owner</label>
          <input 
            type="text"
            class="input"
            bind:value={editingAccount.account_owner}
          />
        </div>
        <div class="flex justify-end gap-3">
          <button 
            type="button"
            class="btn-secondary"
            on:click={() => showEditAccountModal = false}
          >
            Cancel
          </button>
          <button type="submit" class="btn-primary">
            Save Changes
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
