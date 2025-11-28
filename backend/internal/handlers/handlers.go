package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/factory-sagar/notes-droid/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler holds database connection and provides HTTP handlers
type Handler struct {
	db *sql.DB
}

// New creates a new Handler
func New(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// --- Account Handlers ---

func (h *Handler) GetAccounts(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, name, account_owner, budget, est_engineers, created_at, updated_at 
		FROM accounts ORDER BY name ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	accounts := []models.Account{}
	for rows.Next() {
		var a models.Account
		if err := rows.Scan(&a.ID, &a.Name, &a.AccountOwner, &a.Budget, &a.EstEngineers, &a.CreatedAt, &a.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		accounts = append(accounts, a)
	}

	c.JSON(http.StatusOK, accounts)
}

func (h *Handler) GetAccount(c *gin.Context) {
	id := c.Param("id")
	var a models.Account
	err := h.db.QueryRow(`
		SELECT id, name, account_owner, budget, est_engineers, created_at, updated_at 
		FROM accounts WHERE id = ?
	`, id).Scan(&a.ID, &a.Name, &a.AccountOwner, &a.Budget, &a.EstEngineers, &a.CreatedAt, &a.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, a)
}

func (h *Handler) CreateAccount(c *gin.Context) {
	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()
	now := time.Now()

	_, err := h.db.Exec(`
		INSERT INTO accounts (id, name, account_owner, budget, est_engineers, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, req.Name, req.AccountOwner, req.Budget, req.EstEngineers, now, now)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.Account{
		ID:           id,
		Name:         req.Name,
		AccountOwner: req.AccountOwner,
		Budget:       req.Budget,
		EstEngineers: req.EstEngineers,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}

func (h *Handler) UpdateAccount(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build dynamic update query
	updates := []string{}
	args := []interface{}{}

	if req.Name != nil {
		updates = append(updates, "name = ?")
		args = append(args, *req.Name)
	}
	if req.AccountOwner != nil {
		updates = append(updates, "account_owner = ?")
		args = append(args, *req.AccountOwner)
	}
	if req.Budget != nil {
		updates = append(updates, "budget = ?")
		args = append(args, *req.Budget)
	}
	if req.EstEngineers != nil {
		updates = append(updates, "est_engineers = ?")
		args = append(args, *req.EstEngineers)
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	updates = append(updates, "updated_at = ?")
	args = append(args, time.Now())
	args = append(args, id)

	query := "UPDATE accounts SET " + strings.Join(updates, ", ") + " WHERE id = ?"
	result, err := h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	// Return updated account
	h.GetAccount(c)
}

func (h *Handler) DeleteAccount(c *gin.Context) {
	id := c.Param("id")
	result, err := h.db.Exec("DELETE FROM accounts WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted"})
}

// --- Note Handlers ---

func (h *Handler) GetNotes(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT n.id, n.title, n.account_id, n.template_type, n.internal_participants, 
			   n.external_participants, n.content, n.meeting_id, n.meeting_date, 
			   n.created_at, n.updated_at, a.name as account_name
		FROM notes n
		LEFT JOIN accounts a ON n.account_id = a.id
		ORDER BY n.created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	notes := []map[string]interface{}{}
	for rows.Next() {
		var n models.Note
		var internalJSON, externalJSON string
		var accountName sql.NullString

		if err := rows.Scan(&n.ID, &n.Title, &n.AccountID, &n.TemplateType, &internalJSON,
			&externalJSON, &n.Content, &n.MeetingID, &n.MeetingDate, &n.CreatedAt, &n.UpdatedAt, &accountName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		json.Unmarshal([]byte(internalJSON), &n.InternalParticipants)
		json.Unmarshal([]byte(externalJSON), &n.ExternalParticipants)

		note := map[string]interface{}{
			"id":                    n.ID,
			"title":                 n.Title,
			"account_id":            n.AccountID,
			"account_name":          accountName.String,
			"template_type":         n.TemplateType,
			"internal_participants": n.InternalParticipants,
			"external_participants": n.ExternalParticipants,
			"content":               n.Content,
			"meeting_id":            n.MeetingID,
			"meeting_date":          n.MeetingDate,
			"created_at":            n.CreatedAt,
			"updated_at":            n.UpdatedAt,
		}
		notes = append(notes, note)
	}

	c.JSON(http.StatusOK, notes)
}

func (h *Handler) GetNote(c *gin.Context) {
	id := c.Param("id")
	var n models.Note
	var internalJSON, externalJSON string
	var accountName sql.NullString

	err := h.db.QueryRow(`
		SELECT n.id, n.title, n.account_id, n.template_type, n.internal_participants, 
			   n.external_participants, n.content, n.meeting_id, n.meeting_date, 
			   n.created_at, n.updated_at, a.name as account_name
		FROM notes n
		LEFT JOIN accounts a ON n.account_id = a.id
		WHERE n.id = ?
	`, id).Scan(&n.ID, &n.Title, &n.AccountID, &n.TemplateType, &internalJSON,
		&externalJSON, &n.Content, &n.MeetingID, &n.MeetingDate, &n.CreatedAt, &n.UpdatedAt, &accountName)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	json.Unmarshal([]byte(internalJSON), &n.InternalParticipants)
	json.Unmarshal([]byte(externalJSON), &n.ExternalParticipants)

	// Get linked todos
	todoRows, err := h.db.Query(`
		SELECT t.id, t.title, t.description, t.status, t.priority, t.due_date, t.created_at, t.updated_at
		FROM todos t
		JOIN note_todos nt ON t.id = nt.todo_id
		WHERE nt.note_id = ?
	`, id)
	if err == nil {
		defer todoRows.Close()
		for todoRows.Next() {
			var todo models.Todo
			todoRows.Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Status, &todo.Priority, &todo.DueDate, &todo.CreatedAt, &todo.UpdatedAt)
			n.Todos = append(n.Todos, todo)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                    n.ID,
		"title":                 n.Title,
		"account_id":            n.AccountID,
		"account_name":          accountName.String,
		"template_type":         n.TemplateType,
		"internal_participants": n.InternalParticipants,
		"external_participants": n.ExternalParticipants,
		"content":               n.Content,
		"meeting_id":            n.MeetingID,
		"meeting_date":          n.MeetingDate,
		"created_at":            n.CreatedAt,
		"updated_at":            n.UpdatedAt,
		"todos":                 n.Todos,
	})
}

func (h *Handler) CreateNote(c *gin.Context) {
	var req models.CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()
	now := time.Now()

	if req.TemplateType == "" {
		req.TemplateType = "initial"
	}

	internalJSON, _ := json.Marshal(req.InternalParticipants)
	externalJSON, _ := json.Marshal(req.ExternalParticipants)

	var meetingDate *time.Time
	if req.MeetingDate != nil {
		parsed, err := time.Parse(time.RFC3339, *req.MeetingDate)
		if err == nil {
			meetingDate = &parsed
		}
	}

	_, err := h.db.Exec(`
		INSERT INTO notes (id, title, account_id, template_type, internal_participants, external_participants, content, meeting_id, meeting_date, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, req.Title, req.AccountID, req.TemplateType, string(internalJSON), string(externalJSON), req.Content, req.MeetingID, meetingDate, now, now)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":                    id,
		"title":                 req.Title,
		"account_id":            req.AccountID,
		"template_type":         req.TemplateType,
		"internal_participants": req.InternalParticipants,
		"external_participants": req.ExternalParticipants,
		"content":               req.Content,
		"meeting_id":            req.MeetingID,
		"meeting_date":          meetingDate,
		"created_at":            now,
		"updated_at":            now,
	})
}

func (h *Handler) UpdateNote(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := []string{}
	args := []interface{}{}

	if req.Title != nil {
		updates = append(updates, "title = ?")
		args = append(args, *req.Title)
	}
	if req.AccountID != nil {
		updates = append(updates, "account_id = ?")
		args = append(args, *req.AccountID)
	}
	if req.TemplateType != nil {
		updates = append(updates, "template_type = ?")
		args = append(args, *req.TemplateType)
	}
	if req.InternalParticipants != nil {
		internalJSON, _ := json.Marshal(req.InternalParticipants)
		updates = append(updates, "internal_participants = ?")
		args = append(args, string(internalJSON))
	}
	if req.ExternalParticipants != nil {
		externalJSON, _ := json.Marshal(req.ExternalParticipants)
		updates = append(updates, "external_participants = ?")
		args = append(args, string(externalJSON))
	}
	if req.Content != nil {
		updates = append(updates, "content = ?")
		args = append(args, *req.Content)
	}
	if req.MeetingID != nil {
		updates = append(updates, "meeting_id = ?")
		args = append(args, *req.MeetingID)
	}
	if req.MeetingDate != nil {
		parsed, _ := time.Parse(time.RFC3339, *req.MeetingDate)
		updates = append(updates, "meeting_date = ?")
		args = append(args, parsed)
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	updates = append(updates, "updated_at = ?")
	args = append(args, time.Now())
	args = append(args, id)

	query := "UPDATE notes SET " + strings.Join(updates, ", ") + " WHERE id = ?"
	result, err := h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	h.GetNote(c)
}

func (h *Handler) DeleteNote(c *gin.Context) {
	id := c.Param("id")
	result, err := h.db.Exec("DELETE FROM notes WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted"})
}

func (h *Handler) GetNotesByAccount(c *gin.Context) {
	accountID := c.Param("id")
	rows, err := h.db.Query(`
		SELECT id, title, account_id, template_type, internal_participants, 
			   external_participants, content, meeting_id, meeting_date, created_at, updated_at
		FROM notes WHERE account_id = ? ORDER BY created_at DESC
	`, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	notes := []models.Note{}
	for rows.Next() {
		var n models.Note
		var internalJSON, externalJSON string
		if err := rows.Scan(&n.ID, &n.Title, &n.AccountID, &n.TemplateType, &internalJSON,
			&externalJSON, &n.Content, &n.MeetingID, &n.MeetingDate, &n.CreatedAt, &n.UpdatedAt); err != nil {
			continue
		}
		json.Unmarshal([]byte(internalJSON), &n.InternalParticipants)
		json.Unmarshal([]byte(externalJSON), &n.ExternalParticipants)
		notes = append(notes, n)
	}

	c.JSON(http.StatusOK, notes)
}

// --- Todo Handlers ---

func (h *Handler) GetTodos(c *gin.Context) {
	status := c.Query("status")
	query := `
		SELECT id, title, description, status, priority, due_date, created_at, updated_at
		FROM todos
	`
	args := []interface{}{}

	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	todos := []map[string]interface{}{}
	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			continue
		}

		// Get linked notes for this todo
		noteRows, _ := h.db.Query(`
			SELECT n.id, n.title FROM notes n
			JOIN note_todos nt ON n.id = nt.note_id
			WHERE nt.todo_id = ?
		`, t.ID)

		linkedNotes := []map[string]string{}
		if noteRows != nil {
			for noteRows.Next() {
				var noteID, noteTitle string
				noteRows.Scan(&noteID, &noteTitle)
				linkedNotes = append(linkedNotes, map[string]string{"id": noteID, "title": noteTitle})
			}
			noteRows.Close()
		}

		todos = append(todos, map[string]interface{}{
			"id":           t.ID,
			"title":        t.Title,
			"description":  t.Description,
			"status":       t.Status,
			"priority":     t.Priority,
			"due_date":     t.DueDate,
			"created_at":   t.CreatedAt,
			"updated_at":   t.UpdatedAt,
			"linked_notes": linkedNotes,
		})
	}

	c.JSON(http.StatusOK, todos)
}

func (h *Handler) GetTodo(c *gin.Context) {
	id := c.Param("id")
	var t models.Todo
	err := h.db.QueryRow(`
		SELECT id, title, description, status, priority, due_date, created_at, updated_at
		FROM todos WHERE id = ?
	`, id).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get linked notes
	noteRows, _ := h.db.Query(`
		SELECT n.id, n.title FROM notes n
		JOIN note_todos nt ON n.id = nt.note_id
		WHERE nt.todo_id = ?
	`, id)

	linkedNotes := []map[string]string{}
	if noteRows != nil {
		defer noteRows.Close()
		for noteRows.Next() {
			var noteID, noteTitle string
			noteRows.Scan(&noteID, &noteTitle)
			linkedNotes = append(linkedNotes, map[string]string{"id": noteID, "title": noteTitle})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           t.ID,
		"title":        t.Title,
		"description":  t.Description,
		"status":       t.Status,
		"priority":     t.Priority,
		"due_date":     t.DueDate,
		"created_at":   t.CreatedAt,
		"updated_at":   t.UpdatedAt,
		"linked_notes": linkedNotes,
	})
}

func (h *Handler) CreateTodo(c *gin.Context) {
	var req models.CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()
	now := time.Now()

	if req.Status == "" {
		req.Status = "not_started"
	}
	if req.Priority == "" {
		req.Priority = "medium"
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		parsed, err := time.Parse(time.RFC3339, *req.DueDate)
		if err == nil {
			dueDate = &parsed
		}
	}

	_, err := h.db.Exec(`
		INSERT INTO todos (id, title, description, status, priority, due_date, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, id, req.Title, req.Description, req.Status, req.Priority, dueDate, now, now)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Link to note if provided
	if req.NoteID != nil {
		h.db.Exec("INSERT INTO note_todos (note_id, todo_id) VALUES (?, ?)", *req.NoteID, id)
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          id,
		"title":       req.Title,
		"description": req.Description,
		"status":      req.Status,
		"priority":    req.Priority,
		"due_date":    dueDate,
		"created_at":  now,
		"updated_at":  now,
	})
}

func (h *Handler) UpdateTodo(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := []string{}
	args := []interface{}{}

	if req.Title != nil {
		updates = append(updates, "title = ?")
		args = append(args, *req.Title)
	}
	if req.Description != nil {
		updates = append(updates, "description = ?")
		args = append(args, *req.Description)
	}
	if req.Status != nil {
		updates = append(updates, "status = ?")
		args = append(args, *req.Status)
	}
	if req.Priority != nil {
		updates = append(updates, "priority = ?")
		args = append(args, *req.Priority)
	}
	if req.DueDate != nil {
		parsed, _ := time.Parse(time.RFC3339, *req.DueDate)
		updates = append(updates, "due_date = ?")
		args = append(args, parsed)
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	updates = append(updates, "updated_at = ?")
	args = append(args, time.Now())
	args = append(args, id)

	query := "UPDATE todos SET " + strings.Join(updates, ", ") + " WHERE id = ?"
	result, err := h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	h.GetTodo(c)
}

func (h *Handler) DeleteTodo(c *gin.Context) {
	id := c.Param("id")
	result, err := h.db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted"})
}

func (h *Handler) LinkTodoToNote(c *gin.Context) {
	todoID := c.Param("id")
	noteID := c.Param("noteId")

	_, err := h.db.Exec("INSERT OR IGNORE INTO note_todos (note_id, todo_id) VALUES (?, ?)", noteID, todoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo linked to note"})
}

func (h *Handler) UnlinkTodoFromNote(c *gin.Context) {
	todoID := c.Param("id")
	noteID := c.Param("noteId")

	_, err := h.db.Exec("DELETE FROM note_todos WHERE note_id = ? AND todo_id = ?", noteID, todoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo unlinked from note"})
}

// --- Search Handler ---

func (h *Handler) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	results := []models.SearchResult{}

	// Search notes using FTS4
	noteRows, err := h.db.Query(`
		SELECT n.id, n.title, snippet(notes_fts, '<mark>', '</mark>', '...') as snippet
		FROM notes_fts
		JOIN notes n ON notes_fts.docid = n.rowid
		WHERE notes_fts MATCH ?
		LIMIT 20
	`, q)
	if err == nil {
		defer noteRows.Close()
		for noteRows.Next() {
			var id, title, snippet string
			noteRows.Scan(&id, &title, &snippet)
			results = append(results, models.SearchResult{
				Type:    "note",
				ID:      id,
				Title:   title,
				Snippet: snippet,
			})
		}
	}

	// Search accounts by name
	accountRows, err := h.db.Query(`
		SELECT id, name FROM accounts WHERE name LIKE ? LIMIT 10
	`, "%"+q+"%")
	if err == nil {
		defer accountRows.Close()
		for accountRows.Next() {
			var id, name string
			accountRows.Scan(&id, &name)
			results = append(results, models.SearchResult{
				Type:  "account",
				ID:    id,
				Title: name,
			})
		}
	}

	// Search todos by title
	todoRows, err := h.db.Query(`
		SELECT id, title FROM todos WHERE title LIKE ? LIMIT 10
	`, "%"+q+"%")
	if err == nil {
		defer todoRows.Close()
		for todoRows.Next() {
			var id, title string
			todoRows.Scan(&id, &title)
			results = append(results, models.SearchResult{
				Type:  "todo",
				ID:    id,
				Title: title,
			})
		}
	}

	c.JSON(http.StatusOK, results)
}

// --- Analytics Handlers ---

func (h *Handler) GetAnalytics(c *gin.Context) {
	var analytics models.Analytics

	// Total counts
	h.db.QueryRow("SELECT COUNT(*) FROM notes").Scan(&analytics.TotalNotes)
	h.db.QueryRow("SELECT COUNT(*) FROM accounts").Scan(&analytics.TotalAccounts)
	h.db.QueryRow("SELECT COUNT(*) FROM todos").Scan(&analytics.TotalTodos)

	// Todos by status
	analytics.TodosByStatus = map[string]int{}
	statusRows, _ := h.db.Query("SELECT status, COUNT(*) FROM todos GROUP BY status")
	if statusRows != nil {
		defer statusRows.Close()
		for statusRows.Next() {
			var status string
			var count int
			statusRows.Scan(&status, &count)
			analytics.TodosByStatus[status] = count
		}
	}

	// Notes by account
	accountRows, _ := h.db.Query(`
		SELECT a.id, a.name, COUNT(n.id) as note_count
		FROM accounts a
		LEFT JOIN notes n ON a.id = n.account_id
		GROUP BY a.id
		ORDER BY note_count DESC
	`)
	if accountRows != nil {
		defer accountRows.Close()
		for accountRows.Next() {
			var anc models.AccountNoteCount
			accountRows.Scan(&anc.AccountID, &anc.AccountName, &anc.NoteCount)
			analytics.NotesByAccount = append(analytics.NotesByAccount, anc)
		}
	}

	// Count notes with incomplete fields
	h.db.QueryRow(`
		SELECT COUNT(*) FROM notes n
		JOIN accounts a ON n.account_id = a.id
		WHERE a.budget IS NULL OR a.est_engineers IS NULL OR a.account_owner = ''
		   OR n.content = '' OR n.internal_participants = '[]'
	`).Scan(&analytics.IncompleteCount)

	c.JSON(http.StatusOK, analytics)
}

func (h *Handler) GetIncompleteFields(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT n.id, n.title, a.name as account_name, 
			   a.budget, a.est_engineers, a.account_owner,
			   n.content, n.internal_participants
		FROM notes n
		JOIN accounts a ON n.account_id = a.id
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	incomplete := []models.IncompleteField{}
	for rows.Next() {
		var noteID, noteTitle, accountName string
		var budget sql.NullFloat64
		var estEngineers sql.NullInt64
		var accountOwner, content, internalParticipants string

		rows.Scan(&noteID, &noteTitle, &accountName, &budget, &estEngineers, &accountOwner, &content, &internalParticipants)

		missing := []string{}
		if !budget.Valid {
			missing = append(missing, "budget")
		}
		if !estEngineers.Valid {
			missing = append(missing, "est_engineers")
		}
		if accountOwner == "" {
			missing = append(missing, "account_owner")
		}
		if content == "" {
			missing = append(missing, "content")
		}
		if internalParticipants == "[]" {
			missing = append(missing, "internal_participants")
		}

		if len(missing) > 0 {
			incomplete = append(incomplete, models.IncompleteField{
				NoteID:        noteID,
				NoteTitle:     noteTitle,
				AccountName:   accountName,
				MissingFields: missing,
			})
		}
	}

	c.JSON(http.StatusOK, incomplete)
}

// --- PDF Export Handler ---

func (h *Handler) ExportNotePDF(c *gin.Context) {
	// PDF export will be implemented with a proper library
	// For now, return the note data that can be used for client-side PDF generation
	id := c.Param("id")
	exportType := c.DefaultQuery("type", "full") // "full" or "minimal"

	var n models.Note
	var internalJSON, externalJSON string
	var accountName, accountOwner string
	var budget sql.NullFloat64
	var estEngineers sql.NullInt64

	err := h.db.QueryRow(`
		SELECT n.id, n.title, n.account_id, n.template_type, n.internal_participants, 
			   n.external_participants, n.content, n.meeting_id, n.meeting_date, 
			   n.created_at, n.updated_at, a.name, a.account_owner, a.budget, a.est_engineers
		FROM notes n
		LEFT JOIN accounts a ON n.account_id = a.id
		WHERE n.id = ?
	`, id).Scan(&n.ID, &n.Title, &n.AccountID, &n.TemplateType, &internalJSON,
		&externalJSON, &n.Content, &n.MeetingID, &n.MeetingDate, &n.CreatedAt, &n.UpdatedAt,
		&accountName, &accountOwner, &budget, &estEngineers)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	json.Unmarshal([]byte(internalJSON), &n.InternalParticipants)
	json.Unmarshal([]byte(externalJSON), &n.ExternalParticipants)

	response := gin.H{
		"id":           n.ID,
		"title":        n.Title,
		"content":      n.Content,
		"account_name": accountName,
		"meeting_date": n.MeetingDate,
		"export_type":  exportType,
	}

	if exportType == "full" {
		// Get linked todos
		todoRows, _ := h.db.Query(`
			SELECT t.id, t.title, t.status FROM todos t
			JOIN note_todos nt ON t.id = nt.todo_id
			WHERE nt.note_id = ?
		`, id)

		todos := []map[string]string{}
		if todoRows != nil {
			defer todoRows.Close()
			for todoRows.Next() {
				var todoID, todoTitle, todoStatus string
				todoRows.Scan(&todoID, &todoTitle, &todoStatus)
				todos = append(todos, map[string]string{"id": todoID, "title": todoTitle, "status": todoStatus})
			}
		}

		response["account_owner"] = accountOwner
		response["budget"] = budget.Float64
		response["est_engineers"] = estEngineers.Int64
		response["internal_participants"] = n.InternalParticipants
		response["external_participants"] = n.ExternalParticipants
		response["todos"] = todos
	}

	c.JSON(http.StatusOK, response)
}
