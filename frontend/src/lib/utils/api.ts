const API_BASE = 'http://localhost:8080/api';

async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(error.error || `HTTP ${response.status}`);
  }

  return response.json();
}

// Account types
export interface Account {
  id: string;
  name: string;
  account_owner: string;
  budget?: number;
  est_engineers?: number;
  created_at: string;
  updated_at: string;
}

export interface CreateAccountRequest {
  name: string;
  account_owner?: string;
  budget?: number;
  est_engineers?: number;
}

// Note types
export interface Note {
  id: string;
  title: string;
  account_id: string;
  account_name?: string;
  template_type: 'initial' | 'followup';
  internal_participants: string[];
  external_participants: string[];
  content: string;
  meeting_id?: string;
  meeting_date?: string;
  created_at: string;
  updated_at: string;
  todos?: Todo[];
}

export interface CreateNoteRequest {
  title: string;
  account_id: string;
  template_type?: string;
  internal_participants?: string[];
  external_participants?: string[];
  content?: string;
  meeting_id?: string;
  meeting_date?: string;
}

// Todo types
export interface Todo {
  id: string;
  title: string;
  description: string;
  status: 'not_started' | 'in_progress' | 'completed';
  priority: 'low' | 'medium' | 'high';
  due_date?: string;
  created_at: string;
  updated_at: string;
  linked_notes?: { id: string; title: string }[];
}

export interface CreateTodoRequest {
  title: string;
  description?: string;
  status?: string;
  priority?: string;
  due_date?: string;
  note_id?: string;
}

// Analytics types
export interface Analytics {
  total_notes: number;
  total_accounts: number;
  total_todos: number;
  todos_by_status: Record<string, number>;
  notes_by_account: { account_id: string; account_name: string; note_count: number }[];
  incomplete_count: number;
}

export interface IncompleteField {
  note_id: string;
  note_title: string;
  account_name: string;
  missing_fields: string[];
}

export interface SearchResult {
  type: 'note' | 'account' | 'todo';
  id: string;
  title: string;
  snippet?: string;
  account_id?: string;
}

// Calendar types
export interface CalendarConfig {
  connected: boolean;
  email?: string;
}

export interface CalendarEvent {
  id: string;
  title: string;
  description: string;
  start_time: string;
  end_time: string;
  attendees: string[];
  meet_link?: string;
}

export interface ParsedParticipants {
  internal: string[];
  external: string[];
}

// API functions
export const api = {
  // Accounts
  getAccounts: () => request<Account[]>('/accounts'),
  getAccount: (id: string) => request<Account>(`/accounts/${id}`),
  createAccount: (data: CreateAccountRequest) => 
    request<Account>('/accounts', { method: 'POST', body: JSON.stringify(data) }),
  updateAccount: (id: string, data: Partial<CreateAccountRequest>) =>
    request<Account>(`/accounts/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteAccount: (id: string) =>
    request<{ message: string }>(`/accounts/${id}`, { method: 'DELETE' }),

  // Notes
  getNotes: () => request<Note[]>('/notes'),
  getNote: (id: string) => request<Note>(`/notes/${id}`),
  getNotesByAccount: (accountId: string) => request<Note[]>(`/accounts/${accountId}/notes`),
  createNote: (data: CreateNoteRequest) =>
    request<Note>('/notes', { method: 'POST', body: JSON.stringify(data) }),
  updateNote: (id: string, data: Partial<CreateNoteRequest>) =>
    request<Note>(`/notes/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteNote: (id: string) =>
    request<{ message: string }>(`/notes/${id}`, { method: 'DELETE' }),
  exportNote: (id: string, type: 'full' | 'minimal' = 'full') =>
    request<any>(`/notes/${id}/export?type=${type}`),

  // Todos
  getTodos: (status?: string) => 
    request<Todo[]>(`/todos${status ? `?status=${status}` : ''}`),
  getTodo: (id: string) => request<Todo>(`/todos/${id}`),
  createTodo: (data: CreateTodoRequest) =>
    request<Todo>('/todos', { method: 'POST', body: JSON.stringify(data) }),
  updateTodo: (id: string, data: Partial<CreateTodoRequest>) =>
    request<Todo>(`/todos/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteTodo: (id: string) =>
    request<{ message: string }>(`/todos/${id}`, { method: 'DELETE' }),
  linkTodoToNote: (todoId: string, noteId: string) =>
    request<{ message: string }>(`/todos/${todoId}/notes/${noteId}`, { method: 'POST' }),
  unlinkTodoFromNote: (todoId: string, noteId: string) =>
    request<{ message: string }>(`/todos/${todoId}/notes/${noteId}`, { method: 'DELETE' }),

  // Search
  search: (query: string) => request<SearchResult[]>(`/search?q=${encodeURIComponent(query)}`),

  // Analytics
  getAnalytics: () => request<Analytics>('/analytics'),
  getIncompleteFields: () => request<IncompleteField[]>('/analytics/incomplete'),

  // Calendar
  getCalendarAuthURL: () => request<{ url: string }>('/calendar/auth'),
  getCalendarConfig: () => request<CalendarConfig>('/calendar/config'),
  disconnectCalendar: () => request<{ message: string }>('/calendar/disconnect', { method: 'DELETE' }),
  getCalendarEvents: (start?: string, end?: string) => {
    const params = new URLSearchParams();
    if (start) params.append('start', start);
    if (end) params.append('end', end);
    const query = params.toString();
    return request<CalendarEvent[]>(`/calendar/events${query ? `?${query}` : ''}`);
  },
  getCalendarEvent: (eventId: string) => request<CalendarEvent>(`/calendar/events/${eventId}`),
  parseParticipants: (attendees: string[], internalDomain?: string) =>
    request<ParsedParticipants>('/calendar/parse-participants', {
      method: 'POST',
      body: JSON.stringify({ attendees, internal_domain: internalDomain })
    }),
};
