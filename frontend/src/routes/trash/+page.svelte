<script lang="ts">
  import { onMount } from 'svelte';
  import { Trash2, FileText, CheckSquare, Building2, Users, RotateCcw, X, AlertTriangle } from 'lucide-svelte';
  import { api, type Note, type Account, type Contact } from '$lib/utils/api';
  import { addToast } from '$lib/stores';

  type Category = 'all' | 'notes' | 'todos' | 'accounts' | 'contacts';

  let activeTab: Category = 'all';
  let loading = true;

  let deletedNotes: any[] = [];
  let deletedTodos: any[] = [];
  let deletedAccounts: any[] = [];
  let deletedContacts: any[] = [];

  // Confirmation modal
  let showEmptyModal = false;
  let emptyCategory: Category = 'all';

  $: totalItems = deletedNotes.length + deletedTodos.length + deletedAccounts.length + deletedContacts.length;

  $: filteredItems = (() => {
    if (activeTab === 'notes') return deletedNotes.map(n => ({ ...n, type: 'note' }));
    if (activeTab === 'todos') return deletedTodos.map(t => ({ ...t, type: 'todo' }));
    if (activeTab === 'accounts') return deletedAccounts.map(a => ({ ...a, type: 'account' }));
    if (activeTab === 'contacts') return deletedContacts.map(c => ({ ...c, type: 'contact' }));
    return [
      ...deletedNotes.map(n => ({ ...n, type: 'note' })),
      ...deletedTodos.map(t => ({ ...t, type: 'todo' })),
      ...deletedAccounts.map(a => ({ ...a, type: 'account' })),
      ...deletedContacts.map(c => ({ ...c, type: 'contact' }))
    ].sort((a, b) => new Date(b.deleted_at).getTime() - new Date(a.deleted_at).getTime());
  })();

  onMount(async () => {
    await loadAllDeleted();
  });

  async function loadAllDeleted() {
    loading = true;
    try {
      const [notes, todos, accounts, contacts] = await Promise.all([
        api.getDeletedNotes(),
        api.getDeletedTodos(),
        api.getDeletedAccounts(),
        api.getDeletedContacts()
      ]);
      deletedNotes = notes;
      deletedTodos = todos;
      deletedAccounts = accounts;
      deletedContacts = contacts;
    } catch (e) {
      console.error('Failed to load deleted items:', e);
    } finally {
      loading = false;
    }
  }

  async function restoreItem(item: any) {
    try {
      if (item.type === 'note') {
        await api.restoreNote(item.id);
        deletedNotes = deletedNotes.filter(n => n.id !== item.id);
      } else if (item.type === 'todo') {
        await api.restoreTodo(item.id);
        deletedTodos = deletedTodos.filter(t => t.id !== item.id);
      } else if (item.type === 'account') {
        await api.restoreAccount(item.id);
        deletedAccounts = deletedAccounts.filter(a => a.id !== item.id);
      } else if (item.type === 'contact') {
        await api.restoreContact(item.id);
        deletedContacts = deletedContacts.filter(c => c.id !== item.id);
      }
      addToast('success', 'Item restored');
    } catch (e: any) {
      addToast('error', e.message || 'Failed to restore');
    }
  }

  async function permanentDeleteItem(item: any) {
    try {
      if (item.type === 'note') {
        await api.permanentDeleteNote(item.id);
        deletedNotes = deletedNotes.filter(n => n.id !== item.id);
      } else if (item.type === 'todo') {
        await api.permanentDeleteTodo(item.id);
        deletedTodos = deletedTodos.filter(t => t.id !== item.id);
      } else if (item.type === 'account') {
        await api.permanentDeleteAccount(item.id);
        deletedAccounts = deletedAccounts.filter(a => a.id !== item.id);
      } else if (item.type === 'contact') {
        await api.permanentDeleteContact(item.id);
        deletedContacts = deletedContacts.filter(c => c.id !== item.id);
      }
      addToast('success', 'Permanently deleted');
    } catch (e: any) {
      addToast('error', e.message || 'Failed to delete');
    }
  }

  function openEmptyModal(category: Category) {
    emptyCategory = category;
    showEmptyModal = true;
  }

  async function emptyTrash() {
    showEmptyModal = false;
    try {
      if (emptyCategory === 'all' || emptyCategory === 'notes') {
        await api.emptyNotesTrash();
        deletedNotes = [];
      }
      if (emptyCategory === 'all' || emptyCategory === 'todos') {
        await api.emptyTodosTrash();
        deletedTodos = [];
      }
      if (emptyCategory === 'all' || emptyCategory === 'accounts') {
        await api.emptyAccountsTrash();
        deletedAccounts = [];
      }
      if (emptyCategory === 'all' || emptyCategory === 'contacts') {
        await api.emptyContactsTrash();
        deletedContacts = [];
      }
      addToast('success', emptyCategory === 'all' ? 'All trash emptied' : `${emptyCategory} trash emptied`);
    } catch (e: any) {
      addToast('error', e.message || 'Failed to empty trash');
    }
  }

  function getItemIcon(type: string) {
    switch (type) {
      case 'note': return FileText;
      case 'todo': return CheckSquare;
      case 'account': return Building2;
      case 'contact': return Users;
      default: return FileText;
    }
  }

  function getItemTitle(item: any) {
    if (item.type === 'note') return item.title || 'Untitled Note';
    if (item.type === 'todo') return item.title || 'Untitled Todo';
    if (item.type === 'account') return item.name || 'Unnamed Account';
    if (item.type === 'contact') return item.name || item.email || 'Unknown Contact';
    return 'Unknown';
  }

  function getItemSubtitle(item: any) {
    if (item.type === 'note') return item.account_name || '';
    if (item.type === 'todo') return item.status || '';
    if (item.type === 'account') return item.account_owner || '';
    if (item.type === 'contact') return item.email || '';
    return '';
  }

  function formatDate(dateStr: string) {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  }
</script>

<div class="p-6 max-w-5xl mx-auto">
  <div class="flex items-center justify-between mb-6">
    <div>
      <h1 class="page-title flex items-center gap-3">
        <Trash2 class="w-7 h-7" />
        Trash
      </h1>
      <p class="text-[var(--color-muted)] mt-1">
        {totalItems} item{totalItems !== 1 ? 's' : ''} in trash
      </p>
    </div>
    {#if totalItems > 0}
      <button
        class="btn-danger flex items-center gap-2"
        on:click={() => openEmptyModal(activeTab)}
      >
        <Trash2 class="w-4 h-4" />
        Empty {activeTab === 'all' ? 'All' : activeTab.charAt(0).toUpperCase() + activeTab.slice(1)}
      </button>
    {/if}
  </div>

  <!-- Category Tabs -->
  <div class="flex gap-2 mb-6 flex-wrap">
    <button
      class="px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2
        {activeTab === 'all' ? 'bg-[var(--color-accent)] text-white' : 'bg-[var(--color-card)] border border-[var(--color-border)] hover:border-[var(--color-muted)]'}"
      on:click={() => activeTab = 'all'}
    >
      All
      <span class="px-1.5 py-0.5 rounded text-xs {activeTab === 'all' ? 'bg-white/20' : 'bg-[var(--color-border)]'}">{totalItems}</span>
    </button>
    <button
      class="px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2
        {activeTab === 'notes' ? 'bg-[var(--color-accent)] text-white' : 'bg-[var(--color-card)] border border-[var(--color-border)] hover:border-[var(--color-muted)]'}"
      on:click={() => activeTab = 'notes'}
    >
      <FileText class="w-4 h-4" />
      Notes
      <span class="px-1.5 py-0.5 rounded text-xs {activeTab === 'notes' ? 'bg-white/20' : 'bg-[var(--color-border)]'}">{deletedNotes.length}</span>
    </button>
    <button
      class="px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2
        {activeTab === 'todos' ? 'bg-[var(--color-accent)] text-white' : 'bg-[var(--color-card)] border border-[var(--color-border)] hover:border-[var(--color-muted)]'}"
      on:click={() => activeTab = 'todos'}
    >
      <CheckSquare class="w-4 h-4" />
      Todos
      <span class="px-1.5 py-0.5 rounded text-xs {activeTab === 'todos' ? 'bg-white/20' : 'bg-[var(--color-border)]'}">{deletedTodos.length}</span>
    </button>
    <button
      class="px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2
        {activeTab === 'accounts' ? 'bg-[var(--color-accent)] text-white' : 'bg-[var(--color-card)] border border-[var(--color-border)] hover:border-[var(--color-muted)]'}"
      on:click={() => activeTab = 'accounts'}
    >
      <Building2 class="w-4 h-4" />
      Accounts
      <span class="px-1.5 py-0.5 rounded text-xs {activeTab === 'accounts' ? 'bg-white/20' : 'bg-[var(--color-border)]'}">{deletedAccounts.length}</span>
    </button>
    <button
      class="px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2
        {activeTab === 'contacts' ? 'bg-[var(--color-accent)] text-white' : 'bg-[var(--color-card)] border border-[var(--color-border)] hover:border-[var(--color-muted)]'}"
      on:click={() => activeTab = 'contacts'}
    >
      <Users class="w-4 h-4" />
      Contacts
      <span class="px-1.5 py-0.5 rounded text-xs {activeTab === 'contacts' ? 'bg-white/20' : 'bg-[var(--color-border)]'}">{deletedContacts.length}</span>
    </button>
  </div>

  <!-- Items List -->
  {#if loading}
    <div class="flex justify-center py-12">
      <div class="animate-spin w-8 h-8 border-2 border-[var(--color-accent)] border-t-transparent rounded-full"></div>
    </div>
  {:else if filteredItems.length === 0}
    <div class="card p-12 text-center">
      <Trash2 class="w-12 h-12 mx-auto mb-4 text-[var(--color-muted)]" />
      <h3 class="font-medium text-lg mb-2">Trash is empty</h3>
      <p class="text-[var(--color-muted)]">Deleted items will appear here</p>
    </div>
  {:else}
    <div class="space-y-2">
      {#each filteredItems as item (item.id + item.type)}
        {@const Icon = getItemIcon(item.type)}
        <div class="card p-4 flex items-center gap-4">
          <div class="w-10 h-10 rounded-lg bg-[var(--color-border)]/50 flex items-center justify-center shrink-0">
            <Icon class="w-5 h-5 text-[var(--color-muted)]" />
          </div>
          <div class="flex-1 min-w-0">
            <div class="font-medium truncate">{getItemTitle(item)}</div>
            <div class="text-sm text-[var(--color-muted)] flex items-center gap-2">
              <span class="capitalize">{item.type}</span>
              {#if getItemSubtitle(item)}
                <span>·</span>
                <span class="truncate">{getItemSubtitle(item)}</span>
              {/if}
              {#if item.deleted_at}
                <span>·</span>
                <span>Deleted {formatDate(item.deleted_at)}</span>
              {/if}
            </div>
          </div>
          <div class="flex items-center gap-2 shrink-0">
            <button
              class="p-2 hover:bg-[var(--color-border)] rounded-lg text-[var(--color-accent)]"
              on:click={() => restoreItem(item)}
              title="Restore"
            >
              <RotateCcw class="w-4 h-4" />
            </button>
            <button
              class="p-2 hover:bg-red-100 dark:hover:bg-red-900/30 rounded-lg text-red-500"
              on:click={() => permanentDeleteItem(item)}
              title="Delete permanently"
            >
              <X class="w-4 h-4" />
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Empty Trash Confirmation Modal -->
{#if showEmptyModal}
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" on:click={() => showEmptyModal = false}>
    <div class="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-md" on:click|stopPropagation>
      <div class="flex items-center gap-3 text-red-500 mb-4">
        <AlertTriangle class="w-6 h-6" />
        <h2 class="text-lg font-semibold">Empty Trash</h2>
      </div>
      <p class="text-[var(--color-text-secondary)] mb-6">
        {#if emptyCategory === 'all'}
          This will permanently delete all {totalItems} item{totalItems !== 1 ? 's' : ''} in the trash. This action cannot be undone.
        {:else}
          This will permanently delete all deleted {emptyCategory}. This action cannot be undone.
        {/if}
      </p>
      <div class="flex justify-end gap-3">
        <button class="btn-secondary" on:click={() => showEmptyModal = false}>Cancel</button>
        <button class="btn-danger" on:click={emptyTrash}>
          Empty {emptyCategory === 'all' ? 'All' : emptyCategory.charAt(0).toUpperCase() + emptyCategory.slice(1)}
        </button>
      </div>
    </div>
  </div>
{/if}
