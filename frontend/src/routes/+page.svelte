<script lang="ts">
  import { onMount } from 'svelte';
  import { FileText, Users, CheckSquare, AlertCircle, TrendingUp, ArrowRight } from 'lucide-svelte';
  import { api, type Analytics, type IncompleteField } from '$lib/utils/api';

  let analytics: Analytics | null = null;
  let incompleteFields: IncompleteField[] = [];
  let loading = true;
  let error: string | null = null;

  onMount(async () => {
    try {
      const [analyticsData, incompleteData] = await Promise.all([
        api.getAnalytics(),
        api.getIncompleteFields()
      ]);
      analytics = analyticsData;
      incompleteFields = incompleteData;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load analytics';
    } finally {
      loading = false;
    }
  });

  function getCompletionRate(): number {
    if (!analytics || analytics.total_todos === 0) return 0;
    const completed = analytics.todos_by_status['completed'] || 0;
    return Math.round((completed / analytics.total_todos) * 100);
  }
</script>

<svelte:head>
  <title>Dashboard - SE Notes</title>
</svelte:head>

<div class="max-w-7xl mx-auto">
  <div class="mb-8">
    <h1 class="page-title">Dashboard</h1>
    <p class="page-subtitle">Overview of your notes and tasks</p>
  </div>

  {#if loading}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
      {#each [1, 2, 3, 4] as _}
        <div class="card animate-pulse">
          <div class="h-4 bg-[var(--color-border)] rounded w-24 mb-4"></div>
          <div class="h-8 bg-[var(--color-border)] rounded w-16"></div>
        </div>
      {/each}
    </div>
  {:else if error}
    <div class="card border-red-500/50 bg-red-500/10">
      <div class="flex items-center gap-3 text-red-500">
        <AlertCircle class="w-5 h-5" />
        <span>{error}</span>
      </div>
      <p class="mt-2 text-sm text-[var(--color-muted)]">
        Make sure the backend server is running on port 8080.
      </p>
    </div>
  {:else if analytics}
    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
      <div class="card">
        <div class="flex items-center justify-between mb-4">
          <span class="text-[var(--color-muted)] text-sm font-medium">Total Notes</span>
          <div class="p-2 bg-primary-500/10 rounded-lg">
            <FileText class="w-5 h-5 text-primary-500" />
          </div>
        </div>
        <p class="text-3xl font-semibold">{analytics.total_notes}</p>
      </div>

      <div class="card">
        <div class="flex items-center justify-between mb-4">
          <span class="text-[var(--color-muted)] text-sm font-medium">Accounts</span>
          <div class="p-2 bg-green-500/10 rounded-lg">
            <Users class="w-5 h-5 text-green-500" />
          </div>
        </div>
        <p class="text-3xl font-semibold">{analytics.total_accounts}</p>
      </div>

      <div class="card">
        <div class="flex items-center justify-between mb-4">
          <span class="text-[var(--color-muted)] text-sm font-medium">Total Todos</span>
          <div class="p-2 bg-orange-500/10 rounded-lg">
            <CheckSquare class="w-5 h-5 text-orange-500" />
          </div>
        </div>
        <p class="text-3xl font-semibold">{analytics.total_todos}</p>
      </div>

      <div class="card">
        <div class="flex items-center justify-between mb-4">
          <span class="text-[var(--color-muted)] text-sm font-medium">Completion Rate</span>
          <div class="p-2 bg-purple-500/10 rounded-lg">
            <TrendingUp class="w-5 h-5 text-purple-500" />
          </div>
        </div>
        <p class="text-3xl font-semibold">{getCompletionRate()}%</p>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Notes by Account -->
      <div class="card">
        <div class="flex items-center justify-between mb-6">
          <h2 class="text-lg font-semibold">Notes by Account</h2>
          <a href="/notes" class="text-sm text-primary-500 hover:text-primary-600 flex items-center gap-1">
            View all <ArrowRight class="w-4 h-4" />
          </a>
        </div>
        {#if analytics.notes_by_account.length === 0}
          <p class="text-[var(--color-muted)] text-sm">No accounts yet. Create one to get started.</p>
        {:else}
          <div class="space-y-4">
            {#each analytics.notes_by_account.slice(0, 5) as item}
              <div class="flex items-center justify-between">
                <span class="font-medium">{item.account_name || 'Unknown'}</span>
                <span class="text-[var(--color-muted)]">{item.note_count} notes</span>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Todos by Status -->
      <div class="card">
        <div class="flex items-center justify-between mb-6">
          <h2 class="text-lg font-semibold">Todos Status</h2>
          <a href="/todos" class="text-sm text-primary-500 hover:text-primary-600 flex items-center gap-1">
            View all <ArrowRight class="w-4 h-4" />
          </a>
        </div>
        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <div class="w-3 h-3 bg-gray-400 rounded-full"></div>
              <span>Not Started</span>
            </div>
            <span class="font-medium">{analytics.todos_by_status['not_started'] || 0}</span>
          </div>
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <div class="w-3 h-3 bg-blue-500 rounded-full"></div>
              <span>In Progress</span>
            </div>
            <span class="font-medium">{analytics.todos_by_status['in_progress'] || 0}</span>
          </div>
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <div class="w-3 h-3 bg-green-500 rounded-full"></div>
              <span>Completed</span>
            </div>
            <span class="font-medium">{analytics.todos_by_status['completed'] || 0}</span>
          </div>
        </div>
      </div>

      <!-- Incomplete Fields -->
      <div class="card lg:col-span-2">
        <div class="flex items-center justify-between mb-6">
          <div class="flex items-center gap-2">
            <h2 class="text-lg font-semibold">Incomplete Fields</h2>
            {#if incompleteFields.length > 0}
              <span class="px-2 py-0.5 text-xs font-medium bg-orange-500/10 text-orange-500 rounded-full">
                {incompleteFields.length}
              </span>
            {/if}
          </div>
        </div>
        {#if incompleteFields.length === 0}
          <div class="flex items-center gap-2 text-green-500">
            <CheckSquare class="w-5 h-5" />
            <span>All notes have complete information!</span>
          </div>
        {:else}
          <div class="space-y-3">
            {#each incompleteFields.slice(0, 5) as item}
              <a 
                href="/notes/{item.note_id}"
                class="flex items-center justify-between p-3 bg-[var(--color-bg)] rounded-lg hover:bg-[var(--color-border)] transition-colors"
              >
                <div>
                  <p class="font-medium">{item.note_title}</p>
                  <p class="text-sm text-[var(--color-muted)]">{item.account_name}</p>
                </div>
                <div class="flex flex-wrap gap-1 max-w-[200px]">
                  {#each item.missing_fields as field}
                    <span class="px-2 py-0.5 text-xs bg-orange-500/10 text-orange-500 rounded">
                      {field.replace('_', ' ')}
                    </span>
                  {/each}
                </div>
              </a>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>
