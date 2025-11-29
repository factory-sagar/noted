package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/factory-sagar/notes-droid/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) GetAccounts(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, name, account_owner, budget, est_engineers, created_at, updated_at 
		FROM accounts ORDER BY name ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	accounts := []models.Account{}
	for rows.Next() {
		var a models.Account
		var accountOwner sql.NullString
		if err := rows.Scan(&a.ID, &a.Name, &accountOwner, &a.Budget, &a.EstEngineers, &a.CreatedAt, &a.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if accountOwner.Valid {
			a.AccountOwner = accountOwner.String
		}
		accounts = append(accounts, a)
	}

	c.JSON(http.StatusOK, accounts)
}

func (h *Handler) GetAccount(c *gin.Context) {
	id := c.Param("id")
	var a models.Account
	var accountOwner sql.NullString
	err := h.db.QueryRow(`
		SELECT id, name, account_owner, budget, est_engineers, created_at, updated_at 
		FROM accounts WHERE id = ?
	`, id).Scan(&a.ID, &a.Name, &accountOwner, &a.Budget, &a.EstEngineers, &a.CreatedAt, &a.UpdatedAt)

	if accountOwner.Valid {
		a.AccountOwner = accountOwner.String
	}

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, a)
}

func (h *Handler) CreateAccount(c *gin.Context) {
	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account name is required"})
		return
	}

	id := uuid.New().String()
	now := time.Now()

	_, err := h.db.Exec(`
		INSERT INTO accounts (id, name, account_owner, budget, est_engineers, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, req.Name, req.AccountOwner, req.Budget, req.EstEngineers, now, now)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.Account{
		ID:           id,
		Name:         req.Name,
		AccountOwner: req.AccountOwner,
		Budget:       req.Budget,
		EstEngineers: req.EstEngineers,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}

func (h *Handler) UpdateAccount(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build dynamic update query
	updates := []string{}
	args := []interface{}{}

	if req.Name != nil {
		updates = append(updates, "name = ?")
		args = append(args, *req.Name)
	}
	if req.AccountOwner != nil {
		updates = append(updates, "account_owner = ?")
		args = append(args, *req.AccountOwner)
	}
	if req.Budget != nil {
		updates = append(updates, "budget = ?")
		args = append(args, *req.Budget)
	}
	if req.EstEngineers != nil {
		updates = append(updates, "est_engineers = ?")
		args = append(args, *req.EstEngineers)
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	updates = append(updates, "updated_at = ?")
	args = append(args, time.Now())
	args = append(args, id)

	query := "UPDATE accounts SET " + strings.Join(updates, ", ") + " WHERE id = ?"
	result, err := h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	// Return updated account
	h.GetAccount(c)
}

func (h *Handler) DeleteAccount(c *gin.Context) {
	id := c.Param("id")
	result, err := h.db.Exec("DELETE FROM accounts WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted"})
}
