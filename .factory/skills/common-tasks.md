# Skill: Common Tasks

## Overview
Quick reference for common development tasks in SE Notes.

## Starting Development

```bash
# Full setup (first time)
make setup

# Start both servers
make dev

# Or start separately
make dev-backend   # http://localhost:8080
make dev-frontend  # http://localhost:5173
```

## Adding a New Feature

### 1. Backend endpoint
```bash
# Edit these files:
backend/internal/models/models.go      # Add types
backend/internal/db/db.go              # Add migrations
backend/internal/handlers/handlers.go  # Add handlers
backend/cmd/server/main.go             # Add routes
```

### 2. Frontend page/component
```bash
# Create route:
frontend/src/routes/my-page/+page.svelte

# Add to navigation:
frontend/src/routes/+layout.svelte

# Add API client functions:
frontend/src/lib/utils/api.ts
```

### 3. Test and commit
```bash
# Test locally
make dev

# Commit (hooks will validate)
git add .
git commit -m "feat(scope): add new feature"
```

## Database Operations

### Reset database
```bash
rm data/notes.db
# Restart backend - will recreate
```

### Add new table
```go
// backend/internal/db/db.go
migrations := []string{
    // ... existing
    `CREATE TABLE IF NOT EXISTS new_table (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`,
}
```

### Query examples
```go
// Single row
h.db.QueryRow("SELECT * FROM table WHERE id = ?", id).Scan(&fields...)

// Multiple rows
rows, _ := h.db.Query("SELECT * FROM table")
defer rows.Close()
for rows.Next() {
    rows.Scan(&fields...)
}

// Insert
h.db.Exec("INSERT INTO table (id, name) VALUES (?, ?)", id, name)

// Update
h.db.Exec("UPDATE table SET name = ? WHERE id = ?", name, id)

// Delete
h.db.Exec("DELETE FROM table WHERE id = ?", id)
```

## Frontend Patterns

### Fetch data on mount
```svelte
<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/utils/api';
  
  let data = [];
  let loading = true;
  
  onMount(async () => {
    data = await api.getData();
    loading = false;
  });
</script>
```

### Form submission
```svelte
<script>
  let formData = { name: '' };
  
  async function handleSubmit() {
    await api.createItem(formData);
    // Reset or redirect
  }
</script>

<form on:submit|preventDefault={handleSubmit}>
  <input bind:value={formData.name} />
  <button type="submit">Create</button>
</form>
```

### Show toast notification
```svelte
<script>
  import { addToast } from '$lib/stores';
  
  function doSomething() {
    try {
      // action
      addToast('success', 'Done!');
    } catch {
      addToast('error', 'Failed!');
    }
  }
</script>
```

## Docker Operations

```bash
# Build and start
docker-compose up -d --build

# View logs
docker-compose logs -f

# Stop
docker-compose down

# Rebuild single service
docker-compose build backend
docker-compose up -d backend
```

## Git Operations

```bash
# Create feature branch
git checkout -b feat/my-feature

# Commit with conventional format
git commit -m "feat(scope): description"

# Push
git push -u origin feat/my-feature

# Skip hooks (emergency only)
git commit --no-verify -m "message"
```

## Debugging

### Backend
```bash
# Check logs
docker-compose logs backend

# Test endpoint
curl -v http://localhost:8080/api/endpoint

# Check database
sqlite3 data/notes.db
.tables
SELECT * FROM accounts;
```

### Frontend
```bash
# Check browser console
# Check network tab for API calls

# Check logs
docker-compose logs frontend
```

### Hooks
```bash
# Run pre-commit manually
.githooks/pre-commit

# Check hook configuration
git config core.hooksPath
```

## Quick Fixes

### Port already in use
```bash
lsof -ti:8080 | xargs kill  # Backend
lsof -ti:5173 | xargs kill  # Frontend dev
lsof -ti:3000 | xargs kill  # Frontend prod
```

### Node modules issues
```bash
cd frontend
rm -rf node_modules
npm install
```

### Go module issues
```bash
cd backend
go mod tidy
```

### Docker issues
```bash
docker-compose down
docker system prune -f
docker-compose up -d --build
```

## File Locations Quick Reference

| What | Where |
|------|-------|
| API routes | `backend/cmd/server/main.go` |
| API handlers | `backend/internal/handlers/handlers.go` |
| Data models | `backend/internal/models/models.go` |
| Database setup | `backend/internal/db/db.go` |
| Frontend pages | `frontend/src/routes/` |
| API client | `frontend/src/lib/utils/api.ts` |
| Global styles | `frontend/src/app.css` |
| App layout | `frontend/src/routes/+layout.svelte` |
| Git hooks | `.githooks/` |
| Docker config | `docker-compose.yml` |
