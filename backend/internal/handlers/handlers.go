package handlers

import (
	"database/sql"
)

// Handler holds database connection and provides HTTP handlers
type Handler struct {
	db         *sql.DB
	uploadsDir string
}

// New creates a new Handler
func New(db *sql.DB) *Handler {
	return &Handler{db: db, uploadsDir: "./data/uploads"}
}

// NewWithUploadsDir creates a new Handler with custom uploads directory
func NewWithUploadsDir(db *sql.DB, uploadsDir string) *Handler {
	return &Handler{db: db, uploadsDir: uploadsDir}
}
