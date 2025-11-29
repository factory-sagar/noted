<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import {
    Users,
    UserCheck,
    UserX,
    Building2,
    Mail,
    Calendar,
    Check,
    X,
    Link,
    Search,
    Filter,
    Pencil,
    Trash2,
    FileText,
    ChevronRight,
    ArrowLeft
  } from 'lucide-svelte';
  import { api, type Contact, type ContactStats, type Account, type ContactNote } from '$lib/utils/api';
  import { addToast } from '$lib/stores';

  let contacts: Contact[] = [];
  let stats: ContactStats | null = null;
  let accounts: Account[] = [];
  let loading = true;
  let filter: 'all' | 'internal' | 'external' | 'suggestions' = 'all';
  let sortBy: 'name' | 'company' | 'account' | 'recent' = 'name';
  let searchQuery = '';
  
  // Modals
  let linkingContact: Contact | null = null;
  let editingContact: Contact | null = null;
  let deletingContact: Contact | null = null;
  let selectedContact: Contact | null = null;
  let contactNotes: ContactNote[] = [];
  let loadingNotes = false;

  // Edit form
  let editName = '';
  let editCompany = '';

  onMount(async () => {
    await Promise.all([loadContacts(), loadStats(), loadAccounts()]);
    loading = false;
  });

  async function loadContacts() {
    try {
      const filterParam = filter === 'all' ? undefined : filter;
      contacts = await api.getContacts(filterParam as any);
    } catch (e) {
      addToast('error', 'Failed to load contacts');
    }
  }

  async function loadStats() {
    try {
      stats = await api.getContactStats();
    } catch (e) {
      console.error('Failed to load stats:', e);
    }
  }

  async function loadAccounts() {
    try {
      accounts = await api.getAccounts();
    } catch (e) {
      console.error('Failed to load accounts:', e);
    }
  }

  async function handleFilterChange() {
    loading = true;
    await loadContacts();
    loading = false;
  }

  async function confirmSuggestion(contact: Contact, confirm: boolean) {
    try {
      await api.confirmContactSuggestion(contact.id, confirm);
      addToast('success', confirm ? 'Contact linked to account' : 'Suggestion dismissed');
      await Promise.all([loadContacts(), loadStats()]);
    } catch (e) {
      addToast('error', 'Failed to process suggestion');
    }
  }

  async function linkToAccount(contact: Contact, accountId: string) {
    try {
      await api.linkContactToAccount(contact.id, accountId);
      addToast('success', 'Contact linked to account');
      linkingContact = null;
      await Promise.all([loadContacts(), loadStats()]);
    } catch (e) {
      addToast('error', 'Failed to link contact');
    }
  }

  function openEdit(contact: Contact) {
    editingContact = contact;
    editName = contact.name || '';
    editCompany = contact.company || '';
  }

  async function saveEdit() {
    if (!editingContact) return;
    try {
      await api.updateContact(editingContact.id, {
        name: editName,
        company: editCompany
      });
      addToast('success', 'Contact updated');
      editingContact = null;
      await loadContacts();
    } catch (e) {
      addToast('error', 'Failed to update contact');
    }
  }

  async function deleteContact() {
    if (!deletingContact) return;
    try {
      const deletedId = deletingContact.id;
      await api.deleteContact(deletedId);
      addToast('success', 'Contact deleted');
      if (selectedContact?.id === deletedId) {
        selectedContact = null;
      }
      deletingContact = null;
      await Promise.all([loadContacts(), loadStats()]);
    } catch (e) {
      addToast('error', 'Failed to delete contact');
    }
  }

  async function viewContact(contact: Contact) {
    selectedContact = contact;
    loadingNotes = true;
    try {
      contactNotes = await api.getContactNotes(contact.id);
    } catch (e) {
      addToast('error', 'Failed to load meeting history');
      contactNotes = [];
    }
    loadingNotes = false;
  }

  $: filteredContacts = contacts
    .filter(c => {
      if (!searchQuery) return true;
      const q = searchQuery.toLowerCase();
      return c.email.toLowerCase().includes(q) ||
             c.name.toLowerCase().includes(q) ||
             c.company.toLowerCase().includes(q) ||
             c.domain.toLowerCase().includes(q);
    })
    .sort((a, b) => {
      switch (sortBy) {
        case 'name':
          return a.name.localeCompare(b.name);
        case 'company':
          return a.company.localeCompare(b.company);
        case 'account':
          const aAccount = a.account_name || '';
          const bAccount = b.account_name || '';
          if (!aAccount && bAccount) return 1;
          if (aAccount && !bAccount) return -1;
          return aAccount.localeCompare(bAccount);
        case 'recent':
          return new Date(b.last_seen || b.created_at).getTime() - new Date(a.last_seen || a.created_at).getTime();
        default:
          return 0;
      }
    });

  $: internalContacts = filteredContacts.filter(c => c.is_internal);
  $: externalContacts = filteredContacts.filter(c => !c.is_internal);

  function formatDate(dateStr: string) {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  }
</script>

<svelte:head>
  <title>Contacts - Noted</title>
</svelte:head>

<div class="flex h-full">
  <!-- Main List -->
  <div class="flex-1 overflow-auto p-6 {selectedContact ? 'hidden md:block md:border-r md:border-[var(--color-border)]' : ''}">
    <div class="max-w-4xl mx-auto">
      <div class="mb-6 flex items-center justify-between">
        <div>
          <h1 class="page-title">Contacts</h1>
          <p class="page-subtitle">People from your meetings and notes</p>
        </div>
      </div>

      <!-- Stats Cards -->
      {#if stats}
        <div class="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
          <div class="card p-4 text-center">
            <div class="text-2xl font-bold">{stats.total_contacts}</div>
            <div class="text-sm text-[var(--color-muted)]">Total</div>
          </div>
          <div class="card p-4 text-center">
            <div class="text-2xl font-bold text-blue-500">{stats.internal_contacts}</div>
            <div class="text-sm text-[var(--color-muted)]">Internal</div>
          </div>
          <div class="card p-4 text-center">
            <div class="text-2xl font-bold text-green-500">{stats.external_contacts}</div>
            <div class="text-sm text-[var(--color-muted)]">External</div>
          </div>
          <div class="card p-4 text-center">
            <div class="text-2xl font-bold text-purple-500">{stats.linked_contacts}</div>
            <div class="text-sm text-[var(--color-muted)]">Linked</div>
          </div>
          <div class="card p-4 text-center">
            <div class="text-2xl font-bold text-amber-500">{stats.pending_suggestions}</div>
            <div class="text-sm text-[var(--color-muted)]">Suggestions</div>
          </div>
        </div>
      {/if}

      <!-- Filters -->
      <div class="flex items-center gap-4 mb-6">
        <div class="relative flex-1 max-w-md">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--color-muted)]" />
          <input
            type="text"
            placeholder="Search contacts..."
            class="input pl-10"
            bind:value={searchQuery}
          />
        </div>
        <div class="flex items-center gap-2">
          <Filter class="w-4 h-4 text-[var(--color-muted)]" />
          <select class="input w-auto" bind:value={filter} on:change={handleFilterChange}>
            <option value="all">All Contacts</option>
            <option value="internal">Internal Only</option>
            <option value="external">External Only</option>
            <option value="suggestions">Pending Suggestions</option>
          </select>
        </div>
        <div class="flex items-center gap-2">
          <span class="text-sm text-[var(--color-muted)]">Sort:</span>
          <select class="input w-auto" bind:value={sortBy}>
            <option value="name">Name</option>
            <option value="company">Company</option>
            <option value="account">Account</option>
            <option value="recent">Recent</option>
          </select>
        </div>
      </div>

      {#if loading}
        <div class="flex justify-center py-12">
          <div class="animate-spin w-8 h-8 border-2 border-primary-500 border-t-transparent rounded-full"></div>
        </div>
      {:else if filteredContacts.length === 0}
        <div class="card p-12 text-center">
          <Users class="w-12 h-12 mx-auto mb-4 text-[var(--color-muted)]" />
          <h3 class="text-lg font-medium mb-2">No contacts yet</h3>
          <p class="text-[var(--color-muted)]">
            Contacts are automatically created when you add participants to notes or sync calendar events.
          </p>
        </div>
      {:else}
        <!-- Internal Contacts -->
        {#if filter === 'all' || filter === 'internal'}
          {#if internalContacts.length > 0}
            <div class="mb-8">
              <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
                <UserCheck class="w-5 h-5 text-blue-500" />
                Internal Contacts
                <span class="text-sm font-normal text-[var(--color-muted)]">(Your Team)</span>
              </h2>
              <div class="grid gap-3">
                {#each internalContacts as contact}
                  <button
                    class="card p-4 flex items-center justify-between w-full text-left hover:border-primary-500 transition-colors {selectedContact?.id === contact.id ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20' : ''}"
                    on:click={() => viewContact(contact)}
                  >
                    <div class="flex items-center gap-4">
                      <div class="w-10 h-10 rounded-full bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center text-blue-600">
                        {contact.name ? contact.name[0].toUpperCase() : contact.email[0].toUpperCase()}
                      </div>
                      <div>
                        <div class="font-medium">{contact.name || contact.email}</div>
                        <div class="text-sm text-[var(--color-muted)] flex items-center gap-2">
                          <Mail class="w-3 h-3" />
                          {contact.email}
                        </div>
                      </div>
                    </div>
                    <div class="flex items-center gap-3">
                      <div class="text-right text-sm text-[var(--color-muted)]">
                        <div class="flex items-center gap-1">
                          <Calendar class="w-3 h-3" />
                          {contact.meeting_count} meetings
                        </div>
                      </div>
                      <ChevronRight class="w-4 h-4 text-[var(--color-muted)]" />
                    </div>
                  </button>
                {/each}
              </div>
            </div>
          {/if}
        {/if}

        <!-- External Contacts -->
        {#if filter === 'all' || filter === 'external' || filter === 'suggestions'}
          {#if externalContacts.length > 0}
            <div>
              <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
                <UserX class="w-5 h-5 text-green-500" />
                External Contacts
              </h2>
              <div class="grid gap-3">
                {#each externalContacts as contact}
                  <div
                    class="card p-4 w-full text-left hover:border-primary-500 transition-colors cursor-pointer {selectedContact?.id === contact.id ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20' : ''}"
                    role="button"
                    tabindex="0"
                    on:click={() => viewContact(contact)}
                    on:keydown={(e) => e.key === 'Enter' && viewContact(contact)}
                  >
                    <div class="flex items-center justify-between">
                      <div class="flex items-center gap-4">
                        <div class="w-10 h-10 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center text-green-600">
                          {contact.name ? contact.name[0].toUpperCase() : contact.email[0].toUpperCase()}
                        </div>
                        <div>
                          <div class="font-medium">{contact.name || contact.email}</div>
                          <div class="text-sm text-[var(--color-muted)] flex items-center gap-2">
                            <Mail class="w-3 h-3" />
                            {contact.email}
                            <span class="px-1.5 py-0.5 bg-[var(--color-border)] rounded text-xs">{contact.domain}</span>
                          </div>
                        </div>
                      </div>
                      <div class="flex items-center gap-3">
                        {#if contact.account_name}
                          <div class="flex items-center gap-2 text-sm text-purple-500">
                            <Building2 class="w-4 h-4" />
                            {contact.account_name}
                          </div>
                        {:else if contact.suggested_account_name && !contact.suggestion_confirmed}
                          <div class="flex items-center gap-2">
                            <span class="text-sm text-amber-500">Suggested: {contact.suggested_account_name}</span>
                            <button
                              class="p-1 hover:bg-green-100 dark:hover:bg-green-900/30 rounded text-green-500"
                              on:click|stopPropagation={() => confirmSuggestion(contact, true)}
                              title="Confirm"
                            >
                              <Check class="w-4 h-4" />
                            </button>
                            <button
                              class="p-1 hover:bg-red-100 dark:hover:bg-red-900/30 rounded text-red-500"
                              on:click|stopPropagation={() => confirmSuggestion(contact, false)}
                              title="Dismiss"
                            >
                              <X class="w-4 h-4" />
                            </button>
                          </div>
                        {/if}
                        <div class="text-sm text-[var(--color-muted)]">
                          {contact.meeting_count} meetings
                        </div>
                        <ChevronRight class="w-4 h-4 text-[var(--color-muted)]" />
                      </div>
                    </div>
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        {/if}
      {/if}
    </div>
  </div>

  <!-- Contact Detail Panel -->
  {#if selectedContact}
    <div class="w-full md:w-96 lg:w-[28rem] bg-[var(--color-card)] overflow-auto">
      <div class="p-6">
        <!-- Back button (mobile) -->
        <button
          class="md:hidden flex items-center gap-2 text-[var(--color-muted)] mb-4"
          on:click={() => selectedContact = null}
        >
          <ArrowLeft class="w-4 h-4" />
          Back to list
        </button>

        <!-- Contact Header -->
        <div class="flex items-start justify-between mb-6">
          <div class="flex items-center gap-4">
            <div class="w-14 h-14 rounded-full {selectedContact.is_internal ? 'bg-blue-100 dark:bg-blue-900/30 text-blue-600' : 'bg-green-100 dark:bg-green-900/30 text-green-600'} flex items-center justify-center text-xl font-medium">
              {selectedContact.name ? selectedContact.name[0].toUpperCase() : selectedContact.email[0].toUpperCase()}
            </div>
            <div>
              <h2 class="text-xl font-semibold">{selectedContact.name || 'No name'}</h2>
              <div class="text-sm text-[var(--color-muted)]">{selectedContact.email}</div>
            </div>
          </div>
          <div class="flex items-center gap-1">
            <button
              class="p-2 hover:bg-[var(--color-border)] rounded-lg transition-colors"
              on:click={() => selectedContact && openEdit(selectedContact)}
              title="Edit contact"
            >
              <Pencil class="w-4 h-4" />
            </button>
            <button
              class="p-2 hover:bg-red-100 dark:hover:bg-red-900/30 rounded-lg text-red-500 transition-colors"
              on:click={() => selectedContact && (deletingContact = selectedContact)}
              title="Delete contact"
            >
              <Trash2 class="w-4 h-4" />
            </button>
          </div>
        </div>

        <!-- Contact Info -->
        <div class="space-y-4 mb-6">
          {#if selectedContact.company}
            <div class="flex items-center gap-3 text-sm">
              <Building2 class="w-4 h-4 text-[var(--color-muted)]" />
              <span>{selectedContact.company}</span>
            </div>
          {/if}
          <div class="flex items-center gap-3 text-sm">
            <span class="px-2 py-1 rounded text-xs {selectedContact.is_internal ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400' : 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'}">
              {selectedContact.is_internal ? 'Internal' : 'External'}
            </span>
            <span class="text-[var(--color-muted)]">{selectedContact.domain}</span>
          </div>
          {#if selectedContact.account_name}
            <div class="flex items-center gap-3 text-sm">
              <Building2 class="w-4 h-4 text-purple-500" />
              <span class="text-purple-500">Linked to {selectedContact.account_name}</span>
            </div>
          {:else if !selectedContact.is_internal}
            <button
              class="btn-secondary btn-sm flex items-center gap-2"
              on:click={() => linkingContact = selectedContact}
            >
              <Link class="w-3 h-3" />
              Link to Account
            </button>
          {/if}
        </div>

        <!-- Stats -->
        <div class="grid grid-cols-2 gap-4 mb-6">
          <div class="p-3 bg-[var(--color-border)]/50 rounded-lg text-center">
            <div class="text-xl font-bold">{selectedContact.meeting_count}</div>
            <div class="text-xs text-[var(--color-muted)]">Meetings</div>
          </div>
          <div class="p-3 bg-[var(--color-border)]/50 rounded-lg text-center">
            <div class="text-sm font-medium">{formatDate(selectedContact.first_seen)}</div>
            <div class="text-xs text-[var(--color-muted)]">First Seen</div>
          </div>
        </div>

        <!-- Meeting History -->
        <div>
          <h3 class="font-semibold mb-3 flex items-center gap-2">
            <FileText class="w-4 h-4" />
            Meeting History
          </h3>
          {#if loadingNotes}
            <div class="flex justify-center py-8">
              <div class="animate-spin w-6 h-6 border-2 border-primary-500 border-t-transparent rounded-full"></div>
            </div>
          {:else if contactNotes.length === 0}
            <p class="text-sm text-[var(--color-muted)] text-center py-4">No meetings found</p>
          {:else}
            <div class="space-y-2">
              {#each contactNotes as note}
                <button
                  class="w-full p-3 text-left rounded-lg border border-[var(--color-border)] hover:border-primary-500 transition-colors"
                  on:click={() => goto(`/notes/${note.id}`)}
                >
                  <div class="font-medium text-sm truncate">{note.title}</div>
                  <div class="text-xs text-[var(--color-muted)] flex items-center gap-2 mt-1">
                    {#if note.account_name}
                      <span>{note.account_name}</span>
                      <span>-</span>
                    {/if}
                    <span>{formatDate(note.meeting_date || note.created_at)}</span>
                  </div>
                </button>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>

<!-- Link to Account Modal -->
{#if linkingContact}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => linkingContact = null}
      aria-label="Close modal"
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-md animate-slide-up">
      <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
        <Link class="w-5 h-5" />
        Link Contact to Account
      </h2>
      <p class="text-sm text-[var(--color-muted)] mb-4">
        Link <strong>{linkingContact.email}</strong> to an account
      </p>
      <div class="space-y-2 max-h-64 overflow-y-auto">
        {#each accounts as account}
          <button
            class="w-full p-3 text-left rounded-lg border border-[var(--color-border)] hover:border-primary-500 hover:bg-primary-50 dark:hover:bg-primary-900/20 transition-colors"
            on:click={() => linkingContact && linkToAccount(linkingContact, account.id)}
          >
            <div class="font-medium">{account.name}</div>
            {#if account.account_owner}
              <div class="text-sm text-[var(--color-muted)]">{account.account_owner}</div>
            {/if}
          </button>
        {/each}
      </div>
      <div class="mt-4 flex justify-end">
        <button class="btn-secondary" on:click={() => linkingContact = null}>
          Cancel
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Edit Contact Modal -->
{#if editingContact}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => editingContact = null}
      aria-label="Close modal"
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-md animate-slide-up">
      <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
        <Pencil class="w-5 h-5" />
        Edit Contact
      </h2>
      <div class="space-y-4">
        <div>
          <label for="edit-email" class="block text-sm font-medium mb-1">Email</label>
          <input
            id="edit-email"
            type="email"
            class="input bg-[var(--color-border)]/50"
            value={editingContact.email}
            disabled
          />
        </div>
        <div>
          <label for="edit-name" class="block text-sm font-medium mb-1">Name</label>
          <input
            id="edit-name"
            type="text"
            class="input"
            placeholder="Full name"
            bind:value={editName}
          />
        </div>
        <div>
          <label for="edit-company" class="block text-sm font-medium mb-1">Company</label>
          <input
            id="edit-company"
            type="text"
            class="input"
            placeholder="Company name"
            bind:value={editCompany}
          />
        </div>
      </div>
      <div class="mt-6 flex justify-end gap-3">
        <button class="btn-secondary" on:click={() => editingContact = null}>
          Cancel
        </button>
        <button class="btn-primary" on:click={saveEdit}>
          Save Changes
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Delete Confirmation Modal -->
{#if deletingContact}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => deletingContact = null}
      aria-label="Close modal"
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-md animate-slide-up">
      <h2 class="text-lg font-semibold mb-4 flex items-center gap-2 text-red-500">
        <Trash2 class="w-5 h-5" />
        Delete Contact
      </h2>
      <p class="text-[var(--color-muted)] mb-2">
        Are you sure you want to delete this contact?
      </p>
      <p class="font-medium mb-4">{deletingContact.email}</p>
      <p class="text-sm text-[var(--color-muted)] mb-6">
        This action cannot be undone. Meeting history will be preserved in notes.
      </p>
      <div class="flex justify-end gap-3">
        <button class="btn-secondary" on:click={() => deletingContact = null}>
          Cancel
        </button>
        <button
          class="px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors"
          on:click={deleteContact}
        >
          Delete
        </button>
      </div>
    </div>
  </div>
{/if}
