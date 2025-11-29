<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { 
    ChevronLeft, 
    ChevronRight, 
    Plus,
    Calendar as CalendarIcon,
    FileText,
    ExternalLink,
    Users,
    Clock,
    Unlink,
    RefreshCw,
    List,
    Grid3X3,
    X,
    Video,
    MapPin
  } from 'lucide-svelte';
  import { api, type Note, type CalendarEvent } from '$lib/utils/api';
  import { addToast } from '$lib/stores';
  import { 
    calendarStore, 
    calendarEvents, 
    calendarLoading, 
    calendarConnected,
    calendarConfig,
    currentMonth 
  } from '$lib/stores/calendar';

  type ViewMode = 'month' | 'week' | 'agenda';

  let viewMode: ViewMode = 'week'; // Default to week view
  let notes: Note[] = [];
  let notesLoading = true;
  let refreshing = false;
  
  // Modal states
  let showEventModal = false;
  let selectedEvent: CalendarEvent | null = null;
  let creatingNote = false;
  let showDayModal = false;
  let selectedDay: Date | null = null;

  const weekDays = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
  const weekDaysFull = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
  const months = ['January', 'February', 'March', 'April', 'May', 'June', 
                  'July', 'August', 'September', 'October', 'November', 'December'];

  onMount(async () => {
    // Check for OAuth callback
    if ($page.url.searchParams.get('calendar') === 'connected') {
      addToast('success', 'Calendar connected!');
      goto('/calendar', { replaceState: true });
    }
    
    // Initialize calendar store
    await calendarStore.init();
    
    // Load notes
    try {
      notes = await api.getNotes();
    } catch (e) {
      console.error('Failed to load notes:', e);
    } finally {
      notesLoading = false;
    }
  });

  async function refreshCalendar() {
    refreshing = true;
    await calendarStore.refresh();
    try {
      notes = await api.getNotes();
    } catch (e) {
      console.error('Failed to refresh notes:', e);
    }
    refreshing = false;
    addToast('success', 'Calendar refreshed');
  }

  async function connectCalendar() {
    try {
      const result = await api.connectAppleCalendar();
      if (result.success) {
        addToast('success', 'Calendar connected!');
        await calendarStore.init();
      } else {
        addToast('error', result.message || 'Calendar access denied. Check System Settings > Privacy & Security > Calendars');
      }
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : 'Failed to connect calendar';
      addToast('error', message);
    }
  }

  function manageCalendarAccess() {
    addToast('info', 'To manage calendar access, go to System Settings > Privacy & Security > Calendars');
  }

  function getDaysInMonth(date: Date): (Date | null)[] {
    const year = date.getFullYear();
    const month = date.getMonth();
    const firstDay = new Date(year, month, 1);
    const lastDay = new Date(year, month + 1, 0);
    const daysInMonth = lastDay.getDate();
    const startingDay = firstDay.getDay();

    const days: (Date | null)[] = [];
    
    // Add empty cells for days before the first
    for (let i = 0; i < startingDay; i++) {
      days.push(null);
    }
    
    // Add all days of the month
    for (let i = 1; i <= daysInMonth; i++) {
      days.push(new Date(year, month, i));
    }

    // Add empty cells to complete the last row
    const remaining = 7 - (days.length % 7);
    if (remaining < 7) {
      for (let i = 0; i < remaining; i++) {
        days.push(null);
      }
    }

    return days;
  }

  function getWeekDays(date: Date): Date[] {
    const day = date.getDay();
    const diff = date.getDate() - day;
    const days: Date[] = [];
    for (let i = 0; i < 7; i++) {
      days.push(new Date(date.getFullYear(), date.getMonth(), diff + i));
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

  function getEventsForDate(date: Date, events: CalendarEvent[]): CalendarEvent[] {
    return events.filter(event => {
      const eventDate = new Date(event.start_time);
      return eventDate.toDateString() === date.toDateString();
    }).sort((a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime());
  }

  function getUpcomingEvents(events: CalendarEvent[]): CalendarEvent[] {
    const now = new Date();
    now.setHours(0, 0, 0, 0);
    return events
      .filter(e => new Date(e.start_time) >= now)
      .sort((a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime());
  }

  function isToday(date: Date): boolean {
    const today = new Date();
    return date.toDateString() === today.toDateString();
  }

  function isCurrentMonth(date: Date, currentDate: Date): boolean {
    return date.getMonth() === currentDate.getMonth();
  }

  function formatTime(dateStr: string): string {
    const date = new Date(dateStr);
    return date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
  }

  function formatTimeRange(start: string, end: string): string {
    return `${formatTime(start)} - ${formatTime(end)}`;
  }

  function formatDateHeader(dateStr: string): string {
    const date = new Date(dateStr);
    const today = new Date();
    const tomorrow = new Date(today);
    tomorrow.setDate(tomorrow.getDate() + 1);
    
    if (date.toDateString() === today.toDateString()) return 'Today';
    if (date.toDateString() === tomorrow.toDateString()) return 'Tomorrow';
    
    return date.toLocaleDateString('en-US', { weekday: 'long', month: 'short', day: 'numeric' });
  }

  function openEventModal(event: CalendarEvent) {
    selectedEvent = event;
    showEventModal = true;
  }

  function openDayModal(day: Date) {
    selectedDay = day;
    showDayModal = true;
  }

  async function createNoteFromEvent() {
    if (!selectedEvent) return;
    
    try {
      creatingNote = true;
      
      const { internal, external } = await api.parseParticipants(
        selectedEvent.attendees || []
      );
      
      let accountId: string;
      const accounts = await api.getAccounts();
      
      const externalDomain = external[0]?.split('@')[1];
      const matchingAccount = accounts.find(a => 
        a.name.toLowerCase().includes(externalDomain?.split('.')[0] || '')
      );
      
      if (matchingAccount) {
        accountId = matchingAccount.id;
      } else {
        const accountName = externalDomain 
          ? externalDomain.split('.')[0].charAt(0).toUpperCase() + externalDomain.split('.')[0].slice(1)
          : selectedEvent.title;
        
        const newAccount = await api.createAccount({ name: accountName });
        accountId = newAccount.id;
      }
      
      const note = await api.createNote({
        title: selectedEvent.title,
        account_id: accountId,
        template_type: 'initial',
        internal_participants: internal,
        external_participants: external,
        meeting_id: selectedEvent.id,
        meeting_date: selectedEvent.start_time,
        content: '<p>Meeting notes...</p>'
      });
      
      showEventModal = false;
      addToast('success', 'Note created');
      goto(`/notes/${note.id}`);
    } catch (e) {
      addToast('error', 'Failed to create note');
    } finally {
      creatingNote = false;
    }
  }

  function groupEventsByDate(events: CalendarEvent[]): Map<string, CalendarEvent[]> {
    const grouped = new Map<string, CalendarEvent[]>();
    for (const event of events) {
      const dateKey = new Date(event.start_time).toDateString();
      if (!grouped.has(dateKey)) {
        grouped.set(dateKey, []);
      }
      grouped.get(dateKey)!.push(event);
    }
    return grouped;
  }

  function getEventColor(event: CalendarEvent): string {
    // Simple hash-based color assignment
    const colors = [
      'bg-blue-500/20 text-blue-700 dark:text-blue-300 border-l-blue-500',
      'bg-green-500/20 text-green-700 dark:text-green-300 border-l-green-500',
      'bg-purple-500/20 text-purple-700 dark:text-purple-300 border-l-purple-500',
      'bg-orange-500/20 text-orange-700 dark:text-orange-300 border-l-orange-500',
      'bg-pink-500/20 text-pink-700 dark:text-pink-300 border-l-pink-500',
      'bg-cyan-500/20 text-cyan-700 dark:text-cyan-300 border-l-cyan-500',
    ];
    const hash = event.title.split('').reduce((a, b) => a + b.charCodeAt(0), 0);
    return colors[hash % colors.length];
  }

  $: calendarDays = getDaysInMonth($currentMonth);
  $: weekDaysForView = getWeekDays($currentMonth);
  $: monthYear = `${months[$currentMonth.getMonth()]} ${$currentMonth.getFullYear()}`;
  $: upcomingEvents = getUpcomingEvents($calendarEvents);
  $: groupedEvents = groupEventsByDate(upcomingEvents);
</script>

<svelte:head>
  <title>Calendar - Noted</title>
</svelte:head>

<div class="h-full flex flex-col">
  <!-- Header -->
  <div class="flex items-center justify-between mb-4 px-1">
    <div class="flex items-center gap-4">
      <h1 class="text-2xl font-semibold">{monthYear}</h1>
      <div class="flex items-center">
        <button 
          class="p-2 hover:bg-[var(--color-border)] rounded-lg transition-colors" 
          on:click={() => calendarStore.previousMonth()}
          title="Previous"
        >
          <ChevronLeft class="w-5 h-5" />
        </button>
        <button 
          class="p-2 hover:bg-[var(--color-border)] rounded-lg transition-colors" 
          on:click={() => calendarStore.nextMonth()}
          title="Next"
        >
          <ChevronRight class="w-5 h-5" />
        </button>
      </div>
      <button 
        class="px-3 py-1.5 text-sm border border-[var(--color-border)] rounded-lg hover:bg-[var(--color-border)] transition-colors" 
        on:click={() => calendarStore.goToToday()}
      >
        Today
      </button>
    </div>
    
    <div class="flex items-center gap-3">
      <!-- View Toggle -->
      <div class="flex items-center border border-[var(--color-border)] rounded-lg overflow-hidden">
        <button
          class="px-3 py-1.5 text-sm transition-colors {viewMode === 'month' ? 'bg-[var(--color-border)]' : 'hover:bg-[var(--color-border)]/50'}"
          on:click={() => viewMode = 'month'}
        >
          Month
        </button>
        <button
          class="px-3 py-1.5 text-sm transition-colors {viewMode === 'week' ? 'bg-[var(--color-border)]' : 'hover:bg-[var(--color-border)]/50'}"
          on:click={() => viewMode = 'week'}
        >
          Week
        </button>
        <button
          class="px-3 py-1.5 text-sm transition-colors {viewMode === 'agenda' ? 'bg-[var(--color-border)]' : 'hover:bg-[var(--color-border)]/50'}"
          on:click={() => viewMode = 'agenda'}
        >
          Agenda
        </button>
      </div>

      <button 
        class="p-2 hover:bg-[var(--color-border)] rounded-lg transition-colors" 
        on:click={refreshCalendar}
        disabled={refreshing}
        title="Refresh"
      >
        <RefreshCw class="w-5 h-5 {refreshing ? 'animate-spin' : ''}" />
      </button>
      
      {#if $calendarConnected}
        <button 
          class="px-3 py-1.5 text-sm border border-[var(--color-border)] rounded-lg hover:bg-[var(--color-border)] transition-colors"
          on:click={manageCalendarAccess}
        >
          Manage Access
        </button>
      {:else}
        <button 
          class="btn-primary btn-sm"
          on:click={connectCalendar}
        >
          Connect Apple Calendar
        </button>
      {/if}
    </div>
  </div>

  <!-- Calendar Status -->
  {#if $calendarConfig.email}
    <p class="text-sm text-[var(--color-muted)] mb-4 px-1">
      Connected to {$calendarConfig.email}
    </p>
  {:else if $calendarConfig.connected && $calendarConfig.type === 'apple'}
    <p class="text-sm text-[var(--color-muted)] mb-4 px-1">
      Connected to Apple Calendar
    </p>
  {/if}

  <!-- Loading indicator -->
  {#if $calendarLoading && $calendarEvents.length === 0}
    <div class="flex-1 flex items-center justify-center">
      <div class="text-center">
        <RefreshCw class="w-8 h-8 mx-auto mb-3 animate-spin text-[var(--color-muted)]" />
        <p class="text-[var(--color-muted)]">Loading calendar...</p>
      </div>
    </div>
  {:else if !$calendarConnected}
    <!-- Not connected state -->
    <div class="flex-1 flex items-center justify-center">
      <div class="text-center max-w-md">
        <CalendarIcon class="w-16 h-16 mx-auto mb-4 text-[var(--color-muted)]" />
        <h2 class="text-xl font-semibold mb-2">Connect Apple Calendar</h2>
        <p class="text-[var(--color-muted)] mb-6">
          Connect to Apple Calendar to see your meetings, create notes from events, and keep track of your schedule. Your Google Calendar events will appear if synced to Apple Calendar.
        </p>
        <button class="btn-primary" on:click={connectCalendar}>
          <CalendarIcon class="w-4 h-4" />
          Connect Apple Calendar
        </button>
      </div>
    </div>
  {:else if viewMode === 'month'}
    <!-- Month View -->
    <div class="flex-1 border border-[var(--color-border)] rounded-lg overflow-hidden bg-[var(--color-card)]">
      <!-- Week day headers -->
      <div class="grid grid-cols-7 border-b border-[var(--color-border)] bg-[var(--color-bg)]">
        {#each weekDays as day}
          <div class="py-2 text-center text-xs font-medium text-[var(--color-muted)] uppercase tracking-wide">
            {day}
          </div>
        {/each}
      </div>

      <!-- Calendar grid -->
      <div class="grid grid-cols-7 flex-1">
        {#each calendarDays as day, i}
          {@const dayEvents = day ? getEventsForDate(day, $calendarEvents) : []}
          {@const dayNotes = day ? getNotesForDate(day) : []}
          {@const totalItems = dayEvents.length + dayNotes.length}
          
          <div 
            class="min-h-[100px] p-1 border-b border-r border-[var(--color-border)] {day && isToday(day) ? 'bg-blue-50 dark:bg-blue-900/10' : ''}"
          >
            {#if day}
              <div class="flex items-center justify-between mb-1 px-1">
                <span 
                  class="text-sm w-7 h-7 flex items-center justify-center rounded-full {isToday(day) ? 'bg-blue-500 text-white font-semibold' : ''}"
                >
                  {day.getDate()}
                </span>
              </div>
              
              <div class="space-y-0.5">
                {#each dayEvents.slice(0, 3) as event}
                  <button 
                    class="block w-full text-left text-xs px-1.5 py-0.5 rounded truncate border-l-2 {getEventColor(event)} hover:opacity-80 transition-opacity"
                    on:click={() => openEventModal(event)}
                  >
                    {formatTime(event.start_time)} {event.title}
                  </button>
                {/each}
                
                {#if totalItems > 3}
                  <button 
                    class="text-xs text-blue-500 hover:underline px-1.5"
                    on:click={() => day && openDayModal(day)}
                  >
                    +{totalItems - 3} more
                  </button>
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  {:else if viewMode === 'week'}
    <!-- Week View -->
    <div class="flex-1 border border-[var(--color-border)] rounded-lg overflow-hidden bg-[var(--color-card)]">
      <div class="grid grid-cols-7 border-b border-[var(--color-border)] bg-[var(--color-bg)]">
        {#each weekDaysForView as day, i}
          <div class="py-3 text-center border-r border-[var(--color-border)] last:border-r-0 {isToday(day) ? 'bg-blue-50 dark:bg-blue-900/10' : ''}">
            <div class="text-xs font-medium text-[var(--color-muted)] uppercase">{weekDays[i]}</div>
            <div class="text-2xl font-light {isToday(day) ? 'text-blue-500' : ''}">{day.getDate()}</div>
          </div>
        {/each}
      </div>
      
      <div class="grid grid-cols-7 flex-1 min-h-[400px]">
        {#each weekDaysForView as day, i}
          {@const dayEvents = getEventsForDate(day, $calendarEvents)}
          <div class="p-2 border-r border-[var(--color-border)] last:border-r-0 {isToday(day) ? 'bg-blue-50/50 dark:bg-blue-900/5' : ''}">
            <div class="space-y-2">
              {#each dayEvents as event}
                <button 
                  class="block w-full text-left p-2 rounded border-l-2 text-sm {getEventColor(event)} hover:opacity-80 transition-opacity"
                  on:click={() => openEventModal(event)}
                >
                  <div class="font-medium truncate">{event.title}</div>
                  <div class="text-xs opacity-75">{formatTime(event.start_time)}</div>
                </button>
              {/each}
            </div>
          </div>
        {/each}
      </div>
    </div>
  {:else}
    <!-- Agenda View -->
    <div class="flex-1 overflow-auto">
      {#if upcomingEvents.length === 0}
        <div class="text-center py-16">
          <CalendarIcon class="w-12 h-12 mx-auto mb-4 text-[var(--color-muted)]" />
          <h3 class="text-lg font-medium mb-2">No upcoming events</h3>
          <p class="text-[var(--color-muted)]">Your schedule is clear</p>
        </div>
      {:else}
        <div class="space-y-6">
          {#each [...groupedEvents.entries()] as [dateKey, events]}
            <div>
              <h3 class="text-sm font-semibold text-[var(--color-muted)] uppercase tracking-wide mb-3 px-1">
                {formatDateHeader(events[0].start_time)}
              </h3>
              <div class="space-y-2">
                {#each events as event}
                  <div
                    class="flex items-start gap-4 p-4 rounded-lg border border-[var(--color-border)] hover:border-blue-300 dark:hover:border-blue-700 transition-colors cursor-pointer bg-[var(--color-card)]"
                    role="button"
                    tabindex="0"
                    on:click={() => openEventModal(event)}
                    on:keydown={(e) => e.key === 'Enter' && openEventModal(event)}
                  >
                    <div class="text-sm text-[var(--color-muted)] w-20 shrink-0 pt-0.5">
                      {formatTime(event.start_time)}
                    </div>
                    <div class="flex-1 min-w-0">
                      <div class="font-medium">{event.title}</div>
                      {#if event.attendees && event.attendees.length > 0}
                        <div class="text-sm text-[var(--color-muted)] mt-1 flex items-center gap-1">
                          <Users class="w-3.5 h-3.5" />
                          {event.attendees.length} attendees
                        </div>
                      {/if}
                    </div>
                    <button
                      class="btn-primary btn-sm shrink-0"
                      on:click|stopPropagation={() => { selectedEvent = event; createNoteFromEvent(); }}
                    >
                      <Plus class="w-4 h-4" />
                      Note
                    </button>
                  </div>
                {/each}
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>

<!-- Event Modal -->
{#if showEventModal && selectedEvent}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button 
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => showEventModal = false}
      aria-label="Close modal"
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl w-full max-w-lg animate-slide-up overflow-hidden">
      <!-- Header with color bar -->
      <div class="h-2 bg-blue-500"></div>
      <div class="p-6">
        <h2 class="text-xl font-semibold mb-4">{selectedEvent.title}</h2>
        
        <div class="space-y-4 mb-6">
          <div class="flex items-center gap-3 text-sm">
            <Clock class="w-5 h-5 text-[var(--color-muted)]" />
            <div>
              <div>{new Date(selectedEvent.start_time).toLocaleDateString('en-US', {
                weekday: 'long',
                month: 'long',
                day: 'numeric'
              })}</div>
              <div class="text-[var(--color-muted)]">
                {formatTimeRange(selectedEvent.start_time, selectedEvent.end_time)}
              </div>
            </div>
          </div>
          
          {#if selectedEvent.meet_link}
            <a 
              href={selectedEvent.meet_link}
              target="_blank"
              rel="noopener noreferrer"
              class="flex items-center gap-3 text-sm text-blue-500 hover:underline"
            >
              <Video class="w-5 h-5" />
              Join video call
            </a>
          {/if}
          
          {#if selectedEvent.attendees && selectedEvent.attendees.length > 0}
            <div class="flex items-start gap-3 text-sm">
              <Users class="w-5 h-5 text-[var(--color-muted)] mt-0.5" />
              <div class="flex-1">
                <div class="text-[var(--color-muted)] mb-2">{selectedEvent.attendees.length} attendees</div>
                <div class="flex flex-wrap gap-1.5">
                  {#each selectedEvent.attendees.slice(0, 8) as attendee}
                    <span class="px-2 py-1 bg-[var(--color-bg)] rounded text-xs">
                      {attendee}
                    </span>
                  {/each}
                  {#if selectedEvent.attendees.length > 8}
                    <span class="px-2 py-1 text-xs text-[var(--color-muted)]">
                      +{selectedEvent.attendees.length - 8} more
                    </span>
                  {/if}
                </div>
              </div>
            </div>
          {/if}
          
          {#if selectedEvent.description}
            <div class="pt-4 border-t border-[var(--color-border)]">
              <p class="text-sm text-[var(--color-muted)] whitespace-pre-wrap">{selectedEvent.description.slice(0, 500)}{selectedEvent.description.length > 500 ? '...' : ''}</p>
            </div>
          {/if}
        </div>
        
        <div class="flex justify-end gap-3">
          <button 
            class="btn-secondary"
            on:click={() => showEventModal = false}
          >
            Close
          </button>
          <button 
            class="btn-primary"
            on:click={createNoteFromEvent}
            disabled={creatingNote}
          >
            <FileText class="w-4 h-4" />
            {creatingNote ? 'Creating...' : 'Create Note'}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

<!-- Day Detail Modal -->
{#if showDayModal && selectedDay}
  {@const dayEvents = getEventsForDate(selectedDay, $calendarEvents)}
  {@const dayNotes = getNotesForDate(selectedDay)}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button 
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => showDayModal = false}
      aria-label="Close modal"
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-lg animate-slide-up max-h-[80vh] overflow-auto">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold">
          {selectedDay.toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric' })}
        </h2>
        <button class="btn-icon btn-icon-ghost" on:click={() => showDayModal = false}>
          <X class="w-5 h-5" />
        </button>
      </div>
      
      {#if dayEvents.length > 0}
        <div class="mb-4">
          <h3 class="text-sm font-medium text-[var(--color-muted)] mb-2">Events</h3>
          <div class="space-y-2">
            {#each dayEvents as event}
              <button 
                class="w-full text-left p-3 rounded-lg border border-[var(--color-border)] hover:border-blue-500 transition-colors"
                on:click={() => { showDayModal = false; openEventModal(event); }}
              >
                <div class="font-medium text-sm">{event.title}</div>
                <div class="text-xs text-[var(--color-muted)] mt-1">
                  {formatTime(event.start_time)}
                  {#if event.attendees && event.attendees.length > 0}
                    · {event.attendees.length} attendees
                  {/if}
                </div>
              </button>
            {/each}
          </div>
        </div>
      {/if}
      
      {#if dayNotes.length > 0}
        <div>
          <h3 class="text-sm font-medium text-[var(--color-muted)] mb-2">Notes</h3>
          <div class="space-y-2">
            {#each dayNotes as note}
              <a 
                href="/notes/{note.id}"
                class="block p-3 rounded-lg border border-[var(--color-border)] hover:border-primary-500 transition-colors"
              >
                <div class="font-medium text-sm flex items-center gap-2">
                  <FileText class="w-4 h-4 text-primary-500" />
                  {note.title}
                </div>
              </a>
            {/each}
          </div>
        </div>
      {/if}
      
      {#if dayEvents.length === 0 && dayNotes.length === 0}
        <p class="text-center text-[var(--color-muted)] py-8">No events or notes for this day</p>
      {/if}
    </div>
  </div>
{/if}
