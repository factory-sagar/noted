package handlers

import (
	"net/http"
	"time"

	"github.com/factory-sagar/notes-droid/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) GetActivities(c *gin.Context) {
	accountID := c.Param("id")
	limit := c.DefaultQuery("limit", "50")

	rows, err := h.db.Query(`
		SELECT id, account_id, type, title, description, entity_type, entity_id, created_at
		FROM activities
		WHERE account_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, accountID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	activities := []models.Activity{}
	for rows.Next() {
		var a models.Activity
		rows.Scan(&a.ID, &a.AccountID, &a.Type, &a.Title, &a.Description, &a.EntityType, &a.EntityID, &a.CreatedAt)
		activities = append(activities, a)
	}

	c.JSON(http.StatusOK, activities)
}

func (h *Handler) CreateActivity(c *gin.Context) {
	var req models.CreateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()
	now := time.Now()

	_, err := h.db.Exec(`
		INSERT INTO activities (id, account_id, type, title, description, entity_type, entity_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, id, req.AccountID, req.Type, req.Title, req.Description, req.EntityType, req.EntityID, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.Activity{
		ID:          id,
		AccountID:   req.AccountID,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		EntityType:  req.EntityType,
		EntityID:    req.EntityID,
		CreatedAt:   now,
	})
}

// LogActivity is a helper to log activities from other handlers
func (h *Handler) LogActivity(accountID, actType, title, description, entityType, entityID string) {
	id := uuid.New().String()
	h.db.Exec(`
		INSERT INTO activities (id, account_id, type, title, description, entity_type, entity_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, id, accountID, actType, title, description, entityType, entityID, time.Now())
}
