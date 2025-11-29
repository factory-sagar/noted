package handlers

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// ExportAllData exports all data as JSON
func (h *Handler) ExportAllData(c *gin.Context) {
	export := make(map[string]interface{})

	// Export accounts
	accounts := []map[string]interface{}{}
	rows, _ := h.db.Query(`SELECT id, name, account_owner, budget, est_engineers, created_at, updated_at FROM accounts`)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var id, name string
			var accountOwner sql.NullString
			var budget sql.NullFloat64
			var estEngineers sql.NullInt64
			var createdAt, updatedAt string
			rows.Scan(&id, &name, &accountOwner, &budget, &estEngineers, &createdAt, &updatedAt)
			accounts = append(accounts, map[string]interface{}{
				"id":            id,
				"name":          name,
				"account_owner": accountOwner.String,
				"budget":        budget.Float64,
				"est_engineers": estEngineers.Int64,
				"created_at":    createdAt,
				"updated_at":    updatedAt,
			})
		}
	}
	export["accounts"] = accounts

	// Export notes
	notes := []map[string]interface{}{}
	rows2, _ := h.db.Query(`SELECT id, title, account_id, template_type, internal_participants, external_participants, content, meeting_id, meeting_date, pinned, archived, created_at, updated_at FROM notes WHERE deleted_at IS NULL`)
	if rows2 != nil {
		defer rows2.Close()
		for rows2.Next() {
			var id, title, accountID, templateType, internalPart, externalPart, content string
			var meetingID, meetingDate sql.NullString
			var pinned, archived int
			var createdAt, updatedAt string
			rows2.Scan(&id, &title, &accountID, &templateType, &internalPart, &externalPart, &content, &meetingID, &meetingDate, &pinned, &archived, &createdAt, &updatedAt)
			notes = append(notes, map[string]interface{}{
				"id":                    id,
				"title":                 title,
				"account_id":            accountID,
				"template_type":         templateType,
				"internal_participants": internalPart,
				"external_participants": externalPart,
				"content":               content,
				"meeting_id":            meetingID.String,
				"meeting_date":          meetingDate.String,
				"pinned":                pinned == 1,
				"archived":              archived == 1,
				"created_at":            createdAt,
				"updated_at":            updatedAt,
			})
		}
	}
	export["notes"] = notes

	// Export todos
	todos := []map[string]interface{}{}
	rows3, _ := h.db.Query(`SELECT id, title, description, status, priority, due_date, account_id, pinned, created_at, updated_at FROM todos WHERE deleted_at IS NULL`)
	if rows3 != nil {
		defer rows3.Close()
		for rows3.Next() {
			var id, title, description, status, priority string
			var dueDate, accountID sql.NullString
			var pinned int
			var createdAt, updatedAt string
			rows3.Scan(&id, &title, &description, &status, &priority, &dueDate, &accountID, &pinned, &createdAt, &updatedAt)
			todos = append(todos, map[string]interface{}{
				"id":          id,
				"title":       title,
				"description": description,
				"status":      status,
				"priority":    priority,
				"due_date":    dueDate.String,
				"account_id":  accountID.String,
				"pinned":      pinned == 1,
				"created_at":  createdAt,
				"updated_at":  updatedAt,
			})
		}
	}
	export["todos"] = todos

	// Export tags
	tags := []map[string]interface{}{}
	rows4, _ := h.db.Query(`SELECT id, name, color, created_at FROM tags`)
	if rows4 != nil {
		defer rows4.Close()
		for rows4.Next() {
			var id, name, color, createdAt string
			rows4.Scan(&id, &name, &color, &createdAt)
			tags = append(tags, map[string]interface{}{
				"id":         id,
				"name":       name,
				"color":      color,
				"created_at": createdAt,
			})
		}
	}
	export["tags"] = tags

	// Export contacts
	contacts := []map[string]interface{}{}
	rows5, _ := h.db.Query(`SELECT id, email, name, company, domain, is_internal, account_id, meeting_count, created_at FROM contacts`)
	if rows5 != nil {
		defer rows5.Close()
		for rows5.Next() {
			var id, email, name, company, domain string
			var isInternal, meetingCount int
			var accountID sql.NullString
			var createdAt string
			rows5.Scan(&id, &email, &name, &company, &domain, &isInternal, &accountID, &meetingCount, &createdAt)
			contacts = append(contacts, map[string]interface{}{
				"id":            id,
				"email":         email,
				"name":          name,
				"company":       company,
				"domain":        domain,
				"is_internal":   isInternal == 1,
				"account_id":    accountID.String,
				"meeting_count": meetingCount,
				"created_at":    createdAt,
			})
		}
	}
	export["contacts"] = contacts

	export["exported_at"] = time.Now().Format(time.RFC3339)
	export["version"] = "1.0"

	c.JSON(http.StatusOK, export)
}

// ClearAllData deletes all user data
func (h *Handler) ClearAllData(c *gin.Context) {
	// Delete in order to respect foreign key constraints
	tables := []string{
		"note_todos",
		"note_tags",
		"attachments",
		"activities",
		"contacts",
		"todos",
		"notes",
		"tags",
		"accounts",
	}

	for _, table := range tables {
		if _, err := h.db.Exec("DELETE FROM " + table); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear " + table})
			return
		}
	}

	// Clear uploads directory
	files, _ := os.ReadDir(h.uploadsDir)
	for _, f := range files {
		os.Remove(filepath.Join(h.uploadsDir, f.Name()))
	}

	c.JSON(http.StatusOK, gin.H{"message": "All data cleared successfully"})
}
