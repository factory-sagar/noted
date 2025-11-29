import { writable, derived, get } from 'svelte/store';
import { api, type CalendarEvent, type CalendarConfig } from '$lib/utils/api';

interface CalendarState {
  config: CalendarConfig;
  events: CalendarEvent[];
  loading: boolean;
  error: string | null;
  lastFetch: number | null;
  currentMonth: Date;
}

const initialState: CalendarState = {
  config: { connected: false },
  events: [],
  loading: false,
  error: null,
  lastFetch: null,
  currentMonth: new Date()
};

function createCalendarStore() {
  const { subscribe, set, update } = writable<CalendarState>(initialState);

  return {
    subscribe,
    
    async init() {
      update(s => ({ ...s, loading: true, error: null }));
      try {
        const config = await api.getCalendarConfig();
        update(s => ({ ...s, config }));
        
        if (config.connected) {
          await this.fetchEvents();
        }
      } catch (e: any) {
        update(s => ({ ...s, error: e.message }));
      } finally {
        update(s => ({ ...s, loading: false }));
      }
    },

    async fetchEvents(force = false) {
      const state = get({ subscribe });
      
      // Skip if already loading
      if (state.loading && !force) return;
      
      // Skip if fetched recently (within 30 seconds) unless forced
      if (!force && state.lastFetch && Date.now() - state.lastFetch < 30000) {
        return;
      }

      update(s => ({ ...s, loading: true, error: null }));
      
      try {
        const currentMonth = state.currentMonth;
        const start = new Date(currentMonth.getFullYear(), currentMonth.getMonth() - 1, 1);
        const end = new Date(currentMonth.getFullYear(), currentMonth.getMonth() + 2, 0);
        
        const events = await api.getCalendarEvents(
          start.toISOString(),
          end.toISOString()
        );
        
        update(s => ({ 
          ...s, 
          events, 
          lastFetch: Date.now(),
          error: null
        }));
      } catch (e: any) {
        console.error('Failed to fetch calendar events:', e);
        update(s => ({ ...s, error: e.message }));
      } finally {
        update(s => ({ ...s, loading: false }));
      }
    },

    setMonth(date: Date) {
      update(s => ({ ...s, currentMonth: date }));
      this.fetchEvents(true);
    },

    previousMonth() {
      update(s => {
        const newDate = new Date(s.currentMonth.getFullYear(), s.currentMonth.getMonth() - 1, 1);
        return { ...s, currentMonth: newDate };
      });
      this.fetchEvents(true);
    },

    nextMonth() {
      update(s => {
        const newDate = new Date(s.currentMonth.getFullYear(), s.currentMonth.getMonth() + 1, 1);
        return { ...s, currentMonth: newDate };
      });
      this.fetchEvents(true);
    },

    goToToday() {
      update(s => ({ ...s, currentMonth: new Date() }));
      this.fetchEvents(true);
    },

    async refresh() {
      await this.fetchEvents(true);
    },

    async disconnect() {
      try {
        await api.disconnectCalendar();
        set(initialState);
      } catch (e: any) {
        update(s => ({ ...s, error: e.message }));
      }
    },

    reset() {
      set(initialState);
    }
  };
}

export const calendarStore = createCalendarStore();

// Derived stores for convenience
export const calendarEvents = derived(calendarStore, $s => $s.events);
export const calendarLoading = derived(calendarStore, $s => $s.loading);
export const calendarConnected = derived(calendarStore, $s => $s.config.connected);
export const calendarConfig = derived(calendarStore, $s => $s.config);
export const currentMonth = derived(calendarStore, $s => $s.currentMonth);
