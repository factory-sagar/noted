# SE Notes

A minimalist notes application for Solutions Engineers to organize meeting notes, track follow-ups, and manage customer accounts.

![License](https://img.shields.io/badge/license-MIT-blue.svg)

## Features

- **Notes Management**
  - Rich text editor with code blocks
  - Organize notes by customer account
  - Initial call and follow-up templates
  - Auto-populate participants from calendar
  - Export to PDF (full or minimal)

- **Kanban Todo Board**
  - Drag-and-drop cards between columns
  - Link todos to multiple notes
  - Track follow-up items from calls
  - Priority levels and due dates

- **Dashboard & Analytics**
  - Overview of notes, accounts, and todos
  - Incomplete fields tracker
  - Notes by account breakdown
  - Todo completion rate

- **Calendar Integration** *(Coming Soon)*
  - Google Calendar sync
  - Click meeting to create/link notes
  - Auto-detect participants by domain

- **Modern UI**
  - Light and dark mode
  - Smooth animations
  - Minimalist Asana/Factory.ai inspired design
  - Global search (Cmd+K)

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
- Go 1.21+
- Node.js 20+
- SQLite

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
│       ├── db/              # SQLite + FTS5
│       ├── calendar/        # Google Calendar (TBD)
│       └── pdf/             # PDF export (TBD)
├── frontend/                # SvelteKit app
│   └── src/
│       ├── routes/          # Pages
│       └── lib/
│           ├── components/  # UI components
│           ├── stores/      # State management
│           └── utils/       # API client
├── docs/                    # Documentation
├── data/                    # SQLite database (gitignored)
├── docker-compose.yml
└── Makefile
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/accounts` | List all accounts |
| POST | `/api/accounts` | Create account |
| GET | `/api/accounts/:id` | Get account |
| PUT | `/api/accounts/:id` | Update account |
| DELETE | `/api/accounts/:id` | Delete account |
| GET | `/api/notes` | List all notes |
| POST | `/api/notes` | Create note |
| GET | `/api/notes/:id` | Get note |
| PUT | `/api/notes/:id` | Update note |
| DELETE | `/api/notes/:id` | Delete note |
| GET | `/api/notes/:id/export` | Export note |
| GET | `/api/todos` | List all todos |
| POST | `/api/todos` | Create todo |
| PUT | `/api/todos/:id` | Update todo |
| DELETE | `/api/todos/:id` | Delete todo |
| POST | `/api/todos/:id/notes/:noteId` | Link todo to note |
| GET | `/api/search` | Global search |
| GET | `/api/analytics` | Dashboard stats |

## Environment Variables

### Backend
| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | Server port |

### Frontend
| Variable | Default | Description |
|----------|---------|-------------|
| PUBLIC_API_URL | http://localhost:8080 | Backend API URL |

## Tech Stack

- **Frontend**: SvelteKit, Tailwind CSS, TipTap, svelte-dnd-action
- **Backend**: Go, Gin, SQLite with FTS5
- **Containerization**: Docker, Docker Compose

## Roadmap

- [ ] Google Calendar OAuth integration
- [ ] Auto-populate participants from calendar
- [ ] PDF export with proper formatting
- [ ] Full-text search highlighting
- [ ] Note templates customization
- [ ] Multi-user support
- [ ] Mobile responsive improvements

## License

MIT
