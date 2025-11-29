<script lang="ts">
  import { onMount } from 'svelte';
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
    Filter
  } from 'lucide-svelte';
  import { api, type Contact, type ContactStats, type Account } from '$lib/utils/api';
  import { addToast } from '$lib/stores';

  let contacts: Contact[] = [];
  let stats: ContactStats | null = null;
  let accounts: Account[] = [];
  let loading = true;
  let filter: 'all' | 'internal' | 'external' | 'suggestions' = 'all';
  let searchQuery = '';
  let linkingContact: Contact | null = null;

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

  $: filteredContacts = contacts.filter(c => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return c.email.toLowerCase().includes(q) ||
           c.name.toLowerCase().includes(q) ||
           c.company.toLowerCase().includes(q) ||
           c.domain.toLowerCase().includes(q);
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

<div class="max-w-6xl mx-auto">
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
            <span class="text-sm font-normal text-[var(--color-muted)]">(@factory.ai)</span>
          </h2>
          <div class="grid gap-3">
            {#each internalContacts as contact}
              <div class="card p-4 flex items-center justify-between">
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
                <div class="text-right text-sm text-[var(--color-muted)]">
                  <div class="flex items-center gap-1">
                    <Calendar class="w-3 h-3" />
                    {contact.meeting_count} meetings
                  </div>
                  <div>Last seen {formatDate(contact.last_seen)}</div>
                </div>
              </div>
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
              <div class="card p-4">
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
                  <div class="flex items-center gap-4">
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
                          on:click={() => confirmSuggestion(contact, true)}
                          title="Confirm"
                        >
                          <Check class="w-4 h-4" />
                        </button>
                        <button
                          class="p-1 hover:bg-red-100 dark:hover:bg-red-900/30 rounded text-red-500"
                          on:click={() => confirmSuggestion(contact, false)}
                          title="Dismiss"
                        >
                          <X class="w-4 h-4" />
                        </button>
                      </div>
                    {:else}
                      <button
                        class="btn-secondary btn-sm flex items-center gap-1"
                        on:click={() => linkingContact = contact}
                      >
                        <Link class="w-3 h-3" />
                        Link
                      </button>
                    {/if}
                    <div class="text-right text-sm text-[var(--color-muted)]">
                      <div>{contact.meeting_count} meetings</div>
                    </div>
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
