# Skill: Frontend Development

## Overview
This skill covers working with the SvelteKit frontend for SE Notes.

## Tech Stack
- SvelteKit 2.x
- Svelte 5 (with runes)
- Tailwind CSS 3.x
- TipTap (rich text editor)
- svelte-dnd-action (drag and drop)
- lucide-svelte (icons)

## Project Structure
```
frontend/
├── src/
│   ├── routes/                  # Pages (file-based routing)
│   │   ├── +layout.svelte       # App shell, sidebar, theme
│   │   ├── +page.svelte         # Dashboard (/)
│   │   ├── notes/
│   │   │   ├── +page.svelte     # Notes list (/notes)
│   │   │   └── [id]/+page.svelte # Note editor (/notes/:id)
│   │   ├── todos/+page.svelte   # Kanban board (/todos)
│   │   ├── calendar/+page.svelte # Calendar (/calendar)
│   │   └── settings/+page.svelte # Settings (/settings)
│   ├── lib/
│   │   ├── stores/index.ts      # Svelte stores
│   │   └── utils/api.ts         # API client
│   ├── app.css                  # Global styles, Tailwind
│   └── app.html                 # HTML template
├── tailwind.config.js
├── svelte.config.js
└── package.json
```

## Running the Frontend

```bash
cd frontend

# Install dependencies
npm install

# Development (hot reload)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## Adding a New Page

### 1. Create the route file
```svelte
<!-- src/routes/my-page/+page.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { SomeIcon } from 'lucide-svelte';
  import { api } from '$lib/utils/api';
  
  let data = [];
  let loading = true;
  
  onMount(async () => {
    try {
      data = await api.getSomeData();
    } catch (e) {
      console.error(e);
    } finally {
      loading = false;
    }
  });
</script>

<svelte:head>
  <title>My Page - SE Notes</title>
</svelte:head>

<div class="max-w-7xl mx-auto">
  <div class="mb-8">
    <h1 class="page-title">My Page</h1>
    <p class="page-subtitle">Description here</p>
  </div>
  
  {#if loading}
    <div class="card animate-pulse">
      <div class="h-4 bg-[var(--color-border)] rounded w-48"></div>
    </div>
  {:else}
    <!-- Content -->
  {/if}
</div>
```

### 2. Add to navigation
```svelte
<!-- src/routes/+layout.svelte -->
<script>
  const navItems = [
    // ... existing items
    { href: '/my-page', label: 'My Page', icon: SomeIcon },
  ];
</script>
```

## API Client Usage

```typescript
// src/lib/utils/api.ts

// Add new types
export interface MyThing {
  id: string;
  name: string;
}

// Add new API functions
export const api = {
  // ... existing functions
  
  getMyThings: () => request<MyThing[]>('/my-things'),
  
  createMyThing: (data: { name: string }) =>
    request<MyThing>('/my-things', { 
      method: 'POST', 
      body: JSON.stringify(data) 
    }),
};
```

## Styling Patterns

### CSS Variables (Theme)
```css
/* Light mode */
:root {
  --color-bg: #ffffff;
  --color-card: #f9fafb;
  --color-border: #e5e7eb;
  --color-text: #111827;
  --color-muted: #6b7280;
}

/* Dark mode */
.dark {
  --color-bg: #0a0a0a;
  --color-card: #141414;
  --color-border: #262626;
  --color-text: #fafafa;
  --color-muted: #a1a1aa;
}
```

### Component Classes
```html
<!-- Buttons -->
<button class="btn-primary">Primary</button>
<button class="btn-secondary">Secondary</button>
<button class="btn-ghost">Ghost</button>

<!-- Cards -->
<div class="card">Content</div>

<!-- Forms -->
<label class="label">Label</label>
<input class="input" />

<!-- Typography -->
<h1 class="page-title">Title</h1>
<p class="page-subtitle">Subtitle</p>
```

## Svelte Stores

```typescript
// src/lib/stores/index.ts
import { writable, derived } from 'svelte/store';

// Create a store
export const myStore = writable<MyType[]>([]);

// Derived store
export const filteredItems = derived(myStore, ($items) =>
  $items.filter(item => item.active)
);

// Toast notifications
import { addToast } from '$lib/stores';
addToast('success', 'Item created!');
addToast('error', 'Something went wrong');
```

## TipTap Rich Text Editor

```svelte
<script>
  import { onMount, onDestroy } from 'svelte';
  import { Editor } from '@tiptap/core';
  import StarterKit from '@tiptap/starter-kit';
  
  let editor: Editor;
  let editorElement: HTMLElement;
  
  onMount(() => {
    editor = new Editor({
      element: editorElement,
      extensions: [StarterKit],
      content: '<p>Initial content</p>',
    });
  });
  
  onDestroy(() => {
    editor?.destroy();
  });
  
  function getContent() {
    return editor.getHTML();
  }
</script>

<div bind:this={editorElement}></div>
```

## Drag and Drop (Kanban)

```svelte
<script>
  import { dndzone } from 'svelte-dnd-action';
  import { flip } from 'svelte/animate';
  
  let items = [...];
  const flipDurationMs = 200;
  
  function handleDndConsider(e) {
    items = e.detail.items;
  }
  
  function handleDndFinalize(e) {
    items = e.detail.items;
    // Save to API
  }
</script>

<div
  use:dndzone={{ items, flipDurationMs }}
  on:consider={handleDndConsider}
  on:finalize={handleDndFinalize}
>
  {#each items as item (item.id)}
    <div animate:flip={{ duration: flipDurationMs }}>
      {item.title}
    </div>
  {/each}
</div>
```

## Modal Pattern

```svelte
{#if showModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button 
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => showModal = false}
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-md animate-slide-up">
      <h2 class="text-lg font-semibold mb-4">Modal Title</h2>
      <!-- Content -->
      <div class="flex justify-end gap-3">
        <button class="btn-secondary" on:click={() => showModal = false}>
          Cancel
        </button>
        <button class="btn-primary" on:click={handleSubmit}>
          Confirm
        </button>
      </div>
    </div>
  </div>
{/if}
```

## Icons (Lucide)

```svelte
<script>
  import { 
    Plus, 
    Trash2, 
    Edit3, 
    Search,
    FileText,
    CheckSquare,
    Calendar,
    Settings,
    Moon,
    Sun
  } from 'lucide-svelte';
</script>

<Plus class="w-4 h-4" />
<Trash2 class="w-4 h-4 text-red-500" />
```

## Common Issues

### Hydration mismatch
Ensure server and client render the same content. Use `onMount` for browser-only code.

### API not reachable
Check that backend is running on port 8080. API URL is hardcoded in `api.ts`.

### Build errors with TipTap
TipTap needs specific Svelte configuration. Check `svelte.config.js` for preprocessor settings.
