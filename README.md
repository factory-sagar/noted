# Noted

A local-first notes application for Solutions Engineers to organize meeting notes, track follow-ups, and manage customer accounts.

![License](https://img.shields.io/badge/license-MIT-blue.svg)

## Features

- **Notes Management**
  - Rich text editor (TipTap) with code blocks and formatting
  - Organize notes by customer account
  - Initial call and follow-up templates (customizable)
  - Three view modes: Folders, Cards, Organized
  - Pin important notes, archive old ones
  - Soft delete with trash/restore functionality
  - File attachments support
  - Export to PDF (full or minimal)
  - Drag-to-reorder notes within accounts

- **Kanban Todo Board**
  - 4-column workflow: Not Started, In Progress, Stuck, Completed
  - Drag-and-drop cards between columns
  - Link todos to multiple notes
  - Priority levels (High/Medium/Low) with colored badges
  - High priority items auto-sort to top
  - Account tagging on todo cards
  - Pin important todos
  - Trash/restore for deleted todos
  - Filter by account, priority, linked notes
  - Bulk operations (select, move, delete)

- **Dashboard & Analytics**
  - Overview of notes, accounts, and todos
  - Clickable stat cards for quick navigation
  - Incomplete fields tracker
  - Notes by account breakdown
  - Todo status breakdown and completion rate

- **Quick Capture**
  - Fast note or todo creation from anywhere
  - Keyboard shortcut accessible

- **Google Calendar Integration**
  - OAuth-based Google Calendar sync
  - View calendar events
  - Create notes from calendar events
  - Auto-populate participants by email domain

- **Activities Timeline**
  - Track activity history per account
  - Automatic logging of key actions

- **Modern UI**
  - Light and dark mode
  - Smooth animations
  - Minimalist Asana/Factory.ai inspired design
  - Global search (Cmd+K) with fuzzy matching
  - Responsive sidebar

## Quick Start

### Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/factory-sagar/notes-droid.git
cd notes-droid

# Start with Docker Compose
docker-compose up

# App available at http://localhost:3000
```

### Manual Setup

#### Prerequisites
- Go 1.24+
- Node.js 20+
- SQLite (usually pre-installed on macOS/Linux)

#### Backend
```bash
cd backend
go mod download
go run ./cmd/server/
# Server runs on http://localhost:8080
```

#### Frontend
```bash
cd frontend
npm install
npm run dev
# App runs on http://localhost:5173
```

### Using Make
```bash
# Development (runs both)
make dev

# Or separately
make dev-backend
make dev-frontend

# Build for production
make build

# Docker
make docker
```

## Project Structure

```
notes-droid/
├── backend/                 # Go API server
│   ├── cmd/server/          # Entry point
│   └── internal/
│       ├── handlers/        # HTTP handlers
│       ├── models/          # Data models
│       └── db/              # SQLite + FTS4
├── frontend/                # SvelteKit app
│   └── src/
│       ├── routes/          # Pages (dashboard, notes, todos, accounts, calendar, settings)
│       └── lib/
│           ├── components/  # UI components
│           ├── stores/      # State management
│           └── utils/       # API client
├── .githooks/               # Git hooks (pre-commit, commit-msg, post-commit)
├── docs/                    # Documentation
├── data/                    # SQLite database + uploads (gitignored)
├── docker-compose.yml
└── Makefile
```

## API Endpoints

### Core Resources
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/accounts` | List all accounts |
| POST | `/api/accounts` | Create account |
| GET | `/api/accounts/:id` | Get account |
| PUT | `/api/accounts/:id` | Update account |
| DELETE | `/api/accounts/:id` | Delete account |
| GET | `/api/notes` | List all notes |
| POST | `/api/notes` | Create note |
| GET | `/api/notes/:id` | Get note with todos and tags |
| PUT | `/api/notes/:id` | Update note |
| DELETE | `/api/notes/:id` | Soft delete note |
| GET | `/api/todos` | List todos (optional `?status=` filter) |
| POST | `/api/todos` | Create todo |
| GET | `/api/todos/:id` | Get todo |
| PUT | `/api/todos/:id` | Update todo |
| DELETE | `/api/todos/:id` | Soft delete todo |

### Trash/Restore
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/notes/deleted` | List deleted notes |
| POST | `/api/notes/:id/restore` | Restore deleted note |
| DELETE | `/api/notes/:id/permanent` | Permanently delete note |
| GET | `/api/todos/deleted` | List deleted todos |
| POST | `/api/todos/:id/restore` | Restore deleted todo |
| DELETE | `/api/todos/:id/permanent` | Permanently delete todo |

### Pin/Archive
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/notes/:id/pin` | Toggle note pin status |
| POST | `/api/notes/:id/archive` | Toggle note archive status |
| POST | `/api/todos/:id/pin` | Toggle todo pin status |
| GET | `/api/notes/archived` | List archived notes |

### Tags & Linking
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/tags` | List all tags |
| POST | `/api/tags` | Create tag |
| PUT | `/api/tags/:id` | Update tag |
| DELETE | `/api/tags/:id` | Delete tag |
| POST | `/api/notes/:id/tags/:tagId` | Add tag to note |
| DELETE | `/api/notes/:id/tags/:tagId` | Remove tag from note |
| POST | `/api/todos/:id/notes/:noteId` | Link todo to note |
| DELETE | `/api/todos/:id/notes/:noteId` | Unlink todo from note |

### Activities & Attachments
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/accounts/:id/activities` | Get account activity timeline |
| POST | `/api/activities` | Create activity |
| GET | `/api/notes/:id/attachments` | List note attachments |
| POST | `/api/notes/:id/attachments` | Upload attachment |
| DELETE | `/api/notes/:id/attachments/:attachmentId` | Delete attachment |

### Search, Analytics & Utilities
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/search?q=term` | Full-text search (fuzzy, FTS4) |
| GET | `/api/analytics` | Dashboard statistics |
| GET | `/api/analytics/incomplete` | Notes with missing fields |
| GET | `/api/notes/:id/export` | Export note for PDF |
| POST | `/api/quick-capture` | Quick create note or todo |
| POST | `/api/accounts/:id/notes/reorder` | Reorder notes in account |

### Calendar (Google OAuth)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/calendar/auth` | Get OAuth URL |
| GET | `/api/calendar/callback` | OAuth callback |
| GET | `/api/calendar/config` | Get connection status |
| DELETE | `/api/calendar/disconnect` | Disconnect calendar |
| GET | `/api/calendar/events` | List calendar events |
| GET | `/api/calendar/events/:eventId` | Get single event |
| POST | `/api/calendar/parse-participants` | Parse attendees by domain |

## Environment Variables

### Backend
| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | Server port |
| GOOGLE_CLIENT_ID | - | Google OAuth client ID (for calendar) |
| GOOGLE_CLIENT_SECRET | - | Google OAuth client secret |

### Frontend
| Variable | Default | Description |
|----------|---------|-------------|
| PUBLIC_API_URL | http://localhost:8080 | Backend API URL |

## Tech Stack

- **Frontend**: SvelteKit 2.x, Tailwind CSS 3.x, TipTap (rich text), svelte-dnd-action
- **Backend**: Go 1.24+, Gin framework, SQLite with FTS4
- **Containerization**: Docker, Docker Compose

## Roadmap

- [x] Google Calendar OAuth integration
- [x] Auto-populate participants from calendar
- [x] PDF export with proper formatting
- [x] Full-text search with fuzzy matching
- [x] Note templates customization
- [x] Trash/recycle bin with restore
- [x] Pin and archive functionality
- [x] File attachments
- [x] Quick capture
- [x] Activity timeline
- [ ] Multi-user support
- [ ] Mobile responsive improvements
- [ ] Unit and integration tests

## License

MIT


