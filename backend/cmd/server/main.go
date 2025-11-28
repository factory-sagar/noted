package main

import (
	"log"
	"os"

	"github.com/factory-sagar/notes-droid/backend/internal/db"
	"github.com/factory-sagar/notes-droid/backend/internal/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	database, err := db.Initialize("./data/notes.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := db.Migrate(database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize handlers
	h := handlers.New(database)

	// Setup Gin router
	router := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173", "http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(config))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := router.Group("/api")
	{
		// Accounts
		api.GET("/accounts", h.GetAccounts)
		api.GET("/accounts/:id", h.GetAccount)
		api.POST("/accounts", h.CreateAccount)
		api.PUT("/accounts/:id", h.UpdateAccount)
		api.DELETE("/accounts/:id", h.DeleteAccount)

		// Notes
		api.GET("/notes", h.GetNotes)
		api.GET("/notes/:id", h.GetNote)
		api.POST("/notes", h.CreateNote)
		api.PUT("/notes/:id", h.UpdateNote)
		api.DELETE("/notes/:id", h.DeleteNote)
		api.GET("/accounts/:id/notes", h.GetNotesByAccount)

		// Todos
		api.GET("/todos", h.GetTodos)
		api.GET("/todos/:id", h.GetTodo)
		api.POST("/todos", h.CreateTodo)
		api.PUT("/todos/:id", h.UpdateTodo)
		api.DELETE("/todos/:id", h.DeleteTodo)
		api.POST("/todos/:id/notes/:noteId", h.LinkTodoToNote)
		api.DELETE("/todos/:id/notes/:noteId", h.UnlinkTodoFromNote)

		// Search
		api.GET("/search", h.Search)

		// Analytics
		api.GET("/analytics", h.GetAnalytics)
		api.GET("/analytics/incomplete", h.GetIncompleteFields)

		// PDF Export
		api.GET("/notes/:id/export", h.ExportNotePDF)
	}

	// Get port from environment or default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
