<script lang="ts">
  import { onMount } from 'svelte';
  import { 
    Plus, 
    FolderOpen, 
    FileText, 
    ChevronRight, 
    ChevronDown,
    Trash2,
    Trash,
    Search,
    ArrowRightLeft,
    Merge,
    LayoutGrid,
    Building2,
    RefreshCw
  } from 'lucide-svelte';
  import { api, type Account, type Note } from '$lib/utils/api';
  import { addToast } from '$lib/stores';

  let accounts: Account[] = [];
  let notes: Note[] = [];
  let deletedNotes: Note[] = [];
  let loading = true;
  let showTrash = false;
  let expandedAccounts: Set<string> = new Set();
  let showNewAccountModal = false;
  let showNewNoteModal = false;
  let showMoveNoteModal = false;
  let showMergeAccountsModal = false;
  let newAccountName = '';
  let newNoteName = '';
  let newNoteAccountId = '';
  let filterQuery = '';
  let importFileInput: HTMLInputElement;
  
  let moveNoteId = '';
  let moveTargetAccountId = '';
  let mergeSourceAccountId = '';
  let mergeTargetAccountId = '';
  
  type ViewMode = 'folders' | 'cards' | 'organized';
  let viewMode: ViewMode = 'folders';

  onMount(async () => {
    const savedView = localStorage.getItem('defaultNotesView');
    if (savedView && ['folders', 'cards', 'organized'].includes(savedView)) {
      viewMode = savedView as ViewMode;
    }
    await loadData();
  });

  async function loadData() {
    try {
      loading = true;
      const [accountsData, notesData, deleted] = await Promise.all([
        api.getAccounts(),
        api.getNotes(),
        api.getDeletedNotes()
      ]);
      accounts = accountsData;
      notes = notesData;
      deletedNotes = deleted;
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
    return notes
      .filter(n => n.account_id === accountId)
      .sort((a, b) => {
        // Sort 1st call (initial) notes to top
        if (a.template_type === 'initial' && b.template_type !== 'initial') return -1;
        if (a.template_type !== 'initial' && b.template_type === 'initial') return 1;
        // Then sort by date (newest first)
        return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
      });
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
    try {
      const deletedNote = notes.find(n => n.id === noteId);
      await api.deleteNote(noteId);
      notes = notes.filter(n => n.id !== noteId);
      if (deletedNote) {
        deletedNotes = [deletedNote, ...deletedNotes];
      }
      addToast('success', 'Moved to trash');
    } catch (e) {
      addToast('error', 'Failed to delete note');
    }
  }

  async function restoreNote(noteId: string) {
    try {
      await api.restoreNote(noteId);
      const note = deletedNotes.find(n => n.id === noteId);
      if (note) {
        deletedNotes = deletedNotes.filter(n => n.id !== noteId);
        notes = [note, ...notes];
      }
      addToast('success', 'Note restored');
    } catch (e) {
      addToast('error', 'Failed to restore note');
    }
  }

  async function permanentDeleteNote(noteId: string) {
    if (!confirm('Permanently delete this note? This cannot be undone.')) return;
    try {
      await api.permanentDeleteNote(noteId);
      deletedNotes = deletedNotes.filter(n => n.id !== noteId);
      addToast('success', 'Permanently deleted');
    } catch (e) {
      addToast('error', 'Failed to delete');
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
      const targetAccount = accounts.find(a => a.id === moveTargetAccountId);
      notes = notes.map(n => 
        n.id === moveNoteId ? { ...n, account_id: moveTargetAccountId, account_name: targetAccount?.name || '' } : n
      );
      showMoveNoteModal = false;
      addToast('success', 'Note moved');
    } catch (e) {
      addToast('error', 'Failed to move note');
    }
  }

  async function mergeAccounts() {
    if (!mergeSourceAccountId || !mergeTargetAccountId || mergeSourceAccountId === mergeTargetAccountId) {
      addToast('error', 'Please select different accounts');
      return;
    }
    if (!confirm(`Merge all notes and delete source account?`)) return;
    try {
      const notesToMove = notes.filter(n => n.account_id === mergeSourceAccountId);
      await Promise.all(notesToMove.map(note => 
        api.updateNote(note.id, { account_id: mergeTargetAccountId })
      ));
      await api.deleteAccount(mergeSourceAccountId);
      notes = notes.map(n => 
        n.account_id === mergeSourceAccountId ? { ...n, account_id: mergeTargetAccountId } : n
      );
      accounts = accounts.filter(a => a.id !== mergeSourceAccountId);
      showMergeAccountsModal = false;
      mergeSourceAccountId = '';
      mergeTargetAccountId = '';
      addToast('success', `Merged ${notesToMove.length} notes`);
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

  async function handleImport(e: Event) {
    const file = (e.target as HTMLInputElement).files?.[0];
    if (!file) return;

    try {
      const result = await api.importMarkdown(file);
      addToast('success', 'Note imported');
      // Reload data to show new note
      await loadData();
      // Optionally navigate to it
      // goto(`/notes/${result.id}`);
    } catch (e) {
      console.error(e);
      addToast('error', 'Failed to import markdown');
    } finally {
      // Reset input
      if (importFileInput) importFileInput.value = '';
    }
  }

  $: allNotes = (filterQuery
    ? notes.filter(n => 
        n.title.toLowerCase().includes(filterQuery.toLowerCase()) ||
        (n.content && n.content.toLowerCase().includes(filterQuery.toLowerCase()))
      )
    : notes
  ).sort((a, b) => {
    // Sort 1st call (initial) notes to top
    if (a.template_type === 'initial' && b.template_type !== 'initial') return -1;
    if (a.template_type !== 'initial' && b.template_type === 'initial') return 1;
    // Then sort by date (newest first)
    return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
  });
</script>

<svelte:head>
  <title>Notes - Noted</title>
</svelte:head>

<div class="max-w-6xl mx-auto">
  <!-- Header -->
  <div class="flex items-start justify-between mb-12">
    <div class="page-header mb-0">
      <div class="divider-accent mb-6"></div>
      <h1 class="page-title">Notes</h1>
      <p class="page-subtitle">Organize your meeting notes</p>
    </div>
    <div class="flex items-center gap-3">
      <!-- View Toggle -->
      <div class="flex items-center bg-[var(--color-card)] border border-[var(--color-border)] p-1" style="border-radius: 2px;">
        <button 
          class="btn-icon {viewMode === 'folders' ? 'bg-[var(--color-bg)] text-[var(--color-accent)]' : 'text-[var(--color-muted)] hover:text-[var(--color-text)]'}"
          on:click={() => viewMode = 'folders'}
          title="Folder view"
        >
          <FolderOpen class="w-4 h-4" strokeWidth={1.5} />
        </button>
        <button 
          class="btn-icon {viewMode === 'cards' ? 'bg-[var(--color-bg)] text-[var(--color-accent)]' : 'text-[var(--color-muted)] hover:text-[var(--color-text)]'}"
          on:click={() => viewMode = 'cards'}
          title="Cards view"
        >
          <LayoutGrid class="w-4 h-4" strokeWidth={1.5} />
        </button>
        <button 
          class="btn-icon {viewMode === 'organized' ? 'bg-[var(--color-bg)] text-[var(--color-accent)]' : 'text-[var(--color-muted)] hover:text-[var(--color-text)]'}"
          on:click={() => viewMode = 'organized'}
          title="Organized view"
        >
          <Building2 class="w-4 h-4" strokeWidth={1.5} />
        </button>
      </div>
      {#if accounts.length >= 2}
        <button class="btn-ghost" on:click={() => showMergeAccountsModal = true}>
          <Merge class="w-4 h-4" strokeWidth={1.5} />
          <span class="hidden sm:inline">Merge</span>
        </button>
      {/if}
      <button class="btn-secondary" on:click={() => showNewAccountModal = true}>
        <Plus class="w-4 h-4" strokeWidth={1.5} />
        <span class="hidden sm:inline">Account</span>
      </button>
      <button class="btn-secondary" on:click={() => importFileInput.click()}>
        <span class="hidden sm:inline">Import MD</span>
        <span class="sm:hidden">Imp</span>
      </button>
      <input 
        type="file" 
        accept=".md" 
        class="hidden" 
        bind:this={importFileInput}
        on:change={handleImport}
      />
      <button class="btn-primary" on:click={() => showNewNoteModal = true} disabled={accounts.length === 0}>
        <Plus class="w-4 h-4" strokeWidth={1.5} />
        Note
      </button>
    </div>
  </div>

  <!-- Search -->
  <div class="mb-8">
    <div class="relative">
      <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-[var(--color-muted)]" strokeWidth={1.5} />
      <input 
        type="text"
        placeholder="Filter accounts and notes..."
        class="input pl-12"
        bind:value={filterQuery}
      />
    </div>
  </div>

  {#if loading}
    <div class="space-y-4">
      {#each [1, 2, 3] as _}
        <div class="card">
          <div class="skeleton h-6 w-48 mb-4"></div>
          <div class="space-y-2">
            <div class="skeleton h-4 w-full"></div>
            <div class="skeleton h-4 w-3/4"></div>
          </div>
        </div>
      {/each}
    </div>
  {:else if accounts.length === 0}
    <div class="card text-center py-16">
      <div class="w-16 h-16 mx-auto mb-6 flex items-center justify-center bg-[var(--color-bg)] border border-[var(--color-border)]" style="border-radius: 2px;">
        <FolderOpen class="w-8 h-8 text-[var(--color-muted)]" strokeWidth={1.5} />
      </div>
      <h3 class="font-serif text-xl mb-2">No accounts yet</h3>
      <p class="text-[var(--color-muted)] mb-8">Create your first account to start organizing notes.</p>
      <button class="btn-primary" on:click={() => showNewAccountModal = true}>
        <Plus class="w-4 h-4" strokeWidth={1.5} />
        Create Account
      </button>
    </div>
  {:else if viewMode === 'folders'}
    <!-- Folders View -->
    <div class="space-y-4 animate-stagger">
      {#each filteredAccounts as account (account.id)}
        <div class="card p-0 overflow-hidden group">
          <div
            class="w-full flex items-center justify-between p-5 hover:bg-[var(--color-card-hover)] transition-colors cursor-pointer"
            role="button"
            tabindex="0"
            on:click={() => toggleAccount(account.id)}
            on:keypress={(e) => e.key === 'Enter' && toggleAccount(account.id)}
          >
            <div class="flex items-center gap-4">
              {#if expandedAccounts.has(account.id)}
                <ChevronDown class="w-4 h-4 text-[var(--color-muted)]" strokeWidth={1.5} />
              {:else}
                <ChevronRight class="w-4 h-4 text-[var(--color-muted)]" strokeWidth={1.5} />
              {/if}
              <div class="w-10 h-10 flex items-center justify-center bg-[var(--color-bg)] border border-[var(--color-border)]" style="border-radius: 2px;">
                <FolderOpen class="w-5 h-5 text-[var(--color-accent)]" strokeWidth={1.5} />
              </div>
              <div>
                <span class="font-medium">{account.name}</span>
                <span class="text-sm text-[var(--color-muted)] ml-2">
                  {getNotesForAccount(account.id).length} notes
                </span>
              </div>
            </div>
            <div class="flex items-center gap-3">
              {#if account.account_owner}
                <span class="text-sm text-[var(--color-muted)]">{account.account_owner}</span>
              {/if}
              <button 
                class="btn-icon btn-icon-danger opacity-0 group-hover:opacity-100"
                on:click|stopPropagation={() => deleteAccount(account.id)}
                aria-label="Delete account"
              >
                <Trash2 class="w-4 h-4" strokeWidth={1.5} />
              </button>
            </div>
          </div>

          {#if expandedAccounts.has(account.id)}
            <div class="border-t border-[var(--color-border)]">
              {#each getNotesForAccount(account.id) as note (note.id)}
                <a 
                  href="/notes/{note.id}"
                  class="flex items-center justify-between px-5 py-4 pl-16 hover:bg-[var(--color-card-hover)] transition-colors border-b border-[var(--color-border)] last:border-b-0 group/note"
                >
                  <div class="flex items-center gap-4">
                    <FileText class="w-4 h-4 text-[var(--color-muted)]" strokeWidth={1.5} />
                    <span>{note.title}</span>
                    {#if note.template_type === 'initial'}
                      <span class="tag-accent">1st Call</span>
                    {/if}
                  </div>
                  <div class="flex items-center gap-3">
                    <span class="text-sm text-[var(--color-muted)]">{formatDate(note.created_at)}</span>
                    <button 
                      class="btn-icon-sm btn-icon-ghost opacity-0 group-hover/note:opacity-100"
                      title="Move"
                      on:click|preventDefault|stopPropagation={() => openMoveNoteModal(note.id)}
                    >
                      <ArrowRightLeft class="w-4 h-4" strokeWidth={1.5} />
                    </button>
                    <button 
                      class="btn-icon-sm btn-icon-danger opacity-0 group-hover/note:opacity-100"
                      title="Delete"
                      on:click|preventDefault|stopPropagation={() => deleteNote(note.id)}
                    >
                      <Trash2 class="w-4 h-4" strokeWidth={1.5} />
                    </button>
                  </div>
                </a>
              {:else}
                <div class="px-5 py-8 pl-16 text-center text-[var(--color-muted)]">
                  No notes yet.
                  <button 
                    class="text-[var(--color-accent)] hover:underline ml-1"
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
      <div class="card text-center py-16">
        <FileText class="w-12 h-12 mx-auto text-[var(--color-muted)] mb-4" strokeWidth={1.5} />
        <h3 class="font-serif text-xl mb-2">No notes yet</h3>
        <p class="text-[var(--color-muted)] mb-8">Create your first note to get started.</p>
        <button class="btn-primary" on:click={() => showNewNoteModal = true}>
          <Plus class="w-4 h-4" strokeWidth={1.5} />
          Create Note
        </button>
      </div>
    {:else}
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 animate-stagger">
        {#each allNotes as note (note.id)}
          <a href="/notes/{note.id}" class="card-interactive p-5 group hover-lift">
            <div class="flex items-start justify-between mb-3">
              {#if note.template_type === 'initial'}
                <span class="tag-accent">1st Call</span>
              {:else}
                <span class="tag-default">Follow-up</span>
              {/if}
              <button 
                class="btn-icon-sm btn-icon-danger opacity-0 group-hover:opacity-100"
                title="Delete"
                on:click|preventDefault|stopPropagation={() => deleteNote(note.id)}
              >
                <Trash2 class="w-4 h-4" strokeWidth={1.5} />
              </button>
            </div>
            <h3 class="font-serif text-lg mb-2 line-clamp-2">{note.title}</h3>
            <p class="text-sm text-[var(--color-muted)] mb-4 line-clamp-3">
              {note.content ? getPreview(note.content) : 'No content yet...'}
            </p>
            <div class="flex items-center justify-between text-xs text-[var(--color-muted)]">
              <span class="flex items-center gap-1.5">
                <Building2 class="w-3.5 h-3.5" strokeWidth={1.5} />
                {getAccountName(note.account_id)}
              </span>
              <span>{formatDate(note.created_at)}</span>
            </div>
          </a>
        {/each}
      </div>
    {/if}
  {:else if viewMode === 'organized'}
    <!-- Organized View -->
    <div class="space-y-12 animate-stagger">
      {#each filteredAccounts as account (account.id)}
        {@const accountNotes = getNotesForAccount(account.id)}
        <div>
          <div class="flex items-center gap-4 mb-6">
            <div class="w-12 h-12 flex items-center justify-center bg-[var(--color-card)] border border-[var(--color-border)]" style="border-radius: 2px;">
              <Building2 class="w-6 h-6 text-[var(--color-accent)]" strokeWidth={1.5} />
            </div>
            <div>
              <h2 class="font-serif text-xl">{account.name}</h2>
              <p class="text-sm text-[var(--color-muted)]">
                {accountNotes.length} note{accountNotes.length !== 1 ? 's' : ''}
                {#if account.account_owner} · {account.account_owner}{/if}
              </p>
            </div>
          </div>
          {#if accountNotes.length === 0}
            <div class="card text-center py-8 text-[var(--color-muted)] border-dashed">
              No notes yet.
              <button 
                class="text-[var(--color-accent)] hover:underline ml-1"
                on:click={() => { newNoteAccountId = account.id; showNewNoteModal = true; }}
              >
                Create one
              </button>
            </div>
          {:else}
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {#each accountNotes as note (note.id)}
                <a href="/notes/{note.id}" class="card-interactive p-5 group hover-lift">
                  <div class="flex items-start justify-between mb-3">
                    {#if note.template_type === 'initial'}
                      <span class="tag-accent">1st Call</span>
                    {:else}
                      <span class="tag-default">Follow-up</span>
                    {/if}
                    <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button 
                        class="btn-icon-sm btn-icon-ghost"
                        title="Move"
                        on:click|preventDefault|stopPropagation={() => openMoveNoteModal(note.id)}
                      >
                        <ArrowRightLeft class="w-3.5 h-3.5" strokeWidth={1.5} />
                      </button>
                      <button 
                        class="btn-icon-sm btn-icon-danger"
                        title="Delete"
                        on:click|preventDefault|stopPropagation={() => deleteNote(note.id)}
                      >
                        <Trash2 class="w-3.5 h-3.5" strokeWidth={1.5} />
                      </button>
                    </div>
                  </div>
                  <h3 class="font-medium mb-2 line-clamp-2">{note.title}</h3>
                  <p class="text-sm text-[var(--color-muted)] mb-3 line-clamp-2">
                    {note.content ? getPreview(note.content, 80) : 'No content yet...'}
                  </p>
                  <span class="text-xs text-[var(--color-muted)]">{formatDate(note.created_at)}</span>
                </a>
              {/each}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}

  <!-- Trash Section -->
  {#if deletedNotes.length > 0}
    <div class="mt-12">
      <button 
        class="flex items-center gap-2 mb-4 text-[var(--color-muted)] hover:text-[var(--color-text)] transition-colors"
        on:click={() => showTrash = !showTrash}
      >
        <Trash class="w-4 h-4" strokeWidth={1.5} />
        <span class="font-medium">Trash</span>
        <span class="text-sm">({deletedNotes.length})</span>
        <ChevronDown class="w-4 h-4 transition-transform {showTrash ? 'rotate-180' : ''}" strokeWidth={1.5} />
      </button>
      
      {#if showTrash}
        <div class="p-6 bg-[var(--color-card)] border border-dashed border-[var(--color-border)]" style="border-radius: 2px;">
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {#each deletedNotes as note (note.id)}
              <div class="bg-[var(--color-bg)] border border-[var(--color-border)] p-4 opacity-50 hover:opacity-100 transition-opacity group" style="border-radius: 2px;">
                <div class="flex items-start justify-between gap-2 mb-2">
                  <span class="text-sm font-medium line-through truncate">{note.title}</span>
                  <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 flex-shrink-0">
                    <button 
                      class="btn-icon-sm btn-icon-success"
                      title="Restore"
                      on:click={() => restoreNote(note.id)}
                    >
                      <RefreshCw class="w-3.5 h-3.5" strokeWidth={1.5} />
                    </button>
                    <button 
                      class="btn-icon-sm btn-icon-danger"
                      title="Delete permanently"
                      on:click={() => permanentDeleteNote(note.id)}
                    >
                      <Trash2 class="w-3.5 h-3.5" strokeWidth={1.5} />
                    </button>
                  </div>
                </div>
                {#if note.account_name}
                  <span class="tag-default text-[10px]">
                    <Building2 class="w-3 h-3" strokeWidth={1.5} />
                    {note.account_name}
                  </span>
                {/if}
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
          <label class="label" for="notes-new-account-name">Account Name</label>
          <!-- svelte-ignore a11y-autofocus -->
          <input 
            id="notes-new-account-name"
            type="text"
            class="input"
            placeholder="e.g., Acme Corp"
            bind:value={newAccountName}
            autofocus
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

<!-- New Note Modal -->
{#if showNewNoteModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button class="modal-backdrop" on:click={() => showNewNoteModal = false} aria-label="Close modal"></button>
    <div class="relative modal-content animate-scale-in">
      <h2 class="modal-title">New Note</h2>
      <form on:submit|preventDefault={createNote}>
        <div class="mb-6">
          <label class="label" for="new-note-account">Account</label>
          <select id="new-note-account" class="input" bind:value={newNoteAccountId}>
            <option value="">Select an account</option>
            {#each accounts as account}
              <option value={account.id}>{account.name}</option>
            {/each}
          </select>
        </div>
        <div class="mb-6">
          <label class="label" for="new-note-title">Note Title</label>
          <input 
            id="new-note-title"
            type="text"
            class="input"
            placeholder="e.g., Initial Discovery Call"
            bind:value={newNoteName}
          />
        </div>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn-secondary" on:click={() => showNewNoteModal = false}>
            Cancel
          </button>
          <button type="submit" class="btn-primary" disabled={!newNoteName.trim() || !newNoteAccountId}>
            Create
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- Move Note Modal -->
{#if showMoveNoteModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button class="modal-backdrop" on:click={() => showMoveNoteModal = false} aria-label="Close modal"></button>
    <div class="relative modal-content animate-scale-in">
      <h2 class="modal-title flex items-center gap-2">
        <ArrowRightLeft class="w-5 h-5" strokeWidth={1.5} />
        Move Note
      </h2>
      <form on:submit|preventDefault={moveNote}>
        <div class="mb-6">
          <label class="label" for="move-note-account">Move to Account</label>
          <select id="move-note-account" class="input" bind:value={moveTargetAccountId}>
            {#each accounts as account}
              <option value={account.id}>{account.name}</option>
            {/each}
          </select>
        </div>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn-secondary" on:click={() => showMoveNoteModal = false}>
            Cancel
          </button>
          <button type="submit" class="btn-primary">Move</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- Merge Accounts Modal -->
{#if showMergeAccountsModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button class="modal-backdrop" on:click={() => showMergeAccountsModal = false} aria-label="Close modal"></button>
    <div class="relative modal-content animate-scale-in">
      <h2 class="modal-title flex items-center gap-2">
        <Merge class="w-5 h-5" strokeWidth={1.5} />
        Merge Accounts
      </h2>
      <p class="text-sm text-[var(--color-muted)] mb-6">
        Move all notes from one account to another, then delete the source.
      </p>
      <form on:submit|preventDefault={mergeAccounts}>
        <div class="mb-6">
          <label class="label" for="merge-source-account">Source Account (will be deleted)</label>
          <select id="merge-source-account" class="input" bind:value={mergeSourceAccountId}>
            <option value="">Select source</option>
            {#each accounts as account}
              <option value={account.id}>
                {account.name} ({getNotesForAccount(account.id).length} notes)
              </option>
            {/each}
          </select>
        </div>
        <div class="mb-6">
          <label class="label" for="merge-target-account">Target Account</label>
          <select id="merge-target-account" class="input" bind:value={mergeTargetAccountId}>
            <option value="">Select target</option>
            {#each accounts.filter(a => a.id !== mergeSourceAccountId) as account}
              <option value={account.id}>
                {account.name} ({getNotesForAccount(account.id).length} notes)
              </option>
            {/each}
          </select>
        </div>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn-secondary" on:click={() => showMergeAccountsModal = false}>
            Cancel
          </button>
          <button type="submit" class="btn-primary" disabled={!mergeSourceAccountId || !mergeTargetAccountId}>
            Merge
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
