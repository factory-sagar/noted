package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Initialize creates and returns a database connection
func Initialize(dbPath string) (*sql.DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Open database with FTS5 support
	db, err := sql.Open("sqlite3", dbPath+"?_fk=on")
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// columnExists checks if a column exists in a table
func columnExists(db *sql.DB, table, column string) bool {
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue sql.NullString
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			continue
		}
		if name == column {
			return true
		}
	}
	return false
}

// Migrate runs database migrations
func Migrate(db *sql.DB) error {
	migrations := []string{
		// Accounts table
		`CREATE TABLE IF NOT EXISTS accounts (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			account_owner TEXT,
			budget REAL,
			est_engineers INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Notes table
		`CREATE TABLE IF NOT EXISTS notes (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			account_id TEXT NOT NULL,
			template_type TEXT DEFAULT 'initial',
			internal_participants TEXT DEFAULT '[]',
			external_participants TEXT DEFAULT '[]',
			content TEXT DEFAULT '',
			meeting_id TEXT,
			meeting_date DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
		)`,

		// Todos table
		`CREATE TABLE IF NOT EXISTS todos (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT DEFAULT '',
			status TEXT DEFAULT 'not_started',
			priority TEXT DEFAULT 'medium',
			due_date DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Note-Todo junction table (many-to-many)
		`CREATE TABLE IF NOT EXISTS note_todos (
			note_id TEXT NOT NULL,
			todo_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (note_id, todo_id),
			FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE,
			FOREIGN KEY (todo_id) REFERENCES todos(id) ON DELETE CASCADE
		)`,

		// Full-text search for notes (using FTS4 for broader compatibility)
		`CREATE VIRTUAL TABLE IF NOT EXISTS notes_fts USING fts4(
			title,
			content,
			content='notes',
			tokenize=porter
		)`,

		// Triggers to keep FTS in sync
		`CREATE TRIGGER IF NOT EXISTS notes_ai AFTER INSERT ON notes BEGIN
			INSERT INTO notes_fts(docid, title, content) VALUES (NEW.rowid, NEW.title, NEW.content);
		END`,

		`CREATE TRIGGER IF NOT EXISTS notes_ad AFTER DELETE ON notes BEGIN
			DELETE FROM notes_fts WHERE docid = OLD.rowid;
		END`,

		`CREATE TRIGGER IF NOT EXISTS notes_au AFTER UPDATE ON notes BEGIN
			DELETE FROM notes_fts WHERE docid = OLD.rowid;
			INSERT INTO notes_fts(docid, title, content) VALUES (NEW.rowid, NEW.title, NEW.content);
		END`,

		// Settings table for OAuth tokens and preferences
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Tags table
		`CREATE TABLE IF NOT EXISTS tags (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			color TEXT DEFAULT '#6b7280',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Note-Tags junction table (many-to-many)
		`CREATE TABLE IF NOT EXISTS note_tags (
			note_id TEXT NOT NULL,
			tag_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (note_id, tag_id),
			FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
		)`,

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_notes_account_id ON notes(account_id)`,
		`CREATE INDEX IF NOT EXISTS idx_notes_meeting_date ON notes(meeting_date)`,
		`CREATE INDEX IF NOT EXISTS idx_todos_status ON todos(status)`,
		`CREATE INDEX IF NOT EXISTS idx_note_todos_note_id ON note_todos(note_id)`,
		`CREATE INDEX IF NOT EXISTS idx_note_todos_todo_id ON note_todos(todo_id)`,
		`CREATE INDEX IF NOT EXISTS idx_note_tags_note_id ON note_tags(note_id)`,
		`CREATE INDEX IF NOT EXISTS idx_note_tags_tag_id ON note_tags(tag_id)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}

	// Add account_id to todos (migration) - check if column exists first
	if !columnExists(db, "todos", "account_id") {
		if _, err := db.Exec(`ALTER TABLE todos ADD COLUMN account_id TEXT REFERENCES accounts(id)`); err != nil {
			return err
		}
	}
	// Create index for account_id
	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_todos_account_id ON todos(account_id)`); err != nil {
		return err
	}

	return nil
}
