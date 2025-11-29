//go:build !darwin

package calendar

import "errors"

var ErrNotSupported = errors.New("calendar integration only supported on macOS")

func CheckAccess() bool {
	return false
}

func RequestAccess() (string, error) {
	return "", ErrNotSupported
}

func GetCalendars() ([]CalendarInfo, error) {
	return nil, ErrNotSupported
}

func GetEvents(startDate, endDate, calendarID string) ([]EventInfo, error) {
	return nil, ErrNotSupported
}

func GetEvent(eventID string) (*EventInfo, error) {
	return nil, ErrNotSupported
}

func RefreshSources() error {
	return ErrNotSupported
}
