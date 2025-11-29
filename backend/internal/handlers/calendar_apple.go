package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/factory-sagar/notes-droid/backend/internal/calendar"
	"github.com/gin-gonic/gin"
)

// Apple Calendar handlers using EventKit

func (h *Handler) GetAppleCalendarStatus(c *gin.Context) {
	hasAccess := calendar.CheckAccess()
	c.JSON(http.StatusOK, gin.H{
		"connected": hasAccess,
		"type":      "apple",
	})
}

func (h *Handler) RequestAppleCalendarAccess(c *gin.Context) {
	result, err := calendar.RequestAccess()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result == "granted" {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Calendar access granted",
		})
	} else if result == "denied" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Calendar access denied. Please enable in System Settings > Privacy & Security > Calendars",
		})
	} else {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": result,
		})
	}
}

func (h *Handler) GetAppleCalendars(c *gin.Context) {
	if !calendar.CheckAccess() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Calendar access not granted"})
		return
	}

	calendars, err := calendar.GetCalendars()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, calendars)
}

func (h *Handler) GetAppleCalendarEvents(c *gin.Context) {
	if !calendar.CheckAccess() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Calendar access not granted"})
		return
	}

	startStr := c.Query("start")
	endStr := c.Query("end")
	calendarID := c.Query("calendar_id")

	// Default to this week + 4 weeks if not specified
	if startStr == "" {
		now := time.Now()
		weekStart := now.AddDate(0, 0, -int(now.Weekday()))
		startStr = weekStart.Format(time.RFC3339)
	}
	if endStr == "" {
		start, _ := time.Parse(time.RFC3339, startStr)
		if start.IsZero() {
			start = time.Now()
		}
		endStr = start.AddDate(0, 0, 28).Format(time.RFC3339)
	}

	events, err := calendar.GetEvents(startStr, endStr, calendarID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to the format the frontend expects
	result := make([]CalendarEvent, 0, len(events))
	for _, e := range events {
		result = append(result, CalendarEvent{
			ID:          e.ID,
			Title:       e.Title,
			Description: e.Description,
			StartTime:   e.StartTime,
			EndTime:     e.EndTime,
			Attendees:   e.Attendees,
			MeetLink:    e.URL, // Use URL field for meeting link
		})
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetAppleCalendarEvent(c *gin.Context) {
	if !calendar.CheckAccess() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Calendar access not granted"})
		return
	}

	eventID := c.Param("eventId")
	event, err := calendar.GetEvent(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if event == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	c.JSON(http.StatusOK, CalendarEvent{
		ID:          event.ID,
		Title:       event.Title,
		Description: event.Description,
		StartTime:   event.StartTime,
		EndTime:     event.EndTime,
		Attendees:   event.Attendees,
		MeetLink:    event.URL,
	})
}

// ParseParticipantsApple categorizes attendees into internal and external
func (h *Handler) ParseParticipantsApple(c *gin.Context) {
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
		internalDomain = GetInternalDomain()
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
