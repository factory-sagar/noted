package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/factory-sagar/notes-droid/backend/internal/db"
	"github.com/factory-sagar/notes-droid/backend/internal/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type App struct {
	ctx      context.Context
	port     int
	shutdown chan struct{}
}

func NewApp() *App {
	return &App{
		shutdown: make(chan struct{}),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go a.startServer()
}

func (a *App) shutdown_(ctx context.Context) {
	close(a.shutdown)
}

func (a *App) GetServerPort() int {
	return a.port
}

func (a *App) GetDataDir() string {
	return getDataDir()
}

func getDataDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Warning: Could not get home directory: %v", err)
		return "./data"
	}

	dataDir := filepath.Join(homeDir, "Library", "Application Support", "Noted")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Printf("Warning: Could not create data directory: %v", err)
		return "./data"
	}

	return dataDir
}

func (a *App) startServer() {
	dataDir := getDataDir()
	uploadsDir := filepath.Join(dataDir, "uploads")

	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Printf("Warning: Could not create uploads directory: %v", err)
	}

	dbPath := filepath.Join(dataDir, "notes.db")
	database, err := db.Initialize(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := db.Migrate(database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	h := handlers.New(database)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(config))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.Static("/uploads", uploadsDir)

	api := router.Group("/api")
	{
		api.GET("/accounts", h.GetAccounts)
		api.GET("/accounts/:id", h.GetAccount)
		api.POST("/accounts", h.CreateAccount)
		api.PUT("/accounts/:id", h.UpdateAccount)
		api.DELETE("/accounts/:id", h.DeleteAccount)

		api.GET("/notes/archived", h.GetArchivedNotes)
		api.GET("/notes", h.GetNotes)
		api.GET("/notes/:id", h.GetNote)
		api.POST("/notes", h.CreateNote)
		api.PUT("/notes/:id", h.UpdateNote)
		api.DELETE("/notes/:id", h.DeleteNote)
		api.POST("/notes/:id/restore", h.RestoreNote)
		api.DELETE("/notes/:id/permanent", h.PermanentDeleteNote)
		api.GET("/notes/deleted", h.GetDeletedNotes)
		api.GET("/accounts/:id/notes", h.GetNotesByAccount)

		api.GET("/todos", h.GetTodos)
		api.GET("/todos/:id", h.GetTodo)
		api.POST("/todos", h.CreateTodo)
		api.PUT("/todos/:id", h.UpdateTodo)
		api.DELETE("/todos/:id", h.DeleteTodo)
		api.POST("/todos/:id/restore", h.RestoreTodo)
		api.DELETE("/todos/:id/permanent", h.PermanentDeleteTodo)
		api.GET("/todos/deleted", h.GetDeletedTodos)
		api.POST("/todos/:id/notes/:noteId", h.LinkTodoToNote)
		api.DELETE("/todos/:id/notes/:noteId", h.UnlinkTodoFromNote)

		api.GET("/search", h.Search)

		api.GET("/analytics", h.GetAnalytics)
		api.GET("/analytics/incomplete", h.GetIncompleteFields)

		api.GET("/notes/:id/export", h.ExportNotePDF)

		// Apple Calendar (EventKit) - native macOS integration
		api.GET("/calendar/config", h.GetAppleCalendarStatus)
		api.POST("/calendar/connect", h.RequestAppleCalendarAccess)
		api.GET("/calendar/calendars", h.GetAppleCalendars)
		api.GET("/calendar/events", h.GetAppleCalendarEvents)
		api.GET("/calendar/events/:eventId", h.GetAppleCalendarEvent)
		api.POST("/calendar/parse-participants", h.ParseParticipantsApple)

		api.GET("/tags", h.GetTags)
		api.POST("/tags", h.CreateTag)
		api.PUT("/tags/:id", h.UpdateTag)
		api.DELETE("/tags/:id", h.DeleteTag)
		api.GET("/notes/:id/tags", h.GetNoteTags)
		api.POST("/notes/:id/tags/:tagId", h.AddTagToNote)
		api.DELETE("/notes/:id/tags/:tagId", h.RemoveTagFromNote)

		api.GET("/accounts/:id/activities", h.GetActivities)
		api.POST("/activities", h.CreateActivity)

		api.GET("/notes/:id/attachments", h.GetAttachments)
		api.POST("/notes/:id/attachments", h.UploadAttachment)
		api.DELETE("/notes/:id/attachments/:attachmentId", h.DeleteAttachment)

		api.POST("/accounts/:id/notes/reorder", h.ReorderNotes)

		api.POST("/quick-capture", h.QuickCapture)

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
	}

	// Use fixed port 8080 for OAuth compatibility
	a.port = 8080
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", a.port))
	if err != nil {
		// If 8080 is busy, try a random port (OAuth won't work but app will run)
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatalf("Failed to find available port: %v", err)
		}
		a.port = listener.Addr().(*net.TCPAddr).Port
		log.Printf("Warning: Port 8080 unavailable, using %d. Google Calendar OAuth may not work.", a.port)
	}
	log.Printf("Internal server starting on port %d", a.port)

	go func() {
		<-a.shutdown
		listener.Close()
		database.Close()
	}()

	if err := router.RunListener(listener); err != nil {
		log.Printf("Server stopped: %v", err)
	}
}
