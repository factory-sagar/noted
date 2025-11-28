<script lang="ts">
  import { onMount } from 'svelte';
  import { 
    ChevronLeft, 
    ChevronRight, 
    Plus,
    Calendar as CalendarIcon,
    FileText,
    Settings
  } from 'lucide-svelte';
  import { api, type Note } from '$lib/utils/api';
  import { addToast } from '$lib/stores';

  type ViewMode = 'month' | 'week' | 'day';

  let viewMode: ViewMode = 'month';
  let currentDate = new Date();
  let notes: Note[] = [];
  let loading = true;
  let calendarConnected = false;

  const weekDays = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
  const months = ['January', 'February', 'March', 'April', 'May', 'June', 
                  'July', 'August', 'September', 'October', 'November', 'December'];

  onMount(async () => {
    await loadNotes();
  });

  async function loadNotes() {
    try {
      loading = true;
      notes = await api.getNotes();
    } catch (e) {
      addToast('error', 'Failed to load notes');
    } finally {
      loading = false;
    }
  }

  function getDaysInMonth(date: Date): (Date | null)[] {
    const year = date.getFullYear();
    const month = date.getMonth();
    const firstDay = new Date(year, month, 1);
    const lastDay = new Date(year, month + 1, 0);
    const daysInMonth = lastDay.getDate();
    const startingDay = firstDay.getDay();

    const days: (Date | null)[] = [];
    
    // Add empty slots for days before the first day of the month
    for (let i = 0; i < startingDay; i++) {
      days.push(null);
    }
    
    // Add all days of the month
    for (let i = 1; i <= daysInMonth; i++) {
      days.push(new Date(year, month, i));
    }

    return days;
  }

  function getNotesForDate(date: Date): Note[] {
    return notes.filter(note => {
      if (!note.meeting_date) return false;
      const noteDate = new Date(note.meeting_date);
      return noteDate.toDateString() === date.toDateString();
    });
  }

  function previousMonth() {
    currentDate = new Date(currentDate.getFullYear(), currentDate.getMonth() - 1, 1);
  }

  function nextMonth() {
    currentDate = new Date(currentDate.getFullYear(), currentDate.getMonth() + 1, 1);
  }

  function goToToday() {
    currentDate = new Date();
  }

  function isToday(date: Date): boolean {
    const today = new Date();
    return date.toDateString() === today.toDateString();
  }

  function connectCalendar() {
    // TODO: Implement Google Calendar OAuth
    addToast('info', 'Google Calendar integration coming soon!');
  }

  $: calendarDays = getDaysInMonth(currentDate);
  $: monthYear = `${months[currentDate.getMonth()]} ${currentDate.getFullYear()}`;
</script>

<svelte:head>
  <title>Calendar - SE Notes</title>
</svelte:head>

<div class="max-w-7xl mx-auto">
  <div class="flex items-center justify-between mb-8">
    <div>
      <h1 class="page-title">Calendar</h1>
      <p class="page-subtitle">View meetings and linked notes</p>
    </div>
    <div class="flex items-center gap-3">
      {#if !calendarConnected}
        <button class="btn-secondary" on:click={connectCalendar}>
          <CalendarIcon class="w-4 h-4" />
          Connect Google Calendar
        </button>
      {/if}
    </div>
  </div>

  <!-- Calendar Header -->
  <div class="card mb-6">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-4">
        <button class="btn-ghost p-2" on:click={previousMonth}>
          <ChevronLeft class="w-5 h-5" />
        </button>
        <h2 class="text-xl font-semibold min-w-[200px] text-center">{monthYear}</h2>
        <button class="btn-ghost p-2" on:click={nextMonth}>
          <ChevronRight class="w-5 h-5" />
        </button>
      </div>
      <div class="flex items-center gap-2">
        <button class="btn-secondary text-sm" on:click={goToToday}>
          Today
        </button>
        <div class="flex bg-[var(--color-bg)] rounded-lg p-1 border border-[var(--color-border)]">
          <button 
            class="px-3 py-1 text-sm rounded-md transition-colors"
            class:bg-primary-600={viewMode === 'month'}
            class:text-white={viewMode === 'month'}
            on:click={() => viewMode = 'month'}
          >
            Month
          </button>
          <button 
            class="px-3 py-1 text-sm rounded-md transition-colors"
            class:bg-primary-600={viewMode === 'week'}
            class:text-white={viewMode === 'week'}
            on:click={() => viewMode = 'week'}
          >
            Week
          </button>
          <button 
            class="px-3 py-1 text-sm rounded-md transition-colors"
            class:bg-primary-600={viewMode === 'day'}
            class:text-white={viewMode === 'day'}
            on:click={() => viewMode = 'day'}
          >
            Day
          </button>
        </div>
      </div>
    </div>
  </div>

  {#if loading}
    <div class="card animate-pulse">
      <div class="grid grid-cols-7 gap-4">
        {#each Array(35) as _}
          <div class="h-24 bg-[var(--color-border)] rounded"></div>
        {/each}
      </div>
    </div>
  {:else}
    <!-- Month View -->
    {#if viewMode === 'month'}
      <div class="card p-0 overflow-hidden">
        <!-- Week day headers -->
        <div class="grid grid-cols-7 border-b border-[var(--color-border)]">
          {#each weekDays as day}
            <div class="py-3 text-center text-sm font-medium text-[var(--color-muted)]">
              {day}
            </div>
          {/each}
        </div>

        <!-- Calendar grid -->
        <div class="grid grid-cols-7">
          {#each calendarDays as day, i}
            <div 
              class="min-h-[100px] p-2 border-b border-r border-[var(--color-border)] last:border-r-0"
              class:bg-[var(--color-bg)]={day && isToday(day)}
            >
              {#if day}
                <div class="flex items-center justify-between mb-2">
                  <span 
                    class="text-sm font-medium"
                    class:text-primary-500={isToday(day)}
                  >
                    {day.getDate()}
                  </span>
                  {#if isToday(day)}
                    <span class="text-xs bg-primary-500 text-white px-1.5 py-0.5 rounded">Today</span>
                  {/if}
                </div>
                
                <!-- Notes for this day -->
                <div class="space-y-1">
                  {#each getNotesForDate(day).slice(0, 2) as note}
                    <a 
                      href="/notes/{note.id}"
                      class="block text-xs p-1.5 bg-primary-500/10 text-primary-600 dark:text-primary-400 rounded truncate hover:bg-primary-500/20 transition-colors"
                    >
                      {note.title}
                    </a>
                  {/each}
                  {#if getNotesForDate(day).length > 2}
                    <span class="text-xs text-[var(--color-muted)]">
                      +{getNotesForDate(day).length - 2} more
                    </span>
                  {/if}
                </div>
              {/if}
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <!-- Week/Day views placeholder -->
    {#if viewMode === 'week' || viewMode === 'day'}
      <div class="card text-center py-12">
        <CalendarIcon class="w-12 h-12 mx-auto text-[var(--color-muted)] mb-4" />
        <h3 class="text-lg font-medium mb-2">{viewMode === 'week' ? 'Week' : 'Day'} View</h3>
        <p class="text-[var(--color-muted)]">Coming soon with Google Calendar integration.</p>
      </div>
    {/if}
  {/if}

  <!-- Info Card -->
  {#if !calendarConnected}
    <div class="card mt-6 border-primary-500/50 bg-primary-500/5">
      <div class="flex items-start gap-4">
        <div class="p-3 bg-primary-500/10 rounded-lg">
          <CalendarIcon class="w-6 h-6 text-primary-500" />
        </div>
        <div>
          <h3 class="font-medium mb-1">Connect Google Calendar</h3>
          <p class="text-sm text-[var(--color-muted)] mb-3">
            Connect your Google Calendar to automatically sync meetings, see upcoming calls, 
            and quickly create notes from calendar events.
          </p>
          <button class="btn-primary text-sm" on:click={connectCalendar}>
            Connect Now
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>
