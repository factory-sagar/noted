<script lang="ts">
  import { onMount } from 'svelte';
  import { FileText, Users, CheckSquare, AlertCircle, TrendingUp, ArrowRight, ArrowUpRight } from 'lucide-svelte';
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
  <title>Dashboard - Noted</title>
</svelte:head>

<div class="max-w-6xl mx-auto">
  <!-- Header -->
  <div class="page-header">
    <div class="divider-accent mb-6"></div>
    <h1 class="page-title">Dashboard</h1>
    <p class="page-subtitle">Your workspace at a glance</p>
  </div>

  {#if loading}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-12">
      {#each [1, 2, 3, 4] as _}
        <div class="card p-8">
          <div class="skeleton h-4 w-24 mb-4"></div>
          <div class="skeleton h-10 w-16"></div>
        </div>
      {/each}
    </div>
  {:else if error}
    <div class="card border-red-500/30 bg-red-500/5">
      <div class="flex items-start gap-4">
        <div class="p-2 bg-red-500/10 border border-red-500/20" style="border-radius: 2px;">
          <AlertCircle class="w-5 h-5 text-red-500" strokeWidth={1.5} />
        </div>
        <div>
          <h3 class="font-medium text-red-600 dark:text-red-400 mb-1">Connection Error</h3>
          <p class="text-sm text-[var(--color-muted)]">
            Make sure the backend server is running on port 8080.
          </p>
        </div>
      </div>
    </div>
  {:else if analytics}
    <!-- Stats Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-16 animate-stagger">
      <a href="/notes" class="stat-card group hover-lift">
        <div class="flex items-start justify-between mb-6">
          <div class="p-3 bg-[var(--color-bg)] border border-[var(--color-border)] group-hover:border-[var(--color-accent)]/40 transition-colors" style="border-radius: 2px;">
            <FileText class="w-5 h-5 text-[var(--color-accent)]" strokeWidth={1.5} />
          </div>
          <ArrowUpRight class="w-4 h-4 text-[var(--color-muted)] opacity-0 group-hover:opacity-100 transition-opacity" strokeWidth={1.5} />
        </div>
        <p class="stat-value">{analytics.total_notes}</p>
        <p class="stat-label">Total Notes</p>
      </a>

      <a href="/accounts" class="stat-card group hover-lift">
        <div class="flex items-start justify-between mb-6">
          <div class="p-3 bg-[var(--color-bg)] border border-[var(--color-border)] group-hover:border-emerald-500/40 transition-colors" style="border-radius: 2px;">
            <Users class="w-5 h-5 text-emerald-600 dark:text-emerald-400" strokeWidth={1.5} />
          </div>
          <ArrowUpRight class="w-4 h-4 text-[var(--color-muted)] opacity-0 group-hover:opacity-100 transition-opacity" strokeWidth={1.5} />
        </div>
        <p class="stat-value">{analytics.total_accounts}</p>
        <p class="stat-label">Accounts</p>
      </a>

      <a href="/todos" class="stat-card group hover-lift">
        <div class="flex items-start justify-between mb-6">
          <div class="p-3 bg-[var(--color-bg)] border border-[var(--color-border)] group-hover:border-blue-500/40 transition-colors" style="border-radius: 2px;">
            <CheckSquare class="w-5 h-5 text-blue-600 dark:text-blue-400" strokeWidth={1.5} />
          </div>
          <ArrowUpRight class="w-4 h-4 text-[var(--color-muted)] opacity-0 group-hover:opacity-100 transition-opacity" strokeWidth={1.5} />
        </div>
        <p class="stat-value">{analytics.total_todos}</p>
        <p class="stat-label">Total Todos</p>
      </a>

      <a href="/todos" class="stat-card group hover-lift">
        <div class="flex items-start justify-between mb-6">
          <div class="p-3 bg-[var(--color-bg)] border border-[var(--color-border)] group-hover:border-violet-500/40 transition-colors" style="border-radius: 2px;">
            <TrendingUp class="w-5 h-5 text-violet-600 dark:text-violet-400" strokeWidth={1.5} />
          </div>
          <ArrowUpRight class="w-4 h-4 text-[var(--color-muted)] opacity-0 group-hover:opacity-100 transition-opacity" strokeWidth={1.5} />
        </div>
        <p class="stat-value">{getCompletionRate()}<span class="text-2xl">%</span></p>
        <p class="stat-label">Completion Rate</p>
      </a>
    </div>

    <!-- Two Column Layout -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-12">
      <!-- Notes by Account -->
      <div class="card">
        <div class="flex items-center justify-between mb-8">
          <div>
            <h2 class="font-serif text-xl tracking-tight">Notes by Account</h2>
            <p class="text-sm text-[var(--color-muted)] mt-1">Distribution across accounts</p>
          </div>
          <a href="/notes" class="text-sm text-[var(--color-accent)] editorial-underline flex items-center gap-1 hover:gap-2 transition-all">
            View all <ArrowRight class="w-3 h-3" strokeWidth={1.5} />
          </a>
        </div>
        {#if !analytics.notes_by_account || analytics.notes_by_account.length === 0}
          <p class="text-[var(--color-muted)]">No accounts yet. Create one to get started.</p>
        {:else}
          <div class="space-y-4">
            {#each analytics.notes_by_account.slice(0, 5) as item, i}
              <div class="flex items-center gap-4 group">
                <span class="text-xs text-[var(--color-muted)] w-4">{String(i + 1).padStart(2, '0')}</span>
                <div class="flex-1 flex items-center justify-between py-3 border-b border-[var(--color-border)] group-hover:border-[var(--color-accent)]/30 transition-colors">
                  <span class="font-medium">{item.account_name || 'Unknown'}</span>
                  <span class="text-[var(--color-muted)] tabular-nums">{item.note_count}</span>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Todos by Status -->
      <div class="card">
        <div class="flex items-center justify-between mb-8">
          <div>
            <h2 class="font-serif text-xl tracking-tight">Todo Status</h2>
            <p class="text-sm text-[var(--color-muted)] mt-1">Current workflow state</p>
          </div>
          <a href="/todos" class="text-sm text-[var(--color-accent)] editorial-underline flex items-center gap-1 hover:gap-2 transition-all">
            View all <ArrowRight class="w-3 h-3" strokeWidth={1.5} />
          </a>
        </div>
        <div class="space-y-5">
          <div class="flex items-center gap-4">
            <div class="w-3 h-3 bg-charcoal-400" style="border-radius: 1px;"></div>
            <div class="flex-1 flex items-center justify-between">
              <span>Not Started</span>
              <span class="font-serif text-xl tabular-nums">{analytics.todos_by_status['not_started'] || 0}</span>
            </div>
          </div>
          <div class="flex items-center gap-4">
            <div class="w-3 h-3 bg-blue-500" style="border-radius: 1px;"></div>
            <div class="flex-1 flex items-center justify-between">
              <span>In Progress</span>
              <span class="font-serif text-xl tabular-nums">{analytics.todos_by_status['in_progress'] || 0}</span>
            </div>
          </div>
          <div class="flex items-center gap-4">
            <div class="w-3 h-3 bg-red-500" style="border-radius: 1px;"></div>
            <div class="flex-1 flex items-center justify-between">
              <span>Stuck</span>
              <span class="font-serif text-xl tabular-nums">{analytics.todos_by_status['stuck'] || 0}</span>
            </div>
          </div>
          <div class="flex items-center gap-4">
            <div class="w-3 h-3 bg-emerald-500" style="border-radius: 1px;"></div>
            <div class="flex-1 flex items-center justify-between">
              <span>Completed</span>
              <span class="font-serif text-xl tabular-nums">{analytics.todos_by_status['completed'] || 0}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Incomplete Fields -->
    <div class="card-accent">
      <div class="flex items-center justify-between mb-8">
        <div class="flex items-center gap-4">
          <div>
            <h2 class="font-serif text-xl tracking-tight">Incomplete Fields</h2>
            <p class="text-sm text-[var(--color-muted)] mt-1">Notes that need attention</p>
          </div>
          {#if incompleteFields.length > 0}
            <span class="tag-accent">
              {incompleteFields.length}
            </span>
          {/if}
        </div>
      </div>
      {#if incompleteFields.length === 0}
        <div class="flex items-center gap-3 text-emerald-600 dark:text-emerald-400">
          <CheckSquare class="w-5 h-5" strokeWidth={1.5} />
          <span>All notes have complete information</span>
        </div>
      {:else}
        <div class="space-y-3">
          {#each incompleteFields.slice(0, 5) as item}
            <a 
              href="/notes/{item.note_id}"
              class="flex items-center justify-between p-4 bg-[var(--color-bg)] border border-[var(--color-border)] hover:border-[var(--color-accent)]/40 transition-all group"
              style="border-radius: 2px;"
            >
              <div>
                <p class="font-medium group-hover:text-[var(--color-accent)] transition-colors">{item.note_title}</p>
                <p class="text-sm text-[var(--color-muted)]">{item.account_name}</p>
              </div>
              <div class="flex flex-wrap gap-1.5 justify-end">
                {#each item.missing_fields as field}
                  <span class="text-xs px-2 py-0.5 bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-300 rounded whitespace-nowrap">
                    {field.replace('_', ' ')}
                  </span>
                {/each}
              </div>
            </a>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>
