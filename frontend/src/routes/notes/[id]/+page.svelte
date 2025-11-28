<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { 
    ArrowLeft, 
    Save, 
    Download, 
    Plus,
    Trash2,
    Users,
    Building2,
    Calendar,
    DollarSign,
    Users2,
    CheckSquare,
    X
  } from 'lucide-svelte';
  import { api, type Note, type Todo, type Account } from '$lib/utils/api';
  import { addToast } from '$lib/stores';
  import { Editor } from '@tiptap/core';
  import StarterKit from '@tiptap/starter-kit';
  import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight';
  import { common, createLowlight } from 'lowlight';

  let note: Note | null = null;
  let account: Account | null = null;
  let loading = true;
  let saving = false;
  let editor: Editor | null = null;
  let editorElement: HTMLElement;

  // Form fields
  let title = '';
  let templateType = 'initial';
  let internalParticipants: string[] = [];
  let externalParticipants: string[] = [];
  let newInternalParticipant = '';
  let newExternalParticipant = '';

  // Todo section
  let todos: Todo[] = [];
  let showNewTodoModal = false;
  let newTodoTitle = '';

  $: noteId = $page.params.id;

  onMount(async () => {
    await loadNote();
  });

  onDestroy(() => {
    if (editor) {
      editor.destroy();
    }
  });

  async function loadNote() {
    try {
      loading = true;
      note = await api.getNote(noteId);
      title = note.title;
      templateType = note.template_type;
      internalParticipants = note.internal_participants || [];
      externalParticipants = note.external_participants || [];
      todos = note.todos || [];

      // Load account info
      if (note.account_id) {
        account = await api.getAccount(note.account_id);
      }

      // Initialize editor after note is loaded
      initEditor(note.content || '');
    } catch (e) {
      addToast('error', 'Failed to load note');
      goto('/notes');
    } finally {
      loading = false;
    }
  }

  function initEditor(content: string) {
    const lowlight = createLowlight(common);
    
    editor = new Editor({
      element: editorElement,
      extensions: [
        StarterKit.configure({
          codeBlock: false,
        }),
        CodeBlockLowlight.configure({
          lowlight,
        }),
      ],
      content: content || '<p>Start writing your notes here...</p>',
      editorProps: {
        attributes: {
          class: 'prose prose-sm dark:prose-invert max-w-none focus:outline-none min-h-[300px]',
        },
      },
      onTransaction: () => {
        editor = editor; // Trigger reactivity
      },
    });
  }

  async function saveNote() {
    if (!editor || !note) return;
    
    try {
      saving = true;
      await api.updateNote(noteId, {
        title,
        template_type: templateType,
        internal_participants: internalParticipants,
        external_participants: externalParticipants,
        content: editor.getHTML(),
      });
      addToast('success', 'Note saved');
    } catch (e) {
      addToast('error', 'Failed to save note');
    } finally {
      saving = false;
    }
  }

  function addParticipant(type: 'internal' | 'external') {
    if (type === 'internal' && newInternalParticipant.trim()) {
      internalParticipants = [...internalParticipants, newInternalParticipant.trim()];
      newInternalParticipant = '';
    } else if (type === 'external' && newExternalParticipant.trim()) {
      externalParticipants = [...externalParticipants, newExternalParticipant.trim()];
      newExternalParticipant = '';
    }
  }

  function removeParticipant(type: 'internal' | 'external', index: number) {
    if (type === 'internal') {
      internalParticipants = internalParticipants.filter((_, i) => i !== index);
    } else {
      externalParticipants = externalParticipants.filter((_, i) => i !== index);
    }
  }

  async function createTodo() {
    if (!newTodoTitle.trim()) return;
    try {
      const todo = await api.createTodo({
        title: newTodoTitle.trim(),
        note_id: noteId
      });
      todos = [...todos, todo];
      newTodoTitle = '';
      showNewTodoModal = false;
      addToast('success', 'Todo created');
    } catch (e) {
      addToast('error', 'Failed to create todo');
    }
  }

  async function deleteTodo(todoId: string) {
    try {
      await api.deleteTodo(todoId);
      todos = todos.filter(t => t.id !== todoId);
      addToast('success', 'Todo deleted');
    } catch (e) {
      addToast('error', 'Failed to delete todo');
    }
  }

  async function exportPDF(type: 'full' | 'minimal') {
    try {
      const data = await api.exportNote(noteId, type);
      // For now, just download as JSON - PDF generation would need additional library
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${title.replace(/\s+/g, '-')}-${type}.json`;
      a.click();
      URL.revokeObjectURL(url);
      addToast('info', 'Exported as JSON (PDF coming soon)');
    } catch (e) {
      addToast('error', 'Failed to export');
    }
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-US', {
      weekday: 'long',
      month: 'long',
      day: 'numeric',
      year: 'numeric'
    });
  }

  function getStatusColor(status: string): string {
    switch (status) {
      case 'completed': return 'bg-green-500';
      case 'in_progress': return 'bg-blue-500';
      default: return 'bg-gray-400';
    }
  }
</script>

<svelte:head>
  <title>{title || 'Note'} - SE Notes</title>
</svelte:head>

{#if loading}
  <div class="max-w-5xl mx-auto">
    <div class="animate-pulse">
      <div class="h-8 bg-[var(--color-border)] rounded w-48 mb-4"></div>
      <div class="h-4 bg-[var(--color-border)] rounded w-32 mb-8"></div>
      <div class="card">
        <div class="space-y-4">
          <div class="h-4 bg-[var(--color-border)] rounded w-full"></div>
          <div class="h-4 bg-[var(--color-border)] rounded w-3/4"></div>
          <div class="h-4 bg-[var(--color-border)] rounded w-1/2"></div>
        </div>
      </div>
    </div>
  </div>
{:else if note}
  <div class="max-w-5xl mx-auto">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-4">
        <a href="/notes" class="p-2 hover:bg-[var(--color-card)] rounded-lg transition-colors">
          <ArrowLeft class="w-5 h-5" />
        </a>
        <div>
          <input 
            type="text"
            class="text-2xl font-semibold bg-transparent border-none outline-none focus:ring-0 p-0"
            bind:value={title}
            placeholder="Note title"
          />
          {#if account}
            <p class="text-[var(--color-muted)] flex items-center gap-2 mt-1">
              <Building2 class="w-4 h-4" />
              {account.name}
              {#if note.meeting_date}
                <span class="mx-2">•</span>
                <Calendar class="w-4 h-4" />
                {formatDate(note.meeting_date)}
              {/if}
            </p>
          {/if}
        </div>
      </div>
      <div class="flex items-center gap-3">
        <div class="relative group">
          <button class="btn-secondary">
            <Download class="w-4 h-4" />
            Export
          </button>
          <div class="absolute right-0 top-full mt-2 bg-[var(--color-card)] border border-[var(--color-border)] rounded-lg shadow-xl opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all">
            <button 
              class="w-full px-4 py-2 text-left text-sm hover:bg-[var(--color-bg)] rounded-t-lg"
              on:click={() => exportPDF('full')}
            >
              Full Export
            </button>
            <button 
              class="w-full px-4 py-2 text-left text-sm hover:bg-[var(--color-bg)] rounded-b-lg"
              on:click={() => exportPDF('minimal')}
            >
              Minimal (Notes + Account)
            </button>
          </div>
        </div>
        <button class="btn-primary" on:click={saveNote} disabled={saving}>
          <Save class="w-4 h-4" />
          {saving ? 'Saving...' : 'Save'}
        </button>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Main Content -->
      <div class="lg:col-span-2 space-y-6">
        <!-- Editor -->
        <div class="card">
          <div class="flex items-center justify-between mb-4">
            <h3 class="font-medium">Notes</h3>
            <select 
              class="text-sm bg-[var(--color-bg)] border border-[var(--color-border)] rounded px-2 py-1"
              bind:value={templateType}
            >
              <option value="initial">Initial Call</option>
              <option value="followup">Follow-up</option>
            </select>
          </div>
          <div bind:this={editorElement} class="min-h-[300px]"></div>
        </div>

        <!-- Todos -->
        <div class="card">
          <div class="flex items-center justify-between mb-4">
            <h3 class="font-medium flex items-center gap-2">
              <CheckSquare class="w-5 h-5" />
              Follow-up Items
              {#if todos.length > 0}
                <span class="text-sm text-[var(--color-muted)]">({todos.length})</span>
              {/if}
            </h3>
            <button 
              class="btn-secondary text-sm py-1.5"
              on:click={() => showNewTodoModal = true}
            >
              <Plus class="w-4 h-4" />
              Add Todo
            </button>
          </div>
          
          {#if todos.length === 0}
            <p class="text-[var(--color-muted)] text-sm">No todos yet. Add follow-up items from this call.</p>
          {:else}
            <div class="space-y-2">
              {#each todos as todo}
                <div class="flex items-center justify-between p-3 bg-[var(--color-bg)] rounded-lg group">
                  <div class="flex items-center gap-3">
                    <div class="w-2 h-2 rounded-full {getStatusColor(todo.status)}"></div>
                    <span class:line-through={todo.status === 'completed'}>{todo.title}</span>
                  </div>
                  <div class="flex items-center gap-2">
                    <a 
                      href="/todos"
                      class="text-xs text-primary-500 hover:underline"
                    >
                      View in Kanban
                    </a>
                    <button 
                      class="p-1 hover:bg-[var(--color-border)] rounded opacity-0 group-hover:opacity-100 transition-opacity"
                      on:click={() => deleteTodo(todo.id)}
                    >
                      <Trash2 class="w-4 h-4 text-red-500" />
                    </button>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>

      <!-- Sidebar -->
      <div class="space-y-6">
        <!-- Account Info -->
        {#if account}
          <div class="card">
            <h3 class="font-medium mb-4 flex items-center gap-2">
              <Building2 class="w-5 h-5" />
              Account Details
            </h3>
            <div class="space-y-3">
              <div>
                <label class="label">Account Owner</label>
                <p class="text-sm">{account.account_owner || 'Not set'}</p>
              </div>
              <div>
                <label class="label">Budget</label>
                <p class="text-sm flex items-center gap-1">
                  <DollarSign class="w-4 h-4" />
                  {account.budget ? `$${account.budget.toLocaleString()}` : 'Not set'}
                </p>
              </div>
              <div>
                <label class="label">Est. Engineers (POC Size)</label>
                <p class="text-sm flex items-center gap-1">
                  <Users2 class="w-4 h-4" />
                  {account.est_engineers ?? 'Not set'}
                </p>
              </div>
            </div>
          </div>
        {/if}

        <!-- Internal Participants -->
        <div class="card">
          <h3 class="font-medium mb-4 flex items-center gap-2">
            <Users class="w-5 h-5 text-primary-500" />
            Internal Participants
          </h3>
          <div class="space-y-2 mb-3">
            {#each internalParticipants as participant, i}
              <div class="flex items-center justify-between px-3 py-2 bg-primary-500/10 rounded-lg text-sm">
                <span>{participant}</span>
                <button 
                  class="p-1 hover:bg-primary-500/20 rounded"
                  on:click={() => removeParticipant('internal', i)}
                >
                  <X class="w-3 h-3" />
                </button>
              </div>
            {/each}
          </div>
          <div class="flex gap-2">
            <input 
              type="text"
              class="input text-sm"
              placeholder="Add participant"
              bind:value={newInternalParticipant}
              on:keypress={(e) => e.key === 'Enter' && addParticipant('internal')}
            />
            <button 
              class="btn-secondary py-1.5"
              on:click={() => addParticipant('internal')}
            >
              <Plus class="w-4 h-4" />
            </button>
          </div>
        </div>

        <!-- External Participants -->
        <div class="card">
          <h3 class="font-medium mb-4 flex items-center gap-2">
            <Users class="w-5 h-5 text-green-500" />
            Customer Participants
          </h3>
          <div class="space-y-2 mb-3">
            {#each externalParticipants as participant, i}
              <div class="flex items-center justify-between px-3 py-2 bg-green-500/10 rounded-lg text-sm">
                <span>{participant}</span>
                <button 
                  class="p-1 hover:bg-green-500/20 rounded"
                  on:click={() => removeParticipant('external', i)}
                >
                  <X class="w-3 h-3" />
                </button>
              </div>
            {/each}
          </div>
          <div class="flex gap-2">
            <input 
              type="text"
              class="input text-sm"
              placeholder="Add participant"
              bind:value={newExternalParticipant}
              on:keypress={(e) => e.key === 'Enter' && addParticipant('external')}
            />
            <button 
              class="btn-secondary py-1.5"
              on:click={() => addParticipant('external')}
            >
              <Plus class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}

<!-- New Todo Modal -->
{#if showNewTodoModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button 
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => showNewTodoModal = false}
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-md animate-slide-up">
      <h2 class="text-lg font-semibold mb-4">New Follow-up Item</h2>
      <form on:submit|preventDefault={createTodo}>
        <div class="mb-4">
          <label class="label">Todo Title</label>
          <input 
            type="text"
            class="input"
            placeholder="e.g., Send technical documentation"
            bind:value={newTodoTitle}
            autofocus
          />
        </div>
        <div class="flex justify-end gap-3">
          <button 
            type="button"
            class="btn-secondary"
            on:click={() => showNewTodoModal = false}
          >
            Cancel
          </button>
          <button type="submit" class="btn-primary" disabled={!newTodoTitle.trim()}>
            Create Todo
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
