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
	// Create uploads directory
	if err := os.MkdirAll("./data/uploads", 0755); err != nil {
		log.Printf("Warning: Could not create uploads directory: %v", err)
	}

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

	// Static file serving for uploads
	router.Static("/uploads", "./data/uploads")

	// API routes
	api := router.Group("/api")
	{
		// Accounts
		api.GET("/accounts", h.GetAccounts)
		api.GET("/accounts/deleted", h.GetDeletedAccounts)
		api.GET("/accounts/:id", h.GetAccount)
		api.POST("/accounts", h.CreateAccount)
		api.PUT("/accounts/:id", h.UpdateAccount)
		api.DELETE("/accounts/:id", h.DeleteAccount)
		api.POST("/accounts/:id/restore", h.RestoreAccount)
		api.DELETE("/accounts/:id/permanent", h.PermanentDeleteAccount)

		// Notes - archived must come before :id to avoid route capture
		api.GET("/notes/archived", h.GetArchivedNotes)
		api.GET("/notes/deleted", h.GetDeletedNotes)
		api.GET("/notes", h.GetNotes)
		api.GET("/notes/:id", h.GetNote)
		api.POST("/notes", h.CreateNote)
		api.PUT("/notes/:id", h.UpdateNote)
		api.DELETE("/notes/:id", h.DeleteNote)
		api.POST("/notes/:id/restore", h.RestoreNote)
		api.DELETE("/notes/:id/permanent", h.PermanentDeleteNote)
		api.GET("/accounts/:id/notes", h.GetNotesByAccount)

		// Todos
		api.GET("/todos", h.GetTodos)
		api.GET("/todos/deleted", h.GetDeletedTodos)
		api.GET("/todos/:id", h.GetTodo)
		api.POST("/todos", h.CreateTodo)
		api.PUT("/todos/:id", h.UpdateTodo)
		api.DELETE("/todos/:id", h.DeleteTodo)
		api.POST("/todos/:id/restore", h.RestoreTodo)
		api.DELETE("/todos/:id/permanent", h.PermanentDeleteTodo)
		api.POST("/todos/:id/notes/:noteId", h.LinkTodoToNote)
		api.DELETE("/todos/:id/notes/:noteId", h.UnlinkTodoFromNote)

		// Search
		api.GET("/search", h.Search)

		// Analytics
		api.GET("/analytics", h.GetAnalytics)
		api.GET("/analytics/incomplete", h.GetIncompleteFields)

		// Data management
		api.GET("/export", h.ExportAllData)
		api.DELETE("/data", h.ClearAllData)

		// PDF Export
		api.GET("/notes/:id/export", h.ExportNotePDF)

		// Markdown Import/Export
		api.POST("/import/markdown", h.ImportMarkdown)
		api.GET("/notes/:id/export/markdown", h.ExportMarkdown)

		// Calendar
		api.GET("/calendar/auth", h.CalendarAuthHandler) // Renamed from GetCalendarAuthURL
		api.GET("/calendar/callback", h.HandleCalendarCallback)
		api.GET("/calendar/config", h.GetCalendarConfig)
		api.POST("/calendar/connect", h.ConnectCalendar)
		api.DELETE("/calendar/disconnect", h.DisconnectCalendar)
		api.GET("/calendar/events", h.GetCalendarEvents)
		api.GET("/calendar/events/:eventId", h.GetCalendarEvent)
		api.POST("/calendar/parse-participants", h.ParseParticipants)

		// Tags
		api.GET("/tags", h.GetTags)
		api.POST("/tags", h.CreateTag)
		api.PUT("/tags/:id", h.UpdateTag)
		api.DELETE("/tags/:id", h.DeleteTag)
		api.GET("/notes/:id/tags", h.GetNoteTags)
		api.POST("/notes/:id/tags/:tagId", h.AddTagToNote)
		api.DELETE("/notes/:id/tags/:tagId", h.RemoveTagFromNote)

		// Activities
		api.GET("/accounts/:id/activities", h.GetActivities)
		api.POST("/activities", h.CreateActivity)

		// Attachments
		api.GET("/notes/:id/attachments", h.GetAttachments)
		api.POST("/notes/:id/attachments", h.UploadAttachment)
		api.DELETE("/notes/:id/attachments/:attachmentId", h.DeleteAttachment)

		// Reorder notes
		api.POST("/accounts/:id/notes/reorder", h.ReorderNotes)

		// Quick capture
		api.POST("/quick-capture", h.QuickCapture)

		// Pin/Archive
		api.POST("/notes/:id/pin", h.ToggleNotePin)
		api.POST("/notes/:id/archive", h.ToggleNoteArchive)
		api.POST("/todos/:id/pin", h.ToggleTodoPin)

		// Contacts
		api.GET("/contacts", h.GetContacts)
		api.GET("/contacts/stats", h.GetContactStats)
		api.GET("/contacts/:id", h.GetContact)
		api.POST("/contacts", h.CreateContact)
		api.PUT("/contacts/:id", h.UpdateContact)
		api.DELETE("/contacts/:id", h.DeleteContact)
		api.POST("/contacts/:id/confirm-suggestion", h.ConfirmAccountSuggestion)
		api.POST("/contacts/:id/link/:accountId", h.LinkContactToAccount)
		api.GET("/contacts/:id/notes", h.GetContactNotes)
		api.POST("/contacts/bulk", h.BulkContactsOperation)
		api.GET("/contacts/domain-groups", h.GetContactDomainGroups)
		api.POST("/contacts/domain/:domain/link/:accountId", h.LinkDomainToAccount)
		api.POST("/contacts/domain/:domain/create-account", h.CreateAccountFromDomain)
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
