package handlers

import (
	"database/sql"
	"net/http"

	"github.com/factory-sagar/notes-droid/backend/internal/models"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetAnalytics(c *gin.Context) {
	var analytics models.Analytics

	// Total counts - exclude deleted items and Unassigned account
	h.db.QueryRow(`
		SELECT COUNT(*) FROM notes n
		JOIN accounts a ON n.account_id = a.id
		WHERE n.deleted_at IS NULL AND a.name != 'Unassigned'
	`).Scan(&analytics.TotalNotes)
	h.db.QueryRow("SELECT COUNT(*) FROM accounts WHERE name != 'Unassigned'").Scan(&analytics.TotalAccounts)
	h.db.QueryRow("SELECT COUNT(*) FROM todos WHERE deleted_at IS NULL").Scan(&analytics.TotalTodos)

	// Todos by status - exclude deleted
	analytics.TodosByStatus = map[string]int{}
	statusRows, _ := h.db.Query("SELECT status, COUNT(*) FROM todos WHERE deleted_at IS NULL GROUP BY status")
	if statusRows != nil {
		defer statusRows.Close()
		for statusRows.Next() {
			var status string
			var count int
			statusRows.Scan(&status, &count)
			analytics.TodosByStatus[status] = count
		}
	}

	// Notes by account - exclude deleted notes and Unassigned account
	accountRows, _ := h.db.Query(`
		SELECT a.id, a.name, COUNT(n.id) as note_count
		FROM accounts a
		LEFT JOIN notes n ON a.id = n.account_id AND n.deleted_at IS NULL
		WHERE a.name != 'Unassigned'
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

	// Count notes with incomplete fields - exclude deleted and Unassigned
	h.db.QueryRow(`
		SELECT COUNT(*) FROM notes n
		JOIN accounts a ON n.account_id = a.id
		WHERE n.deleted_at IS NULL AND a.name != 'Unassigned'
		  AND (a.budget IS NULL OR a.est_engineers IS NULL OR a.account_owner = ''
		   OR n.content = '' OR n.internal_participants = '[]')
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
		WHERE n.deleted_at IS NULL
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
