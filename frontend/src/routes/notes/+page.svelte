<script lang="ts">
  import { onMount } from 'svelte';
  import { 
    Plus, 
    FolderOpen, 
    FileText, 
    ChevronRight, 
    ChevronDown,
    MoreVertical,
    Trash2,
    Edit3,
    Search,
    ArrowRightLeft,
    Merge,
    LayoutList,
    LayoutGrid,
    Building2
  } from 'lucide-svelte';
  import { api, type Account, type Note } from '$lib/utils/api';
  import { addToast } from '$lib/stores';

  let accounts: Account[] = [];
  let notes: Note[] = [];
  let loading = true;
  let expandedAccounts: Set<string> = new Set();
  let selectedNoteId: string | null = null;
  let showNewAccountModal = false;
  let showNewNoteModal = false;
  let showMoveNoteModal = false;
  let showMergeAccountsModal = false;
  let newAccountName = '';
  let newNoteName = '';
  let newNoteAccountId = '';
  let filterQuery = '';
  
  // Move note state
  let moveNoteId = '';
  let moveTargetAccountId = '';
  
  // Merge accounts state
  let mergeSourceAccountId = '';
  let mergeTargetAccountId = '';
  
  // View mode: 'folders' | 'cards' | 'organized'
  type ViewMode = 'folders' | 'cards' | 'organized';
  let viewMode: ViewMode = 'folders';

  onMount(async () => {
    await loadData();
  });

  async function loadData() {
    try {
      loading = true;
      const [accountsData, notesData] = await Promise.all([
        api.getAccounts(),
        api.getNotes()
      ]);
      accounts = accountsData;
      notes = notesData;
      // Expand all accounts by default
      expandedAccounts = new Set(accounts.map(a => a.id));
    } catch (e) {
      addToast('error', 'Failed to load data');
    } finally {
      loading = false;
    }
  }

  function toggleAccount(accountId: string) {
    if (expandedAccounts.has(accountId)) {
      expandedAccounts.delete(accountId);
    } else {
      expandedAccounts.add(accountId);
    }
    expandedAccounts = expandedAccounts;
  }

  function getNotesForAccount(accountId: string): Note[] {
    return notes.filter(n => n.account_id === accountId);
  }

  async function createAccount() {
    if (!newAccountName.trim()) return;
    try {
      const account = await api.createAccount({ name: newAccountName.trim() });
      accounts = [...accounts, account];
      expandedAccounts.add(account.id);
      expandedAccounts = expandedAccounts;
      newAccountName = '';
      showNewAccountModal = false;
      addToast('success', 'Account created');
    } catch (e) {
      addToast('error', 'Failed to create account');
    }
  }

  async function createNote() {
    if (!newNoteName.trim() || !newNoteAccountId) return;
    try {
      const note = await api.createNote({
        title: newNoteName.trim(),
        account_id: newNoteAccountId,
        template_type: 'initial'
      });
      notes = [note, ...notes];
      newNoteName = '';
      newNoteAccountId = '';
      showNewNoteModal = false;
      addToast('success', 'Note created');
      // Navigate to the new note
      window.location.href = `/notes/${note.id}`;
    } catch (e) {
      addToast('error', 'Failed to create note');
    }
  }

  async function deleteAccount(accountId: string) {
    if (!confirm('Delete this account and all its notes?')) return;
    try {
      await api.deleteAccount(accountId);
      accounts = accounts.filter(a => a.id !== accountId);
      notes = notes.filter(n => n.account_id !== accountId);
      addToast('success', 'Account deleted');
    } catch (e) {
      addToast('error', 'Failed to delete account');
    }
  }

  async function deleteNote(noteId: string) {
    if (!confirm('Delete this note?')) return;
    try {
      await api.deleteNote(noteId);
      notes = notes.filter(n => n.id !== noteId);
      addToast('success', 'Note deleted');
    } catch (e) {
      addToast('error', 'Failed to delete note');
    }
  }

  function openMoveNoteModal(noteId: string) {
    moveNoteId = noteId;
    const note = notes.find(n => n.id === noteId);
    moveTargetAccountId = note?.account_id || '';
    showMoveNoteModal = true;
  }

  async function moveNote() {
    if (!moveNoteId || !moveTargetAccountId) return;
    
    const note = notes.find(n => n.id === moveNoteId);
    if (!note || note.account_id === moveTargetAccountId) {
      showMoveNoteModal = false;
      return;
    }
    
    try {
      await api.updateNote(moveNoteId, { account_id: moveTargetAccountId });
      
      // Update local state
      notes = notes.map(n => 
        n.id === moveNoteId 
          ? { ...n, account_id: moveTargetAccountId }
          : n
      );
      
      showMoveNoteModal = false;
      addToast('success', 'Note moved to new account');
    } catch (e) {
      addToast('error', 'Failed to move note');
    }
  }

  async function mergeAccounts() {
    if (!mergeSourceAccountId || !mergeTargetAccountId || mergeSourceAccountId === mergeTargetAccountId) {
      addToast('error', 'Please select different accounts');
      return;
    }
    
    if (!confirm(`Move all notes from "${accounts.find(a => a.id === mergeSourceAccountId)?.name}" to "${accounts.find(a => a.id === mergeTargetAccountId)?.name}" and delete the source account?`)) {
      return;
    }
    
    try {
      // Move all notes from source to target
      const notesToMove = notes.filter(n => n.account_id === mergeSourceAccountId);
      
      await Promise.all(
        notesToMove.map(note => 
          api.updateNote(note.id, { account_id: mergeTargetAccountId })
        )
      );
      
      // Delete source account
      await api.deleteAccount(mergeSourceAccountId);
      
      // Update local state
      notes = notes.map(n => 
        n.account_id === mergeSourceAccountId 
          ? { ...n, account_id: mergeTargetAccountId }
          : n
      );
      accounts = accounts.filter(a => a.id !== mergeSourceAccountId);
      
      showMergeAccountsModal = false;
      mergeSourceAccountId = '';
      mergeTargetAccountId = '';
      addToast('success', `Merged ${notesToMove.length} notes and deleted source account`);
    } catch (e) {
      addToast('error', 'Failed to merge accounts');
    }
  }

  $: filteredAccounts = filterQuery 
    ? accounts.filter(a => 
        a.name.toLowerCase().includes(filterQuery.toLowerCase()) ||
        getNotesForAccount(a.id).some(n => n.title.toLowerCase().includes(filterQuery.toLowerCase()))
      )
    : accounts;

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  }

  function stripHtml(html: string): string {
    const tmp = document.createElement('div');
    tmp.innerHTML = html;
    return tmp.textContent || tmp.innerText || '';
  }

  function getPreview(content: string, maxLength: number = 120): string {
    const text = stripHtml(content);
    return text.length > maxLength ? text.substring(0, maxLength) + '...' : text;
  }

  function getAccountName(accountId: string): string {
    return accounts.find(a => a.id === accountId)?.name || 'Unknown';
  }

  $: allNotes = filterQuery
    ? notes.filter(n => 
        n.title.toLowerCase().includes(filterQuery.toLowerCase()) ||
        (n.content && n.content.toLowerCase().includes(filterQuery.toLowerCase()))
      )
    : notes;
</script>

<svelte:head>
  <title>Notes - SE Notes</title>
</svelte:head>

<div class="max-w-7xl mx-auto">
  <div class="flex items-center justify-between mb-8">
    <div>
      <h1 class="page-title">Notes</h1>
      <p class="page-subtitle">Organize your meeting notes by account</p>
    </div>
    <div class="flex items-center gap-3">
      <!-- View Mode Toggle -->
      <div class="flex items-center bg-[var(--color-bg)] rounded-lg p-1 border border-[var(--color-border)]">
        <button 
          class="p-1.5 rounded-md transition-colors"
          class:bg-[var(--color-card)]={viewMode === 'folders'}
          class:shadow-sm={viewMode === 'folders'}
          on:click={() => viewMode = 'folders'}
          title="Folder view"
        >
          <FolderOpen class="w-4 h-4" />
        </button>
        <button 
          class="p-1.5 rounded-md transition-colors"
          class:bg-[var(--color-card)]={viewMode === 'cards'}
          class:shadow-sm={viewMode === 'cards'}
          on:click={() => viewMode = 'cards'}
          title="Cards view"
        >
          <LayoutGrid class="w-4 h-4" />
        </button>
        <button 
          class="p-1.5 rounded-md transition-colors"
          class:bg-[var(--color-card)]={viewMode === 'organized'}
          class:shadow-sm={viewMode === 'organized'}
          on:click={() => viewMode = 'organized'}
          title="Organized by account"
        >
          <Building2 class="w-4 h-4" />
        </button>
      </div>
      {#if accounts.length >= 2}
        <button 
          class="btn-ghost"
          on:click={() => showMergeAccountsModal = true}
        >
          <Merge class="w-4 h-4" />
          Merge
        </button>
      {/if}
      <button 
        class="btn-secondary"
        on:click={() => showNewAccountModal = true}
      >
        <Plus class="w-4 h-4" />
        Account
      </button>
      <button 
        class="btn-primary"
        on:click={() => showNewNoteModal = true}
        disabled={accounts.length === 0}
      >
        <Plus class="w-4 h-4" />
        Note
      </button>
    </div>
  </div>

  <!-- Search -->
  <div class="mb-6">
    <div class="relative">
      <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-[var(--color-muted)]" />
      <input 
        type="text"
        placeholder="Filter accounts and notes..."
        class="input pl-10"
        bind:value={filterQuery}
      />
    </div>
  </div>

  {#if loading}
    <div class="space-y-4">
      {#each [1, 2, 3] as _}
        <div class="card animate-pulse">
          <div class="h-6 bg-[var(--color-border)] rounded w-48 mb-4"></div>
          <div class="space-y-2">
            <div class="h-4 bg-[var(--color-border)] rounded w-full"></div>
            <div class="h-4 bg-[var(--color-border)] rounded w-3/4"></div>
          </div>
        </div>
      {/each}
    </div>
  {:else if accounts.length === 0}
    <div class="card text-center py-12">
      <FolderOpen class="w-12 h-12 mx-auto text-[var(--color-muted)] mb-4" />
      <h3 class="text-lg font-medium mb-2">No accounts yet</h3>
      <p class="text-[var(--color-muted)] mb-6">Create your first account to start organizing notes.</p>
      <button class="btn-primary" on:click={() => showNewAccountModal = true}>
        <Plus class="w-4 h-4" />
        Create Account
      </button>
    </div>
  {:else if viewMode === 'folders'}
    <!-- Folders View (Original) -->
    <div class="space-y-4">
      {#each filteredAccounts as account (account.id)}
        <div class="card p-0 overflow-hidden group">
          <!-- Account Header -->
          <div 
            class="w-full flex items-center justify-between p-4 hover:bg-[var(--color-bg)] transition-colors cursor-pointer"
            on:click={() => toggleAccount(account.id)}
            on:keypress={(e) => e.key === 'Enter' && toggleAccount(account.id)}
            role="button"
            tabindex="0"
          >
            <div class="flex items-center gap-3">
              {#if expandedAccounts.has(account.id)}
                <ChevronDown class="w-5 h-5 text-[var(--color-muted)]" />
              {:else}
                <ChevronRight class="w-5 h-5 text-[var(--color-muted)]" />
              {/if}
              <FolderOpen class="w-5 h-5 text-primary-500" />
              <span class="font-medium">{account.name}</span>
              <span class="text-sm text-[var(--color-muted)]">
                ({getNotesForAccount(account.id).length} notes)
              </span>
            </div>
            <div class="flex items-center gap-2">
              {#if account.account_owner}
                <span class="text-sm text-[var(--color-muted)]">{account.account_owner}</span>
              {/if}
              <button 
                class="p-1.5 hover:bg-[var(--color-border)] rounded-lg opacity-0 group-hover:opacity-100 transition-opacity"
                on:click|stopPropagation={() => deleteAccount(account.id)}
              >
                <Trash2 class="w-4 h-4 text-red-500" />
              </button>
            </div>
          </div>

          <!-- Notes List -->
          {#if expandedAccounts.has(account.id)}
            <div class="border-t border-[var(--color-border)]">
              {#each getNotesForAccount(account.id) as note (note.id)}
                <a 
                  href="/notes/{note.id}"
                  class="flex items-center justify-between px-4 py-3 pl-12 hover:bg-[var(--color-bg)] transition-colors border-b border-[var(--color-border)] last:border-b-0 group"
                >
                  <div class="flex items-center gap-3">
                    <FileText class="w-4 h-4 text-[var(--color-muted)]" />
                    <span>{note.title}</span>
                    <span class="px-2 py-0.5 text-xs rounded bg-[var(--color-border)] text-[var(--color-muted)]">
                      {note.template_type}
                    </span>
                  </div>
                  <div class="flex items-center gap-2">
                    <span class="text-sm text-[var(--color-muted)] mr-2">
                      {formatDate(note.created_at)}
                    </span>
                    <button 
                      class="p-1.5 hover:bg-[var(--color-border)] rounded-lg opacity-0 group-hover:opacity-100 transition-opacity"
                      title="Move to another account"
                      on:click|preventDefault|stopPropagation={() => openMoveNoteModal(note.id)}
                    >
                      <ArrowRightLeft class="w-4 h-4 text-[var(--color-muted)]" />
                    </button>
                    <button 
                      class="p-1.5 hover:bg-[var(--color-border)] rounded-lg opacity-0 group-hover:opacity-100 transition-opacity"
                      title="Delete note"
                      on:click|preventDefault|stopPropagation={() => deleteNote(note.id)}
                    >
                      <Trash2 class="w-4 h-4 text-red-500" />
                    </button>
                  </div>
                </a>
              {:else}
                <div class="px-4 py-8 pl-12 text-center text-[var(--color-muted)]">
                  No notes in this account yet.
                  <button 
                    class="text-primary-500 hover:underline ml-1"
                    on:click={() => { newNoteAccountId = account.id; showNewNoteModal = true; }}
                  >
                    Create one
                  </button>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {:else if viewMode === 'cards'}
    <!-- Cards View -->
    {#if allNotes.length === 0}
      <div class="card text-center py-12">
        <FileText class="w-12 h-12 mx-auto text-[var(--color-muted)] mb-4" />
        <h3 class="text-lg font-medium mb-2">No notes yet</h3>
        <p class="text-[var(--color-muted)] mb-6">Create your first note to get started.</p>
        <button class="btn-primary" on:click={() => showNewNoteModal = true}>
          <Plus class="w-4 h-4" />
          Create Note
        </button>
      </div>
    {:else}
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
        {#each allNotes as note (note.id)}
          <a 
            href="/notes/{note.id}"
            class="card p-4 hover:border-primary-500/50 transition-colors group"
          >
            <div class="flex items-start justify-between mb-2">
              <span class="px-2 py-0.5 text-xs rounded bg-primary-500/10 text-primary-500">
                {note.template_type}
              </span>
              <button 
                class="p-1 hover:bg-[var(--color-border)] rounded opacity-0 group-hover:opacity-100 transition-opacity"
                title="Delete note"
                on:click|preventDefault|stopPropagation={() => deleteNote(note.id)}
              >
                <Trash2 class="w-3.5 h-3.5 text-red-500" />
              </button>
            </div>
            <h3 class="font-semibold mb-1 line-clamp-2">{note.title}</h3>
            <p class="text-sm text-[var(--color-muted)] mb-3 line-clamp-3">
              {note.content ? getPreview(note.content) : 'No content yet...'}
            </p>
            <div class="flex items-center justify-between text-xs text-[var(--color-muted)]">
              <span class="flex items-center gap-1">
                <Building2 class="w-3 h-3" />
                {getAccountName(note.account_id)}
              </span>
              <span>{formatDate(note.created_at)}</span>
            </div>
          </a>
        {/each}
      </div>
    {/if}
  {:else if viewMode === 'organized'}
    <!-- Organized View (Grouped by Account, No Collapse) -->
    <div class="space-y-8">
      {#each filteredAccounts as account (account.id)}
        {@const accountNotes = getNotesForAccount(account.id)}
        <div>
          <div class="flex items-center gap-3 mb-4">
            <div class="w-10 h-10 rounded-lg bg-primary-500/10 flex items-center justify-center">
              <Building2 class="w-5 h-5 text-primary-500" />
            </div>
            <div>
              <h2 class="font-semibold">{account.name}</h2>
              <p class="text-sm text-[var(--color-muted)]">
                {accountNotes.length} note{accountNotes.length !== 1 ? 's' : ''}
                {#if account.account_owner}
                  · {account.account_owner}
                {/if}
              </p>
            </div>
          </div>
          {#if accountNotes.length === 0}
            <div class="card text-center py-6 text-[var(--color-muted)]">
              No notes in this account yet.
              <button 
                class="text-primary-500 hover:underline ml-1"
                on:click={() => { newNoteAccountId = account.id; showNewNoteModal = true; }}
              >
                Create one
              </button>
            </div>
          {:else}
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {#each accountNotes as note (note.id)}
                <a 
                  href="/notes/{note.id}"
                  class="card p-4 hover:border-primary-500/50 transition-colors group"
                >
                  <div class="flex items-start justify-between mb-2">
                    <span class="px-2 py-0.5 text-xs rounded bg-primary-500/10 text-primary-500">
                      {note.template_type}
                    </span>
                    <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button 
                        class="p-1 hover:bg-[var(--color-border)] rounded"
                        title="Move to another account"
                        on:click|preventDefault|stopPropagation={() => openMoveNoteModal(note.id)}
                      >
                        <ArrowRightLeft class="w-3.5 h-3.5 text-[var(--color-muted)]" />
                      </button>
                      <button 
                        class="p-1 hover:bg-[var(--color-border)] rounded"
                        title="Delete note"
                        on:click|preventDefault|stopPropagation={() => deleteNote(note.id)}
                      >
                        <Trash2 class="w-3.5 h-3.5 text-red-500" />
                      </button>
                    </div>
                  </div>
                  <h3 class="font-semibold mb-1 line-clamp-2">{note.title}</h3>
                  <p class="text-sm text-[var(--color-muted)] mb-3 line-clamp-2">
                    {note.content ? getPreview(note.content, 80) : 'No content yet...'}
                  </p>
                  <div class="text-xs text-[var(--color-muted)]">
                    {formatDate(note.created_at)}
                  </div>
                </a>
              {/each}
            </div>
          {/if}
        </div>
      {/each}
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

<!-- New Note Modal -->
{#if showNewNoteModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button 
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => showNewNoteModal = false}
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-md animate-slide-up">
      <h2 class="text-lg font-semibold mb-4">New Note</h2>
      <form on:submit|preventDefault={createNote}>
        <div class="mb-4">
          <label class="label">Account</label>
          <select class="input" bind:value={newNoteAccountId}>
            <option value="">Select an account</option>
            {#each accounts as account}
              <option value={account.id}>{account.name}</option>
            {/each}
          </select>
        </div>
        <div class="mb-4">
          <label class="label">Note Title</label>
          <input 
            type="text"
            class="input"
            placeholder="e.g., Initial Discovery Call"
            bind:value={newNoteName}
          />
        </div>
        <div class="flex justify-end gap-3">
          <button 
            type="button"
            class="btn-secondary"
            on:click={() => showNewNoteModal = false}
          >
            Cancel
          </button>
          <button 
            type="submit" 
            class="btn-primary" 
            disabled={!newNoteName.trim() || !newNoteAccountId}
          >
            Create Note
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- Move Note Modal -->
{#if showMoveNoteModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button 
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => showMoveNoteModal = false}
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-md animate-slide-up">
      <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
        <ArrowRightLeft class="w-5 h-5" />
        Move Note
      </h2>
      <form on:submit|preventDefault={moveNote}>
        <div class="mb-4">
          <label class="label">Move to Account</label>
          <select class="input" bind:value={moveTargetAccountId}>
            {#each accounts as account}
              <option value={account.id}>{account.name}</option>
            {/each}
          </select>
        </div>
        <div class="flex justify-end gap-3">
          <button 
            type="button"
            class="btn-secondary"
            on:click={() => showMoveNoteModal = false}
          >
            Cancel
          </button>
          <button type="submit" class="btn-primary">
            Move Note
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- Merge Accounts Modal -->
{#if showMergeAccountsModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button 
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => showMergeAccountsModal = false}
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-md animate-slide-up">
      <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
        <Merge class="w-5 h-5" />
        Merge Accounts
      </h2>
      <p class="text-sm text-[var(--color-muted)] mb-4">
        Move all notes from one account to another, then delete the source account.
      </p>
      <form on:submit|preventDefault={mergeAccounts}>
        <div class="mb-4">
          <label class="label">Source Account (will be deleted)</label>
          <select class="input" bind:value={mergeSourceAccountId}>
            <option value="">Select account to merge from</option>
            {#each accounts as account}
              <option value={account.id}>
                {account.name} ({getNotesForAccount(account.id).length} notes)
              </option>
            {/each}
          </select>
        </div>
        <div class="mb-4">
          <label class="label">Target Account (notes will be moved here)</label>
          <select class="input" bind:value={mergeTargetAccountId}>
            <option value="">Select account to merge into</option>
            {#each accounts.filter(a => a.id !== mergeSourceAccountId) as account}
              <option value={account.id}>
                {account.name} ({getNotesForAccount(account.id).length} notes)
              </option>
            {/each}
          </select>
        </div>
        <div class="flex justify-end gap-3">
          <button 
            type="button"
            class="btn-secondary"
            on:click={() => showMergeAccountsModal = false}
          >
            Cancel
          </button>
          <button 
            type="submit" 
            class="btn-primary"
            disabled={!mergeSourceAccountId || !mergeTargetAccountId}
          >
            Merge Accounts
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
