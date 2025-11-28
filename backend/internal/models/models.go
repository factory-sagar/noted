package models

import (
	"time"
)

// Account represents a customer account
type Account struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	AccountOwner string    `json:"account_owner"` // Sales rep
	Budget       *float64  `json:"budget,omitempty"`
	EstEngineers *int      `json:"est_engineers,omitempty"` // Estimated POC size
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Note represents a meeting note
type Note struct {
	ID                   string     `json:"id"`
	Title                string     `json:"title"`
	AccountID            string     `json:"account_id"`
	Account              *Account   `json:"account,omitempty"`
	TemplateType         string     `json:"template_type"` // "initial" or "followup"
	InternalParticipants []string   `json:"internal_participants"`
	ExternalParticipants []string   `json:"external_participants"`
	Content              string     `json:"content"` // Rich text JSON from TipTap
	MeetingID            *string    `json:"meeting_id,omitempty"`
	MeetingDate          *time.Time `json:"meeting_date,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	Todos                []Todo     `json:"todos,omitempty"`
	Tags                 []Tag      `json:"tags,omitempty"`
}

// Todo represents a task/follow-up item
type Todo struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      string     `json:"status"`             // "not_started", "in_progress", "completed"
	Priority    string     `json:"priority,omitempty"` // "low", "medium", "high"
	DueDate     *time.Time `json:"due_date,omitempty"`
	AccountID   *string    `json:"account_id,omitempty"`   // Optional account tag
	AccountName string     `json:"account_name,omitempty"` // Populated from join
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Notes       []Note     `json:"notes,omitempty"` // Linked notes
}

// Participant represents a meeting participant
type Participant struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Title      string `json:"title,omitempty"`
	IsInternal bool   `json:"is_internal"`
}

// Tag represents a tag for notes
type Tag struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateTagRequest for creating a tag
type CreateTagRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

// UpdateTagRequest for updating a tag
type UpdateTagRequest struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

// CreateAccountRequest for creating an account
type CreateAccountRequest struct {
	Name         string   `json:"name" binding:"required"`
	AccountOwner string   `json:"account_owner"`
	Budget       *float64 `json:"budget"`
	EstEngineers *int     `json:"est_engineers"`
}

// UpdateAccountRequest for updating an account
type UpdateAccountRequest struct {
	Name         *string  `json:"name"`
	AccountOwner *string  `json:"account_owner"`
	Budget       *float64 `json:"budget"`
	EstEngineers *int     `json:"est_engineers"`
}

// CreateNoteRequest for creating a note
type CreateNoteRequest struct {
	Title                string   `json:"title" binding:"required"`
	AccountID            string   `json:"account_id" binding:"required"`
	TemplateType         string   `json:"template_type"`
	InternalParticipants []string `json:"internal_participants"`
	ExternalParticipants []string `json:"external_participants"`
	Content              string   `json:"content"`
	MeetingID            *string  `json:"meeting_id"`
	MeetingDate          *string  `json:"meeting_date"`
}

// UpdateNoteRequest for updating a note
type UpdateNoteRequest struct {
	Title                *string  `json:"title"`
	AccountID            *string  `json:"account_id"`
	TemplateType         *string  `json:"template_type"`
	InternalParticipants []string `json:"internal_participants"`
	ExternalParticipants []string `json:"external_participants"`
	Content              *string  `json:"content"`
	MeetingID            *string  `json:"meeting_id"`
	MeetingDate          *string  `json:"meeting_date"`
}

// CreateTodoRequest for creating a todo
type CreateTodoRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	DueDate     *string `json:"due_date"`
	NoteID      *string `json:"note_id"`    // Optional: link to a note on creation
	AccountID   *string `json:"account_id"` // Optional: tag with account
}

// UpdateTodoRequest for updating a todo
type UpdateTodoRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Priority    *string `json:"priority"`
	DueDate     *string `json:"due_date"`
	AccountID   *string `json:"account_id"`
}

// Analytics response
type Analytics struct {
	TotalNotes      int                `json:"total_notes"`
	TotalAccounts   int                `json:"total_accounts"`
	TotalTodos      int                `json:"total_todos"`
	TodosByStatus   map[string]int     `json:"todos_by_status"`
	NotesByAccount  []AccountNoteCount `json:"notes_by_account"`
	IncompleteCount int                `json:"incomplete_count"`
}

// AccountNoteCount for analytics
type AccountNoteCount struct {
	AccountID   string `json:"account_id"`
	AccountName string `json:"account_name"`
	NoteCount   int    `json:"note_count"`
}

// IncompleteField represents a note with incomplete fields
type IncompleteField struct {
	NoteID        string   `json:"note_id"`
	NoteTitle     string   `json:"note_title"`
	AccountName   string   `json:"account_name"`
	MissingFields []string `json:"missing_fields"`
}

// SearchResult represents a search result
type SearchResult struct {
	Type      string `json:"type"` // "note", "account", "todo"
	ID        string `json:"id"`
	Title     string `json:"title"`
	Snippet   string `json:"snippet,omitempty"`
	AccountID string `json:"account_id,omitempty"`
}
