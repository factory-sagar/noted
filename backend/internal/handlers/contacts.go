package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const InternalDomain = "factory.ai"

type Contact struct {
	ID                  string     `json:"id"`
	Email               string     `json:"email"`
	Name                string     `json:"name"`
	Company             string     `json:"company"`
	Domain              string     `json:"domain"`
	IsInternal          bool       `json:"is_internal"`
	AccountID           *string    `json:"account_id,omitempty"`
	AccountName         string     `json:"account_name,omitempty"`
	SuggestedAccountID  *string    `json:"suggested_account_id,omitempty"`
	SuggestedAccountName string    `json:"suggested_account_name,omitempty"`
	SuggestionConfirmed bool       `json:"suggestion_confirmed"`
	Source              string     `json:"source"`
	FirstSeen           time.Time  `json:"first_seen"`
	LastSeen            time.Time  `json:"last_seen"`
	MeetingCount        int        `json:"meeting_count"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type CreateContactRequest struct {
	Email   string  `json:"email" binding:"required,email"`
	Name    string  `json:"name"`
	Company string  `json:"company"`
	Source  string  `json:"source"`
}

func extractDomain(email string) string {
	parts := strings.Split(strings.ToLower(email), "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

func isInternalEmail(email string) bool {
	return strings.HasSuffix(strings.ToLower(email), "@"+InternalDomain)
}

// GetContacts returns all contacts with optional filtering
func (h *Handler) GetContacts(c *gin.Context) {
	filter := c.Query("filter") // "internal", "external", "unlinked", "suggestions"
	accountID := c.Query("account_id")

	query := `
		SELECT c.id, c.email, c.name, c.company, c.domain, c.is_internal,
		       c.account_id, a.name, c.suggested_account_id, sa.name,
		       c.suggestion_confirmed, c.source, c.first_seen, c.last_seen,
		       c.meeting_count, c.created_at, c.updated_at
		FROM contacts c
		LEFT JOIN accounts a ON c.account_id = a.id
		LEFT JOIN accounts sa ON c.suggested_account_id = sa.id
		WHERE 1=1
	`
	args := []interface{}{}

	if filter == "internal" {
		query += " AND c.is_internal = 1"
	} else if filter == "external" {
		query += " AND c.is_internal = 0"
	} else if filter == "unlinked" {
		query += " AND c.account_id IS NULL AND c.is_internal = 0"
	} else if filter == "suggestions" {
		query += " AND c.suggested_account_id IS NOT NULL AND c.suggestion_confirmed = 0"
	}

	if accountID != "" {
		query += " AND c.account_id = ?"
		args = append(args, accountID)
	}

	query += " ORDER BY c.last_seen DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	contacts := []Contact{}
	for rows.Next() {
		var contact Contact
		var accountID, accountName, suggestedAccountID, suggestedAccountName sql.NullString
		var isInternal, suggestionConfirmed int

		err := rows.Scan(
			&contact.ID, &contact.Email, &contact.Name, &contact.Company, &contact.Domain,
			&isInternal, &accountID, &accountName, &suggestedAccountID, &suggestedAccountName,
			&suggestionConfirmed, &contact.Source, &contact.FirstSeen, &contact.LastSeen,
			&contact.MeetingCount, &contact.CreatedAt, &contact.UpdatedAt,
		)
		if err != nil {
			continue
		}

		contact.IsInternal = isInternal == 1
		contact.SuggestionConfirmed = suggestionConfirmed == 1
		if accountID.Valid {
			contact.AccountID = &accountID.String
			contact.AccountName = accountName.String
		}
		if suggestedAccountID.Valid {
			contact.SuggestedAccountID = &suggestedAccountID.String
			contact.SuggestedAccountName = suggestedAccountName.String
		}

		contacts = append(contacts, contact)
	}

	c.JSON(http.StatusOK, contacts)
}

// GetContact returns a single contact
func (h *Handler) GetContact(c *gin.Context) {
	id := c.Param("id")

	var contact Contact
	var accountID, accountName, suggestedAccountID, suggestedAccountName sql.NullString
	var isInternal, suggestionConfirmed int

	err := h.db.QueryRow(`
		SELECT c.id, c.email, c.name, c.company, c.domain, c.is_internal,
		       c.account_id, a.name, c.suggested_account_id, sa.name,
		       c.suggestion_confirmed, c.source, c.first_seen, c.last_seen,
		       c.meeting_count, c.created_at, c.updated_at
		FROM contacts c
		LEFT JOIN accounts a ON c.account_id = a.id
		LEFT JOIN accounts sa ON c.suggested_account_id = sa.id
		WHERE c.id = ?
	`, id).Scan(
		&contact.ID, &contact.Email, &contact.Name, &contact.Company, &contact.Domain,
		&isInternal, &accountID, &accountName, &suggestedAccountID, &suggestedAccountName,
		&suggestionConfirmed, &contact.Source, &contact.FirstSeen, &contact.LastSeen,
		&contact.MeetingCount, &contact.CreatedAt, &contact.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	contact.IsInternal = isInternal == 1
	contact.SuggestionConfirmed = suggestionConfirmed == 1
	if accountID.Valid {
		contact.AccountID = &accountID.String
		contact.AccountName = accountName.String
	}
	if suggestedAccountID.Valid {
		contact.SuggestedAccountID = &suggestedAccountID.String
		contact.SuggestedAccountName = suggestedAccountName.String
	}

	c.JSON(http.StatusOK, contact)
}

// CreateContact creates a new contact manually
func (h *Handler) CreateContact(c *gin.Context) {
	var req CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	domain := extractDomain(email)
	isInternal := isInternalEmail(email)

	source := req.Source
	if source == "" {
		source = "manual"
	}

	id := uuid.New().String()
	_, err := h.db.Exec(`
		INSERT INTO contacts (id, email, name, company, domain, is_internal, source)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, email, req.Name, req.Company, domain, isInternal, source)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "Contact already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Try to suggest an account
	h.suggestAccountForContact(id, domain)

	c.JSON(http.StatusCreated, gin.H{"id": id, "email": email})
}

// UpdateContact updates a contact
func (h *Handler) UpdateContact(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name      *string `json:"name"`
		Company   *string `json:"company"`
		AccountID *string `json:"account_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		h.db.Exec(`UPDATE contacts SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, *req.Name, id)
	}
	if req.Company != nil {
		h.db.Exec(`UPDATE contacts SET company = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, *req.Company, id)
	}
	if req.AccountID != nil {
		if *req.AccountID == "" {
			h.db.Exec(`UPDATE contacts SET account_id = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
		} else {
			h.db.Exec(`UPDATE contacts SET account_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, *req.AccountID, id)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contact updated"})
}

// DeleteContact deletes a contact
func (h *Handler) DeleteContact(c *gin.Context) {
	id := c.Param("id")

	result, err := h.db.Exec(`DELETE FROM contacts WHERE id = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contact deleted"})
}

// ConfirmAccountSuggestion confirms or rejects an account suggestion
func (h *Handler) ConfirmAccountSuggestion(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Confirm bool `json:"confirm"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Confirm {
		// Move suggested_account_id to account_id
		_, err := h.db.Exec(`
			UPDATE contacts
			SET account_id = suggested_account_id,
			    suggestion_confirmed = 1,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		// Clear the suggestion
		_, err := h.db.Exec(`
			UPDATE contacts
			SET suggested_account_id = NULL,
			    suggestion_confirmed = 1,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Suggestion processed"})
}

// LinkContactToAccount links a contact to an account
func (h *Handler) LinkContactToAccount(c *gin.Context) {
	contactID := c.Param("id")
	accountID := c.Param("accountId")

	_, err := h.db.Exec(`
		UPDATE contacts
		SET account_id = ?, suggestion_confirmed = 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, accountID, contactID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contact linked to account"})
}

// UpsertContactFromEmail creates or updates a contact from an email
func (h *Handler) UpsertContactFromEmail(email, name, source string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil
	}

	domain := extractDomain(email)
	isInternal := isInternalEmail(email)

	// Check if contact exists
	var existingID string
	err := h.db.QueryRow(`SELECT id FROM contacts WHERE email = ?`, email).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Create new contact
		id := uuid.New().String()
		_, err = h.db.Exec(`
			INSERT INTO contacts (id, email, name, domain, is_internal, source, meeting_count)
			VALUES (?, ?, ?, ?, ?, ?, 1)
		`, id, email, name, domain, isInternal, source)
		if err != nil {
			return err
		}
		// Try to suggest an account
		h.suggestAccountForContact(id, domain)
	} else if err == nil {
		// Update existing contact
		if name != "" {
			h.db.Exec(`
				UPDATE contacts
				SET name = CASE WHEN name = '' THEN ? ELSE name END,
				    last_seen = CURRENT_TIMESTAMP,
				    meeting_count = meeting_count + 1,
				    updated_at = CURRENT_TIMESTAMP
				WHERE id = ?
			`, name, existingID)
		} else {
			h.db.Exec(`
				UPDATE contacts
				SET last_seen = CURRENT_TIMESTAMP,
				    meeting_count = meeting_count + 1,
				    updated_at = CURRENT_TIMESTAMP
				WHERE id = ?
			`, existingID)
		}
	}

	return nil
}

// suggestAccountForContact tries to match a contact's domain to an existing account
func (h *Handler) suggestAccountForContact(contactID, domain string) {
	if domain == "" || domain == InternalDomain {
		return
	}

	// Try to find an account with a matching domain in its name
	// e.g., domain "nvidia.com" might match account "NVIDIA"
	domainBase := strings.TrimSuffix(domain, ".com")
	domainBase = strings.TrimSuffix(domainBase, ".io")
	domainBase = strings.TrimSuffix(domainBase, ".ai")
	domainBase = strings.TrimSuffix(domainBase, ".co")

	var accountID string
	err := h.db.QueryRow(`
		SELECT id FROM accounts
		WHERE LOWER(name) LIKE ?
		LIMIT 1
	`, "%"+strings.ToLower(domainBase)+"%").Scan(&accountID)

	if err == nil && accountID != "" {
		h.db.Exec(`
			UPDATE contacts
			SET suggested_account_id = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND account_id IS NULL
		`, accountID, contactID)
	}
}

// ExtractContactsFromNote extracts contacts from note participants
func (h *Handler) ExtractContactsFromNote(internalParticipants, externalParticipants []string) {
	for _, email := range internalParticipants {
		h.UpsertContactFromEmail(email, "", "note")
	}
	for _, email := range externalParticipants {
		h.UpsertContactFromEmail(email, "", "note")
	}
}

// GetContactStats returns contact statistics
func (h *Handler) GetContactStats(c *gin.Context) {
	var stats struct {
		TotalContacts    int `json:"total_contacts"`
		InternalContacts int `json:"internal_contacts"`
		ExternalContacts int `json:"external_contacts"`
		LinkedContacts   int `json:"linked_contacts"`
		PendingSuggestions int `json:"pending_suggestions"`
	}

	h.db.QueryRow(`SELECT COUNT(*) FROM contacts`).Scan(&stats.TotalContacts)
	h.db.QueryRow(`SELECT COUNT(*) FROM contacts WHERE is_internal = 1`).Scan(&stats.InternalContacts)
	h.db.QueryRow(`SELECT COUNT(*) FROM contacts WHERE is_internal = 0`).Scan(&stats.ExternalContacts)
	h.db.QueryRow(`SELECT COUNT(*) FROM contacts WHERE account_id IS NOT NULL`).Scan(&stats.LinkedContacts)
	h.db.QueryRow(`SELECT COUNT(*) FROM contacts WHERE suggested_account_id IS NOT NULL AND suggestion_confirmed = 0`).Scan(&stats.PendingSuggestions)

	c.JSON(http.StatusOK, stats)
}
