# Skill: Backend Development

## Overview
This skill covers working with the Go backend for SE Notes.

## Tech Stack
- Go 1.21+
- Gin web framework
- SQLite with FTS4
- CGO enabled (required for SQLite)

## Project Structure
```
backend/
├── cmd/server/main.go      # Entry point, router setup, middleware
├── internal/
│   ├── handlers/handlers.go # All HTTP handlers
│   ├── models/models.go     # Data models, request/response types
│   └── db/db.go            # Database connection, migrations
├── go.mod                   # Dependencies
└── Dockerfile              # Production build
```

## Running the Backend

```bash
cd backend

# Development
go run ./cmd/server/

# Build binary
go build -o server ./cmd/server/

# Run binary
./server
```

## Adding a New Endpoint

### 1. Define the model (if needed)
```go
// internal/models/models.go
type MyNewThing struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}

type CreateMyNewThingRequest struct {
    Name string `json:"name" binding:"required"`
}
```

### 2. Add database table (if needed)
```go
// internal/db/db.go - in Migrate function
migrations := []string{
    // ... existing migrations
    `CREATE TABLE IF NOT EXISTS my_new_things (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`,
}
```

### 3. Create handler
```go
// internal/handlers/handlers.go
func (h *Handler) GetMyNewThings(c *gin.Context) {
    rows, err := h.db.Query("SELECT id, name, created_at FROM my_new_things")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()
    
    things := []models.MyNewThing{}
    for rows.Next() {
        var t models.MyNewThing
        rows.Scan(&t.ID, &t.Name, &t.CreatedAt)
        things = append(things, t)
    }
    c.JSON(http.StatusOK, things)
}

func (h *Handler) CreateMyNewThing(c *gin.Context) {
    var req models.CreateMyNewThingRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    id := uuid.New().String()
    now := time.Now()
    
    _, err := h.db.Exec(
        "INSERT INTO my_new_things (id, name, created_at) VALUES (?, ?, ?)",
        id, req.Name, now,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, models.MyNewThing{
        ID: id, Name: req.Name, CreatedAt: now,
    })
}
```

### 4. Register routes
```go
// cmd/server/main.go - in main function
api := router.Group("/api")
{
    // ... existing routes
    api.GET("/my-new-things", h.GetMyNewThings)
    api.POST("/my-new-things", h.CreateMyNewThing)
}
```

## Database Patterns

### Query single row
```go
var item Model
err := h.db.QueryRow("SELECT ... WHERE id = ?", id).Scan(&item.Field1, &item.Field2)
if err == sql.ErrNoRows {
    c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
    return
}
```

### Query multiple rows
```go
rows, err := h.db.Query("SELECT ... FROM table")
if err != nil { /* handle */ }
defer rows.Close()

items := []Model{}
for rows.Next() {
    var item Model
    rows.Scan(&item.Field1, &item.Field2)
    items = append(items, item)
}
```

### Dynamic updates
```go
updates := []string{}
args := []interface{}{}

if req.Name != nil {
    updates = append(updates, "name = ?")
    args = append(args, *req.Name)
}
// ... more fields

updates = append(updates, "updated_at = ?")
args = append(args, time.Now())
args = append(args, id)

query := "UPDATE table SET " + strings.Join(updates, ", ") + " WHERE id = ?"
result, err := h.db.Exec(query, args...)
```

## Full-Text Search (FTS4)

The notes table has FTS4 indexing:

```go
// Search query
rows, err := h.db.Query(`
    SELECT n.id, n.title, snippet(notes_fts, '<mark>', '</mark>', '...') as snippet
    FROM notes_fts
    JOIN notes n ON notes_fts.docid = n.rowid
    WHERE notes_fts MATCH ?
    LIMIT 20
`, searchTerm)
```

## Testing API

```bash
# Health check
curl http://localhost:8080/health

# Create account
curl -X POST http://localhost:8080/api/accounts \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Corp"}'

# List accounts
curl http://localhost:8080/api/accounts
```

## Common Issues

### CGO errors
SQLite requires CGO. Ensure `CGO_ENABLED=1` when building:
```bash
CGO_ENABLED=1 go build -o server ./cmd/server/
```

### FTS module not found
Default macOS SQLite may not have FTS5. Use FTS4 instead.

### Port already in use
```bash
lsof -ti:8080 | xargs kill
```
