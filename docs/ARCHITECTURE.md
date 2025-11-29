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
│  │Accounts │ │  Notes  │ │  Todos  │ │ Search  │ │Calendar ││
│  │Handler  │ │ Handler │ │ Handler │ │ Handler │ │Handler  ││
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬───┘│
│       │           │           │           │           │      │
│       └───────────┴───────────┴───────────┴───────────┘      │
│                           │                                   │
│                    SQLite + FTS4                              │
│                    (Database)                                 │
│                                                               │
│                         Backend                               │
│                       (Go + Gin)                              │
└──────────────────────────────────────────────────────────────┘
```

## Data Model

```
┌─────────────┐       ┌─────────────────────────────┐
│   Account   │       │            Note             │
├─────────────┤       ├─────────────────────────────┤
│ id          │◄──────│ account_id                  │
│ name        │       │ id                          │
│ account_owner       │ title                       │
│ budget      │       │ template_type               │
│ est_engineers       │ internal_participants       │
│ created_at  │       │ external_participants       │
│ updated_at  │       │ content                     │
└─────────────┘       │ meeting_id                  │
      │               │ meeting_date                │
      │               │ pinned                      │
      │               │ archived                    │
      │               │ deleted_at (soft delete)    │
      │               │ display_order               │
      │               │ created_at                  │
      │               │ updated_at                  │
      │               └──────┬──────────────────────┘
      │                      │
      │                      │ many-to-many
      │                      │
      │               ┌──────┴──────┐
      │               │ note_todos  │
      │               ├─────────────┤
      │               │ note_id     │
      │               │ todo_id     │
      │               │ created_at  │
      │               └──────┬──────┘
      │                      │
      │               ┌──────┴──────────────────────┐
      │               │            Todo             │
      │               ├─────────────────────────────┤
      └──────────────►│ account_id (optional)       │
                      │ id                          │
                      │ title                       │
                      │ description                 │
                      │ status (not_started,        │
                      │   in_progress, stuck,       │
                      │   completed)                │
                      │ priority (low, medium, high)│
                      │ due_date                    │
                      │ pinned                      │
                      │ deleted_at (soft delete)    │
                      │ created_at                  │
                      │ updated_at                  │
                      └─────────────────────────────┘

┌─────────────┐       ┌─────────────┐
│    Tag      │       │  note_tags  │
├─────────────┤       ├─────────────┤
│ id          │◄──────│ tag_id      │
│ name        │       │ note_id     │
│ color       │       │ created_at  │
│ created_at  │       └─────────────┘
└─────────────┘

┌─────────────────────────────┐
│         Activity            │
├─────────────────────────────┤
│ id                          │
│ account_id                  │
│ type                        │
│ title                       │
│ description                 │
│ entity_type                 │
│ entity_id                   │
│ created_at                  │
└─────────────────────────────┘

┌─────────────────────────────┐
│        Attachment           │
├─────────────────────────────┤
│ id                          │
│ note_id                     │
│ filename                    │
│ original_name               │
│ mime_type                   │
│ size                        │
│ created_at                  │
└─────────────────────────────┘
```

## Key Design Decisions

### 1. SQLite with FTS4
- **Why**: Zero setup, single file, excellent for local/personal use
- **FTS4**: Full-text search on notes without external search service
- **Trade-off**: Single-user focused, would need migration for multi-user

### 2. Soft Delete Pattern
- **Why**: User safety - prevent accidental data loss
- **Implementation**: `deleted_at` timestamp field; NULL = active, non-NULL = deleted
- **Trash**: Deleted items can be restored or permanently removed

### 3. SvelteKit
- **Why**: Fast, minimal JS bundle, excellent DX
- **Animations**: svelte-motion for smooth transitions
- **Trade-off**: Smaller ecosystem than React

### 4. TipTap Editor
- **Why**: Modern ProseMirror wrapper, extensible, works well with Svelte
- **Features**: Rich text, code blocks, markdown-like shortcuts
- **Storage**: HTML string in database (simple, exportable)

### 5. Kanban with svelte-dnd-action
- **Why**: Native Svelte drag-and-drop, good accessibility
- **UX**: Immediate optimistic updates, then sync to server
- **Layout**: 4 columns (Not Started, In Progress, Stuck) + full-width Completed

### 6. Many-to-Many Relationships
- **Todos ↔ Notes**: Junction table `note_todos` - todos can span multiple calls
- **Notes ↔ Tags**: Junction table `note_tags` - flexible categorization

### 7. Pin & Archive
- **Pin**: Boolean field, pinned items sort to top
- **Archive**: Boolean field, archived items hidden from main lists

## Request Flow

### Creating a Note
```
1. User fills form in frontend
2. POST /api/notes with JSON body
3. Handler validates request
4. Generate UUID, set timestamps
5. Insert into SQLite
6. FTS4 trigger auto-updates search index
7. Return created note
8. Frontend updates store
9. Navigate to note editor
```

### Searching
```
1. User types in search bar (debounced)
2. GET /api/search?q=term
3. Handler queries FTS4: notes_fts MATCH term
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

### Soft Delete Flow
```
1. User clicks delete
2. DELETE /api/notes/:id
3. Handler sets deleted_at = NOW()
4. Note removed from main queries
5. Note appears in GET /api/notes/deleted
6. User can POST /api/notes/:id/restore
7. Or DELETE /api/notes/:id/permanent
```

## File Storage

### Attachments
- **Location**: `./data/uploads/`
- **Naming**: UUID prefix + original filename
- **Serving**: Static file route at `/uploads/:filename`
- **Metadata**: Stored in `attachments` table

## Google Calendar Integration

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

Environment variables required:
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`

## Future Considerations

### Multi-User Support
Would require:
- PostgreSQL migration
- Authentication system (JWT/sessions)
- Account/user relationship
- Permission model
- Row-level security

### Performance Optimization
- Pagination for large note lists
- Virtual scrolling for kanban
- Background sync for offline support
- Service worker caching
