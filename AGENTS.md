# Agent Guide - Noted

This document provides context for AI agents working on this repository.

## Project Overview

**Noted** is a local-first notes application designed for Solutions Engineers to manage customer meeting notes, track follow-ups, and organize accounts.

### Tech Stack
- **Frontend**: SvelteKit 2.x + Tailwind CSS 3.x + TipTap (rich text editor)
- **Backend**: Go 1.24+ with Gin framework
- **Database**: SQLite with FTS4 (full-text search)
- **Containerization**: Docker + Docker Compose

### Architecture
```
┌─────────────────┐     HTTP/JSON      ┌─────────────────┐
│    Frontend     │◄──────────────────►│     Backend     │
│  (SvelteKit)    │    localhost:5173  │    (Go/Gin)     │
│  Port: 5173     │    or :3000        │   Port: 8080    │
└─────────────────┘                    └────────┬────────┘
                                                │
                                       ┌────────▼────────┐
                                       │     SQLite      │
                                       │  ./data/notes.db│
                                       └─────────────────┘
```

## Directory Structure

```
notes-droid/
├── backend/                    # Go API server
│   ├── cmd/server/main.go      # Entry point, routes
│   └── internal/
│       ├── handlers/           # HTTP request handlers
│       │   ├── handlers.go     # Base Handler struct
│       │   ├── account.go      # Account handlers
│       │   ├── note.go         # Note handlers
│       │   ├── todo.go         # Todo handlers
│       │   ├── tag.go          # Tag handlers
│       │   ├── activity.go     # Activity handlers
│       │   ├── attachment.go   # Attachment handlers
│       │   ├── search.go       # Search handlers
│       │   ├── analytics.go    # Analytics handlers
│       │   ├── data.go         # Data management handlers
│       │   ├── quick_capture.go # Quick Capture handlers
│       │   └── calendar.go     # Google Calendar OAuth
│       ├── models/             # Data structures & request/response types
│       └── db/                 # Database setup, migrations, FTS4
├── frontend/                   # SvelteKit application
│   └── src/
│       ├── routes/             # Page components
│       │   ├── +page.svelte    # Dashboard (clickable stats)
│       │   ├── +layout.svelte  # App shell, sidebar, global search
│       │   ├── notes/          # Notes list & editor
│       │   ├── accounts/       # Account management page
│       │   ├── todos/          # Kanban board (4 columns)
│       │   ├── calendar/       # Calendar view with OAuth
│       │   └── settings/       # App settings, tags, templates
│       └── lib/
│           ├── stores/         # Svelte stores (state)
│           ├── editor/         # TipTap extensions
│           └── utils/
│               ├── api.ts      # API client with types
│               └── pdf.ts      # PDF generation (jsPDF)
├── tests/                      # End-to-End Tests (Playwright)
├── .githooks/                  # Git hooks for code quality
│   ├── pre-commit              # Linting and security checks
│   ├── commit-msg              # Conventional commit validation
│   └── post-commit             # Commit tracking
├── docs/                       # Additional documentation
├── docker-compose.yml          # Container orchestration
└── Makefile                    # Development commands
```

## Key Data Models

### Account
Customer account that groups related notes.
```go
type Account struct {
    ID           string    // UUID
    Name         string    // "Acme Corp"
    AccountOwner string    // Sales rep name
    Budget       *float64  // Customer budget
    EstEngineers *int      // POC team size
}
```

### Note
Meeting notes linked to an account, can have tags.
```go
type Note struct {
    ID                   string
    Title                string
    AccountID            string      // Foreign key
    TemplateType         string      // "initial" or "followup"
    InternalParticipants []string    // Internal team emails
    ExternalParticipants []string    // Customer emails
    Content              string      // HTML from TipTap
    MeetingID            *string     // Google Calendar event ID
    MeetingDate          *time.Time
    Pinned               bool        // Pin to top
    Archived             bool        // Archive (hide from main list)
    DeletedAt            *time.Time  // Soft delete timestamp
    DisplayOrder         int         // Custom ordering within account
    Tags                 []Tag       // Many-to-many relationship
    Todos                []Todo      // Linked todos
    Attachments          []Attachment // File attachments
}
```

### Todo
Follow-up items with 4-status kanban workflow.
```go
type Todo struct {
    ID          string
    Title       string
    Description string
    Status      string    // "not_started", "in_progress", "stuck", "completed"
    Priority    string    // "low", "medium", "high"
    DueDate     *time.Time
    AccountID   *string   // Optional account tag
    AccountName string    // Populated from join
    Pinned      bool      // Pin to top
    DeletedAt   *time.Time // Soft delete timestamp
    Notes       []Note    // Many-to-many relationship
}
```

### Activity
Activity timeline for accounts.
```go
type Activity struct {
    ID          string
    AccountID   string
    Type        string    // "note_created", "todo_completed", etc.
    Title       string
    Description string
    EntityType  string    // "note", "todo", etc.
    EntityID    string
    CreatedAt   time.Time
}
```

### Attachment
File attachments for notes.
```go
type Attachment struct {
    ID           string
    NoteID       string
    Filename     string    // UUID-prefixed filename
    OriginalName string    // Original upload name
    MimeType     string
    Size         int64
    CreatedAt    time.Time
}
```

### Tag
Tags for organizing notes.
```go
type Tag struct {
    ID        string
    Name      string    // "Follow-up", "Demo", "Urgent"
    Color     string    // Hex color "#ef4444"
    CreatedAt time.Time
}
```

## API Endpoints

### Accounts
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/accounts` | List all accounts |
| POST | `/api/accounts` | Create account |
| GET | `/api/accounts/:id` | Get account by ID |
| PUT | `/api/accounts/:id` | Update account |
| DELETE | `/api/accounts/:id` | Delete account |

### Notes
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/notes` | List all notes |
| POST | `/api/notes` | Create note |
| GET | `/api/notes/:id` | Get note with linked todos and tags |
| PUT | `/api/notes/:id` | Update note |
| DELETE | `/api/notes/:id` | Soft delete note |
| GET | `/api/accounts/:id/notes` | Get notes by account |
| GET | `/api/notes/:id/export` | Export note for PDF |
| GET | `/api/notes/deleted` | List deleted notes |
| POST | `/api/notes/:id/restore` | Restore deleted note |
| DELETE | `/api/notes/:id/permanent` | Permanently delete note |
| GET | `/api/notes/archived` | List archived notes |
| POST | `/api/notes/:id/pin` | Toggle pin status |
| POST | `/api/notes/:id/archive` | Toggle archive status |
| POST | `/api/accounts/:id/notes/reorder` | Reorder notes in account |

### Todos
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/todos` | List todos (optional `?status=` filter) |
| POST | `/api/todos` | Create todo |
| GET | `/api/todos/:id` | Get todo by ID |
| PUT | `/api/todos/:id` | Update todo (status for kanban) |
| DELETE | `/api/todos/:id` | Soft delete todo |
| POST | `/api/todos/:id/notes/:noteId` | Link todo to note |
| DELETE | `/api/todos/:id/notes/:noteId` | Unlink todo from note |
| GET | `/api/todos/deleted` | List deleted todos |
| POST | `/api/todos/:id/restore` | Restore deleted todo |
| DELETE | `/api/todos/:id/permanent` | Permanently delete todo |
| POST | `/api/todos/:id/pin` | Toggle pin status |

### Tags
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/tags` | List all tags |
| POST | `/api/tags` | Create tag |
| PUT | `/api/tags/:id` | Update tag |
| DELETE | `/api/tags/:id` | Delete tag |
| GET | `/api/notes/:id/tags` | Get tags for a note |
| POST | `/api/notes/:id/tags/:tagId` | Add tag to note |
| DELETE | `/api/notes/:id/tags/:tagId` | Remove tag from note |

### Activities
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/accounts/:id/activities` | Get account activity timeline |
| POST | `/api/activities` | Create activity |

### Attachments
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/notes/:id/attachments` | List note attachments |
| POST | `/api/notes/:id/attachments` | Upload attachment (multipart/form-data) |
| DELETE | `/api/notes/:id/attachments/:attachmentId` | Delete attachment |
| GET | `/uploads/:filename` | Access uploaded file |

### Search & Analytics
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/search?q=term` | Full-text search (fuzzy, FTS4) |
| GET | `/api/analytics` | Dashboard statistics |
| GET | `/api/analytics/incomplete` | Notes with missing fields |
| POST | `/api/quick-capture` | Quick create note or todo |

### Calendar (Google OAuth)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/calendar/auth` | Get OAuth URL |
| GET | `/api/calendar/callback` | OAuth callback |
| GET | `/api/calendar/config` | Get connection status |
| DELETE | `/api/calendar/disconnect` | Disconnect calendar |
| GET | `/api/calendar/events` | List calendar events |
| GET | `/api/calendar/events/:eventId` | Get single event |
| POST | `/api/calendar/parse-participants` | Parse attendees by domain |

## Frontend Pages

### Dashboard (`/`)
- Clickable stat cards (Notes, Accounts, Todos, Completion Rate)
- Notes by account breakdown
- Todos by status (Not Started, In Progress, Stuck, Completed)
- Incomplete fields tracker

### Notes (`/notes`)
- **Three view modes**: Folders, Cards, Organized (toggle in header)
- Rich text editor with TipTap
- PDF export (full/minimal)
- Move notes between accounts
- Merge accounts functionality
- Pin important notes (sort to top)
- Archive notes (hide from main list)
- Soft delete with trash/restore
- File attachments
- Drag-to-reorder notes within accounts

### Accounts (`/accounts`)
- Split view: accounts list + detail panel
- Account stats (notes, todos, completed/pending)
- Recent notes and todos per account
- Create, edit, delete accounts
- Activity timeline per account

### Todos (`/todos`)
- **4-column Kanban**: Not Started, In Progress, Stuck, Completed
- Completed section is full-width at bottom
- Priority indicator badges (H/M/L colored squares)
- High priority auto-sorts to top of each column
- Filter by: account, priority, linked notes
- Sort by: date, priority, title
- Bulk operations (select, move, delete)
- Account tags on cards
- Pin important todos
- Soft delete with trash/restore

### Calendar (`/calendar`)
- Google OAuth integration
- Event list view
- Create notes from calendar events
- Auto-populate participants from attendees

### Settings (`/settings`)
- **Default Views**: Notes view, Todos view, Accounts view preferences
- **Tags Management**: Create/edit/delete tags with color picker
- **Templates**: Customize note templates
- **Calendar**: Connect/disconnect Google Calendar
- **Data**: Export all data, delete all data
- **Trash**: View/restore/permanently delete items
- Theme toggle (light/dark)

### Quick Capture
- Accessible via keyboard shortcut or button
- Fast creation of notes or todos
- Optional account assignment
- Appears as modal overlay

## Development Commands

```bash
# Start development servers
make dev              # Both backend and frontend
make dev-backend      # Backend only (port 8080)
make dev-frontend     # Frontend only (port 5173)

# Build
make build            # Build both
make build-backend    # Build Go binary
make build-frontend   # Build SvelteKit

# Docker
make docker           # Build and run containers
make docker-down      # Stop containers

# Setup
make setup            # Install deps + configure hooks
make setup-hooks      # Configure git hooks only
```

## Git Hooks

This repo uses custom git hooks for code quality and security checks.

### Pre-commit Hook (`.githooks/pre-commit`)

**Purpose**: Validates code quality and security before allowing commits.

**Checks performed**:
1. **Security Checks**
   - Scans for secrets (API keys, tokens, passwords, private keys)
   - Detects sensitive file types (.env, .pem, .key, credentials.json)
   - Warns about large files (>1MB)

2. **Go Backend Checks**
   - `gofmt` - Format validation
   - `go vet` - Static analysis
   - `go build` - Compilation check
   - Common issues (fmt.Println usage, TODO comments)

3. **Frontend Checks**
   - `svelte-check` - TypeScript/Svelte validation
   - `console.log` detection

4. **General Checks**
   - Merge conflict markers
   - Trailing whitespace
   - Debugger statements

### Commit-msg Hook (`.githooks/commit-msg`)

**Purpose**: Validates conventional commit message format.

**Format**: `type(scope): description`

**Valid types**:
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation
- `style` - Formatting
- `refactor` - Code restructuring
- `test` - Tests
- `chore` - Maintenance
- `init` - Initial commit
- `perf` - Performance
- `ci` - CI/CD
- `build` - Build system
- `revert` - Revert commit

**Validation**:
- Checks message matches pattern
- Warns if >72 characters
- Warns if <10 characters
- Skips merge commits

### Post-commit Hook (`.githooks/post-commit`)

**Purpose**: Tracks successful commits.

**Note**: Post-commit never blocks - always exits 0.

## Common Tasks for Agents

### Adding a New API Endpoint
1. Add handler function in `backend/internal/handlers/handlers.go`
2. Add route in `backend/cmd/server/main.go`
3. Add corresponding function in `frontend/src/lib/utils/api.ts`
4. Update types if needed

### Adding a New Page
1. Create route in `frontend/src/routes/[page-name]/+page.svelte`
2. Add navigation link in `frontend/src/routes/+layout.svelte` (navItems array)
3. Import any needed stores from `frontend/src/lib/stores/`

### Modifying Database Schema
1. Update models in `backend/internal/models/models.go`
2. Add migration SQL in `backend/internal/db/db.go`
3. Use `columnExists()` helper for ALTER TABLE migrations
4. Update handlers to use new fields

### Adding a New Todo Status
1. Update `columns` array in `frontend/src/routes/todos/+page.svelte`
2. Add color to `getColumnColorClass()` and `getColumnOutlineColor()`
3. Update dashboard status list in `frontend/src/routes/+page.svelte`
4. Update `filterAndSortTodos()` if needed

### Working with Tags
1. Backend: Tags table + note_tags junction table
2. API: `/api/tags` CRUD + `/api/notes/:id/tags/:tagId` linking
3. Frontend: Tag management in Settings, tag display/selection in notes

### Working with Soft Delete
1. Delete sets `deleted_at` timestamp (not NULL = deleted)
2. Main queries filter `WHERE deleted_at IS NULL`
3. `/deleted` endpoints return items where `deleted_at IS NOT NULL`
4. Restore sets `deleted_at = NULL`
5. Permanent delete removes row from database

### Working with Attachments
1. Upload via multipart/form-data to `/api/notes/:id/attachments`
2. Files stored in `./data/uploads/` with UUID-prefixed names
3. Access via `/uploads/:filename` static route
4. Metadata stored in `attachments` table

### Working with Activities
1. Activities are linked to accounts via `account_id`
2. Auto-created on certain actions or manually via API
3. Displayed as timeline in account detail view

## Environment

### Backend
- Port: 8080 (configurable via `PORT` env var)
- Database: `./data/notes.db` (auto-created)
- Uploads: `./data/uploads/` (auto-created)
- No authentication (local use only)
- Google Calendar: Requires `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET`

### Frontend
- Dev port: 5173
- Production port: 3000
- API URL: `http://localhost:8080` (hardcoded in `api.ts`)

## Settings Storage

User preferences stored in localStorage:
- `darkMode` - Theme preference
- `autoSave` - Auto-save notes
- `defaultTemplate` - Default note template
- `defaultNotesView` - Notes page view mode (folders/cards/organized)
- `defaultTodosView` - Todos page view mode (kanban/list)
- `defaultAccountsView` - Accounts page view mode (split/grid)

## Known Limitations

1. **Single user**: No authentication, designed for local use
2. **FTS4 vs FTS5**: Using FTS4 for broader SQLite compatibility
3. **Google Calendar**: Requires OAuth setup (GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET)

## Git Workflow & PR Guidelines

### Branch Structure
```
main (production - protected)
  └── dev (development integration)
        ├── feature/* (new features)
        ├── fix/* (bug fixes)
        └── chore/* (maintenance)
```

### Creating Feature Branches
```bash
# Always branch from dev
git checkout dev
git pull origin dev
git checkout -b feature/feature-name

# Or for fixes
git checkout -b fix/bug-description
```

### Commit Guidelines
- Use conventional commits: `feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `test:`
- Keep commits focused (single feature/fix per commit)
- Write descriptive commit messages explaining "why" not just "what"
- Use `--no-verify` only for acceptable warnings (e.g., macOS API deprecations)

### Creating Pull Requests
```bash
# Push branch
git push -u origin feature/feature-name

# Create PR targeting dev branch
gh pr create --base dev --title "feat: Description" --body "..."
```

### PR Description Template
```markdown
## Summary
Brief description of what this PR does.

## Changes
- List of changes made
- Include new endpoints, UI changes, etc.

## Testing
1. Step-by-step testing instructions
2. Expected behavior
```

### Merging Flow
1. Feature branches → `dev` (via PR review)
2. `dev` → `main` (release to production)

### Pre-commit Hooks
- Security checks (secrets, sensitive files)
- Go: `gofmt`, `go vet`, `go build`
- Frontend: `svelte-check`
- Use `--no-verify` to bypass for acceptable warnings only

## Native macOS App (Wails)

### Build Commands
```bash
make wails-build    # Build Noted.app (~10-20 sec)
make wails-install  # Copy to /Applications
```

### App Location
- Build output: `backend/cmd/wails/build/bin/Noted.app`
- Data storage: `~/Library/Application Support/Noted/`

### EventKit (Apple Calendar)
- Only works in native app (not browser dev mode)
- Requires Calendar permission in System Settings
- Uses deprecated APIs for backward compatibility (macOS 10.13+)

## Internal Domain
- Internal contacts identified by email domain (default: `example.com`)
- Configure via `INTERNAL_DOMAIN` environment variable
- Example: `INTERNAL_DOMAIN=mycompany.com`
