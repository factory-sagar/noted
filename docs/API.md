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

### Get Account Activities
```
GET /accounts/:id/activities?limit=50
```

Response:
```json
[
  {
    "id": "uuid",
    "account_id": "account-uuid",
    "type": "note_created",
    "title": "Created note: Discovery Call",
    "description": "Initial meeting with engineering team",
    "entity_type": "note",
    "entity_id": "note-uuid",
    "created_at": "2024-01-15T10:00:00Z"
  }
]
```

### Reorder Notes in Account
```
POST /accounts/:id/notes/reorder
Content-Type: application/json

{
  "note_ids": ["note-uuid-1", "note-uuid-2", "note-uuid-3"]
}
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
    "internal_participants": ["john@acme.com"],
    "external_participants": ["jane@acme.com"],
    "content": "<p>Meeting notes...</p>",
    "meeting_id": "google-calendar-id",
    "meeting_date": "2024-01-15T10:00:00Z",
    "pinned": false,
    "archived": false,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

### Get Notes by Account
```
GET /accounts/:id/notes
```

### Create Note
```
POST /notes
Content-Type: application/json

{
  "title": "Initial Discovery Call",
  "account_id": "account-uuid",
  "template_type": "initial",
  "internal_participants": ["john@acme.com"],
  "external_participants": ["jane@acme.com"],
  "content": "<p>Meeting notes...</p>",
  "meeting_date": "2024-01-15T10:00:00Z"
}
```

### Get Note
```
GET /notes/:id
```

Returns note with linked todos and tags.

### Update Note
```
PUT /notes/:id
Content-Type: application/json

{
  "title": "Updated Title",
  "content": "<p>Updated content...</p>"
}
```

### Delete Note (Soft Delete)
```
DELETE /notes/:id
```

Moves note to trash. Can be restored.

### Restore Note
```
POST /notes/:id/restore
```

Restores note from trash.

### Permanently Delete Note
```
DELETE /notes/:id/permanent
```

Permanently deletes note. Cannot be undone.

### Get Deleted Notes
```
GET /notes/deleted
```

Returns all soft-deleted notes.

### Get Archived Notes
```
GET /notes/archived
```

Returns all archived notes.

### Toggle Note Pin
```
POST /notes/:id/pin
```

Response:
```json
{
  "pinned": true
}
```

### Toggle Note Archive
```
POST /notes/:id/archive
```

Response:
```json
{
  "archived": true
}
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
GET /todos?status=stuck
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
    "account_id": "account-uuid",
    "account_name": "Acme Corp",
    "pinned": false,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "linked_notes": [
      {"id": "note-uuid", "title": "Discovery Call"}
    ]
  }
]
```

**Status values:** `not_started`, `in_progress`, `stuck`, `completed`

**Priority values:** `low`, `medium`, `high`

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
  "note_id": "note-uuid",
  "account_id": "account-uuid"
}
```

### Get Todo
```
GET /todos/:id
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

### Delete Todo (Soft Delete)
```
DELETE /todos/:id
```

### Restore Todo
```
POST /todos/:id/restore
```

### Permanently Delete Todo
```
DELETE /todos/:id/permanent
```

### Get Deleted Todos
```
GET /todos/deleted
```

### Toggle Todo Pin
```
POST /todos/:id/pin
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

## Tags

### List Tags
```
GET /tags
```

Response:
```json
[
  {
    "id": "uuid",
    "name": "Follow-up",
    "color": "#ef4444",
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

### Create Tag
```
POST /tags
Content-Type: application/json

{
  "name": "Follow-up",
  "color": "#ef4444"
}
```

### Update Tag
```
PUT /tags/:id
Content-Type: application/json

{
  "name": "Urgent Follow-up",
  "color": "#dc2626"
}
```

### Delete Tag
```
DELETE /tags/:id
```

### Get Note Tags
```
GET /notes/:id/tags
```

### Add Tag to Note
```
POST /notes/:id/tags/:tagId
```

### Remove Tag from Note
```
DELETE /notes/:id/tags/:tagId
```

---

## Attachments

### List Note Attachments
```
GET /notes/:id/attachments
```

Response:
```json
[
  {
    "id": "uuid",
    "note_id": "note-uuid",
    "filename": "abc123-document.pdf",
    "original_name": "document.pdf",
    "mime_type": "application/pdf",
    "size": 102400,
    "created_at": "2024-01-15T10:00:00Z"
  }
]
```

### Upload Attachment
```
POST /notes/:id/attachments
Content-Type: multipart/form-data

file: <binary>
```

### Delete Attachment
```
DELETE /notes/:id/attachments/:attachmentId
```

### Access Attachment File
```
GET /uploads/:filename
```

Static file serving for uploaded attachments.

---

## Activities

### Get Account Activities
```
GET /accounts/:id/activities?limit=50
```

### Create Activity
```
POST /activities
Content-Type: application/json

{
  "account_id": "account-uuid",
  "type": "note_created",
  "title": "Created note",
  "description": "Optional description",
  "entity_type": "note",
  "entity_id": "note-uuid"
}
```

---

## Quick Capture

### Create Quick Note or Todo
```
POST /quick-capture
Content-Type: application/json

{
  "type": "note",
  "title": "Quick note title",
  "content": "Note content",
  "account_id": "account-uuid"
}
```

Or for todo:
```json
{
  "type": "todo",
  "title": "Quick todo",
  "description": "Todo description",
  "priority": "high",
  "account_id": "account-uuid"
}
```

Response:
```json
{
  "id": "created-uuid",
  "type": "note",
  "title": "Quick note title"
}
```

---

## Search

### Global Search
```
GET /search?q=search+term
```

Searches across notes (using FTS4 with fuzzy matching), accounts, and todos.

Response:
```json
[
  {
    "type": "note",
    "id": "uuid",
    "title": "Discovery Call",
    "snippet": "...matching search term..."
  },
  {
    "type": "account",
    "id": "uuid",
    "title": "Acme Corp"
  },
  {
    "type": "todo",
    "id": "uuid",
    "title": "Send documentation"
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
    "stuck": 5,
    "completed": 15
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

## Calendar (Google OAuth)

### Get Auth URL
```
GET /calendar/auth
```

Response:
```json
{
  "url": "https://accounts.google.com/o/oauth2/auth?..."
}
```

### OAuth Callback
```
GET /calendar/callback?code=...&state=...
```

Handled automatically by OAuth flow.

### Get Calendar Config
```
GET /calendar/config
```

Response:
```json
{
  "connected": true,
  "email": "user@gmail.com"
}
```

### Disconnect Calendar
```
DELETE /calendar/disconnect
```

### List Calendar Events
```
GET /calendar/events?start=2024-01-01&end=2024-01-31
```

Response:
```json
[
  {
    "id": "google-event-id",
    "title": "Team Standup",
    "description": "Daily standup meeting",
    "start_time": "2024-01-15T09:00:00Z",
    "end_time": "2024-01-15T09:30:00Z",
    "attendees": ["john@acme.com", "jane@example.com"],
    "meet_link": "https://meet.google.com/..."
  }
]
```

### Get Single Event
```
GET /calendar/events/:eventId
```

### Parse Participants
```
POST /calendar/parse-participants
Content-Type: application/json

{
  "attendees": ["john@acme.com", "jane@example.com", "bob@example.com"],
  "internal_domain": "example.com"
}
```

Response:
```json
{
  "internal": ["jane@example.com", "bob@example.com"],
  "external": ["john@acme.com"]
}
```

---

## Health Check

### Health
```
GET /health
```

Response:
```json
{
  "status": "ok"
}
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
