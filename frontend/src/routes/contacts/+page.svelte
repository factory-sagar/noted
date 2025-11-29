<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import {
    Users,
    Building2,
    Mail,
    Calendar,
    Link,
    Search,
    Pencil,
    Trash2,
    FileText,
    ChevronDown,
    ChevronRight,
    Plus,
    Check,
    X,
    Globe,
    ArrowRight
  } from 'lucide-svelte';
  import { api, type Contact, type ContactStats, type Account, type ContactNote, type DomainGroup } from '$lib/utils/api';
  import { addToast } from '$lib/stores';

  let contacts: Contact[] = [];
  let domainGroups: DomainGroup[] = [];
  let stats: ContactStats | null = null;
  let accounts: Account[] = [];
  let loading = true;
  let searchQuery = '';
  let activeTab: 'suggestions' | 'all' | 'external' | 'internal' | 'linked' = 'all';
  
  // Detail panel
  let selectedContact: Contact | null = null;
  let contactNotes: ContactNote[] = [];
  let loadingNotes = false;

  // Modals
  let editingContact: Contact | null = null;
  let editName = '';
  let editCompany = '';
  let deletingContact: Contact | null = null;
  let linkingDomain: DomainGroup | null = null;
  let expandedDomains: Set<string> = new Set();

  onMount(async () => {
    await loadAll();
  });

  async function loadAll() {
    loading = true;
    try {
      const [contactsData, groupsData, statsData, accountsData] = await Promise.all([
        api.getContacts(),
        api.getContactDomainGroups('unlinked'),
        api.getContactStats(),
        api.getAccounts()
      ]);
      contacts = contactsData;
      domainGroups = groupsData;
      stats = statsData;
      accounts = accountsData;
    } catch (e) {
      addToast('error', 'Failed to load contacts');
    }
    loading = false;
  }

  async function linkDomainToAccount(domain: string, accountId: string) {
    try {
      const result = await api.linkDomainToAccount(domain, accountId);
      addToast('success', `Linked ${result.contacts_updated} contacts to account`);
      linkingDomain = null;
      await loadAll();
    } catch (e) {
      addToast('error', 'Failed to link contacts');
    }
  }

  async function createAccountFromDomain(domain: string) {
    try {
      const result = await api.createAccountFromDomain(domain);
      addToast('success', `Created "${result.account_name}" and linked ${result.contacts_updated} contacts`);
      await loadAll();
    } catch (e) {
      addToast('error', 'Failed to create account');
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
      await api.updateContact(editingContact.id, { name: editName, company: editCompany });
      addToast('success', 'Contact updated');
      editingContact = null;
      await loadAll();
    } catch (e) {
      addToast('error', 'Failed to update contact');
    }
  }

  async function confirmDelete() {
    if (!deletingContact) return;
    try {
      await api.deleteContact(deletingContact.id);
      addToast('success', 'Contact deleted');
      if (selectedContact?.id === deletingContact.id) selectedContact = null;
      deletingContact = null;
      await loadAll();
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
      contactNotes = [];
    }
    loadingNotes = false;
  }

  function toggleDomain(domain: string) {
    if (expandedDomains.has(domain)) {
      expandedDomains.delete(domain);
    } else {
      expandedDomains.add(domain);
    }
    expandedDomains = expandedDomains;
  }

  $: filteredContacts = contacts.filter(c => {
    // Search overrides tab filter
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      return c.email.toLowerCase().includes(q) ||
             c.name.toLowerCase().includes(q) ||
             c.company.toLowerCase().includes(q);
    }
    // Tab filters
    if (activeTab === 'all') return true;
    if (activeTab === 'internal') return c.is_internal;
    if (activeTab === 'external') return !c.is_internal;
    if (activeTab === 'linked') return c.account_id != null;
    return false;
  });

  $: unlinkedGroups = domainGroups.filter(g => !g.linked_account_id);

  function formatDate(dateStr: string) {
    return new Date(dateStr).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  }

  function capitalizeFirst(str: string) {
    return str.charAt(0).toUpperCase() + str.slice(1);
  }
</script>

<svelte:head>
  <title>Contacts - Noted</title>
</svelte:head>

<div class="flex h-full overflow-hidden">
  <!-- Main Content -->
  <div class="flex-1 overflow-y-auto p-6 {selectedContact ? 'hidden md:block md:border-r md:border-[var(--color-border)]' : ''}">
    <div class="max-w-4xl mx-auto">
      <!-- Header -->
      <div class="mb-6">
        <h1 class="page-title">Contacts</h1>
        <p class="page-subtitle">People from your meetings and calendar events</p>
      </div>

      <!-- Stats - Clickable filters -->
      {#if stats}
        <div class="grid grid-cols-4 gap-3 mb-6">
          <button 
            class="card p-4 text-center transition-all hover:border-primary-500 {activeTab === 'all' ? 'border-primary-500 ring-1 ring-primary-500' : ''}"
            on:click={() => activeTab = 'all'}
          >
            <div class="text-2xl font-bold">{stats.total_contacts}</div>
            <div class="text-sm text-[var(--color-muted)]">Total</div>
          </button>
          <button 
            class="card p-4 text-center transition-all hover:border-blue-500 {activeTab === 'internal' ? 'border-blue-500 ring-1 ring-blue-500' : ''}"
            on:click={() => activeTab = 'internal'}
          >
            <div class="text-2xl font-bold text-blue-500">{stats.internal_contacts}</div>
            <div class="text-sm text-[var(--color-muted)]">Internal</div>
          </button>
          <button 
            class="card p-4 text-center transition-all hover:border-green-500 {activeTab === 'external' ? 'border-green-500 ring-1 ring-green-500' : ''}"
            on:click={() => activeTab = 'external'}
          >
            <div class="text-2xl font-bold text-green-500">{stats.external_contacts}</div>
            <div class="text-sm text-[var(--color-muted)]">External</div>
          </button>
          <button 
            class="card p-4 text-center transition-all hover:border-purple-500 {activeTab === 'linked' ? 'border-purple-500 ring-1 ring-purple-500' : ''}"
            on:click={() => activeTab = 'linked'}
          >
            <div class="text-2xl font-bold text-purple-500">{stats.linked_contacts}</div>
            <div class="text-sm text-[var(--color-muted)]">Linked</div>
          </button>
        </div>
      {/if}

      <!-- Search -->
      <div class="relative mb-4">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--color-muted)]" />
        <input
          type="text"
          placeholder="Search all contacts..."
          class="input pl-10 w-full"
          bind:value={searchQuery}
        />
      </div>

      <!-- Suggestions Banner -->
      {#if unlinkedGroups.length > 0 && activeTab !== 'suggestions'}
        <button
          class="w-full mb-4 p-3 bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg text-left flex items-center justify-between hover:bg-amber-100 dark:hover:bg-amber-900/30 transition-colors"
          on:click={() => activeTab = 'suggestions'}
        >
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 rounded-full bg-amber-100 dark:bg-amber-900/50 flex items-center justify-center">
              <Globe class="w-4 h-4 text-amber-600" />
            </div>
            <div>
              <div class="font-medium text-amber-800 dark:text-amber-200">{unlinkedGroups.length} domain{unlinkedGroups.length !== 1 ? 's' : ''} need linking</div>
              <div class="text-sm text-amber-600 dark:text-amber-400">Click to review and link to accounts</div>
            </div>
          </div>
          <ChevronRight class="w-5 h-5 text-amber-500" />
        </button>
      {/if}

      {#if loading}
        <div class="flex justify-center py-12">
          <div class="animate-spin w-8 h-8 border-2 border-primary-500 border-t-transparent rounded-full"></div>
        </div>
      {:else if activeTab === 'suggestions'}
        <!-- Back button -->
        <button
          class="flex items-center gap-2 text-[var(--color-muted)] hover:text-[var(--color-text)] mb-4"
          on:click={() => activeTab = 'all'}
        >
          <ChevronRight class="w-4 h-4 rotate-180" />
          Back to all contacts
        </button>
        <!-- Domain Groups with Suggestions -->
        {#if unlinkedGroups.length === 0}
          <div class="card p-8 text-center">
            <Check class="w-12 h-12 mx-auto mb-4 text-green-500" />
            <h3 class="text-lg font-medium mb-2">All caught up!</h3>
            <p class="text-[var(--color-muted)]">All external contacts are linked to accounts.</p>
          </div>
        {:else}
          <div class="space-y-3">
            {#each unlinkedGroups as group}
              <div class="card overflow-hidden">
                <div class="p-4">
                  <div class="flex flex-col sm:flex-row sm:items-center gap-3">
                    <button 
                      class="flex items-center gap-3 text-left flex-1 min-w-0"
                      on:click={() => toggleDomain(group.domain)}
                    >
                      {#if expandedDomains.has(group.domain)}
                        <ChevronDown class="w-4 h-4 text-[var(--color-muted)] shrink-0" />
                      {:else}
                        <ChevronRight class="w-4 h-4 text-[var(--color-muted)] shrink-0" />
                      {/if}
                      <Globe class="w-5 h-5 text-[var(--color-muted)] shrink-0" />
                      <div class="min-w-0">
                        <div class="font-medium truncate">{group.domain}</div>
                        <div class="text-sm text-[var(--color-muted)]">{group.contact_count} contact{group.contact_count !== 1 ? 's' : ''}</div>
                      </div>
                    </button>
                    
                    <div class="flex items-center gap-2 shrink-0 flex-wrap">
                      {#if group.suggested_account}
                        {@const suggested = group.suggested_account}
                        <button
                          class="btn-primary btn-sm flex items-center gap-2 whitespace-nowrap"
                          on:click={() => linkDomainToAccount(group.domain, suggested.id)}
                        >
                          <Link class="w-3 h-3" />
                          <span class="truncate max-w-[120px]">Link to {suggested.name}</span>
                        </button>
                      {:else}
                        <button
                          class="btn-secondary btn-sm whitespace-nowrap"
                          on:click={() => linkingDomain = group}
                        >
                          Link to Account
                        </button>
                      {/if}
                      <button
                        class="btn-secondary btn-sm flex items-center gap-2 whitespace-nowrap"
                        on:click={() => createAccountFromDomain(group.domain)}
                      >
                        <Plus class="w-3 h-3" />
                        Create
                      </button>
                    </div>
                  </div>
                </div>
                
                {#if expandedDomains.has(group.domain)}
                  <div class="border-t border-[var(--color-border)] bg-[var(--color-bg)] p-3">
                    <div class="space-y-2">
                      {#each contacts.filter(c => c.domain === group.domain) as contact}
                        <button
                          class="w-full flex items-center gap-3 p-2 rounded-lg hover:bg-[var(--color-card)] text-left transition-colors"
                          on:click={() => viewContact(contact)}
                        >
                          <div class="w-8 h-8 rounded-full bg-green-100 dark:bg-green-900/30 text-green-600 flex items-center justify-center text-sm font-medium">
                            {contact.name ? contact.name[0].toUpperCase() : contact.email[0].toUpperCase()}
                          </div>
                          <div class="min-w-0 flex-1">
                            <div class="font-medium text-sm truncate">{contact.name || contact.email}</div>
                            <div class="text-xs text-[var(--color-muted)] truncate">{contact.email}</div>
                          </div>
                          <div class="text-xs text-[var(--color-muted)]">{contact.meeting_count} meetings</div>
                        </button>
                      {/each}
                    </div>
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        {/if}
      {:else}
        <!-- Contact List Header -->
        <div class="flex items-center justify-between mb-4">
          <div class="text-sm text-[var(--color-muted)]">
            {filteredContacts.length} contact{filteredContacts.length !== 1 ? 's' : ''}
            {#if searchQuery}matching "{searchQuery}"{/if}
          </div>
        </div>
        
        {#if filteredContacts.length === 0}
          <div class="card p-8 text-center">
            <Users class="w-12 h-12 mx-auto mb-4 text-[var(--color-muted)]" />
            <h3 class="text-lg font-medium mb-2">No contacts found</h3>
            <p class="text-[var(--color-muted)]">
              {searchQuery ? 'Try a different search term' : 'Contacts are created when you add participants to notes'}
            </p>
          </div>
        {:else}
          <div class="space-y-2">
            {#each filteredContacts as contact}
              <div 
                class="card p-3 flex items-center gap-4 group hover:border-primary-500 transition-all cursor-pointer {selectedContact?.id === contact.id ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20' : ''}"
                on:click={() => viewContact(contact)}
                on:keydown={(e) => e.key === 'Enter' && viewContact(contact)}
                role="button"
                tabindex="0"
              >
                <div class="w-10 h-10 rounded-full shrink-0 {contact.is_internal ? 'bg-blue-100 text-blue-600 dark:bg-blue-900/30' : 'bg-green-100 text-green-600 dark:bg-green-900/30'} flex items-center justify-center font-medium">
                  {contact.name ? contact.name[0].toUpperCase() : contact.email[0].toUpperCase()}
                </div>
                <div class="min-w-0 flex-1">
                  <div class="font-medium truncate">{contact.name || contact.email}</div>
                  <div class="text-sm text-[var(--color-muted)] flex items-center gap-2 truncate">
                    <Mail class="w-3 h-3 shrink-0" />
                    <span class="truncate">{contact.email}</span>
                  </div>
                </div>
                <div class="hidden md:flex items-center gap-4 shrink-0">
                  {#if contact.account_name}
                    <span class="text-sm text-purple-500 flex items-center gap-1">
                      <Building2 class="w-4 h-4" />
                      {contact.account_name}
                    </span>
                  {:else if contact.company}
                    <span class="text-sm text-[var(--color-muted)]">{contact.company}</span>
                  {/if}
                  <span class="text-sm text-[var(--color-muted)] flex items-center gap-1">
                    <Calendar class="w-3 h-3" />
                    {contact.meeting_count}
                  </span>
                </div>
                <ChevronRight class="w-4 h-4 text-[var(--color-muted)] opacity-0 group-hover:opacity-100 transition-opacity" />
              </div>
            {/each}
          </div>
        {/if}
      {/if}
    </div>
  </div>

  <!-- Detail Panel -->
  {#if selectedContact}
    <div class="w-full md:w-96 bg-[var(--color-card)] overflow-y-auto h-full border-l border-[var(--color-border)]">
      <div class="p-6">
        <div class="flex items-center justify-between mb-6">
          <button
            class="text-[var(--color-muted)] hover:text-[var(--color-text)]"
            on:click={() => selectedContact = null}
          >
            <X class="w-5 h-5" />
          </button>
          <div class="flex items-center gap-1">
            <button
              class="p-2 hover:bg-[var(--color-border)] rounded-lg"
              on:click={() => selectedContact && openEdit(selectedContact)}
              title="Edit"
            >
              <Pencil class="w-4 h-4" />
            </button>
            <button
              class="p-2 hover:bg-red-100 dark:hover:bg-red-900/30 rounded-lg text-red-500"
              on:click={() => selectedContact && (deletingContact = selectedContact)}
              title="Delete"
            >
              <Trash2 class="w-4 h-4" />
            </button>
          </div>
        </div>

        <div class="flex items-center gap-4 mb-6">
          <div class="w-14 h-14 rounded-full {selectedContact.is_internal ? 'bg-blue-100 dark:bg-blue-900/30 text-blue-600' : 'bg-green-100 dark:bg-green-900/30 text-green-600'} flex items-center justify-center text-xl font-medium">
            {selectedContact.name ? selectedContact.name[0].toUpperCase() : selectedContact.email[0].toUpperCase()}
          </div>
          <div>
            <h2 class="text-xl font-semibold">{selectedContact.name || 'No name'}</h2>
            <div class="text-sm text-[var(--color-muted)]">{selectedContact.email}</div>
          </div>
        </div>

        <div class="space-y-3 mb-6">
          {#if selectedContact.company}
            <div class="flex items-center gap-2 text-sm">
              <Building2 class="w-4 h-4 text-[var(--color-muted)]" />
              <span>{selectedContact.company}</span>
            </div>
          {/if}
          <div class="flex items-center gap-2 text-sm">
            <span class="px-2 py-1 rounded text-xs {selectedContact.is_internal ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400' : 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'}">
              {selectedContact.is_internal ? 'Internal' : 'External'}
            </span>
            <span class="text-[var(--color-muted)]">{selectedContact.domain}</span>
          </div>
          {#if selectedContact.account_name}
            <div class="flex items-center gap-2 text-sm text-purple-500">
              <Link class="w-4 h-4" />
              <span>Linked to {selectedContact.account_name}</span>
            </div>
          {/if}
        </div>

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

        <div>
          <h3 class="font-semibold mb-3 flex items-center gap-2">
            <FileText class="w-4 h-4" />
            Meeting History
          </h3>
          {#if loadingNotes}
            <div class="flex justify-center py-4">
              <div class="animate-spin w-5 h-5 border-2 border-primary-500 border-t-transparent rounded-full"></div>
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
                  <div class="text-xs text-[var(--color-muted)]">{formatDate(note.meeting_date || note.created_at)}</div>
                </button>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>

<!-- Link Domain to Account Modal -->
{#if linkingDomain}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => linkingDomain = null}
      aria-label="Close"
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-md">
      <h2 class="text-lg font-semibold mb-2">Link {linkingDomain.contact_count} contacts</h2>
      <p class="text-sm text-[var(--color-muted)] mb-4">
        Choose an account for all <strong>{linkingDomain.domain}</strong> contacts
      </p>
      <div class="space-y-2 max-h-64 overflow-y-auto mb-4">
        {#each accounts as account}
          <button
            class="w-full p-3 text-left rounded-lg border border-[var(--color-border)] hover:border-primary-500 transition-colors"
            on:click={() => linkingDomain && linkDomainToAccount(linkingDomain.domain, account.id)}
          >
            <div class="font-medium">{account.name}</div>
          </button>
        {/each}
      </div>
      <div class="flex justify-end">
        <button class="btn-secondary" on:click={() => linkingDomain = null}>Cancel</button>
      </div>
    </div>
  </div>
{/if}

<!-- Edit Modal -->
{#if editingContact}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => editingContact = null}
      aria-label="Close"
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-md">
      <h2 class="text-lg font-semibold mb-4">Edit Contact</h2>
      <div class="space-y-4">
        <div>
          <label for="edit-email" class="block text-sm font-medium mb-1">Email</label>
          <input id="edit-email" type="email" class="input bg-[var(--color-border)]/50" value={editingContact.email} disabled />
        </div>
        <div>
          <label for="edit-name" class="block text-sm font-medium mb-1">Name</label>
          <input id="edit-name" type="text" class="input" placeholder="Full name" bind:value={editName} />
        </div>
        <div>
          <label for="edit-company" class="block text-sm font-medium mb-1">Company</label>
          <input id="edit-company" type="text" class="input" placeholder="Company" bind:value={editCompany} />
        </div>
      </div>
      <div class="mt-6 flex justify-end gap-3">
        <button class="btn-secondary" on:click={() => editingContact = null}>Cancel</button>
        <button class="btn-primary" on:click={saveEdit}>Save</button>
      </div>
    </div>
  </div>
{/if}

<!-- Delete Modal -->
{#if deletingContact}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => deletingContact = null}
      aria-label="Close"
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-md">
      <h2 class="text-lg font-semibold mb-2 text-red-500">Delete Contact?</h2>
      <p class="text-[var(--color-muted)] mb-4">{deletingContact.email}</p>
      <p class="text-sm text-[var(--color-muted)] mb-6">Meeting history will be preserved in notes.</p>
      <div class="flex justify-end gap-3">
        <button class="btn-secondary" on:click={() => deletingContact = null}>Cancel</button>
        <button class="btn-danger" on:click={confirmDelete}>Delete</button>
      </div>
    </div>
  </div>
{/if}
