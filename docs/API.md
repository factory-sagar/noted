# API Documentation

Base URL: `http://localhost:8080/api`

## Accounts

### List Accounts
```
GET /accounts
```

Response:
```json
[
  {
    "id": "uuid",
    "name": "Acme Corp",
    "account_owner": "John Sales",
    "budget": 50000,
    "est_engineers": 5,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

### Create Account
```
POST /accounts
Content-Type: application/json

{
  "name": "Acme Corp",
  "account_owner": "John Sales",
  "budget": 50000,
  "est_engineers": 5
}
```

### Get Account
```
GET /accounts/:id
```

### Update Account
```
PUT /accounts/:id
Content-Type: application/json

{
  "name": "Acme Corporation",
  "budget": 75000
}
```

### Delete Account
```
DELETE /accounts/:id
```

---

## Notes

### List Notes
```
GET /notes
```

Response:
```json
[
  {
    "id": "uuid",
    "title": "Initial Discovery Call",
    "account_id": "account-uuid",
    "account_name": "Acme Corp",
    "template_type": "initial",
    "internal_participants": ["john@factory.ai"],
    "external_participants": ["jane@acme.com"],
    "content": "<p>Meeting notes...</p>",
    "meeting_id": "google-calendar-id",
    "meeting_date": "2024-01-15T10:00:00Z",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

### Create Note
```
POST /notes
Content-Type: application/json

{
  "title": "Initial Discovery Call",
  "account_id": "account-uuid",
  "template_type": "initial",
  "internal_participants": ["john@factory.ai"],
  "external_participants": ["jane@acme.com"],
  "content": "<p>Meeting notes...</p>",
  "meeting_date": "2024-01-15T10:00:00Z"
}
```

### Get Note
```
GET /notes/:id
```

Returns note with linked todos.

### Update Note
```
PUT /notes/:id
Content-Type: application/json

{
  "title": "Updated Title",
  "content": "<p>Updated content...</p>"
}
```

### Delete Note
```
DELETE /notes/:id
```

### Get Notes by Account
```
GET /accounts/:id/notes
```

### Export Note
```
GET /notes/:id/export?type=full
GET /notes/:id/export?type=minimal
```

`full` - Includes all metadata, participants, and linked todos
`minimal` - Only note content and account name

---

## Todos

### List Todos
```
GET /todos
GET /todos?status=not_started
GET /todos?status=in_progress
GET /todos?status=completed
```

Response:
```json
[
  {
    "id": "uuid",
    "title": "Send documentation",
    "description": "Technical overview docs",
    "status": "not_started",
    "priority": "high",
    "due_date": "2024-01-20T00:00:00Z",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "linked_notes": [
      {"id": "note-uuid", "title": "Discovery Call"}
    ]
  }
]
```

### Create Todo
```
POST /todos
Content-Type: application/json

{
  "title": "Send documentation",
  "description": "Technical overview docs",
  "status": "not_started",
  "priority": "high",
  "due_date": "2024-01-20T00:00:00Z",
  "note_id": "note-uuid"  // Optional: link to note on creation
}
```

### Update Todo
```
PUT /todos/:id
Content-Type: application/json

{
  "status": "in_progress",
  "priority": "medium"
}
```

### Delete Todo
```
DELETE /todos/:id
```

### Link Todo to Note
```
POST /todos/:id/notes/:noteId
```

### Unlink Todo from Note
```
DELETE /todos/:id/notes/:noteId
```

---

## Search

### Global Search
```
GET /search?q=search+term
```

Searches across notes (using FTS5), accounts, and todos.

Response:
```json
[
  {
    "type": "note",
    "id": "uuid",
    "title": "Discovery Call",
    "snippet": "...matching <mark>search term</mark>..."
  },
  {
    "type": "account",
    "id": "uuid",
    "title": "Acme Corp"
  }
]
```

---

## Analytics

### Get Analytics
```
GET /analytics
```

Response:
```json
{
  "total_notes": 25,
  "total_accounts": 10,
  "total_todos": 45,
  "todos_by_status": {
    "not_started": 15,
    "in_progress": 10,
    "completed": 20
  },
  "notes_by_account": [
    {"account_id": "uuid", "account_name": "Acme Corp", "note_count": 5}
  ],
  "incomplete_count": 3
}
```

### Get Incomplete Fields
```
GET /analytics/incomplete
```

Returns notes with missing required fields.

Response:
```json
[
  {
    "note_id": "uuid",
    "note_title": "Discovery Call",
    "account_name": "Acme Corp",
    "missing_fields": ["budget", "est_engineers"]
  }
]
```

---

## Error Responses

All errors return JSON:
```json
{
  "error": "Error message here"
}
```

HTTP Status Codes:
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `404` - Not Found
- `500` - Internal Server Error
