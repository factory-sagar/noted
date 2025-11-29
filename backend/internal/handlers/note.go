package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/factory-sagar/notes-droid/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) GetNotes(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT n.id, n.title, n.account_id, n.template_type, n.internal_participants, 
			   n.external_participants, n.content, n.meeting_id, n.meeting_date, 
			   n.created_at, n.updated_at, a.name as account_name
		FROM notes n
		LEFT JOIN accounts a ON n.account_id = a.id
		WHERE n.deleted_at IS NULL
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

		if err := json.Unmarshal([]byte(internalJSON), &n.InternalParticipants); err != nil {
			log.Printf("Error unmarshalling internal participants for note %s: %v", n.ID, err)
		}
		if err := json.Unmarshal([]byte(externalJSON), &n.ExternalParticipants); err != nil {
			log.Printf("Error unmarshalling external participants for note %s: %v", n.ID, err)
		}

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
		// Log the error for debugging
		log.Printf("GetNote error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := json.Unmarshal([]byte(internalJSON), &n.InternalParticipants); err != nil {
		// log.Printf("Error unmarshalling internal participants for note %s: %v", n.ID, err)
	}
	if err := json.Unmarshal([]byte(externalJSON), &n.ExternalParticipants); err != nil {
		// log.Printf("Error unmarshalling external participants for note %s: %v", n.ID, err)
	}

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

	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Note title is required"})
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
		if err != nil {
			// Try parsing apple format (no colon in offset)
			parsed, err = time.Parse("2006-01-02T15:04:05-0700", *req.MeetingDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meeting date format"})
				return
			}
		}
		meetingDate = &parsed
	}

	_, err := h.db.Exec(`
		INSERT INTO notes (id, title, account_id, template_type, internal_participants, external_participants, content, meeting_id, meeting_date, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, req.Title, req.AccountID, req.TemplateType, string(internalJSON), string(externalJSON), req.Content, req.MeetingID, meetingDate, now, now)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Auto-extract contacts from participants
	go h.ExtractContactsFromNote(req.InternalParticipants, req.ExternalParticipants)

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
		parsed, err := time.Parse(time.RFC3339, *req.MeetingDate)
		if err != nil {
			// Try parsing apple format
			parsed, err = time.Parse("2006-01-02T15:04:05-0700", *req.MeetingDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meeting date format"})
				return
			}
		}
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
	// Soft delete - set deleted_at timestamp
	result, err := h.db.Exec("UPDATE notes SET deleted_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL", id)
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

func (h *Handler) RestoreNote(c *gin.Context) {
	id := c.Param("id")
	result, err := h.db.Exec("UPDATE notes SET deleted_at = NULL WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Note restored"})
}

func (h *Handler) PermanentDeleteNote(c *gin.Context) {
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
	c.JSON(http.StatusOK, gin.H{"message": "Note permanently deleted"})
}

func (h *Handler) GetDeletedNotes(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT n.id, n.title, n.account_id, n.template_type, n.created_at, n.updated_at, 
		       n.deleted_at, COALESCE(a.name, '') as account_name
		FROM notes n
		LEFT JOIN accounts a ON n.account_id = a.id
		WHERE n.deleted_at IS NOT NULL
		ORDER BY n.deleted_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	notes := []map[string]interface{}{}
	for rows.Next() {
		var id, title, accountID, templateType, accountName string
		var deletedAt sql.NullString
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &title, &accountID, &templateType, &createdAt, &updatedAt, &deletedAt, &accountName); err != nil {
			continue
		}
		notes = append(notes, map[string]interface{}{
			"id":            id,
			"title":         title,
			"account_id":    accountID,
			"account_name":  accountName,
			"template_type": templateType,
			"created_at":    createdAt,
			"updated_at":    updatedAt,
			"deleted_at":    deletedAt.String,
		})
	}
	c.JSON(http.StatusOK, notes)
}

func (h *Handler) GetNotesByAccount(c *gin.Context) {
	accountID := c.Param("id")
	rows, err := h.db.Query(`
		SELECT id, title, account_id, template_type, internal_participants, 
			   external_participants, content, meeting_id, meeting_date, created_at, updated_at
		FROM notes WHERE account_id = ? AND deleted_at IS NULL ORDER BY created_at DESC
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
		if err := json.Unmarshal([]byte(internalJSON), &n.InternalParticipants); err != nil {
			log.Printf("Error unmarshalling internal participants for note %s: %v", n.ID, err)
		}
		if err := json.Unmarshal([]byte(externalJSON), &n.ExternalParticipants); err != nil {
			log.Printf("Error unmarshalling external participants for note %s: %v", n.ID, err)
		}
		notes = append(notes, n)
	}

	c.JSON(http.StatusOK, notes)
}

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

func (h *Handler) ImportMarkdown(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer src.Close()

	// Read content
	contentBytes := make([]byte, file.Size)
	if _, err := src.Read(contentBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}
	content := string(contentBytes)

	// Basic Markdown parsing (very simple for now)
	// Assume Title is first line # Title, or filename if not present
	lines := strings.Split(content, "\n")
	title := strings.TrimSuffix(file.Filename, ".md")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "# ") {
		title = strings.TrimPrefix(lines[0], "# ")
		// Remove title from content if we extracted it
		content = strings.Join(lines[1:], "\n")
	}

	// Convert Markdown to HTML (simple conversion or store raw)
	// For this app, the editor expects HTML. 
	// Since we don't have a full MD->HTML parser here, we'll just wrap paragraphs.
	// Ideally, use a library like blackfriday or goldmark.
	// For now, let's just store it as raw text wrapped in p tags to avoid breaking the editor completely
	// Or better, just treat it as raw content if the editor supports it.
	// TipTap can handle some HTML. Let's just replace newlines with <br> or wrap in <p>
	
	// Simple naive conversion for lines
	var htmlBuilder strings.Builder
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "# ") {
			// Skip title if we already extracted it, or convert to h1
			continue 
		}
		if strings.HasPrefix(line, "## ") {
			htmlBuilder.WriteString("<h2>" + strings.TrimPrefix(line, "## ") + "</h2>")
		} else if strings.HasPrefix(line, "- ") {
			htmlBuilder.WriteString("<ul><li>" + strings.TrimPrefix(line, "- ") + "</li></ul>")
		} else {
			htmlBuilder.WriteString("<p>" + line + "</p>")
		}
	}
	htmlContent := htmlBuilder.String()

	id := uuid.New().String()
	now := time.Now()

	// Create note
	_, err = h.db.Exec(`
		INSERT INTO notes (id, title, content, template_type, created_at, updated_at)
		VALUES (?, ?, ?, 'imported', ?, ?)
	`, id, title, htmlContent, now, now)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    id,
		"title": title,
	})
}

func (h *Handler) ExportMarkdown(c *gin.Context) {
	id := c.Param("id")
	var n models.Note
	err := h.db.QueryRow("SELECT title, content FROM notes WHERE id = ?", id).Scan(&n.Title, &n.Content)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	// Simple HTML to Markdown conversion (naive)
	// Replace tags with MD equivalents
	md := n.Content
	md = strings.ReplaceAll(md, "<h1>", "# ")
	md = strings.ReplaceAll(md, "</h1>", "\n\n")
	md = strings.ReplaceAll(md, "<h2>", "## ")
	md = strings.ReplaceAll(md, "</h2>", "\n\n")
	md = strings.ReplaceAll(md, "<h3>", "### ")
	md = strings.ReplaceAll(md, "</h3>", "\n\n")
	md = strings.ReplaceAll(md, "<p>", "")
	md = strings.ReplaceAll(md, "</p>", "\n\n")
	md = strings.ReplaceAll(md, "<b>", "**")
	md = strings.ReplaceAll(md, "</b>", "**")
	md = strings.ReplaceAll(md, "<strong>", "**")
	md = strings.ReplaceAll(md, "</strong>", "**")
	md = strings.ReplaceAll(md, "<i>", "*")
	md = strings.ReplaceAll(md, "</i>", "*")
	md = strings.ReplaceAll(md, "<em>", "*")
	md = strings.ReplaceAll(md, "</em>", "*")
	md = strings.ReplaceAll(md, "<ul>", "")
	md = strings.ReplaceAll(md, "</ul>", "")
	md = strings.ReplaceAll(md, "<li>", "- ")
	md = strings.ReplaceAll(md, "</li>", "\n")
	md = strings.ReplaceAll(md, "<br>", "\n")
	
	// Add title at top
	finalMD := "# " + n.Title + "\n\n" + md

	c.Header("Content-Disposition", "attachment; filename="+n.Title+".md")
	c.Data(http.StatusOK, "text/markdown", []byte(finalMD))
}

func (h *Handler) ToggleNotePin(c *gin.Context) {
	id := c.Param("id")

	var pinned int
	err := h.db.QueryRow("SELECT pinned FROM notes WHERE id = ?", id).Scan(&pinned)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	newPinned := 1
	if pinned == 1 {
		newPinned = 0
	}

	h.db.Exec("UPDATE notes SET pinned = ?, updated_at = ? WHERE id = ?", newPinned, time.Now(), id)

	c.JSON(http.StatusOK, gin.H{"pinned": newPinned == 1})
}

func (h *Handler) ToggleNoteArchive(c *gin.Context) {
	id := c.Param("id")

	var archived int
	err := h.db.QueryRow("SELECT archived FROM notes WHERE id = ?", id).Scan(&archived)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	newArchived := 1
	if archived == 1 {
		newArchived = 0
	}

	h.db.Exec("UPDATE notes SET archived = ?, updated_at = ? WHERE id = ?", newArchived, time.Now(), id)

	c.JSON(http.StatusOK, gin.H{"archived": newArchived == 1})
}

func (h *Handler) GetArchivedNotes(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT n.id, n.title, n.account_id, n.template_type, n.created_at, n.updated_at,
		       COALESCE(a.name, '') as account_name
		FROM notes n
		LEFT JOIN accounts a ON n.account_id = a.id
		WHERE n.archived = 1
		ORDER BY n.updated_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	notes := []map[string]interface{}{}
	for rows.Next() {
		var id, title, accountID, templateType, accountName string
		var createdAt, updatedAt time.Time
		rows.Scan(&id, &title, &accountID, &templateType, &createdAt, &updatedAt, &accountName)
		notes = append(notes, map[string]interface{}{
			"id":            id,
			"title":         title,
			"account_id":    accountID,
			"account_name":  accountName,
			"template_type": templateType,
			"created_at":    createdAt,
			"updated_at":    updatedAt,
		})
	}

	c.JSON(http.StatusOK, notes)
}

func (h *Handler) ReorderNotes(c *gin.Context) {
	accountID := c.Param("id")
	var req models.ReorderNotesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i, noteID := range req.NoteIDs {
		_, err := tx.Exec("UPDATE notes SET sort_order = ? WHERE id = ? AND account_id = ?", i, noteID, accountID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notes reordered"})
}

// EmptyNotesTrash permanently deletes all soft-deleted notes
func (h *Handler) EmptyNotesTrash(c *gin.Context) {
	result, err := h.db.Exec(`DELETE FROM notes WHERE deleted_at IS NOT NULL`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, _ := result.RowsAffected()
	c.JSON(http.StatusOK, gin.H{"message": "Trash emptied", "count": rows})
}
