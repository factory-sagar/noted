package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/factory-sagar/notes-droid/backend/internal/models"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Enable parseTime for correct time.Time handling
	db, err := sql.Open("sqlite3", ":memory:?parseTime=true")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Initialize schema
	schema := `
	CREATE TABLE accounts (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		account_owner TEXT,
		budget REAL,
		est_engineers INTEGER,
		created_at DATETIME,
		updated_at DATETIME
	);
	CREATE TABLE notes (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		account_id TEXT,
		template_type TEXT,
		internal_participants TEXT,
		external_participants TEXT,
		content TEXT,
		meeting_id TEXT,
		meeting_date DATETIME,
		pinned INTEGER DEFAULT 0,
		archived INTEGER DEFAULT 0,
		deleted_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME,
		sort_order INTEGER DEFAULT 0
	);
	CREATE TABLE todos (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT,
		status TEXT,
		priority TEXT,
		due_date DATETIME,
		account_id TEXT,
		pinned INTEGER DEFAULT 0,
		deleted_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME
	);
	CREATE TABLE note_todos (
		note_id TEXT,
		todo_id TEXT,
		PRIMARY KEY (note_id, todo_id)
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

func TestCreateAccount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	h := New(db)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/accounts", h.CreateAccount)

	t.Run("Success", func(t *testing.T) {
		reqBody := models.CreateAccountRequest{
			Name:         "Acme Corp",
			AccountOwner: "Alice",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/accounts", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp models.Account
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Acme Corp", resp.Name)
		assert.NotEmpty(t, resp.ID)
	})

	t.Run("Validation Failure", func(t *testing.T) {
		reqBody := models.CreateAccountRequest{
			Name: "", // Empty name
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/accounts", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCreateNote(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	h := New(db)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/notes", h.CreateNote)

	// Create a dummy account first
	db.Exec("INSERT INTO accounts (id, name) VALUES ('acc-1', 'Test Account')")

	t.Run("Success", func(t *testing.T) {
		reqBody := models.CreateNoteRequest{
			Title:     "Discovery Call",
			AccountID: "acc-1",
			Content:   "Notes content",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/notes", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Discovery Call", resp["title"])
	})

	t.Run("Invalid Date", func(t *testing.T) {
		badDate := "invalid-date"
		reqBody := models.CreateNoteRequest{
			Title:       "Bad Date Note",
			AccountID:   "acc-1",
			MeetingDate: &badDate,
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/notes", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetTodos_NPlusOne(t *testing.T) {
	// This test verifies the N+1 fix by ensuring linked notes are returned correctly
	db := setupTestDB(t)
	defer db.Close()
	h := New(db)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/todos", h.GetTodos)

	// Setup Data
	todoID := "todo-1"
	// Explicitly provide empty description and status/priority to avoid Scan error
	_, err := db.Exec("INSERT INTO todos (id, title, description, status, priority, created_at, updated_at, deleted_at) VALUES (?, 'Follow up', '', 'not_started', 'low', ?, ?, NULL)", todoID, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to insert todo: %v", err)
	}

	noteID := "note-1"
	_, err = db.Exec("INSERT INTO notes (id, title, created_at, updated_at, deleted_at) VALUES (?, 'Meeting Notes', ?, ?, ?)", noteID, time.Now(), time.Now(), nil)
	if err != nil {
		t.Fatalf("Failed to insert note: %v", err)
	}

	_, err = db.Exec("INSERT INTO note_todos (note_id, todo_id) VALUES (?, ?)", noteID, todoID)
	if err != nil {
		t.Fatalf("Failed to link note and todo: %v", err)
	}

	req, _ := http.NewRequest("GET", "/todos", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var todos []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &todos)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	assert.Len(t, todos, 1)
	if len(todos) > 0 {
		linkedNotes := todos[0]["linked_notes"].([]interface{})
		assert.Len(t, linkedNotes, 1)
		
		if len(linkedNotes) > 0 {
			firstNote := linkedNotes[0].(map[string]interface{})
			assert.Equal(t, "note-1", firstNote["id"])
			assert.Equal(t, "Meeting Notes", firstNote["title"])
		}
	}
}

// Mock Row for testing errors (advanced) - simplified for this scope
type errRow struct{}
func (r *errRow) Scan(dest ...interface{}) error { return errors.New("mock scan error") }
