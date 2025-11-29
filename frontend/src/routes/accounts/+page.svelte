<script lang="ts">
  import { onMount } from 'svelte';
  import { 
    Building2, 
    FileText, 
    CheckSquare, 
    Plus,
    ChevronRight,
    ChevronDown,
    Users,
    DollarSign,
    Trash2,
    Trash,
    RefreshCw,
    Edit3,
    Search
  } from 'lucide-svelte';
  import { api, type Account, type Note, type Todo } from '$lib/utils/api';
  import { addToast } from '$lib/stores';

  let accounts: Account[] = [];
  let deletedAccounts: Account[] = [];
  let notes: Note[] = [];
  let todos: Todo[] = [];
  let loading = true;
  let showTrash = false;
  let searchQuery = '';
  let selectedAccountId: string | null = null;
  
  let showNewAccountModal = false;
  let showEditAccountModal = false;
  let editingAccount: Account | null = null;
  let newAccountName = '';
  let newAccountOwner = '';
  let deletingAccountId: string | null = null;
  let permanentDeletingAccountId: string | null = null;

  onMount(async () => {
    await loadData();
  });

  async function loadData() {
    try {
      loading = true;
      
      // Load each resource independently to prevent partial failures
      const [accountsResult, deletedResult, notesResult, todosResult] = await Promise.allSettled([
        api.getAccounts(),
        api.getDeletedAccounts(),
        api.getNotes(),
        api.getTodos()
      ]);

      if (accountsResult.status === 'fulfilled') accounts = accountsResult.value;
      if (deletedResult.status === 'fulfilled') deletedAccounts = deletedResult.value;
      if (notesResult.status === 'fulfilled') notes = notesResult.value;
      if (todosResult.status === 'fulfilled') todos = todosResult.value;

      // If critical data fails, show error but try to display what we have
      if (accountsResult.status === 'rejected' || notesResult.status === 'rejected') {
        console.error('Failed to load critical data');
        addToast('error', 'Some data failed to load');
      }
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

  async function confirmDeleteAccount() {
    if (!deletingAccountId) return;
    
    try {
      await api.deleteAccount(deletingAccountId);
      
      const deletedAccount = accounts.find(a => a.id === deletingAccountId);
      accounts = accounts.filter(a => a.id !== deletingAccountId);
      notes = notes.filter(n => n.account_id !== deletingAccountId);
      todos = todos.filter(t => t.account_id !== deletingAccountId);
      
      if (deletedAccount) {
        deletedAccounts = [deletedAccount, ...deletedAccounts];
      }
      
      if (selectedAccountId === deletingAccountId) {
        selectedAccountId = null;
      }
      
      addToast('success', 'Account deleted');
      deletingAccountId = null;
    } catch (e) {
      addToast('error', 'Failed to delete account');
    }
  }

  async function restoreAccount(accountId: string) {
    try {
      const account = deletedAccounts.find(a => a.id === accountId);
      // Optimistic update
      deletedAccounts = deletedAccounts.filter(a => a.id !== accountId);
      if (account) {
        accounts = [...accounts, account];
      }
      addToast('success', 'Account restored');
      
      await api.restoreAccount(accountId);
    } catch (e) {
      addToast('error', 'Failed to restore account');
      await loadData(); // Revert
    }
  }

  async function confirmPermanentDeleteAccount() {
    if (!permanentDeletingAccountId) return;
    try {
      await api.permanentDeleteAccount(permanentDeletingAccountId);
      deletedAccounts = deletedAccounts.filter(a => a.id !== permanentDeletingAccountId);
      addToast('success', 'Permanently deleted');
      permanentDeletingAccountId = null;
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
      case 'completed': return 'bg-emerald-500';
      case 'in_progress': return 'bg-blue-500';
      case 'stuck': return 'bg-red-500';
      default: return 'bg-charcoal-400';
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
  <title>Accounts - Noted</title>
</svelte:head>

<div class="max-w-6xl mx-auto">
  <!-- Header -->
  <div class="flex items-start justify-between mb-12">
    <div class="page-header mb-0">
      <div class="divider-accent mb-6"></div>
      <h1 class="page-title">Accounts</h1>
      <p class="page-subtitle">Manage customer accounts</p>
    </div>
    <button class="btn-primary" on:click={() => showNewAccountModal = true}>
      <Plus class="w-4 h-4" strokeWidth={1.5} />
      New Account
    </button>
  </div>

  <!-- Search -->
  <div class="mb-8">
    <div class="relative">
      <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-[var(--color-muted)]" strokeWidth={1.5} />
      <input 
        type="text"
        placeholder="Search accounts..."
        class="input pl-12"
        bind:value={searchQuery}
      />
    </div>
  </div>

  {#if loading}
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <div class="lg:col-span-1">
        <div class="card">
          <div class="space-y-4">
            {#each [1, 2, 3] as _}
              <div class="skeleton h-20"></div>
            {/each}
          </div>
        </div>
      </div>
      <div class="lg:col-span-2">
        <div class="card skeleton h-96"></div>
      </div>
    </div>
  {:else}
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <!-- Accounts List -->
      <div class="lg:col-span-1">
        <div class="card p-3">
          {#if filteredAccounts.length === 0}
            <div class="text-center py-12 text-[var(--color-muted)]">
              {searchQuery ? 'No accounts found' : 'No accounts yet'}
            </div>
          {:else}
            <div class="space-y-1">
              {#each filteredAccounts as account (account.id)}
                {@const stats = getAccountStats(account.id)}
                <button
                  class="w-full flex items-center justify-between p-4 transition-all text-left group hover:bg-[var(--color-card-hover)] {selectedAccountId === account.id ? 'bg-[var(--color-accent)]/5 border-l-2 border-l-[var(--color-accent)]' : ''}"
                  style="border-radius: 2px;"
                  on:click={() => selectedAccountId = account.id}
                >
                  <div class="flex items-center gap-4 min-w-0">
                    <div class="w-10 h-10 flex items-center justify-center bg-[var(--color-bg)] border border-[var(--color-border)] flex-shrink-0" style="border-radius: 2px;">
                      <Building2 class="w-5 h-5 text-[var(--color-accent)]" strokeWidth={1.5} />
                    </div>
                    <div class="min-w-0">
                      <p class="font-medium truncate">{account.name}</p>
                      <p class="text-xs text-[var(--color-muted)]">
                        {stats.noteCount} notes · {stats.todoCount} todos
                      </p>
                    </div>
                  </div>
                  <ChevronRight class="w-4 h-4 text-[var(--color-muted)] opacity-0 group-hover:opacity-100 flex-shrink-0" strokeWidth={1.5} />
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
          <div class="space-y-6 animate-stagger">
            <!-- Account Header -->
            <div class="card">
              <div class="flex items-start justify-between mb-8">
                <div class="flex items-center gap-5">
                  <div class="w-16 h-16 flex items-center justify-center bg-[var(--color-bg)] border border-[var(--color-border)]" style="border-radius: 2px;">
                    <Building2 class="w-8 h-8 text-[var(--color-accent)]" strokeWidth={1.5} />
                  </div>
                  <div>
                    <h2 class="font-serif text-2xl tracking-tight">{selectedAccount.name}</h2>
                    {#if selectedAccount.account_owner}
                      <p class="text-[var(--color-muted)] flex items-center gap-2 mt-1">
                        <Users class="w-4 h-4" strokeWidth={1.5} />
                        {selectedAccount.account_owner}
                      </p>
                    {/if}
                  </div>
                </div>
                <div class="flex items-center gap-2">
                  <button 
                    class="btn-icon btn-icon-ghost"
                    on:click={() => openEditModal(selectedAccount)}
                  >
                    <Edit3 class="w-4 h-4" strokeWidth={1.5} />
                  </button>
                  <button 
                    class="btn-icon btn-icon-danger"
                    on:click={() => deletingAccountId = selectedAccount.id}
                  >
                    <Trash2 class="w-4 h-4" strokeWidth={1.5} />
                  </button>
                </div>
              </div>

              <!-- Stats -->
              <div class="grid grid-cols-4 gap-4">
                <div class="text-center p-4 bg-[var(--color-bg)] border border-[var(--color-border)]" style="border-radius: 2px;">
                  <p class="font-serif text-2xl">{stats.noteCount}</p>
                  <p class="text-xs text-[var(--color-muted)] uppercase tracking-wider mt-1">Notes</p>
                </div>
                <div class="text-center p-4 bg-[var(--color-bg)] border border-[var(--color-border)]" style="border-radius: 2px;">
                  <p class="font-serif text-2xl">{stats.todoCount}</p>
                  <p class="text-xs text-[var(--color-muted)] uppercase tracking-wider mt-1">Todos</p>
                </div>
                <div class="text-center p-4 bg-[var(--color-bg)] border border-[var(--color-border)]" style="border-radius: 2px;">
                  <p class="font-serif text-2xl text-emerald-600 dark:text-emerald-400">{stats.completedTodos}</p>
                  <p class="text-xs text-[var(--color-muted)] uppercase tracking-wider mt-1">Done</p>
                </div>
                <div class="text-center p-4 bg-[var(--color-bg)] border border-[var(--color-border)]" style="border-radius: 2px;">
                  <p class="font-serif text-2xl text-[var(--color-accent)]">{stats.pendingTodos}</p>
                  <p class="text-xs text-[var(--color-muted)] uppercase tracking-wider mt-1">Pending</p>
                </div>
              </div>

              {#if selectedAccount.budget || selectedAccount.est_engineers}
                <div class="flex gap-6 mt-6 pt-6 border-t border-[var(--color-border)]">
                  {#if selectedAccount.budget}
                    <div class="flex items-center gap-2 text-sm">
                      <DollarSign class="w-4 h-4 text-[var(--color-muted)]" strokeWidth={1.5} />
                      <span>Budget: ${selectedAccount.budget.toLocaleString()}</span>
                    </div>
                  {/if}
                  {#if selectedAccount.est_engineers}
                    <div class="flex items-center gap-2 text-sm">
                      <Users class="w-4 h-4 text-[var(--color-muted)]" strokeWidth={1.5} />
                      <span>Est. Engineers: {selectedAccount.est_engineers}</span>
                    </div>
                  {/if}
                </div>
              {/if}
            </div>

            <!-- Notes Section -->
            <div class="card">
              <div class="flex items-center justify-between mb-6">
                <h3 class="font-serif text-lg flex items-center gap-2">
                  <FileText class="w-5 h-5 text-[var(--color-accent)]" strokeWidth={1.5} />
                  Notes ({selectedNotes.length})
                </h3>
                <a href="/notes" class="text-sm text-[var(--color-accent)] editorial-underline">
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
                      class="flex items-center justify-between p-4 bg-[var(--color-bg)] border border-[var(--color-border)] hover:border-[var(--color-accent)]/40 transition-colors group"
                      style="border-radius: 2px;"
                    >
                      <div class="flex items-center gap-4">
                        <FileText class="w-4 h-4 text-[var(--color-muted)]" strokeWidth={1.5} />
                        <div>
                          <p class="font-medium text-sm group-hover:text-[var(--color-accent)] transition-colors">{note.title}</p>
                          <p class="text-xs text-[var(--color-muted)]">
                            {note.template_type} · {formatDate(note.created_at)}
                          </p>
                        </div>
                      </div>
                      <ChevronRight class="w-4 h-4 text-[var(--color-muted)] opacity-0 group-hover:opacity-100" strokeWidth={1.5} />
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
              <div class="flex items-center justify-between mb-6">
                <h3 class="font-serif text-lg flex items-center gap-2">
                  <CheckSquare class="w-5 h-5 text-blue-500" strokeWidth={1.5} />
                  Todos ({selectedTodos.length})
                </h3>
                <a href="/todos" class="text-sm text-[var(--color-accent)] editorial-underline">
                  View All
                </a>
              </div>
              {#if selectedTodos.length === 0}
                <p class="text-[var(--color-muted)] text-sm py-4 text-center">No todos for this account yet.</p>
              {:else}
                <div class="space-y-2">
                  {#each selectedTodos.slice(0, 5) as todo}
                    <div class="flex items-center justify-between p-4 bg-[var(--color-bg)] border border-[var(--color-border)]" style="border-radius: 2px;">
                      <div class="flex items-center gap-4">
                        <div class="w-2 h-2 {getStatusColor(todo.status)}" style="border-radius: 1px;"></div>
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
              <div class="w-16 h-16 mx-auto mb-6 flex items-center justify-center bg-[var(--color-bg)] border border-[var(--color-border)]" style="border-radius: 2px;">
                <Building2 class="w-8 h-8 opacity-50" strokeWidth={1.5} />
              </div>
              <p>Select an account to view details</p>
            </div>
          </div>
        {/if}
      </div>
    </div>
  {/if}

  <!-- Trash Section -->
  {#if deletedAccounts.length > 0}
    <div class="mt-12">
      <button 
        class="flex items-center gap-2 mb-4 text-[var(--color-muted)] hover:text-[var(--color-text)] transition-colors"
        on:click={() => showTrash = !showTrash}
      >
        <Trash class="w-4 h-4" strokeWidth={1.5} />
        <span class="font-medium">Trash</span>
        <span class="text-sm">({deletedAccounts.length})</span>
        <ChevronDown class="w-4 h-4 transition-transform {showTrash ? 'rotate-180' : ''}" strokeWidth={1.5} />
      </button>
      
      {#if showTrash}
        <div class="p-6 bg-[var(--color-card)] border border-dashed border-[var(--color-border)]" style="border-radius: 2px;">
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {#each deletedAccounts as account (account.id)}
              <div class="bg-[var(--color-bg)] border border-[var(--color-border)] p-4 opacity-60 hover:opacity-100 transition-opacity group" style="border-radius: 2px;">
                <div class="flex items-start justify-between gap-2">
                  <div>
                    <span class="font-medium line-through">{account.name}</span>
                    {#if account.account_owner}
                      <p class="text-xs text-[var(--color-muted)] mt-1">{account.account_owner}</p>
                    {/if}
                  </div>
                  <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 flex-shrink-0">
                    <button 
                      class="btn-icon-sm btn-icon-success"
                      title="Restore"
                      on:click={() => restoreAccount(account.id)}
                    >
                      <RefreshCw class="w-3.5 h-3.5" strokeWidth={1.5} />
                    </button>
                    <button 
                      class="btn-icon-sm btn-icon-danger"
                      title="Delete permanently"
                      on:click={() => permanentDeletingAccountId = account.id}
                    >
                      <Trash2 class="w-3.5 h-3.5" strokeWidth={1.5} />
                    </button>
                  </div>
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>

<!-- New Account Modal -->
{#if showNewAccountModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button class="modal-backdrop" on:click={() => showNewAccountModal = false} aria-label="Close modal"></button>
    <div class="relative modal-content animate-scale-in">
      <h2 class="modal-title">New Account</h2>
      <form on:submit|preventDefault={createAccount}>
        <div class="mb-6">
          <label class="label" for="new-account-name">Account Name</label>
          <!-- svelte-ignore a11y-autofocus -->
          <input 
            id="new-account-name"
            type="text"
            class="input"
            placeholder="e.g., Acme Corp"
            bind:value={newAccountName}
            autofocus
          />
        </div>
        <div class="mb-6">
          <label class="label" for="new-account-owner">Account Owner (optional)</label>
          <input 
            id="new-account-owner"
            type="text"
            class="input"
            placeholder="Sales rep name"
            bind:value={newAccountOwner}
          />
        </div>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn-secondary" on:click={() => showNewAccountModal = false}>
            Cancel
          </button>
          <button type="submit" class="btn-primary" disabled={!newAccountName.trim()}>
            Create
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- Edit Account Modal -->
{#if showEditAccountModal && editingAccount}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button class="modal-backdrop" on:click={() => showEditAccountModal = false} aria-label="Close modal"></button>
    <div class="relative modal-content animate-scale-in">
      <h2 class="modal-title">Edit Account</h2>
      <form on:submit|preventDefault={updateAccount}>
        <div class="mb-6">
          <label class="label" for="edit-account-name">Account Name</label>
          <!-- svelte-ignore a11y-autofocus -->
          <input 
            id="edit-account-name"
            type="text"
            class="input"
            bind:value={editingAccount.name}
            autofocus
          />
        </div>
        <div class="mb-6">
          <label class="label" for="edit-account-owner">Account Owner</label>
          <input 
            id="edit-account-owner"
            type="text"
            class="input"
            bind:value={editingAccount.account_owner}
          />
        </div>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn-secondary" on:click={() => showEditAccountModal = false}>
            Cancel
          </button>
          <button type="submit" class="btn-primary">
            Save
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- Delete Account Confirmation Modal -->
{#if deletingAccountId}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button 
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => deletingAccountId = null}
      aria-label="Close"
    ></button>
    <div class="relative bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-sm p-6">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">Delete Account?</h3>
      <p class="text-gray-600 dark:text-gray-400 mb-6">This will delete the account and all its notes.</p>
      <div class="flex justify-end gap-3">
        <button class="btn-secondary" on:click={() => deletingAccountId = null}>Cancel</button>
        <button class="btn-danger" on:click={confirmDeleteAccount}>Delete</button>
      </div>
    </div>
  </div>
{/if}

<!-- Permanent Delete Account Confirmation Modal -->
{#if permanentDeletingAccountId}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button 
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => permanentDeletingAccountId = null}
      aria-label="Close"
    ></button>
    <div class="relative bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-sm p-6">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">Permanently Delete?</h3>
      <p class="text-gray-600 dark:text-gray-400 mb-6">This cannot be undone.</p>
      <div class="flex justify-end gap-3">
        <button class="btn-secondary" on:click={() => permanentDeletingAccountId = null}>Cancel</button>
        <button class="btn-danger" on:click={confirmPermanentDeleteAccount}>Delete Forever</button>
      </div>
    </div>
  </div>
{/if}
