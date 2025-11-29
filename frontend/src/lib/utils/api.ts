// Detect if running in Wails (embedded in native app)
declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          GetServerPort: () => Promise<number>;
        };
      };
    };
  }
}

let cachedPort: number | null = null;

async function getApiBase(): Promise<string> {
  // Browser development mode
  if (typeof window === 'undefined' || !window.go?.main?.App) {
    return 'http://localhost:8080/api';
  }

  // Wails mode - get dynamic port from Go backend
  if (cachedPort === null) {
    try {
      cachedPort = await window.go.main.App.GetServerPort();
    } catch {
      cachedPort = 8080;
    }
  }
  return `http://127.0.0.1:${cachedPort}/api`;
}

// For sync access (fallback to default)
function getApiBaseSync(): string {
  if (cachedPort !== null) {
    return `http://127.0.0.1:${cachedPort}/api`;
  }
  return 'http://localhost:8080/api';
}

async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const apiBase = await getApiBase();
  const response = await fetch(`${apiBase}${endpoint}`, {
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
  status: 'not_started' | 'in_progress' | 'stuck' | 'completed';
  priority: 'low' | 'medium' | 'high';
  due_date?: string;
  account_id?: string;
  account_name?: string;
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
  account_id?: string;
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
  type?: 'google' | 'apple';
}

export interface AppleCalendar {
  id: string;
  title: string;
  color: string;
  type: string;
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

// Tag types
export interface Tag {
  id: string;
  name: string;
  color: string;
  created_at: string;
}

export interface CreateTagRequest {
  name: string;
  color?: string;
}

// Activity types
export interface Activity {
  id: string;
  account_id: string;
  type: string;
  title: string;
  description?: string;
  entity_type?: string;
  entity_id?: string;
  created_at: string;
}

export interface CreateActivityRequest {
  account_id: string;
  type: string;
  title: string;
  description?: string;
  entity_type?: string;
  entity_id?: string;
}

// Attachment types
export interface Attachment {
  id: string;
  note_id: string;
  filename: string;
  original_name: string;
  mime_type: string;
  size: number;
  created_at: string;
}

// Quick Capture types
export interface QuickCaptureRequest {
  type: 'note' | 'todo';
  title: string;
  content?: string;
  account_id?: string;
  priority?: string;
  description?: string;
}

// Contact types
export interface Contact {
  id: string;
  email: string;
  name: string;
  company: string;
  domain: string;
  is_internal: boolean;
  account_id?: string;
  account_name?: string;
  suggested_account_id?: string;
  suggested_account_name?: string;
  suggestion_confirmed: boolean;
  source: string;
  first_seen: string;
  last_seen: string;
  meeting_count: number;
  created_at: string;
  updated_at: string;
}

export interface ContactStats {
  total_contacts: number;
  internal_contacts: number;
  external_contacts: number;
  linked_contacts: number;
  pending_suggestions: number;
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
  restoreAccount: (id: string) =>
    request<{ message: string }>(`/accounts/${id}/restore`, { method: 'POST' }),
  permanentDeleteAccount: (id: string) =>
    request<{ message: string }>(`/accounts/${id}/permanent`, { method: 'DELETE' }),
  getDeletedAccounts: () => request<Account[]>('/accounts/deleted'),

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
  restoreNote: (id: string) =>
    request<{ message: string }>(`/notes/${id}/restore`, { method: 'POST' }),
  permanentDeleteNote: (id: string) =>
    request<{ message: string }>(`/notes/${id}/permanent`, { method: 'DELETE' }),
  getDeletedNotes: () => request<Note[]>('/notes/deleted'),
  exportNote: (id: string, type: 'full' | 'minimal' = 'full') =>
    request<any>(`/notes/${id}/export?type=${type}`),
  
  // Markdown
  exportNoteMarkdown: (id: string) => {
    // Return URL for direct download
    return `${getApiBaseSync()}/notes/${id}/export/markdown`;
  },
  importMarkdown: async (file: File): Promise<{ id: string; title: string }> => {
    const formData = new FormData();
    formData.append('file', file);
    const apiBase = await getApiBase();
    const response = await fetch(`${apiBase}/import/markdown`, {
      method: 'POST',
      body: formData,
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Import failed' }));
      throw new Error(error.error);
    }
    return response.json();
  },

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
  restoreTodo: (id: string) =>
    request<{ message: string }>(`/todos/${id}/restore`, { method: 'POST' }),
  permanentDeleteTodo: (id: string) =>
    request<{ message: string }>(`/todos/${id}/permanent`, { method: 'DELETE' }),
  getDeletedTodos: () => request<Todo[]>('/todos/deleted'),
  linkTodoToNote: (todoId: string, noteId: string) =>
    request<{ message: string }>(`/todos/${todoId}/notes/${noteId}`, { method: 'POST' }),
  unlinkTodoFromNote: (todoId: string, noteId: string) =>
    request<{ message: string }>(`/todos/${todoId}/notes/${noteId}`, { method: 'DELETE' }),

  // Search
  search: (query: string) => request<SearchResult[]>(`/search?q=${encodeURIComponent(query)}`),

  // Analytics
  getAnalytics: () => request<Analytics>('/analytics'),
  getIncompleteFields: () => request<IncompleteField[]>('/analytics/incomplete'),

  // Data management
  exportAllData: () => request<Record<string, unknown>>('/export'),
  clearAllData: () => request<{ message: string }>('/data', { method: 'DELETE' }),

  // Calendar (supports both Google OAuth and Apple EventKit)
  getCalendarAuthURL: () => request<{ url: string }>('/calendar/auth'),
  getCalendarConfig: () => request<CalendarConfig>('/calendar/config'),
  disconnectCalendar: () => request<{ message: string }>('/calendar/disconnect', { method: 'DELETE' }),
  connectAppleCalendar: () => request<{ success: boolean; message: string }>('/calendar/connect', { method: 'POST' }),
  getAppleCalendars: () => request<AppleCalendar[]>('/calendar/calendars'),
  getCalendarEvents: (start?: string, end?: string, calendarId?: string) => {
    const params = new URLSearchParams();
    if (start) params.append('start', start);
    if (end) params.append('end', end);
    if (calendarId) params.append('calendar_id', calendarId);
    const query = params.toString();
    return request<CalendarEvent[]>(`/calendar/events${query ? `?${query}` : ''}`);
  },
  getCalendarEvent: (eventId: string) => request<CalendarEvent>(`/calendar/events/${eventId}`),
  parseParticipants: (attendees: string[], internalDomain?: string) =>
    request<ParsedParticipants>('/calendar/parse-participants', {
      method: 'POST',
      body: JSON.stringify({ attendees, internal_domain: internalDomain })
    }),

  // Tags
  getTags: () => request<Tag[]>('/tags'),
  createTag: (data: CreateTagRequest) =>
    request<Tag>('/tags', { method: 'POST', body: JSON.stringify(data) }),
  updateTag: (id: string, data: Partial<CreateTagRequest>) =>
    request<Tag>(`/tags/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteTag: (id: string) =>
    request<{ message: string }>(`/tags/${id}`, { method: 'DELETE' }),
  getNoteTags: (noteId: string) => request<Tag[]>(`/notes/${noteId}/tags`),
  addTagToNote: (noteId: string, tagId: string) =>
    request<{ message: string }>(`/notes/${noteId}/tags/${tagId}`, { method: 'POST' }),
  removeTagFromNote: (noteId: string, tagId: string) =>
    request<{ message: string }>(`/notes/${noteId}/tags/${tagId}`, { method: 'DELETE' }),

  // Activities
  getActivities: (accountId: string, limit?: number) =>
    request<Activity[]>(`/accounts/${accountId}/activities${limit ? `?limit=${limit}` : ''}`),
  createActivity: (data: CreateActivityRequest) =>
    request<Activity>('/activities', { method: 'POST', body: JSON.stringify(data) }),

  // Attachments
  getAttachments: (noteId: string) => request<Attachment[]>(`/notes/${noteId}/attachments`),
  uploadAttachment: async (noteId: string, file: File): Promise<Attachment> => {
    const formData = new FormData();
    formData.append('file', file);
    const apiBase = await getApiBase();
    const response = await fetch(`${apiBase}/notes/${noteId}/attachments`, {
      method: 'POST',
      body: formData,
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Upload failed' }));
      throw new Error(error.error);
    }
    return response.json();
  },
  deleteAttachment: (noteId: string, attachmentId: string) =>
    request<{ message: string }>(`/notes/${noteId}/attachments/${attachmentId}`, { method: 'DELETE' }),

  // Reorder notes
  reorderNotes: (accountId: string, noteIds: string[]) =>
    request<{ message: string }>(`/accounts/${accountId}/notes/reorder`, {
      method: 'POST',
      body: JSON.stringify({ note_ids: noteIds })
    }),

  // Quick capture
  quickCapture: (data: QuickCaptureRequest) =>
    request<{ id: string; type: string; title: string }>('/quick-capture', {
      method: 'POST',
      body: JSON.stringify(data)
    }),

  // Pin/Archive
  toggleNotePin: (noteId: string) =>
    request<{ pinned: boolean }>(`/notes/${noteId}/pin`, { method: 'POST' }),
  toggleNoteArchive: (noteId: string) =>
    request<{ archived: boolean }>(`/notes/${noteId}/archive`, { method: 'POST' }),
  toggleTodoPin: (todoId: string) =>
    request<{ pinned: boolean }>(`/todos/${todoId}/pin`, { method: 'POST' }),
  getArchivedNotes: () => request<Note[]>('/notes/archived'),

  // Contacts
  getContacts: (filter?: 'internal' | 'external' | 'unlinked' | 'suggestions', accountId?: string) => {
    const params = new URLSearchParams();
    if (filter) params.append('filter', filter);
    if (accountId) params.append('account_id', accountId);
    const query = params.toString();
    return request<Contact[]>(`/contacts${query ? `?${query}` : ''}`);
  },
  getContactStats: () => request<ContactStats>('/contacts/stats'),
  getContact: (id: string) => request<Contact>(`/contacts/${id}`),
  createContact: (data: { email: string; name?: string; company?: string }) =>
    request<{ id: string; email: string }>('/contacts', { method: 'POST', body: JSON.stringify(data) }),
  updateContact: (id: string, data: { name?: string; company?: string; account_id?: string }) =>
    request<{ message: string }>(`/contacts/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteContact: (id: string) =>
    request<{ message: string }>(`/contacts/${id}`, { method: 'DELETE' }),
  confirmContactSuggestion: (id: string, confirm: boolean) =>
    request<{ message: string }>(`/contacts/${id}/confirm-suggestion`, { method: 'POST', body: JSON.stringify({ confirm }) }),
  linkContactToAccount: (contactId: string, accountId: string) =>
    request<{ message: string }>(`/contacts/${contactId}/link/${accountId}`, { method: 'POST' }),
  getContactNotes: (id: string) =>
    request<ContactNote[]>(`/contacts/${id}/notes`),
  bulkContactsOperation: (data: { contact_ids: string[]; action: string; value?: Record<string, any> }) =>
    request<{ message: string }>('/contacts/bulk', { method: 'POST', body: JSON.stringify(data) }),
};

export interface ContactNote {
  id: string;
  title: string;
  account_id?: string;
  account_name?: string;
  meeting_date?: string;
  created_at: string;
}

// Helper for attachment download URL
export const getAttachmentUrl = (filename: string) => {
  const base = getApiBaseSync().replace('/api', '');
  return `${base}/uploads/${filename}`;
};
