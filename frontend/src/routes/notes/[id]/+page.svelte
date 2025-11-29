<script lang="ts">
  import { onMount, onDestroy, tick } from 'svelte';
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
    X,
    Bold,
    Italic,
    List,
    ListOrdered,
    Code,
    Quote,
    Heading2,
    Minus,
    Image as ImageIcon
  } from 'lucide-svelte';
  import { api, type Note, type Todo, type Account } from '$lib/utils/api';
  import { addToast } from '$lib/stores';
  import { generateNotePDF } from '$lib/utils/pdf';
  import { Editor } from '@tiptap/core';
  import StarterKit from '@tiptap/starter-kit';
  import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight';
  import Image from '@tiptap/extension-image';
  import Link from '@tiptap/extension-link';
  import { WikiLink } from '$lib/editor/extensions/WikiLink';
  import WikiLinkList from '$lib/editor/components/WikiLinkList.svelte';
  import { SvelteRenderer } from 'svelte-tiptap';
  import tippy from 'tippy.js';
  import { common, createLowlight } from 'lowlight';

  let note: Note | null = null;
  let account: Account | null = null;
  let loading = true;
  let saving = false;
  let editor: Editor | null = null;
  let editorElement: HTMLElement;
  let imageInput: HTMLInputElement;

  // Form fields
  let title = '';
  let isFirstCall = true;
  let internalParticipants: string[] = [];
  let externalParticipants: string[] = [];
  let newInternalParticipant = '';
  let newExternalParticipant = '';

  // Todo section
  let todos: Todo[] = [];
  let showNewTodoModal = false;
  let newTodoTitle = '';
  let newTodoDescription = '';

  $: noteId = $page.params.id as string;

  onMount(async () => {
    if (!noteId) {
      goto('/notes');
      return;
    }
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
      isFirstCall = note.template_type === 'initial';
      internalParticipants = note.internal_participants || [];
      externalParticipants = note.external_participants || [];
      todos = note.todos || [];

      // Load account info
      if (note.account_id) {
        account = await api.getAccount(note.account_id);
      }

      // Set loading false BEFORE tick so DOM renders the editor element
      loading = false;
      
      // Wait for DOM to update (editor element to be rendered)
      await tick();
      
      // Initialize editor after DOM is ready
      initEditor(note.content || '');
    } catch (e) {
      addToast('error', 'Failed to load note');
      goto('/notes');
    }
  }

  function initEditor(content: string) {
    if (!editorElement) {
      console.error('Editor element not found');
      return;
    }
    
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
        Image.configure({
          inline: true,
          allowBase64: true, // Fallback, ideally we upload
        }),
        Link.configure({
          openOnClick: true,
          linkOnPaste: true, // This is very useful
        }),
        WikiLink.configure({
          suggestion: {
            items: async ({ query }) => {
              try {
                if (query.length === 0) return [];
                const results = await api.search(query);
                return results.slice(0, 5); // Limit to 5 suggestions
              } catch (e) {
                return [];
              }
            },
            render: () => {
              let component: any;
              let popup: any;

              return {
                onStart: (props: any) => {
                  component = new WikiLinkList({
                    target: document.body,
                    props: {
                      items: props.items,
                      command: (item: any) => {
                        props.command(item);
                      }
                    }
                  });

                  popup = tippy('body', {
                    getReferenceClientRect: props.clientRect,
                    appendTo: () => document.body,
                    content: component.$$.root, // Use the root element of the Svelte component
                    showOnCreate: true,
                    interactive: true,
                    trigger: 'manual',
                    placement: 'bottom-start',
                  });
                },
                onUpdate(props: any) {
                  component.$set({ items: props.items });
                  popup[0].setProps({
                    getReferenceClientRect: props.clientRect,
                  });
                },
                onKeyDown(props: any) {
                  if (props.event.key === 'Escape') {
                    popup[0].hide();
                    return true;
                  }
                  return component.onKeyDown(props);
                },
                onExit() {
                  popup[0].destroy();
                  component.$destroy();
                },
              };
            },
          },
        }),
      ],
      content: content || '',
      editable: true,
      autofocus: 'end',
      onTransaction: () => {
        editor = editor; // Trigger reactivity
      },
      editorProps: {
        handlePaste: (view, event, slice) => {
          const items = Array.from(event.clipboardData?.items || []);
          const images = items.filter(item => item.type.indexOf('image') === 0);

          if (images.length === 0) return false;

          event.preventDefault();
          
          images.forEach(item => {
            const file = item.getAsFile();
            if (file) {
              handleImageUpload(file);
            }
          });
          return true;
        },
        handleDrop: (view, event, slice, moved) => {
          if (!moved && event.dataTransfer && event.dataTransfer.files && event.dataTransfer.files.length > 0) {
            const files = Array.from(event.dataTransfer.files);
            const images = files.filter(file => file.type.indexOf('image') === 0);
            
            if (images.length === 0) return false;

            event.preventDefault();
            images.forEach(file => {
              handleImageUpload(file);
            });
            return true;
          }
          return false;
        }
      }
    });
  }

  async function handleImageUpload(file: File) {
    if (!note || !editor) return;
    
    // Insert placeholder
    // We can't easily insert a loading state in standard TipTap Image without a custom node view
    // So we'll just upload then insert.
    
    const toastId = addToast('loading', 'Uploading image...');
    
    try {
      const attachment = await api.uploadAttachment(note.id, file);
      // Assuming backend serves uploads at /uploads/filename
      // We need a helper to get full URL
      const url = `/uploads/${attachment.filename}`;
      
      editor.chain().focus().setImage({ src: url, alt: file.name }).run();
      addToast('success', 'Image uploaded');
    } catch (e) {
      console.error(e);
      addToast('error', 'Failed to upload image');
    }
  }

  function triggerImageUpload() {
    imageInput.click();
  }

  async function handleImageSelect(e: Event) {
    const files = (e.target as HTMLInputElement).files;
    if (files && files.length > 0) {
      await handleImageUpload(files[0]);
    }
    if (imageInput) imageInput.value = '';
  }

  async function saveNote() {
    if (!editor || !note) return;
    
    try {
      saving = true;
      await api.updateNote(noteId, {
        title,
        template_type: isFirstCall ? 'initial' : 'followup',
        internal_participants: internalParticipants,
        external_participants: externalParticipants,
        content: editor.getHTML(),
      });
      addToast('success', 'Note saved');
      goto('/notes');
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
        description: newTodoDescription.trim() || undefined,
        note_id: noteId
      });
      todos = [...todos, todo];
      newTodoTitle = '';
      newTodoDescription = '';
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

  async function deleteNote() {
    try {
      await api.deleteNote(noteId);
      addToast('success', 'Note moved to trash');
      goto('/notes');
    } catch (e) {
      addToast('error', 'Failed to delete note');
    }
  }

  async function saveAccountDetails() {
    if (!account) return;
    try {
      await api.updateAccount(account.id, {
        name: account.name,
        account_owner: account.account_owner || '',
        budget: account.budget || undefined,
        est_engineers: account.est_engineers || undefined
      });
      addToast('success', 'Account updated');
    } catch (e) {
      addToast('error', 'Failed to update account');
    }
  }

  async function exportPDF(type: 'full' | 'minimal') {
    if (!note || !editor) return;
    
    try {
      // Save current content before export
      const currentContent = editor.getHTML();
      
      await generateNotePDF({
        note: { ...note, content: currentContent, title },
        account: account || undefined,
        todos: type === 'full' ? todos : undefined
      }, type);
      
      addToast('success', `Exported ${type === 'full' ? 'full' : 'minimal'} PDF`);
    } catch (e) {
      console.error('PDF export error:', e);
      addToast('error', 'Failed to export PDF');
    }
  }

  function exportMarkdown() {
    if (!note) return;
    const url = api.exportNoteMarkdown(note.id);
    window.open(url, '_blank');
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
  <title>{title || 'Note'} - Noted</title>
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
              class="w-full px-4 py-2 text-left text-sm hover:bg-[var(--color-bg)]"
              on:click={() => exportPDF('minimal')}
            >
              Minimal (Notes + Account)
            </button>
            <button 
              class="w-full px-4 py-2 text-left text-sm hover:bg-[var(--color-bg)] rounded-b-lg"
              on:click={exportMarkdown}
            >
              Markdown
            </button>
          </div>
        </div>
        <button class="btn-primary" on:click={saveNote} disabled={saving}>
          <Save class="w-4 h-4" />
          {saving ? 'Saving...' : 'Save'}
        </button>
        <button 
          class="btn-icon btn-icon-danger"
          title="Delete note"
          on:click={deleteNote}
        >
          <Trash2 class="w-5 h-5" />
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
            <label class="flex items-center gap-2 cursor-pointer">
              <input 
                type="checkbox" 
                bind:checked={isFirstCall}
                class="w-4 h-4 accent-[var(--color-accent)]"
              />
              <span class="text-sm">1st Call</span>
            </label>
          </div>
          
          <!-- Editor Toolbar -->
          {#if editor}
            <div class="flex items-center gap-1 mb-3 p-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-sm">
              <button
                type="button"
                class="p-2 rounded-sm hover:bg-[var(--color-card-hover)] transition-colors {editor.isActive('bold') ? 'bg-[var(--color-accent)]/10 text-[var(--color-accent)]' : ''}"
                on:click={() => editor?.chain().focus().toggleBold().run()}
                title="Bold (Ctrl+B)"
              >
                <Bold class="w-4 h-4" strokeWidth={2} />
              </button>
              <button
                type="button"
                class="p-2 rounded-sm hover:bg-[var(--color-card-hover)] transition-colors {editor.isActive('italic') ? 'bg-[var(--color-accent)]/10 text-[var(--color-accent)]' : ''}"
                on:click={() => editor?.chain().focus().toggleItalic().run()}
                title="Italic (Ctrl+I)"
              >
                <Italic class="w-4 h-4" strokeWidth={2} />
              </button>
              
              <div class="w-px h-5 bg-[var(--color-border)] mx-1"></div>
              
              <button
                type="button"
                class="p-2 rounded-sm hover:bg-[var(--color-card-hover)] transition-colors {editor.isActive('heading', { level: 2 }) ? 'bg-[var(--color-accent)]/10 text-[var(--color-accent)]' : ''}"
                on:click={() => editor?.chain().focus().toggleHeading({ level: 2 }).run()}
                title="Heading"
              >
                <Heading2 class="w-4 h-4" strokeWidth={2} />
              </button>
              
              <div class="w-px h-5 bg-[var(--color-border)] mx-1"></div>
              
              <button
                type="button"
                class="p-2 rounded-sm hover:bg-[var(--color-card-hover)] transition-colors {editor.isActive('bulletList') ? 'bg-[var(--color-accent)]/10 text-[var(--color-accent)]' : ''}"
                on:click={() => editor?.chain().focus().toggleBulletList().run()}
                title="Bullet List"
              >
                <List class="w-4 h-4" strokeWidth={2} />
              </button>
              <button
                type="button"
                class="p-2 rounded-sm hover:bg-[var(--color-card-hover)] transition-colors {editor.isActive('orderedList') ? 'bg-[var(--color-accent)]/10 text-[var(--color-accent)]' : ''}"
                on:click={() => editor?.chain().focus().toggleOrderedList().run()}
                title="Numbered List"
              >
                <ListOrdered class="w-4 h-4" strokeWidth={2} />
              </button>
              
              <div class="w-px h-5 bg-[var(--color-border)] mx-1"></div>
              
              <button
                type="button"
                class="p-2 rounded-sm hover:bg-[var(--color-card-hover)] transition-colors {editor.isActive('blockquote') ? 'bg-[var(--color-accent)]/10 text-[var(--color-accent)]' : ''}"
                on:click={() => editor?.chain().focus().toggleBlockquote().run()}
                title="Quote"
              >
                <Quote class="w-4 h-4" strokeWidth={2} />
              </button>
              <button
                type="button"
                class="p-2 rounded-sm hover:bg-[var(--color-card-hover)] transition-colors {editor.isActive('code') ? 'bg-[var(--color-accent)]/10 text-[var(--color-accent)]' : ''}"
                on:click={() => editor?.chain().focus().toggleCode().run()}
                title="Inline Code"
              >
                <Code class="w-4 h-4" strokeWidth={2} />
              </button>
              
              <div class="w-px h-5 bg-[var(--color-border)] mx-1"></div>
              
              <button
                type="button"
                class="p-2 rounded-sm hover:bg-[var(--color-card-hover)] transition-colors"
                on:click={() => editor?.chain().focus().setHorizontalRule().run()}
                title="Horizontal Rule"
              >
                <Minus class="w-4 h-4" strokeWidth={2} />
              </button>
              
              <div class="w-px h-5 bg-[var(--color-border)] mx-1"></div>
              
              <button
                type="button"
                class="p-2 rounded-sm hover:bg-[var(--color-card-hover)] transition-colors"
                on:click={triggerImageUpload}
                title="Insert Image"
              >
                <ImageIcon class="w-4 h-4" strokeWidth={2} />
              </button>
              <input
                type="file"
                accept="image/*"
                class="hidden"
                bind:this={imageInput}
                on:change={handleImageSelect}
              />
            </div>
          {/if}
          
          <div bind:this={editorElement} class="editor-content"></div>
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
              class="btn-secondary btn-sm"
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
                      class="btn-icon-sm btn-icon-danger opacity-0 group-hover:opacity-100"
                      on:click={() => deleteTodo(todo.id)}
                    >
                      <Trash2 class="w-4 h-4" />
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
                <label class="label" for="note-account-owner">Account Owner</label>
                <input 
                  id="note-account-owner"
                  type="text"
                  class="input text-sm"
                  bind:value={account.account_owner}
                  placeholder="Account owner name"
                  on:change={saveAccountDetails}
                />
              </div>
              <div>
                <label class="label" for="note-account-budget">Budget</label>
                <div class="relative">
                  <DollarSign class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--color-muted)]" />
                  <input 
                    id="note-account-budget"
                    type="number"
                    class="input text-sm pl-8"
                    bind:value={account.budget}
                    placeholder="Budget amount"
                    on:change={saveAccountDetails}
                  />
                </div>
              </div>
              <div>
                <label class="label" for="note-account-engineers">Est. Engineers (POC Size)</label>
                <input 
                  id="note-account-engineers"
                  type="number"
                  class="input text-sm"
                  bind:value={account.est_engineers}
                  placeholder="Team size"
                  on:change={saveAccountDetails}
                />
              </div>
            </div>
          </div>
        {/if}

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
                  class="btn-icon-sm hover:bg-green-500/20"
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
              class="btn-secondary btn-sm"
              on:click={() => addParticipant('external')}
            >
              <Plus class="w-4 h-4" />
            </button>
          </div>
        </div>

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
                  class="btn-icon-sm hover:bg-primary-500/20"
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
              class="btn-secondary btn-sm"
              on:click={() => addParticipant('internal')}
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
      aria-label="Close modal"
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-md animate-slide-up">
      <h2 class="text-lg font-semibold mb-4">New Follow-up Item</h2>
      <form on:submit|preventDefault={createTodo}>
        <div class="mb-4">
          <label class="label" for="note-new-todo-title">Todo Title</label>
          <!-- svelte-ignore a11y-autofocus -->
          <input 
            id="note-new-todo-title"
            type="text"
            class="input"
            placeholder="e.g., Send technical documentation"
            bind:value={newTodoTitle}
            autofocus
          />
        </div>
        <div class="mb-4">
          <label class="label" for="note-new-todo-desc">Description (optional)</label>
          <textarea
            id="note-new-todo-desc"
            class="input"
            rows="3"
            placeholder="Add more details about this follow-up..."
            bind:value={newTodoDescription}
          ></textarea>
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
