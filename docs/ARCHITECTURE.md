# Architecture Overview

## System Design

```
┌─────────────────────────────────────────────────────────────┐
│                         Frontend                              │
│                    (SvelteKit + Tailwind)                    │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌────────┐│
│  │Dashboard│ │  Notes  │ │  Todos  │ │Calendar │ │Settings││
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬───┘│
│       │           │           │           │           │      │
│       └───────────┴───────────┴───────────┴───────────┘      │
│                           │                                   │
│                    Svelte Stores                              │
│                    (State Mgmt)                               │
│                           │                                   │
│                    API Client                                 │
│                    (fetch)                                    │
└───────────────────────────┼──────────────────────────────────┘
                            │ HTTP/JSON
                            │
┌───────────────────────────┼──────────────────────────────────┐
│                           │                                   │
│                    REST API                                   │
│                    (Gin Router)                               │
│                           │                                   │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌────────┐│
│  │Accounts │ │  Notes  │ │  Todos  │ │ Search  │ │Analytics││
│  │Handler  │ │ Handler │ │ Handler │ │ Handler │ │Handler  ││
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬───┘│
│       │           │           │           │           │      │
│       └───────────┴───────────┴───────────┴───────────┘      │
│                           │                                   │
│                    SQLite + FTS5                              │
│                    (Database)                                 │
│                                                               │
│                         Backend                               │
│                       (Go + Gin)                              │
└──────────────────────────────────────────────────────────────┘
```

## Data Model

```
┌─────────────┐       ┌─────────────┐
│   Account   │       │    Note     │
├─────────────┤       ├─────────────┤
│ id          │◄──────│ account_id  │
│ name        │       │ id          │
│ account_owner       │ title       │
│ budget      │       │ template_type
│ est_engineers       │ internal_participants
│ created_at  │       │ external_participants
│ updated_at  │       │ content     │
└─────────────┘       │ meeting_id  │
                      │ meeting_date│
                      │ created_at  │
                      │ updated_at  │
                      └──────┬──────┘
                             │
                             │ many-to-many
                             │
                      ┌──────┴──────┐
                      │ note_todos  │
                      ├─────────────┤
                      │ note_id     │
                      │ todo_id     │
                      │ created_at  │
                      └──────┬──────┘
                             │
                      ┌──────┴──────┐
                      │    Todo     │
                      ├─────────────┤
                      │ id          │
                      │ title       │
                      │ description │
                      │ status      │
                      │ priority    │
                      │ due_date    │
                      │ created_at  │
                      │ updated_at  │
                      └─────────────┘
```

## Key Design Decisions

### 1. SQLite with FTS5
- **Why**: Zero setup, single file, excellent for local/personal use
- **FTS5**: Full-text search on notes without external search service
- **Trade-off**: Single-user focused, would need migration for multi-user

### 2. SvelteKit
- **Why**: Fast, minimal JS bundle, excellent DX
- **Animations**: svelte-motion for smooth transitions
- **Trade-off**: Smaller ecosystem than React

### 3. TipTap Editor
- **Why**: Modern ProseMirror wrapper, extensible, works well with Svelte
- **Features**: Rich text, code blocks, markdown-like shortcuts
- **Storage**: HTML string in database (simple, exportable)

### 4. Kanban with svelte-dnd-action
- **Why**: Native Svelte drag-and-drop, good accessibility
- **UX**: Immediate optimistic updates, then sync to server

### 5. Many-to-Many Todos
- **Why**: Todos can span multiple calls (carry-over items)
- **Implementation**: Junction table `note_todos`

## Request Flow

### Creating a Note
```
1. User fills form in frontend
2. POST /api/notes with JSON body
3. Handler validates request
4. Generate UUID, set timestamps
5. Insert into SQLite
6. FTS5 trigger auto-updates search index
7. Return created note
8. Frontend updates store
9. Navigate to note editor
```

### Searching
```
1. User types in search bar (debounced)
2. GET /api/search?q=term
3. Handler queries FTS5: notes_fts MATCH term
4. Also queries accounts/todos with LIKE
5. Returns combined results with snippets
6. Frontend displays results with highlighting
```

### Drag-and-Drop Todo
```
1. User drags card to new column
2. svelte-dnd-action fires finalize event
3. Local state updated immediately (optimistic)
4. PUT /api/todos/:id with new status
5. Server updates todo
6. On error, revert local state
```

## Future Considerations

### Google Calendar Integration
```
┌──────────┐     OAuth 2.0    ┌─────────────────┐
│ Frontend │◄────────────────►│ Google OAuth    │
└────┬─────┘                  └────────┬────────┘
     │                                  │
     │ Store tokens                     │ Access token
     ▼                                  ▼
┌──────────┐                  ┌─────────────────┐
│ Backend  │◄────────────────►│ Google Calendar │
│          │  Calendar API    │      API        │
└──────────┘                  └─────────────────┘
```

### Multi-User Support
Would require:
- PostgreSQL migration
- Authentication system
- Account/user relationship
- Permission model
