package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/factory-sagar/notes-droid/backend/internal/models"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	results := []models.SearchResult{}
	seen := make(map[string]bool) // Prevent duplicates

	// Create fuzzy search term for FTS (prefix matching)
	ftsQuery := q + "*"
	likeQuery := "%" + q + "%"

	// Search notes using FTS4 with prefix matching
	noteRows, err := h.db.Query(`
		SELECT n.id, n.title, n.account_id, COALESCE(a.name, '') as account_name,
		       snippet(notes_fts, '<mark>', '</mark>', '...', -1, 64) as snippet
		FROM notes_fts
		JOIN notes n ON notes_fts.docid = n.rowid
		LEFT JOIN accounts a ON n.account_id = a.id
		WHERE notes_fts MATCH ?
		LIMIT 20
	`, ftsQuery)
	if err == nil {
		defer noteRows.Close()
		for noteRows.Next() {
			var id, title, accountID, accountName, snippet string
			noteRows.Scan(&id, &title, &accountID, &accountName, &snippet)
			if !seen["note:"+id] {
				seen["note:"+id] = true
				results = append(results, models.SearchResult{
					Type:      "note",
					ID:        id,
					Title:     title,
					Snippet:   snippet,
					AccountID: accountID,
				})
			}
		}
	}

	// Also search notes by participants (not in FTS)
	participantRows, _ := h.db.Query(`
		SELECT n.id, n.title, n.account_id, COALESCE(a.name, '') as account_name
		FROM notes n
		LEFT JOIN accounts a ON n.account_id = a.id
		WHERE n.internal_participants LIKE ? OR n.external_participants LIKE ?
		LIMIT 10
	`, likeQuery, likeQuery)
	if participantRows != nil {
		defer participantRows.Close()
		for participantRows.Next() {
			var id, title, accountID, accountName string
			participantRows.Scan(&id, &title, &accountID, &accountName)
			if !seen["note:"+id] {
				seen["note:"+id] = true
				results = append(results, models.SearchResult{
					Type:      "note",
					ID:        id,
					Title:     title,
					Snippet:   "Match in participants",
					AccountID: accountID,
				})
			}
		}
	}

	// Search accounts by name and owner
	accountRows, err := h.db.Query(`
		SELECT id, name, account_owner FROM accounts 
		WHERE name LIKE ? OR account_owner LIKE ? 
		LIMIT 10
	`, likeQuery, likeQuery)
	if err == nil {
		defer accountRows.Close()
		for accountRows.Next() {
			var id, name string
			var owner sql.NullString
			accountRows.Scan(&id, &name, &owner)
			if !seen["account:"+id] {
				seen["account:"+id] = true
				snippet := ""
				if owner.Valid && owner.String != "" && strings.Contains(strings.ToLower(owner.String), strings.ToLower(q)) {
					snippet = "Owner: " + owner.String
				}
				results = append(results, models.SearchResult{
					Type:    "account",
					ID:      id,
					Title:   name,
					Snippet: snippet,
				})
			}
		}
	}

	// Search todos by title and description
	todoRows, err := h.db.Query(`
		SELECT t.id, t.title, t.description, t.account_id, COALESCE(a.name, '') as account_name
		FROM todos t
		LEFT JOIN accounts a ON t.account_id = a.id
		WHERE t.title LIKE ? OR t.description LIKE ?
		LIMIT 10
	`, likeQuery, likeQuery)
	if err == nil {
		defer todoRows.Close()
		for todoRows.Next() {
			var id, title, description string
			var accountID sql.NullString
			var accountName string
			todoRows.Scan(&id, &title, &description, &accountID, &accountName)
			if !seen["todo:"+id] {
				seen["todo:"+id] = true
				snippet := ""
				if description != "" && strings.Contains(strings.ToLower(description), strings.ToLower(q)) {
					if len(description) > 100 {
						snippet = description[:100] + "..."
					} else {
						snippet = description
					}
				}
				if accountName != "" {
					if snippet != "" {
						snippet += " | "
					}
					snippet += "Account: " + accountName
				}
				results = append(results, models.SearchResult{
					Type:    "todo",
					ID:      id,
					Title:   title,
					Snippet: snippet,
				})
			}
		}
	}

	c.JSON(http.StatusOK, results)
}
