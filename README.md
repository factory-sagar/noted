<p align="center">
  <a href="https://github.com/factory-sagar/notes-droid">
    <img src="frontend/src/lib/assets/logo-notion-style.svg" width="120" height="120" alt="Noted Logo">
  </a>
</p>

<h1 align="center">Noted</h1>

<p align="center">
  <strong>The Local-First Intelligence Workspace for Solutions Engineers</strong>
</p>

<p align="center">
  <a href="https://github.com/factory-sagar/notes-droid/actions"><img src="https://img.shields.io/github/actions/workflow/status/factory-sagar/notes-droid/test.yml?branch=main&style=flat-square" alt="Build Status"></a>
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8.svg?style=flat-square&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/SvelteKit-2.x-FF3E00.svg?style=flat-square&logo=svelte" alt="SvelteKit">
  <img src="https://img.shields.io/badge/license-MIT-green.svg?style=flat-square" alt="License">
  <img src="https://img.shields.io/badge/platform-macOS-lightgrey.svg?style=flat-square&logo=apple" alt="Platform">
</p>

<p align="center">
  <a href="#-features">Features</a> •
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-tech-stack">Tech Stack</a> •
  <a href="#-development">Development</a> •
  <a href="#-api-reference">API</a>
</p>

<br>

## 🚀 Overview

**Noted** is an open-source, high-performance workspace built specifically for **Solutions Engineers** (SEs) and **Sales Engineers**. 

Generic note-taking apps treat every document the same. Noted understands the SE workflow: you don't just take notes—you manage **Accounts**, drive **Opportunities**, and track **Follow-ups** across dozens of meetings.

### Why Noted?

> *"Stop losing context between meetings. Keep everything organized, searchable, and actionable."*

*   **⚡️ Local-First Speed**: Zero latency. Your data lives on your machine (SQLite). No cloud lock-in.
*   **🧠 Context-Aware**: Every note is linked to an Account and a Meeting.
*   **🔄 Workflow Integrated**: Seamlessly bridges your Calendar, Notes, and Tasks.

---

## ✨ Features

### 📝 Intelligent Note Taking
*   **Rich Text Editor**: Powered by TipTap with support for code blocks, markdown, and images.
*   **Template System**: Pre-built templates for *Initial Discovery*, *Technical Deep Dive*, and *POC Planning*.
*   **Account Linking**: Automatically associate notes with customer accounts.

### ✅ Kanban Task Management
*   **Integrated Workflow**: Don't just write "follow up"—track it. 
*   **4-Stage Board**: `Not Started` → `In Progress` → `Stuck` → `Completed`.
*   **Prioritization**: Visual H/M/L priority indicators.

### 📅 Calendar Intelligence
*   **Apple Calendar Integration**: Native macOS EventKit integration.
*   **One-Click Notes**: Create a pre-filled note from any calendar event instantly.
*   **Participant Extraction**: Automatically captures attendee details.

### 👥 Contact Management
*   **Auto-Extraction**: Contacts automatically captured from meeting participants.
*   **Domain Intelligence**: Smart account suggestions based on email domains.
*   **Internal/External Classification**: Mark contacts as internal team or external customers.
*   **Bulk Operations**: Manage multiple contacts efficiently.

### 🗑️ Unified Trash
*   **Chrome-Style Management**: View all deleted items in one place.
*   **Category Tabs**: Filter by Notes, Todos, Accounts, or Contacts.
*   **Restore or Delete**: Recover items or permanently remove them.
*   **Empty by Category**: Clear trash selectively or all at once.

### 🎨 Theme System
*   **11 Built-in Themes**: Modern, Minimal, Nordic, Noir, Cyber, Monokai, Retro, Dracula, Solarized, Ocean, Forest.
*   **Dark/Light Mode**: Each theme supports both modes.
*   **Instant Switching**: Change themes without restart.

### 📊 Account Hub & Analytics
*   **360° View**: See all notes, tasks, and activity for a specific customer in one place.
*   **Dashboard**: Track your completion rates, meeting volume, and "incomplete data" warnings.
*   **Search**: Instant full-text search (FTS4) across every data point.

---

## 🛠️ Tech Stack

Built with performance and developer experience in mind.

| Layer | Technology | Description |
|-------|------------|-------------|
| **Frontend** | ![Svelte](https://img.shields.io/badge/-SvelteKit-FF3E00?style=flat-square&logo=svelte&logoColor=white) | Reactive UI, TypeScript, Tailwind CSS |
| **Backend** | ![Go](https://img.shields.io/badge/-Go-00ADD8?style=flat-square&logo=go&logoColor=white) | High-performance REST API (Gin Framework) |
| **Database** | ![SQLite](https://img.shields.io/badge/-SQLite-003B57?style=flat-square&logo=sqlite&logoColor=white) | Embedded SQL with FTS4 full-text search |
| **Testing** | ![Playwright](https://img.shields.io/badge/-Playwright-2EAD33?style=flat-square&logo=playwright&logoColor=white) | E2E Testing & Vitest for Unit Tests |
| **Desktop** | ![Wails](https://img.shields.io/badge/-Wails-CC3534?style=flat-square&logo=wails&logoColor=white) | Native macOS application wrapper |

---

## 🏁 Quick Start

### Prerequisites
*   **Go**: 1.24+
*   **Node.js**: 20+
*   **Make**: (Optional, for build scripts)

### ⚡️ Instant Dev Environment

1.  **Clone & Setup**
    ```bash
    git clone https://github.com/factory-sagar/notes-droid.git
    cd notes-droid
    make setup
    ```

2.  **Run**
    ```bash
    make dev
    ```
    
    | Service | URL |
    |---------|-----|
    | **Frontend** | `http://localhost:5173` |
    | **Backend API** | `http://localhost:8080` |

### 📦 Native Mac App (Wails)

For the best experience on macOS (including full Calendar integration), build the native app.

1.  **Install Wails CLI**
    ```bash
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    ```

2.  **Build & Install**
    ```bash
    make wails-build
    make wails-install
    ```
    
    This compiles the app and installs `Noted.app` to your `/Applications` folder.

3.  **Run**
    Launch **Noted** from Spotlight or your Applications folder.

    > **Note**: The first launch may require you to approve "Calendar Access" in System Settings.

---

## 💻 Development

We strictly enforce code quality. All PRs must pass our pre-commit hooks.

### Common Commands

```bash
make dev            # Web development (fast refresh)
make wails-dev      # Native app development (hot reload)
make build          # Compile production binaries (web)
make test           # Run all tests (Backend + Frontend)
```

### 🛡️ Quality Gates

Our `pre-commit` hook runs automatically to ensure:
*   ✅ **Go**: `go vet`, `gofmt`, `go test`
*   ✅ **Frontend**: `svelte-check`, `vitest`
*   ✅ **Security**: Secret scanning & sensitive file detection

> **Pro Tip**: Need to bypass hooks for a WIP commit? Use `git commit --no-verify`.

---

## 📂 Project Structure

```
noted/
├── backend/
│   ├── cmd/server/       # 🚀 API Entry point
│   ├── internal/
│   │   ├── handlers/     # 🎮 HTTP Controllers
│   │   ├── models/       # 📦 Data Structures
│   │   └── db/           # 💾 Database & Migrations
│   └── data/             # 📂 Local storage (gitignored)
├── frontend/
│   ├── src/
│   │   ├── routes/       # 🌐 SvelteKit Pages
│   │   ├── lib/          # 🧩 Components & Stores
│   │   └── app.html      # 📄 Root HTML
├── .githooks/            # 🪝 Git Hooks
└── Makefile              # 🛠️ Task Automation
```

---

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

<br>

<p align="center">
  <sub>Built with ❤️ by <a href="https://github.com/factory-sagar">Sagar</a> for Solutions Engineers everywhere.</sub>
</p>
