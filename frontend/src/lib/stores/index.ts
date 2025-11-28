import { writable, derived } from 'svelte/store';
import type { Account, Note, Todo, Analytics } from '$lib/utils/api';

// Accounts store
export const accounts = writable<Account[]>([]);
export const accountsLoading = writable(false);
export const accountsError = writable<string | null>(null);

// Notes store
export const notes = writable<Note[]>([]);
export const notesLoading = writable(false);
export const notesError = writable<string | null>(null);
export const selectedNote = writable<Note | null>(null);

// Todos store
export const todos = writable<Todo[]>([]);
export const todosLoading = writable(false);
export const todosError = writable<string | null>(null);

// Derived stores for kanban columns
export const todosByStatus = derived(todos, ($todos) => ({
  not_started: $todos.filter(t => t.status === 'not_started'),
  in_progress: $todos.filter(t => t.status === 'in_progress'),
  completed: $todos.filter(t => t.status === 'completed'),
}));

// Analytics store
export const analytics = writable<Analytics | null>(null);
export const analyticsLoading = writable(false);

// UI state
export const searchQuery = writable('');
export const searchResults = writable<any[]>([]);
export const searchLoading = writable(false);

// Notes grouped by account
export const notesByAccount = derived([notes, accounts], ([$notes, $accounts]) => {
  const grouped: Record<string, { account: Account; notes: Note[] }> = {};
  
  for (const account of $accounts) {
    grouped[account.id] = {
      account,
      notes: $notes.filter(n => n.account_id === account.id)
    };
  }
  
  return grouped;
});

// Toast notifications
export interface Toast {
  id: string;
  type: 'success' | 'error' | 'info';
  message: string;
}

export const toasts = writable<Toast[]>([]);

export function addToast(type: Toast['type'], message: string) {
  const id = crypto.randomUUID();
  toasts.update(t => [...t, { id, type, message }]);
  setTimeout(() => {
    toasts.update(t => t.filter(toast => toast.id !== id));
  }, 3000);
}
