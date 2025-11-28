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
```

### 4. Commit with conventional commits
```bash
git add .
git commit -m "feat(scope): description"
```

The pre-commit hook will:
- Check for secrets
- Lint Go code
- Validate commit message
- Report to Dashcode

### 5. Push and create PR
```bash
git push -u origin feat/my-feature
```

## Commit Message Format

```
type(scope): short description

[optional longer description]

[optional footer]
```

### Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting
- `refactor`: Code restructuring
- `test`: Adding tests
- `chore`: Maintenance

### Examples
```
feat(notes): add export to PDF functionality
fix(api): handle empty participant arrays
docs(readme): add Docker instructions
refactor(handlers): extract validation logic
```

## Code Style

### Go (Backend)
- Use `gofmt` for formatting
- Run `go vet` before committing
- Keep handlers focused and small
- Use meaningful variable names

### Svelte/TypeScript (Frontend)
- Use TypeScript for type safety
- Follow existing component patterns
- Use Tailwind utility classes
- Keep components under 300 lines

## Testing

### Backend
```bash
cd backend
go test ./...
```

### Frontend
```bash
cd frontend
npm run check    # Type checking
npm run lint     # Linting
```

### API Testing
```bash
# Health check
curl http://localhost:8080/health

# Create test data
curl -X POST http://localhost:8080/api/accounts \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Account"}'
```

## Project Structure

```
notes-droid/
├── backend/           # Go API
│   ├── cmd/server/    # Entry point
│   └── internal/      # Business logic
├── frontend/          # SvelteKit
│   └── src/
│       ├── routes/    # Pages
│       └── lib/       # Shared code
├── .githooks/         # Git hooks
├── docs/              # Documentation
└── docker-compose.yml
```

## Need Help?

- Check `AGENTS.md` for detailed context
- Check `docs/` for API and setup guides
- Check `.factory/skills/` for specific tasks
