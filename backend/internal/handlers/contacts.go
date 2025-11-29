package handlers

import (
	"database/sql"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetInternalDomain returns the internal email domain for identifying internal contacts.
// Set INTERNAL_DOMAIN environment variable to customize (default: "example.com")
func GetInternalDomain() string {
	if domain := os.Getenv("INTERNAL_DOMAIN"); domain != "" {
		return domain
	}
	return "example.com"
}

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
	return strings.HasSuffix(strings.ToLower(email), "@"+GetInternalDomain())
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
	if domain == "" || domain == GetInternalDomain() {
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

// GetContactNotes returns notes where this contact participated
func (h *Handler) GetContactNotes(c *gin.Context) {
	id := c.Param("id")

	// First get the contact's email
	var email string
	err := h.db.QueryRow(`SELECT email FROM contacts WHERE id = ?`, id).Scan(&email)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Find notes where this email appears in participants
	rows, err := h.db.Query(`
		SELECT n.id, n.title, n.account_id, a.name, n.meeting_date, n.created_at
		FROM notes n
		LEFT JOIN accounts a ON n.account_id = a.id
		WHERE n.deleted_at IS NULL
		  AND (n.internal_participants LIKE ? OR n.external_participants LIKE ?)
		ORDER BY COALESCE(n.meeting_date, n.created_at) DESC
		LIMIT 50
	`, "%"+email+"%", "%"+email+"%")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type NoteRef struct {
		ID          string     `json:"id"`
		Title       string     `json:"title"`
		AccountID   *string    `json:"account_id,omitempty"`
		AccountName string     `json:"account_name,omitempty"`
		MeetingDate *time.Time `json:"meeting_date,omitempty"`
		CreatedAt   time.Time  `json:"created_at"`
	}

	notes := []NoteRef{}
	for rows.Next() {
		var note NoteRef
		var accountID, accountName sql.NullString
		var meetingDate sql.NullTime

		err := rows.Scan(&note.ID, &note.Title, &accountID, &accountName, &meetingDate, &note.CreatedAt)
		if err != nil {
			continue
		}

		if accountID.Valid {
			note.AccountID = &accountID.String
			note.AccountName = accountName.String
		}
		if meetingDate.Valid {
			note.MeetingDate = &meetingDate.Time
		}

		notes = append(notes, note)
	}

	c.JSON(http.StatusOK, notes)
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

type BulkContactsRequest struct {
	ContactIDs []string               `json:"contact_ids" binding:"required"`
	Action     string                 `json:"action" binding:"required"` // "delete", "set_internal", "set_account"
	Value      map[string]interface{} `json:"value"`
}

// BulkContactsOperation handles bulk actions on contacts
func (h *Handler) BulkContactsOperation(c *gin.Context) {
	var req BulkContactsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.ContactIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No contacts selected"})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	// Prepare IDs for query
	placeholders := strings.Repeat("?,", len(req.ContactIDs))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]interface{}, len(req.ContactIDs))
	for i, id := range req.ContactIDs {
		args[i] = id
	}

	switch req.Action {
	case "delete":
		query := "DELETE FROM contacts WHERE id IN (" + placeholders + ")"
		if _, err := tx.Exec(query, args...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	case "set_internal":
		isInternal, ok := req.Value["is_internal"].(bool)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid value for is_internal"})
			return
		}
		query := "UPDATE contacts SET is_internal = ?, updated_at = CURRENT_TIMESTAMP WHERE id IN (" + placeholders + ")"
		execArgs := append([]interface{}{isInternal}, args...)
		if _, err := tx.Exec(query, execArgs...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	case "set_account":
		accountID, ok := req.Value["account_id"].(string)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid value for account_id"})
			return
		}
		query := "UPDATE contacts SET account_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id IN (" + placeholders + ")"
		execArgs := append([]interface{}{accountID}, args...)
		if _, err := tx.Exec(query, execArgs...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bulk operation completed"})
}

// DomainGroup represents contacts grouped by domain with suggestions
type DomainGroup struct {
	Domain           string    `json:"domain"`
	ContactCount     int       `json:"contact_count"`
	ContactIDs       []string  `json:"contact_ids"`
	IsInternal       bool      `json:"is_internal"`
	LinkedAccountID  *string   `json:"linked_account_id,omitempty"`
	LinkedAccountName string   `json:"linked_account_name,omitempty"`
	SuggestedAccount *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"suggested_account,omitempty"`
	Contacts []Contact `json:"contacts,omitempty"`
}

// GetContactDomainGroups returns contacts grouped by domain with smart suggestions
func (h *Handler) GetContactDomainGroups(c *gin.Context) {
	includeContacts := c.Query("include_contacts") == "true"
	filter := c.Query("filter") // "unlinked", "all"

	// Get all external contacts grouped by domain
	query := `
		SELECT c.domain, COUNT(*) as count, c.is_internal,
		       GROUP_CONCAT(c.id) as contact_ids,
		       MAX(c.account_id) as account_id
		FROM contacts c
		WHERE c.is_internal = 0
	`
	if filter == "unlinked" {
		query += " AND c.account_id IS NULL"
	}
	query += `
		GROUP BY c.domain
		ORDER BY count DESC
	`

	rows, err := h.db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	groups := []DomainGroup{}
	for rows.Next() {
		var group DomainGroup
		var isInternal int
		var contactIDsStr string
		var accountID sql.NullString

		if err := rows.Scan(&group.Domain, &group.ContactCount, &isInternal, &contactIDsStr, &accountID); err != nil {
			continue
		}

		group.IsInternal = isInternal == 1
		group.ContactIDs = strings.Split(contactIDsStr, ",")

		if accountID.Valid {
			group.LinkedAccountID = &accountID.String
			// Get account name
			h.db.QueryRow("SELECT name FROM accounts WHERE id = ?", accountID.String).Scan(&group.LinkedAccountName)
		} else {
			// Try to find a matching account by domain
			var suggestedID, suggestedName string
			// Look for accounts that already have contacts with this domain
			err := h.db.QueryRow(`
				SELECT a.id, a.name FROM accounts a
				INNER JOIN contacts c ON c.account_id = a.id
				WHERE c.domain = ? AND a.deleted_at IS NULL
				GROUP BY a.id
				ORDER BY COUNT(*) DESC
				LIMIT 1
			`, group.Domain).Scan(&suggestedID, &suggestedName)
			
			if err == nil {
				group.SuggestedAccount = &struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				}{ID: suggestedID, Name: suggestedName}
			} else {
				// Try matching by account name containing domain
				domainParts := strings.Split(group.Domain, ".")
				if len(domainParts) > 0 {
					companyName := domainParts[0]
					err := h.db.QueryRow(`
						SELECT id, name FROM accounts 
						WHERE LOWER(name) LIKE ? AND deleted_at IS NULL
						LIMIT 1
					`, "%"+companyName+"%").Scan(&suggestedID, &suggestedName)
					
					if err == nil {
						group.SuggestedAccount = &struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						}{ID: suggestedID, Name: suggestedName}
					}
				}
			}
		}

		groups = append(groups, group)
	}

	// Optionally include full contact details
	if includeContacts {
		for i := range groups {
			groups[i].Contacts = h.getContactsByDomain(groups[i].Domain)
		}
	}

	c.JSON(http.StatusOK, groups)
}

func (h *Handler) getContactsByDomain(domain string) []Contact {
	rows, err := h.db.Query(`
		SELECT c.id, c.email, c.name, c.company, c.domain, c.is_internal,
		       c.account_id, a.name, c.suggested_account_id, sa.name,
		       c.suggestion_confirmed, c.source, c.first_seen, c.last_seen,
		       c.meeting_count, c.created_at, c.updated_at
		FROM contacts c
		LEFT JOIN accounts a ON c.account_id = a.id
		LEFT JOIN accounts sa ON c.suggested_account_id = sa.id
		WHERE c.domain = ?
		ORDER BY c.name ASC
	`, domain)
	if err != nil {
		return []Contact{}
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
	return contacts
}

// LinkDomainToAccount links all contacts with a domain to an account
func (h *Handler) LinkDomainToAccount(c *gin.Context) {
	domain := c.Param("domain")
	accountID := c.Param("accountId")

	if domain == "" || accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain and account ID required"})
		return
	}

	result, err := h.db.Exec(`
		UPDATE contacts 
		SET account_id = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE domain = ? AND is_internal = 0
	`, accountID, domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	c.JSON(http.StatusOK, gin.H{
		"message": "Domain linked to account",
		"contacts_updated": rowsAffected,
	})
}

// CreateAccountFromDomain creates a new account and links all domain contacts to it
func (h *Handler) CreateAccountFromDomain(c *gin.Context) {
	domain := c.Param("domain")

	var req struct {
		AccountName string `json:"account_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// Use domain as account name if not provided
		domainParts := strings.Split(domain, ".")
		if len(domainParts) > 0 {
			req.AccountName = strings.Title(domainParts[0])
		} else {
			req.AccountName = domain
		}
	}

	// Create account
	accountID := uuid.New().String()
	_, err := h.db.Exec(`
		INSERT INTO accounts (id, name) VALUES (?, ?)
	`, accountID, req.AccountName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Link all contacts with this domain
	result, err := h.db.Exec(`
		UPDATE contacts 
		SET account_id = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE domain = ? AND is_internal = 0
	`, accountID, domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	c.JSON(http.StatusOK, gin.H{
		"message": "Account created and contacts linked",
		"account_id": accountID,
		"account_name": req.AccountName,
		"contacts_updated": rowsAffected,
	})
}
