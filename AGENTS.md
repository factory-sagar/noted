# Agent Guide - SE Notes

This document provides context for AI agents working on this repository.

## Project Overview

**SE Notes** is a local-first notes application designed for Solutions Engineers to manage customer meeting notes, track follow-ups, and organize accounts.

### Tech Stack
- **Frontend**: SvelteKit 2.x + Tailwind CSS 3.x + TipTap (rich text editor)
- **Backend**: Go 1.21+ with Gin framework
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
│       ├── models/             # Data structures & request/response types
│       └── db/                 # Database setup, migrations, FTS4
├── frontend/                   # SvelteKit application
│   └── src/
│       ├── routes/             # Page components
│       │   ├── +page.svelte    # Dashboard
│       │   ├── notes/          # Notes list & editor
│       │   ├── todos/          # Kanban board
│       │   ├── calendar/       # Calendar view
│       │   └── settings/       # App settings
│       └── lib/
│           ├── stores/         # Svelte stores (state)
│           └── utils/api.ts    # API client
├── .githooks/                  # Git hooks with dashcode integration
│   ├── pre-commit              # Linting, security checks
│   ├── commit-msg              # Conventional commit validation
│   └── post-commit             # Dashcode tracking
├── docs/                       # Additional documentation
│   ├── API.md                  # API endpoint reference
│   ├── SETUP.md                # Setup instructions
│   └── ARCHITECTURE.md         # System design
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
Meeting notes linked to an account.
```go
type Note struct {
    ID                   string
    Title                string
    AccountID            string      // Foreign key
    TemplateType         string      // "initial" or "followup"
    InternalParticipants []string    // @factory.ai emails
    ExternalParticipants []string    // Customer emails
    Content              string      // HTML from TipTap
    MeetingID            *string     // Google Calendar event ID
    MeetingDate          *time.Time
}
```

### Todo
Follow-up items, can link to multiple notes.
```go
type Todo struct {
    ID          string
    Title       string
    Description string
    Status      string  // "not_started", "in_progress", "completed"
    Priority    string  // "low", "medium", "high"
    DueDate     *time.Time
    Notes       []Note  // Many-to-many relationship
}
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/accounts` | List all accounts |
| POST | `/api/accounts` | Create account |
| GET | `/api/accounts/:id` | Get account by ID |
| PUT | `/api/accounts/:id` | Update account |
| DELETE | `/api/accounts/:id` | Delete account |
| GET | `/api/notes` | List all notes |
| POST | `/api/notes` | Create note |
| GET | `/api/notes/:id` | Get note with linked todos |
| PUT | `/api/notes/:id` | Update note |
| DELETE | `/api/notes/:id` | Delete note |
| GET | `/api/todos` | List todos (optional `?status=` filter) |
| POST | `/api/todos` | Create todo |
| PUT | `/api/todos/:id` | Update todo (including status for kanban) |
| DELETE | `/api/todos/:id` | Delete todo |
| POST | `/api/todos/:id/notes/:noteId` | Link todo to note |
| DELETE | `/api/todos/:id/notes/:noteId` | Unlink todo from note |
| GET | `/api/search?q=term` | Full-text search |
| GET | `/api/analytics` | Dashboard statistics |
| GET | `/api/analytics/incomplete` | Notes with missing fields |

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

## Git Hooks & Dashcode

This repo uses custom git hooks that report to **Dashcode** (localhost:3001).

### Pre-commit Hook
- Security: Scans for secrets, API keys, sensitive files
- Go: Format check (gofmt), vet, build verification
- General: Merge conflicts, trailing whitespace, debugger statements
- Reports results to Dashcode

### Commit-msg Hook
- Validates conventional commit format
- Required format: `type(scope): description`
- Valid types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `init`, `perf`, `ci`, `build`, `revert`

### Post-commit Hook
- Posts commit metadata to Dashcode for tracking
- Includes: author, message, files changed, branch

## Testing Locally

```bash
# 1. Start backend
cd backend && go run ./cmd/server/

# 2. In another terminal, start frontend
cd frontend && npm run dev

# 3. Open http://localhost:5173

# OR use Docker
docker-compose up
# Open http://localhost:3000
```

## Common Tasks for Agents

### Adding a New API Endpoint
1. Add handler function in `backend/internal/handlers/handlers.go`
2. Add route in `backend/cmd/server/main.go`
3. Add corresponding function in `frontend/src/lib/utils/api.ts`
4. Update `docs/API.md`

### Adding a New Page
1. Create route in `frontend/src/routes/[page-name]/+page.svelte`
2. Add navigation link in `frontend/src/routes/+layout.svelte`
3. Import any needed stores from `frontend/src/lib/stores/`

### Modifying Database Schema
1. Update models in `backend/internal/models/models.go`
2. Add migration SQL in `backend/internal/db/db.go` (Migrate function)
3. Update handlers to use new fields
4. Delete `data/notes.db` to reset (or write migration logic)

### Adding a New Field to Notes
1. Add field to `Note` struct in `backend/internal/models/models.go`
2. Add to `CreateNoteRequest` and `UpdateNoteRequest`
3. Update SQL in `db.go` migration
4. Update `GetNote`, `CreateNote`, `UpdateNote` handlers
5. Add to frontend form in `frontend/src/routes/notes/[id]/+page.svelte`
6. Update API types in `frontend/src/lib/utils/api.ts`

## Environment

### Backend
- Port: 8080 (configurable via `PORT` env var)
- Database: `./data/notes.db` (auto-created)
- No authentication (local use only)

### Frontend
- Dev port: 5173
- Production port: 3000
- API URL: `http://localhost:8080` (hardcoded in `api.ts`)

## Known Limitations

1. **Single user**: No authentication, designed for local use
2. **Google Calendar**: Placeholder only, OAuth not implemented
3. **PDF Export**: Returns JSON, client-side PDF generation needed
4. **FTS4 vs FTS5**: Using FTS4 for broader SQLite compatibility

## Future Roadmap

- [ ] Google Calendar OAuth integration
- [ ] Auto-populate participants from calendar
- [ ] Proper PDF export with formatting
- [ ] Note templates customization
- [ ] Search result highlighting
- [ ] Mobile responsive improvements
