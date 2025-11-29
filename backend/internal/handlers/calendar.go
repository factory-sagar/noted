package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/calendar/v3"
)

type CalendarEvent struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	StartTime   string   `json:"start_time"`
	EndTime     string   `json:"end_time"`
	Attendees   []string `json:"attendees"`
	MeetLink    string   `json:"meet_link,omitempty"`
}

type CalendarConfig struct {
	Connected bool   `json:"connected"`
	Email     string `json:"email,omitempty"`
	Type      string `json:"type,omitempty"` // "apple" or "google"
}

// Deprecated: getOAuthConfig is removed as we rely on Apple Calendar
func (h *Handler) getOAuthConfig() interface{} {
	return nil
}

func (h *Handler) CalendarAuthHandler(c *gin.Context) {
	// Apple Calendar integration does not require server-side OAuth flow initiation
	// The frontend handles the redirect/interaction directly or via Wails bridge
	c.JSON(http.StatusOK, gin.H{
		"message": "Use client-side integration for Apple Calendar",
	})
}

func (h *Handler) HandleCalendarCallback(c *gin.Context) {
	// Callback logic if needed for future integrations
	c.JSON(http.StatusOK, gin.H{"message": "Callback received"})
}

func (h *Handler) GetCalendarConfig(c *gin.Context) {
	var tokenStr string
	err := h.db.QueryRow(`SELECT value FROM settings WHERE key = 'google_oauth_token'`).Scan(&tokenStr)

	if err != nil || tokenStr == "" {
		c.JSON(http.StatusOK, CalendarConfig{Connected: false})
		return
	}

	var token interface{} // Changed from oauth2.Token to interface{} since oauth2 is removed
	if err := json.Unmarshal([]byte(tokenStr), &token); err != nil {
		c.JSON(http.StatusOK, CalendarConfig{Connected: false})
		return
	}

	// Mock validation for now
	// Check if token is valid
	// if !token.Valid() && token.RefreshToken == "" {
	// 	c.JSON(http.StatusOK, CalendarConfig{Connected: false})
	// 	return
	// }

	// Get user email from stored token info
	// In a real implementation we might get this from the system or a config file
	// for Apple Calendar, we might not have it unless we ask for it
	// For now, we'll just return connected status

	c.JSON(http.StatusOK, CalendarConfig{
		Connected: true,
		Email:     "", // Email not always available for local cal
	})
}

func (h *Handler) DisconnectCalendar(c *gin.Context) {
	// For Apple Calendar, we can't really "disconnect" since it's a system permission
	// But we can clear any stored preferences if we had them
	h.db.Exec(`DELETE FROM settings WHERE key IN ('apple_calendar_enabled')`)
	c.JSON(http.StatusOK, gin.H{"message": "Calendar disconnected"})
}

// ConnectCalendar is a fallback for non-Apple platforms (tells user to use OAuth)
func (h *Handler) ConnectCalendar(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"message": "Please use 'Connect Apple Calendar' button",
	})
}

func (h *Handler) getCalendarService() (*calendar.Service, error) {
	// Mock implementation for Apple Calendar
	// In a real app, this would interface with EventKit via Wails/Cgo
	return nil, nil
}

func (h *Handler) GetCalendarEvents(c *gin.Context) {
	// For Apple Calendar, we rely on the frontend/Wails bridge to fetch events
	// But if we were to implement a fallback or server-side fetching:

	// Return empty list for now as this is handled client-side
	c.JSON(http.StatusOK, []CalendarEvent{})
}

func (h *Handler) GetCalendarEvent(c *gin.Context) {
	// Similarly, handled client-side
	c.JSON(http.StatusNotFound, gin.H{"error": "Event not found via API"})
}

// ParseParticipants categorizes attendees into internal and external
func (h *Handler) ParseParticipants(c *gin.Context) {
	var req struct {
		Attendees      []string `json:"attendees"`
		InternalDomain string   `json:"internal_domain"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	internalDomain := req.InternalDomain
	if internalDomain == "" {
		internalDomain = "factory.ai"
	}

	internal := make([]string, 0)
	external := make([]string, 0)

	for _, email := range req.Attendees {
		email = strings.TrimSpace(strings.ToLower(email))
		if email == "" {
			continue
		}

		if strings.HasSuffix(email, "@"+internalDomain) {
			internal = append(internal, email)
		} else {
			external = append(external, email)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"internal": internal,
		"external": external,
	})
}
