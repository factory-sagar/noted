<p align="center">
  <img src="https://raw.githubusercontent.com/factory-sagar/noted/main/frontend/src/lib/assets/favicon.svg" width="80" height="80" alt="Noted Logo">
</p>

<h1 align="center">Noted</h1>

<p align="center">
  <strong>The Local-First Workspace for Solutions Engineers</strong>
</p>

<p align="center">
  Manage accounts • Track follow-ups • Organize meeting notes • Sync calendars
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#getting-started">Getting Started</a> •
  <a href="#tech-stack">Tech Stack</a> •
  <a href="#development">Development</a> •
  <a href="#api">API</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/version-1.1.0-blue.svg" alt="Version">
  <img src="https://img.shields.io/badge/license-MIT-green.svg" alt="License">
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8.svg?logo=go" alt="Go">
  <img src="https://img.shields.io/badge/SvelteKit-2.x-FF3E00.svg?logo=svelte" alt="SvelteKit">
  <img src="https://img.shields.io/badge/SQLite-FTS4-003B57.svg?logo=sqlite" alt="SQLite">
  <a href="https://github.com/factory-sagar/noted/actions"><img src="https://img.shields.io/github/actions/workflow/status/factory-sagar/noted/ci.yml?branch=main" alt="Build Status"></a>
</p>

---

## 🚀 Overview

**Noted** is a high-performance, local-first application designed for the unique workflow of Solutions Engineers. It bridges the gap between your calendar, your customer accounts, and your daily tasks.

Unlike generic note-taking apps, Noted understands **Accounts**, **Meetings**, and **Follow-ups**.

### Key Capabilities

*   **📝 Context-Aware Notes**: Link notes to specific accounts and meetings.
*   **📊 Kanban Workflow**: Built-in board to track todo status (Not Started → Complete).
*   **📅 Calendar Sync**: Deep integration with Apple Calendar & Google Calendar.
*   **🔍 Instant Search**: Full-text search across notes, accounts, and todos (powered by SQLite FTS4).
*   **🛡️ Local-First**: Your data lives on your machine. No cloud lock-in.

---

## ✨ Features

| Feature | Description |
|---------|-------------|
| **Smart Editor** | Rich text editing with TipTap, code blocks, and markdown support. |
| **Kanban Board** | Drag-and-drop task management with priority levels (H/M/L). |
| **Account Hub** | Centralized view of customer details, budgets, and engagement stats. |
| **Quick Capture** | `⌘+Shift+C` to capture thoughts instantly without breaking flow. |
| **Attachments** | Drag-and-drop file management for every note. |
| **Analytics** | Dashboard with completion rates, incomplete field tracking, and activity timelines. |
| **Theming** | Multiple built-in themes (SaaS, Nordic, Retro, Dark Mode). |

---

## 🛠️ Tech Stack

*   **Frontend**: [SvelteKit](https://kit.svelte.dev/) (TypeScript, Tailwind CSS)
*   **Backend**: [Go](https://go.dev/) (Gin Framework)
*   **Database**: SQLite (with FTS4)
*   **Desktop App**: [Wails](https://wails.io/) (for macOS native build)

---

## 🏁 Getting Started

### Prerequisites

*   **Go**: 1.24 or higher
*   **Node.js**: 20 or higher
*   **Make**: For running build commands

### Installation

1.  **Clone the repository**
    ```bash
    git clone https://github.com/factory-sagar/noted.git
    cd noted
    ```

2.  **Setup dependencies & hooks**
    ```bash
    make setup
    ```

3.  **Start Development Servers**
    ```bash
    make dev
    ```
    *   Frontend: `http://localhost:5173`
    *   Backend: `http://localhost:8080`

### Docker

Prefer containers? Run the full stack with Docker Compose:

```bash
make docker
# Access app at http://localhost:3000
```

---

## 💻 Development Workflow

We use a `Makefile` to streamline common tasks.

| Command | Description |
|---------|-------------|
| `make dev` | Run both backend and frontend in watch mode. |
| `make build` | Compile production binaries for both. |
| `make setup-hooks` | Configure Git hooks (Pre-commit analysis). |
| `make clean` | Remove build artifacts and temp files. |

### Code Quality & Hooks

This project enforces code quality via **Git Hooks**:

*   **Pre-commit**: Runs `go vet`, `gofmt`, `svelte-check`, and security scans.
*   **Commit-msg**: Enforces [Conventional Commits](https://www.conventionalcommits.org/) format (e.g., `feat: add new sidebar`).

> **Note**: If you need to bypass hooks (e.g., for WIP commits), use `git commit --no-verify`.

---

## 📂 Project Structure

```
noted/
├── backend/
│   ├── cmd/server/       # API Entry point
│   ├── internal/
│   │   ├── handlers/     # HTTP Controllers (Business Logic)
│   │   ├── models/       # Data Structures
│   │   └── db/           # Database & Migrations
│   └── data/             # SQLite DB storage (gitignored)
├── frontend/
│   ├── src/
│   │   ├── routes/       # SvelteKit Pages
│   │   ├── lib/          # Components & Utilities
│   │   └── app.html      # Root HTML
├── .githooks/            # Custom Git Hooks
└── Makefile              # Task Automation
```

---

## 🔌 API Reference

The backend provides a RESTful API.

*   **Notes**: `/api/notes`
*   **Accounts**: `/api/accounts`
*   **Todos**: `/api/todos`
*   **Search**: `/api/search?q={term}`
*   **Calendar**: `/api/calendar/events`

See [docs/API.md](docs/API.md) for the full OpenAPI specification.

---

## 🤝 Contributing

1.  Fork the project.
2.  Create your feature branch (`git checkout -b feat/amazing-feature`).
3.  Commit your changes (`git commit -m 'feat: add amazing feature'`).
4.  Push to the branch (`git push origin feat/amazing-feature`).
5.  Open a Pull Request.

---

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.
