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
    X
  } from 'lucide-svelte';
  import { api, type Note, type CalendarEvent, type CalendarConfig } from '$lib/utils/api';
  import { addToast } from '$lib/stores';

  type ViewMode = 'month' | 'agenda';

  let viewMode: ViewMode = 'month';
  let currentDate = new Date();
  let notes: Note[] = [];
  let calendarEvents: CalendarEvent[] = [];
  let loading = true;
  let loadingEvents = false;
  let refreshing = false;
  let calendarConfig: CalendarConfig = { connected: false };
  
  // Modal states
  let showEventModal = false;
  let selectedEvent: CalendarEvent | null = null;
  let creatingNote = false;
  let showDayModal = false;
  let selectedDay: Date | null = null;

  const weekDays = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
  const months = ['January', 'February', 'March', 'April', 'May', 'June', 
                  'July', 'August', 'September', 'October', 'November', 'December'];

  onMount(async () => {
    if ($page.url.searchParams.get('calendar') === 'connected') {
      addToast('success', 'Google Calendar connected!');
      goto('/calendar', { replaceState: true });
    }
    
    await Promise.all([
      loadCalendarConfig(),
      loadNotes()
    ]);
  });

  async function loadCalendarConfig() {
    try {
      calendarConfig = await api.getCalendarConfig();
      if (calendarConfig.connected) {
        await loadCalendarEvents();
      }
    } catch (e) {
      console.error('Failed to load calendar config:', e);
    }
  }

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

  async function loadCalendarEvents() {
    if (!calendarConfig.connected) return;
    
    try {
      loadingEvents = true;
      const start = new Date(currentDate.getFullYear(), currentDate.getMonth(), 1);
      const end = new Date(currentDate.getFullYear(), currentDate.getMonth() + 2, 0);
      
      calendarEvents = await api.getCalendarEvents(
        start.toISOString(),
        end.toISOString()
      );
    } catch (e) {
      console.error('Failed to load calendar events:', e);
    } finally {
      loadingEvents = false;
    }
  }

  async function refreshCalendar() {
    refreshing = true;
    await Promise.all([loadCalendarEvents(), loadNotes()]);
    refreshing = false;
    addToast('success', 'Calendar refreshed');
  }

  async function connectCalendar() {
    try {
      const { url } = await api.getCalendarAuthURL();
      window.location.href = url;
    } catch (e: any) {
      if (e.message.includes('not configured')) {
        addToast('error', 'Google OAuth not configured. Set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET.');
      } else {
        addToast('error', 'Failed to connect calendar');
      }
    }
  }

  async function disconnectCalendar() {
    try {
      await api.disconnectCalendar();
      calendarConfig = { connected: false };
      calendarEvents = [];
      addToast('success', 'Calendar disconnected');
    } catch (e) {
      addToast('error', 'Failed to disconnect calendar');
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
    
    for (let i = 0; i < startingDay; i++) {
      days.push(null);
    }
    
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

  function getEventsForDate(date: Date): CalendarEvent[] {
    return calendarEvents.filter(event => {
      const eventDate = new Date(event.start_time);
      return eventDate.toDateString() === date.toDateString();
    });
  }

  function getUpcomingEvents(): CalendarEvent[] {
    const now = new Date();
    now.setHours(0, 0, 0, 0);
    return calendarEvents
      .filter(e => new Date(e.start_time) >= now)
      .sort((a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime());
  }

  function previousMonth() {
    currentDate = new Date(currentDate.getFullYear(), currentDate.getMonth() - 1, 1);
    if (calendarConfig.connected) {
      loadCalendarEvents();
    }
  }

  function nextMonth() {
    currentDate = new Date(currentDate.getFullYear(), currentDate.getMonth() + 1, 1);
    if (calendarConfig.connected) {
      loadCalendarEvents();
    }
  }

  function goToToday() {
    currentDate = new Date();
    if (calendarConfig.connected) {
      loadCalendarEvents();
    }
  }

  function isToday(date: Date): boolean {
    const today = new Date();
    return date.toDateString() === today.toDateString();
  }

  function formatTime(dateStr: string): string {
    const date = new Date(dateStr);
    return date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
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
        selectedEvent.attendees || [],
        'factory.ai'
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
        content: selectedEvent.description || '<p>Meeting notes...</p>'
      });
      
      showEventModal = false;
      addToast('success', 'Note created from calendar event');
      goto(`/notes/${note.id}`);
    } catch (e) {
      addToast('error', 'Failed to create note');
    } finally {
      creatingNote = false;
    }
  }

  // Group events by date for agenda view
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

  $: calendarDays = getDaysInMonth(currentDate);
  $: monthYear = `${months[currentDate.getMonth()]} ${currentDate.getFullYear()}`;
  $: upcomingEvents = getUpcomingEvents();
  $: groupedEvents = groupEventsByDate(upcomingEvents);
</script>

<svelte:head>
  <title>Calendar - Noted</title>
</svelte:head>

<div class="max-w-7xl mx-auto">
  <div class="flex items-center justify-between mb-8">
    <div>
      <h1 class="page-title">Calendar</h1>
      <p class="page-subtitle">
        {#if calendarConfig.connected && calendarConfig.email}
          Connected to {calendarConfig.email}
        {:else}
          View meetings and linked notes
        {/if}
      </p>
    </div>
    <div class="flex items-center gap-3">
      {#if calendarConfig.connected}
        <button 
          class="btn-secondary" 
          on:click={refreshCalendar}
          disabled={refreshing}
          title="Refresh calendar"
        >
          <RefreshCw class="w-4 h-4 {refreshing ? 'animate-spin' : ''}" />
          {refreshing ? 'Refreshing...' : 'Refresh'}
        </button>
        <button class="btn-danger" on:click={disconnectCalendar}>
          <Unlink class="w-4 h-4" />
          Disconnect
        </button>
      {:else}
        <button class="btn-secondary" on:click={connectCalendar}>
          <CalendarIcon class="w-4 h-4" />
          Connect Google Calendar
        </button>
      {/if}
    </div>
  </div>

  <!-- Calendar Header -->
  <div class="card mb-6">
    <div class="flex items-center justify-between flex-wrap gap-4">
      <div class="flex items-center gap-4">
        <button class="btn-icon btn-icon-ghost" on:click={previousMonth} title="Previous month">
          <ChevronLeft class="w-5 h-5" />
        </button>
        <h2 class="text-xl font-semibold min-w-[200px] text-center">{monthYear}</h2>
        <button class="btn-icon btn-icon-ghost" on:click={nextMonth} title="Next month">
          <ChevronRight class="w-5 h-5" />
        </button>
      </div>
      <div class="flex items-center gap-2">
        <!-- View Toggle -->
        <div class="flex items-center border border-[var(--color-border)] rounded-lg overflow-hidden">
          <button
            class="px-3 py-1.5 text-sm flex items-center gap-1.5 transition-colors {viewMode === 'month' ? 'bg-primary-500 text-white' : 'hover:bg-[var(--color-border)]'}"
            on:click={() => viewMode = 'month'}
          >
            <Grid3X3 class="w-4 h-4" />
            Month
          </button>
          <button
            class="px-3 py-1.5 text-sm flex items-center gap-1.5 transition-colors {viewMode === 'agenda' ? 'bg-primary-500 text-white' : 'hover:bg-[var(--color-border)]'}"
            on:click={() => viewMode = 'agenda'}
          >
            <List class="w-4 h-4" />
            Agenda
          </button>
        </div>
        <button class="btn-secondary btn-sm" on:click={goToToday}>
          Today
        </button>
        {#if loadingEvents}
          <span class="text-sm text-[var(--color-muted)]">Loading...</span>
        {/if}
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
  {:else if viewMode === 'month'}
    <!-- Month View -->
    <div class="card p-0 overflow-hidden">
      <div class="grid grid-cols-7 border-b border-[var(--color-border)]">
        {#each weekDays as day}
          <div class="py-3 text-center text-sm font-medium text-[var(--color-muted)]">
            {day}
          </div>
        {/each}
      </div>

      <div class="grid grid-cols-7">
        {#each calendarDays as day, i}
          <div 
            class="min-h-[100px] p-2 border-b border-r border-[var(--color-border)] last:border-r-0"
            class:bg-[var(--color-bg)]={day && isToday(day)}
          >
            {#if day}
              {@const dayEvents = getEventsForDate(day)}
              {@const dayNotes = getNotesForDate(day)}
              {@const totalItems = dayEvents.length + dayNotes.length}
              
              <div class="flex items-center justify-between mb-1">
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
              
              <div class="space-y-1">
                {#each dayEvents.slice(0, 2) as event}
                  <button 
                    class="block w-full text-left text-xs p-1 bg-blue-500/10 text-blue-600 dark:text-blue-400 rounded truncate hover:bg-blue-500/20 transition-colors"
                    on:click={() => openEventModal(event)}
                  >
                    <span class="font-medium">{formatTime(event.start_time)}</span> {event.title}
                  </button>
                {/each}
                
                {#each dayNotes.slice(0, dayEvents.length >= 2 ? 0 : 1) as note}
                  <a 
                    href="/notes/{note.id}"
                    class="block text-xs p-1 bg-primary-500/10 text-primary-600 dark:text-primary-400 rounded truncate hover:bg-primary-500/20 transition-colors"
                  >
                    <FileText class="w-3 h-3 inline mr-1" />
                    {note.title}
                  </a>
                {/each}
                
                {#if totalItems > 2}
                  <button 
                    class="text-xs text-[var(--color-muted)] hover:text-primary-500 transition-colors"
                    on:click={() => openDayModal(day)}
                  >
                    +{totalItems - 2} more
                  </button>
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  {:else}
    <!-- Agenda View -->
    <div class="space-y-4">
      {#if upcomingEvents.length === 0}
        <div class="card p-12 text-center">
          <CalendarIcon class="w-12 h-12 mx-auto mb-4 text-[var(--color-muted)]" />
          <h3 class="text-lg font-medium mb-2">No upcoming events</h3>
          <p class="text-[var(--color-muted)]">
            {#if calendarConfig.connected}
              No events found for this month
            {:else}
              Connect your calendar to see events
            {/if}
          </p>
        </div>
      {:else}
        {#each [...groupedEvents.entries()] as [dateKey, events]}
          <div class="card">
            <h3 class="font-semibold text-lg mb-4 pb-2 border-b border-[var(--color-border)]">
              {formatDateHeader(events[0].start_time)}
            </h3>
            <div class="space-y-3">
              {#each events as event}
                <div
                  class="w-full text-left p-4 rounded-lg border border-[var(--color-border)] hover:border-primary-500 hover:bg-primary-50 dark:hover:bg-primary-900/20 transition-colors cursor-pointer"
                  role="button"
                  tabindex="0"
                  on:click={() => openEventModal(event)}
                  on:keydown={(e) => e.key === 'Enter' && openEventModal(event)}
                >
                  <div class="flex items-start justify-between gap-4">
                    <div class="flex-1 min-w-0">
                      <div class="font-medium truncate">{event.title}</div>
                      <div class="text-sm text-[var(--color-muted)] flex items-center gap-2 mt-1">
                        <Clock class="w-3.5 h-3.5" />
                        {formatTime(event.start_time)}
                        {#if event.attendees && event.attendees.length > 0}
                          <span class="mx-1">-</span>
                          <Users class="w-3.5 h-3.5" />
                          {event.attendees.length} attendees
                        {/if}
                      </div>
                    </div>
                    <button
                      class="btn-primary btn-sm shrink-0"
                      on:click|stopPropagation={() => { selectedEvent = event; createNoteFromEvent(); }}
                    >
                      <Plus class="w-3.5 h-3.5" />
                      Note
                    </button>
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {/each}
      {/if}
    </div>
  {/if}

  <!-- Info Card when not connected -->
  {#if !calendarConfig.connected}
    <div class="card mt-6 border-primary-500/50 bg-primary-500/5">
      <div class="flex items-start gap-4">
        <div class="p-3 bg-primary-500/10 rounded-lg">
          <CalendarIcon class="w-6 h-6 text-primary-500" />
        </div>
        <div>
          <h3 class="font-medium mb-1">Connect Google Calendar</h3>
          <p class="text-sm text-[var(--color-muted)] mb-3">
            Connect your Google Calendar to automatically sync meetings, see upcoming calls, 
            and quickly create notes from calendar events with auto-populated participants.
          </p>
          <button class="btn-primary btn-sm" on:click={connectCalendar}>
            Connect Now
          </button>
        </div>
      </div>
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
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-lg animate-slide-up">
      <h2 class="text-lg font-semibold mb-2">{selectedEvent.title}</h2>
      
      <div class="space-y-3 mb-6">
        <div class="flex items-center gap-2 text-sm text-[var(--color-muted)]">
          <Clock class="w-4 h-4" />
          {new Date(selectedEvent.start_time).toLocaleString('en-US', {
            weekday: 'long',
            month: 'long',
            day: 'numeric',
            hour: 'numeric',
            minute: '2-digit'
          })}
        </div>
        
        {#if selectedEvent.attendees && selectedEvent.attendees.length > 0}
          <div class="flex items-start gap-2 text-sm">
            <Users class="w-4 h-4 mt-0.5 text-[var(--color-muted)]" />
            <div class="flex-1">
              <p class="text-[var(--color-muted)] mb-1">Attendees:</p>
              <div class="flex flex-wrap gap-1">
                {#each selectedEvent.attendees.slice(0, 5) as attendee}
                  <span class="px-2 py-0.5 bg-[var(--color-bg)] rounded text-xs">
                    {attendee}
                  </span>
                {/each}
                {#if selectedEvent.attendees.length > 5}
                  <span class="px-2 py-0.5 text-xs text-[var(--color-muted)]">
                    +{selectedEvent.attendees.length - 5} more
                  </span>
                {/if}
              </div>
            </div>
          </div>
        {/if}
        
        {#if selectedEvent.meet_link}
          <a 
            href={selectedEvent.meet_link}
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex items-center gap-2 text-sm text-primary-500 hover:underline"
          >
            <ExternalLink class="w-4 h-4" />
            Join Meeting
          </a>
        {/if}
        
        {#if selectedEvent.description}
          <div class="text-sm text-[var(--color-muted)] border-t border-[var(--color-border)] pt-3 mt-3">
            <p class="whitespace-pre-wrap">{selectedEvent.description.slice(0, 300)}{selectedEvent.description.length > 300 ? '...' : ''}</p>
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
          <Plus class="w-4 h-4" />
          {creatingNote ? 'Creating...' : 'Create Note'}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Day Detail Modal -->
{#if showDayModal && selectedDay}
  {@const dayEvents = getEventsForDate(selectedDay)}
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
                    - {event.attendees.length} attendees
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
    </div>
  </div>
{/if}
