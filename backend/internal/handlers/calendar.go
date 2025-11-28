package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
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
}

func (h *Handler) getOAuthConfig() *oauth2.Config {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	if redirectURL == "" {
		redirectURL = "http://localhost:8080/api/calendar/callback"
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar.readonly",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
}

func (h *Handler) GetCalendarAuthURL(c *gin.Context) {
	config := h.getOAuthConfig()

	if config.ClientID == "" || config.ClientSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Google OAuth not configured",
			"message": "Set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET environment variables",
		})
		return
	}

	state := "se-notes-calendar" // In production, use a random state
	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *Handler) HandleCalendarCallback(c *gin.Context) {
	config := h.getOAuthConfig()

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No authorization code"})
		return
	}

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token: " + err.Error()})
		return
	}

	// Store token in database
	tokenJSON, _ := json.Marshal(token)

	_, err = h.db.Exec(`
		INSERT OR REPLACE INTO settings (key, value) 
		VALUES ('google_oauth_token', ?)
	`, string(tokenJSON))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store token"})
		return
	}

	// Redirect to frontend settings page
	c.Redirect(http.StatusFound, "http://localhost:5173/settings?calendar=connected")
}

func (h *Handler) GetCalendarConfig(c *gin.Context) {
	var tokenStr string
	err := h.db.QueryRow(`SELECT value FROM settings WHERE key = 'google_oauth_token'`).Scan(&tokenStr)

	if err != nil || tokenStr == "" {
		c.JSON(http.StatusOK, CalendarConfig{Connected: false})
		return
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenStr), &token); err != nil {
		c.JSON(http.StatusOK, CalendarConfig{Connected: false})
		return
	}

	// Check if token is valid
	if !token.Valid() && token.RefreshToken == "" {
		c.JSON(http.StatusOK, CalendarConfig{Connected: false})
		return
	}

	// Get user email from stored token info
	var email string
	h.db.QueryRow(`SELECT value FROM settings WHERE key = 'google_user_email'`).Scan(&email)

	c.JSON(http.StatusOK, CalendarConfig{
		Connected: true,
		Email:     email,
	})
}

func (h *Handler) DisconnectCalendar(c *gin.Context) {
	h.db.Exec(`DELETE FROM settings WHERE key IN ('google_oauth_token', 'google_user_email')`)
	c.JSON(http.StatusOK, gin.H{"message": "Calendar disconnected"})
}

func (h *Handler) getCalendarService() (*calendar.Service, error) {
	var tokenStr string
	err := h.db.QueryRow(`SELECT value FROM settings WHERE key = 'google_oauth_token'`).Scan(&tokenStr)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenStr), &token); err != nil {
		return nil, err
	}

	config := h.getOAuthConfig()
	client := config.Client(context.Background(), &token)

	// If token was refreshed, save the new one
	newToken, err := config.TokenSource(context.Background(), &token).Token()
	if err == nil && newToken.AccessToken != token.AccessToken {
		tokenJSON, _ := json.Marshal(newToken)
		h.db.Exec(`UPDATE settings SET value = ? WHERE key = 'google_oauth_token'`, string(tokenJSON))
	}

	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return srv, nil
}

func (h *Handler) GetCalendarEvents(c *gin.Context) {
	srv, err := h.getCalendarService()
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Calendar not connected"})
		return
	}

	// Get time range from query params or default to this week
	startStr := c.Query("start")
	endStr := c.Query("end")

	var timeMin, timeMax time.Time
	if startStr != "" {
		timeMin, _ = time.Parse(time.RFC3339, startStr)
	} else {
		// Default: start of this week
		now := time.Now()
		timeMin = now.AddDate(0, 0, -int(now.Weekday()))
		timeMin = time.Date(timeMin.Year(), timeMin.Month(), timeMin.Day(), 0, 0, 0, 0, timeMin.Location())
	}

	if endStr != "" {
		timeMax, _ = time.Parse(time.RFC3339, endStr)
	} else {
		// Default: 4 weeks from start
		timeMax = timeMin.AddDate(0, 0, 28)
	}

	events, err := srv.Events.List("primary").
		TimeMin(timeMin.Format(time.RFC3339)).
		TimeMax(timeMax.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		MaxResults(100).
		Do()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events: " + err.Error()})
		return
	}

	// Convert to our format
	result := make([]CalendarEvent, 0)
	for _, item := range events.Items {
		if item.Start == nil {
			continue
		}

		event := CalendarEvent{
			ID:          item.Id,
			Title:       item.Summary,
			Description: item.Description,
		}

		// Handle all-day vs timed events
		if item.Start.DateTime != "" {
			event.StartTime = item.Start.DateTime
		} else {
			event.StartTime = item.Start.Date
		}

		if item.End != nil {
			if item.End.DateTime != "" {
				event.EndTime = item.End.DateTime
			} else {
				event.EndTime = item.End.Date
			}
		}

		// Get attendees
		for _, attendee := range item.Attendees {
			if attendee.Email != "" {
				event.Attendees = append(event.Attendees, attendee.Email)
			}
		}

		// Get meet link if available
		if item.ConferenceData != nil && item.ConferenceData.EntryPoints != nil {
			for _, ep := range item.ConferenceData.EntryPoints {
				if ep.EntryPointType == "video" {
					event.MeetLink = ep.Uri
					break
				}
			}
		}

		result = append(result, event)
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetCalendarEvent(c *gin.Context) {
	eventID := c.Param("eventId")

	srv, err := h.getCalendarService()
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Calendar not connected"})
		return
	}

	event, err := srv.Events.Get("primary", eventID).Do()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	result := CalendarEvent{
		ID:          event.Id,
		Title:       event.Summary,
		Description: event.Description,
	}

	if event.Start != nil {
		if event.Start.DateTime != "" {
			result.StartTime = event.Start.DateTime
		} else {
			result.StartTime = event.Start.Date
		}
	}

	if event.End != nil {
		if event.End.DateTime != "" {
			result.EndTime = event.End.DateTime
		} else {
			result.EndTime = event.End.Date
		}
	}

	for _, attendee := range event.Attendees {
		if attendee.Email != "" {
			result.Attendees = append(result.Attendees, attendee.Email)
		}
	}

	c.JSON(http.StatusOK, result)
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
