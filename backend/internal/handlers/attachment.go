package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/factory-sagar/notes-droid/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) GetAttachments(c *gin.Context) {
	noteID := c.Param("id")

	rows, err := h.db.Query(`
		SELECT id, note_id, filename, original_name, mime_type, size, created_at
		FROM attachments
		WHERE note_id = ?
		ORDER BY created_at DESC
	`, noteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	attachments := []models.Attachment{}
	for rows.Next() {
		var a models.Attachment
		rows.Scan(&a.ID, &a.NoteID, &a.Filename, &a.OriginalName, &a.MimeType, &a.Size, &a.CreatedAt)
		attachments = append(attachments, a)
	}

	c.JSON(http.StatusOK, attachments)
}

func (h *Handler) UploadAttachment(c *gin.Context) {
	noteID := c.Param("id")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	id := uuid.New().String()
	filename := id + "_" + file.Filename

	// Save file to uploads directory
	uploadPath := filepath.Join(h.uploadsDir, filename)
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	now := time.Now()
	_, err = h.db.Exec(`
		INSERT INTO attachments (id, note_id, filename, original_name, mime_type, size, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, noteID, filename, file.Filename, file.Header.Get("Content-Type"), file.Size, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.Attachment{
		ID:           id,
		NoteID:       noteID,
		Filename:     filename,
		OriginalName: file.Filename,
		MimeType:     file.Header.Get("Content-Type"),
		Size:         file.Size,
		CreatedAt:    now,
	})
}

func (h *Handler) DeleteAttachment(c *gin.Context) {
	id := c.Param("attachmentId")

	var filename string
	err := h.db.QueryRow("SELECT filename FROM attachments WHERE id = ?", id).Scan(&filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attachment not found"})
		return
	}

	// Delete from database
	h.db.Exec("DELETE FROM attachments WHERE id = ?", id)

	// Delete file from disk
	filePath := filepath.Join(h.uploadsDir, filename)
	os.Remove(filePath) // Ignore error - file may already be deleted

	c.JSON(http.StatusOK, gin.H{"message": "Attachment deleted"})
}
