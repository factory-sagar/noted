package handlers

import (
	"net/http"
	"github.com/factory-sagar/notes-droid/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) GetTags(c *gin.Context) {
	rows, err := h.db.Query("SELECT id, name, color, created_at FROM tags ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	tags := []models.Tag{}
	for rows.Next() {
		var tag models.Tag
		rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt)
		tags = append(tags, tag)
	}

	c.JSON(http.StatusOK, tags)
}

func (h *Handler) CreateTag(c *gin.Context) {
	var req models.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()
	color := req.Color
	if color == "" {
		color = "#6b7280"
	}

	_, err := h.db.Exec(
		"INSERT INTO tags (id, name, color) VALUES (?, ?, ?)",
		id, req.Name, color,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tag := models.Tag{ID: id, Name: req.Name, Color: color}
	c.JSON(http.StatusCreated, tag)
}

func (h *Handler) UpdateTag(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		h.db.Exec("UPDATE tags SET name = ? WHERE id = ?", *req.Name, id)
	}
	if req.Color != nil {
		h.db.Exec("UPDATE tags SET color = ? WHERE id = ?", *req.Color, id)
	}

	var tag models.Tag
	h.db.QueryRow("SELECT id, name, color, created_at FROM tags WHERE id = ?", id).
		Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt)

	c.JSON(http.StatusOK, tag)
}

func (h *Handler) DeleteTag(c *gin.Context) {
	id := c.Param("id")
	_, err := h.db.Exec("DELETE FROM tags WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tag deleted"})
}

func (h *Handler) AddTagToNote(c *gin.Context) {
	noteID := c.Param("id")
	tagID := c.Param("tagId")

	_, err := h.db.Exec("INSERT OR IGNORE INTO note_tags (note_id, tag_id) VALUES (?, ?)", noteID, tagID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tag added to note"})
}

func (h *Handler) RemoveTagFromNote(c *gin.Context) {
	noteID := c.Param("id")
	tagID := c.Param("tagId")

	_, err := h.db.Exec("DELETE FROM note_tags WHERE note_id = ? AND tag_id = ?", noteID, tagID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tag removed from note"})
}

func (h *Handler) GetNoteTags(c *gin.Context) {
	noteID := c.Param("id")

	rows, err := h.db.Query(`
		SELECT t.id, t.name, t.color, t.created_at
		FROM tags t
		JOIN note_tags nt ON t.id = nt.tag_id
		WHERE nt.note_id = ?
		ORDER BY t.name
	`, noteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	tags := []models.Tag{}
	for rows.Next() {
		var tag models.Tag
		rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt)
		tags = append(tags, tag)
	}

	c.JSON(http.StatusOK, tags)
}
