package calendar

type CalendarInfo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Color string `json:"color"`
	Type  string `json:"type"`
}

type EventInfo struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	StartTime     string   `json:"start_time"`
	EndTime       string   `json:"end_time"`
	Location      string   `json:"location"`
	AllDay        bool     `json:"all_day"`
	CalendarID    string   `json:"calendar_id"`
	CalendarTitle string   `json:"calendar_title"`
	Attendees     []string `json:"attendees"`
	URL           string   `json:"url"`
}
