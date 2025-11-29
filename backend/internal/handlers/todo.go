package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/factory-sagar/notes-droid/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) GetTodos(c *gin.Context) {
	status := c.Query("status")
	query := `
		SELECT t.id, t.title, t.description, t.status, t.priority, t.due_date, t.account_id, 
		       COALESCE(a.name, '') as account_name, t.created_at, t.updated_at
		FROM todos t
		LEFT JOIN accounts a ON t.account_id = a.id
		WHERE t.deleted_at IS NULL
	`
	args := []interface{}{}

	if status != "" {
		query += " AND t.status = ?"
		args = append(args, status)
	}
	query += " ORDER BY t.created_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	todos := []map[string]interface{}{}
	var todoIDs []interface{}

	for rows.Next() {
		var t models.Todo
		var accountName string
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.AccountID, &accountName, &t.CreatedAt, &t.UpdatedAt); err != nil {
			continue
		}

		todos = append(todos, map[string]interface{}{
			"id":           t.ID,
			"title":        t.Title,
			"description":  t.Description,
			"status":       t.Status,
			"priority":     t.Priority,
			"due_date":     t.DueDate,
			"account_id":   t.AccountID,
			"account_name": accountName,
			"created_at":   t.CreatedAt,
			"updated_at":   t.UpdatedAt,
			"linked_notes": []map[string]string{},
		})
		todoIDs = append(todoIDs, t.ID)
	}

	if len(todos) == 0 {
		c.JSON(http.StatusOK, todos)
		return
	}

	// Batch fetch linked notes
	placeholders := strings.Repeat("?,", len(todoIDs))
	placeholders = placeholders[:len(placeholders)-1]

	// Correctly cast todoIDs to []interface{} for Query
	queryArgs := make([]interface{}, len(todoIDs))
	for i, v := range todoIDs {
		queryArgs[i] = v
	}

	notesQuery := `
		SELECT nt.todo_id, n.id, n.title
		FROM note_todos nt
		JOIN notes n ON nt.note_id = n.id
		WHERE nt.todo_id IN (` + placeholders + `)
	`

	noteRows, err := h.db.Query(notesQuery, queryArgs...)
	if err != nil {
		// Return todos even if linked notes fail
		c.JSON(http.StatusOK, todos)
		return
	}
	defer noteRows.Close()

	linkedNotesMap := make(map[string][]map[string]string)
	for noteRows.Next() {
		var todoID, noteID, noteTitle string
		if err := noteRows.Scan(&todoID, &noteID, &noteTitle); err != nil {
			continue
		}
		if _, ok := linkedNotesMap[todoID]; !ok {
			linkedNotesMap[todoID] = []map[string]string{}
		}
		linkedNotesMap[todoID] = append(linkedNotesMap[todoID], map[string]string{"id": noteID, "title": noteTitle})
	}

	// Attach linked notes to todos
	for i, todo := range todos {
		id := todo["id"].(string)
		if notes, ok := linkedNotesMap[id]; ok {
			todos[i]["linked_notes"] = notes
		}
	}

	c.JSON(http.StatusOK, todos)
}

func (h *Handler) GetTodo(c *gin.Context) {
	id := c.Param("id")
	var t models.Todo
	var accountName string
	err := h.db.QueryRow(`
		SELECT t.id, t.title, t.description, t.status, t.priority, t.due_date, t.account_id,
		       COALESCE(a.name, '') as account_name, t.created_at, t.updated_at
		FROM todos t
		LEFT JOIN accounts a ON t.account_id = a.id
		WHERE t.id = ?
	`, id).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.AccountID, &accountName, &t.CreatedAt, &t.UpdatedAt)

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
		"account_id":   t.AccountID,
		"account_name": accountName,
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

	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Todo title is required"})
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
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid due date format"})
			return
		}
		dueDate = &parsed
	}

	_, err := h.db.Exec(`
		INSERT INTO todos (id, title, description, status, priority, due_date, account_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, req.Title, req.Description, req.Status, req.Priority, dueDate, req.AccountID, now, now)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Link to note if provided
	if req.NoteID != nil {
		h.db.Exec("INSERT INTO note_todos (note_id, todo_id) VALUES (?, ?)", *req.NoteID, id)
	}

	// Get account name if account_id was provided
	var accountName string
	if req.AccountID != nil {
		h.db.QueryRow("SELECT name FROM accounts WHERE id = ?", *req.AccountID).Scan(&accountName)
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":           id,
		"title":        req.Title,
		"description":  req.Description,
		"status":       req.Status,
		"priority":     req.Priority,
		"due_date":     dueDate,
		"account_id":   req.AccountID,
		"account_name": accountName,
		"created_at":   now,
		"updated_at":   now,
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
		parsed, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid due date format"})
			return
		}
		updates = append(updates, "due_date = ?")
		args = append(args, parsed)
	}
	if req.AccountID != nil {
		updates = append(updates, "account_id = ?")
		args = append(args, *req.AccountID)
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
	// Soft delete - set deleted_at timestamp
	result, err := h.db.Exec("UPDATE todos SET deleted_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL", id)
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

func (h *Handler) RestoreTodo(c *gin.Context) {
	id := c.Param("id")
	result, err := h.db.Exec("UPDATE todos SET deleted_at = NULL WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Todo restored"})
}

func (h *Handler) PermanentDeleteTodo(c *gin.Context) {
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
	c.JSON(http.StatusOK, gin.H{"message": "Todo permanently deleted"})
}

func (h *Handler) GetDeletedTodos(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT t.id, t.title, t.description, t.status, t.priority, t.due_date, 
			   t.account_id, COALESCE(a.name, '') as account_name, t.created_at, t.updated_at, t.deleted_at
		FROM todos t
		LEFT JOIN accounts a ON t.account_id = a.id
		WHERE t.deleted_at IS NOT NULL
		ORDER BY t.deleted_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	todos := []map[string]interface{}{}
	for rows.Next() {
		var id, title, description, status, priority, accountName string
		var dueDate, accountID, deletedAt sql.NullString
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &title, &description, &status, &priority, &dueDate, &accountID, &accountName, &createdAt, &updatedAt, &deletedAt); err != nil {
			continue
		}
		todo := map[string]interface{}{
			"id":           id,
			"title":        title,
			"description":  description,
			"status":       status,
			"priority":     priority,
			"account_name": accountName,
			"created_at":   createdAt,
			"updated_at":   updatedAt,
			"deleted_at":   deletedAt.String,
		}
		if dueDate.Valid {
			todo["due_date"] = dueDate.String
		}
		if accountID.Valid {
			todo["account_id"] = accountID.String
		}
		todos = append(todos, todo)
	}
	c.JSON(http.StatusOK, todos)
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

func (h *Handler) ToggleTodoPin(c *gin.Context) {
	id := c.Param("id")

	var pinned int
	err := h.db.QueryRow("SELECT pinned FROM todos WHERE id = ?", id).Scan(&pinned)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	newPinned := 1
	if pinned == 1 {
		newPinned = 0
	}

	h.db.Exec("UPDATE todos SET pinned = ?, updated_at = ? WHERE id = ?", newPinned, time.Now(), id)

	c.JSON(http.StatusOK, gin.H{"pinned": newPinned == 1})
}
