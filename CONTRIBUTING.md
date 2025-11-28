# Contributing to SE Notes

## Quick Start

```bash
# Clone
git clone https://github.com/factory-sagar/notes-droid.git
cd notes-droid

# Setup (installs deps + configures hooks)
make setup

# Start development
make dev
```

## Development Workflow

### 1. Create a branch
```bash
git checkout -b feat/my-feature
```

### 2. Make changes
- Backend: `backend/`
- Frontend: `frontend/`
- Hooks: `.githooks/`

### 3. Test locally
```bash
# Run both servers
make dev

# Or separately
make dev-backend   # Terminal 1
make dev-frontend  # Terminal 2

# Or use Docker
docker-compose up
```

### 4. Commit with conventional commits
```bash
git add .
git commit -m "feat(scope): description"
```

### 5. Push and create PR
```bash
git push -u origin feat/my-feature
```

## Git Hooks

All hooks are in `.githooks/` and are configured via `make setup-hooks`.

### Pre-commit Hook

Runs automatically before every commit to ensure code quality.

**Checks performed:**
1. **Security Checks**
   - Secret detection (API keys, tokens, passwords)
   - Sensitive file detection (.env, .pem, .key)
   - Large file warnings (>1MB)

2. **Go Backend Checks**
   - `gofmt` format validation
   - `go vet` static analysis
   - `go build` compilation check

3. **Frontend Checks**
   - `svelte-check` TypeScript validation
   - `console.log` detection

4. **General Checks**
   - Merge conflict markers
   - Trailing whitespace
   - Debugger statements

**Bypass (not recommended):**
```bash
git commit --no-verify
```

### Commit-msg Hook

Validates commit message format.

**Required format:**
```
type(scope): description

[optional body]

[optional footer]
```

**Valid types:**
| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation changes |
| `style` | Code style (formatting) |
| `refactor` | Code restructuring |
| `test` | Adding/updating tests |
| `chore` | Maintenance tasks |
| `init` | Initial commit |
| `perf` | Performance improvements |
| `ci` | CI/CD changes |
| `build` | Build system changes |
| `revert` | Revert previous commit |

**Examples:**
```
feat(notes): add export to PDF functionality
fix(api): handle empty participant arrays
docs(readme): add Docker instructions
refactor(handlers): extract validation logic
chore(deps): update Go dependencies
```

### Post-commit Hook

Tracks successful commits to Dashcode for analytics.

**Data sent:**
- Commit hash and short hash
- Commit message
- Author name and email
- Branch name
- Files changed count

**Note:** This hook never blocks commits - it always exits successfully.

## Dashcode Integration

Both pre-commit and post-commit hooks report to Dashcode at `localhost:3001`.

### Endpoint
```
POST http://localhost:3001/api/hooks/report
Content-Type: application/json
```

### Pre-commit Payload
```json
{
  "repoId": "e8450e26-c573-4aee-af1b-248405af0acc",
  "commitHash": "abc123...",
  "branch": "main",
  "trigger": "pre-commit",
  "status": "success",
  "durationMs": 0,
  "meta": {
    "author": "John Doe",
    "stagedFiles": 5,
    "errors": 0,
    "warnings": 1
  },
  "results": [
    {"type": "security", "status": "pass", "output": "...", "durationMs": 0, "exitCode": 0},
    {"type": "general", "status": "pass", "output": "...", "durationMs": 0, "exitCode": 0}
  ]
}
```

### Post-commit Payload
```json
{
  "repoName": "notes-droid",
  "commitHash": "abc123def456...",
  "branch": "main",
  "trigger": "post-commit",
  "status": "success",
  "durationMs": 100,
  "meta": {
    "author": "John Doe",
    "email": "john@example.com",
    "message": "feat: add new feature",
    "shortHash": "abc123",
    "filesChanged": 5
  },
  "results": [
    {
      "type": "commit",
      "status": "pass",
      "output": "Commit abc123 by John Doe",
      "durationMs": 100,
      "exitCode": 0
    }
  ]
}
```

### If Dashcode is Unavailable
- Pre-commit: Shows warning but allows commit (if no errors)
- Post-commit: Shows warning, commit is already complete

## Code Style

### Go (Backend)
- Use `gofmt` for formatting
- Run `go vet` before committing
- Keep handlers focused and small
- Use meaningful variable names
- Handle all errors explicitly

### Svelte/TypeScript (Frontend)
- Use TypeScript for type safety
- Follow existing component patterns
- Use Tailwind utility classes
- Keep components under 300 lines
- Use reactive statements ($:) appropriately

## Testing

### Backend
```bash
cd backend
go test ./...
go vet ./...
```

### Frontend
```bash
cd frontend
npm run check    # Type checking
npm run build    # Production build
```

### API Testing
```bash
# Health check
curl http://localhost:8080/health

# List accounts
curl http://localhost:8080/api/accounts

# Create account
curl -X POST http://localhost:8080/api/accounts \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Account"}'

# List tags
curl http://localhost:8080/api/tags

# Search
curl "http://localhost:8080/api/search?q=test"
```

## Project Structure

```
notes-droid/
├── backend/              # Go API
│   ├── cmd/server/       # Entry point
│   └── internal/         # Business logic
│       ├── handlers/     # HTTP handlers
│       ├── models/       # Data types
│       └── db/           # Database
├── frontend/             # SvelteKit
│   └── src/
│       ├── routes/       # Pages
│       └── lib/          # Shared code
├── .githooks/            # Git hooks
│   ├── pre-commit        # Security + linting
│   ├── commit-msg        # Message validation
│   └── post-commit       # Dashcode tracking
├── docs/                 # Documentation
├── .factory/skills/      # Agent skills
└── docker-compose.yml
```

## Adding New Features

### New API Endpoint
1. Add handler in `backend/internal/handlers/handlers.go`
2. Add route in `backend/cmd/server/main.go`
3. Add frontend function in `frontend/src/lib/utils/api.ts`

### New Page
1. Create `frontend/src/routes/[page]/+page.svelte`
2. Add to navItems in `frontend/src/routes/+layout.svelte`

### New Database Table
1. Add model in `backend/internal/models/models.go`
2. Add migration in `backend/internal/db/db.go`
3. Use `columnExists()` helper for ALTER TABLE

## Need Help?

- Check `AGENTS.md` for detailed AI agent context
- Check `PROJECT_STATUS.md` for feature status
- Check `docs/` for API documentation
- Check `.factory/skills/` for specific task guides
