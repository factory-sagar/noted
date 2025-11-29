package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/factory-sagar/notes-droid/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) QuickCapture(c *gin.Context) {
	var req models.QuickCaptureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()
	now := time.Now()

	if req.Type == "note" {
		accountID := req.AccountID
		if accountID == nil {
			// Create a default "Unassigned" account if none provided
			var defaultAccountID string
			err := h.db.QueryRow("SELECT id FROM accounts WHERE name = 'Unassigned'").Scan(&defaultAccountID)
			if err == sql.ErrNoRows {
				defaultAccountID = uuid.New().String()
				h.db.Exec("INSERT INTO accounts (id, name, created_at, updated_at) VALUES (?, 'Unassigned', ?, ?)",
					defaultAccountID, now, now)
			}
			accountID = &defaultAccountID
		}

		_, err := h.db.Exec(`
			INSERT INTO notes (id, title, account_id, template_type, content, created_at, updated_at)
			VALUES (?, ?, ?, 'quick', ?, ?, ?)
		`, id, req.Title, *accountID, req.Content, now, now)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":         id,
			"type":       "note",
			"title":      req.Title,
			"account_id": *accountID,
			"created_at": now,
		})
	} else if req.Type == "todo" {
		priority := req.Priority
		if priority == "" {
			priority = "medium"
		}

		_, err := h.db.Exec(`
			INSERT INTO todos (id, title, description, status, priority, account_id, created_at, updated_at)
			VALUES (?, ?, ?, 'not_started', ?, ?, ?, ?)
		`, id, req.Title, req.Description, priority, req.AccountID, now, now)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":         id,
			"type":       "todo",
			"title":      req.Title,
			"priority":   priority,
			"created_at": now,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type, must be 'note' or 'todo'"})
	}
}
