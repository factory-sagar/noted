package calendar

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework EventKit -framework AppKit -framework CoreGraphics

#import <Foundation/Foundation.h>
#import <EventKit/EventKit.h>

static EKEventStore *eventStore = nil;

const char* requestCalendarAccess() {
    if (eventStore == nil) {
        eventStore = [[EKEventStore alloc] init];
    }

    __block const char* result = NULL;
    dispatch_semaphore_t semaphore = dispatch_semaphore_create(0);

    if (@available(macOS 14.0, *)) {
        [eventStore requestFullAccessToEventsWithCompletion:^(BOOL granted, NSError *error) {
            if (granted) {
                result = "granted";
            } else if (error) {
                result = [[error localizedDescription] UTF8String];
            } else {
                result = "denied";
            }
            dispatch_semaphore_signal(semaphore);
        }];
    } else {
        [eventStore requestAccessToEntityType:EKEntityTypeEvent completion:^(BOOL granted, NSError *error) {
            if (granted) {
                result = "granted";
            } else if (error) {
                result = [[error localizedDescription] UTF8String];
            } else {
                result = "denied";
            }
            dispatch_semaphore_signal(semaphore);
        }];
    }

    dispatch_semaphore_wait(semaphore, DISPATCH_TIME_FOREVER);
    return result ? strdup(result) : strdup("unknown");
}

int checkCalendarAccess() {
    EKAuthorizationStatus status = [EKEventStore authorizationStatusForEntityType:EKEntityTypeEvent];
    if (@available(macOS 14.0, *)) {
        return (status == EKAuthorizationStatusFullAccess) ? 1 : 0;
    }
    return (status == EKAuthorizationStatusAuthorized) ? 1 : 0;
}

const char* getCalendars() {
    if (eventStore == nil) {
        eventStore = [[EKEventStore alloc] init];
    }

    // Refresh from remote sources (CalDAV, Exchange, etc.)
    [eventStore refreshSourcesIfNecessary];

    NSArray<EKCalendar *> *calendars = [eventStore calendarsForEntityType:EKEntityTypeEvent];
    NSMutableArray *calendarList = [NSMutableArray array];

    for (EKCalendar *cal in calendars) {
        NSString *colorHex = @"#6b7280"; // Default gray
        CGColorRef cgColor = cal.CGColor;
        if (cgColor) {
            const CGFloat *components = CGColorGetComponents(cgColor);
            size_t numComponents = CGColorGetNumberOfComponents(cgColor);
            if (numComponents >= 3) {
                colorHex = [NSString stringWithFormat:@"#%02X%02X%02X",
                    (int)(components[0] * 255),
                    (int)(components[1] * 255),
                    (int)(components[2] * 255)];
            }
        }
        NSDictionary *calDict = @{
            @"id": cal.calendarIdentifier,
            @"title": cal.title,
            @"color": colorHex,
            @"type": cal.source.title ?: @"Local"
        };
        [calendarList addObject:calDict];
    }

    NSError *error;
    NSData *jsonData = [NSJSONSerialization dataWithJSONObject:calendarList options:0 error:&error];
    if (error) {
        return strdup("[]");
    }

    NSString *jsonString = [[NSString alloc] initWithData:jsonData encoding:NSUTF8StringEncoding];
    return strdup([jsonString UTF8String]);
}

const char* getEvents(const char* startDate, const char* endDate, const char* calendarId) {
    if (eventStore == nil) {
        eventStore = [[EKEventStore alloc] init];
    }

    // Refresh from remote sources (CalDAV, Exchange, etc.)
    [eventStore refreshSourcesIfNecessary];

    NSDateFormatter *formatter = [[NSDateFormatter alloc] init];
    [formatter setDateFormat:@"yyyy-MM-dd'T'HH:mm:ssZ"];

    NSDate *start = [formatter dateFromString:[NSString stringWithUTF8String:startDate]];
    NSDate *end = [formatter dateFromString:[NSString stringWithUTF8String:endDate]];

    if (!start || !end) {
        // Fallback: try ISO8601
        NSISO8601DateFormatter *isoFormatter = [[NSISO8601DateFormatter alloc] init];
        start = [isoFormatter dateFromString:[NSString stringWithUTF8String:startDate]];
        end = [isoFormatter dateFromString:[NSString stringWithUTF8String:endDate]];
    }

    if (!start) {
        start = [NSDate date];
    }
    if (!end) {
        end = [start dateByAddingTimeInterval:60*60*24*28]; // 4 weeks
    }

    NSArray<EKCalendar *> *calendars = nil;
    if (calendarId && strlen(calendarId) > 0) {
        NSString *calId = [NSString stringWithUTF8String:calendarId];
        EKCalendar *cal = [eventStore calendarWithIdentifier:calId];
        if (cal) {
            calendars = @[cal];
        }
    }

    if (!calendars) {
        calendars = [eventStore calendarsForEntityType:EKEntityTypeEvent];
    }

    NSPredicate *predicate = [eventStore predicateForEventsWithStartDate:start endDate:end calendars:calendars];
    NSArray<EKEvent *> *events = [eventStore eventsMatchingPredicate:predicate];

    NSMutableArray *eventList = [NSMutableArray array];
    NSDateFormatter *outputFormatter = [[NSDateFormatter alloc] init];
    [outputFormatter setDateFormat:@"yyyy-MM-dd'T'HH:mm:ssZ"];

    for (EKEvent *event in events) {
        NSMutableArray *attendees = [NSMutableArray array];
        for (EKParticipant *participant in event.attendees) {
            if (participant.URL) {
                NSString *email = [[participant.URL absoluteString] stringByReplacingOccurrencesOfString:@"mailto:" withString:@""];
                [attendees addObject:email];
            }
        }

        NSDictionary *eventDict = @{
            @"id": event.eventIdentifier ?: @"",
            @"title": event.title ?: @"(No Title)",
            @"description": event.notes ?: @"",
            @"start_time": [outputFormatter stringFromDate:event.startDate] ?: @"",
            @"end_time": [outputFormatter stringFromDate:event.endDate] ?: @"",
            @"location": event.location ?: @"",
            @"all_day": @(event.allDay),
            @"calendar_id": event.calendar.calendarIdentifier ?: @"",
            @"calendar_title": event.calendar.title ?: @"",
            @"attendees": attendees,
            @"url": event.URL ? [event.URL absoluteString] : @""
        };
        [eventList addObject:eventDict];
    }

    // Sort by start time
    [eventList sortUsingComparator:^NSComparisonResult(NSDictionary *a, NSDictionary *b) {
        return [a[@"start_time"] compare:b[@"start_time"]];
    }];

    NSError *error;
    NSData *jsonData = [NSJSONSerialization dataWithJSONObject:eventList options:0 error:&error];
    if (error) {
        return strdup("[]");
    }

    NSString *jsonString = [[NSString alloc] initWithData:jsonData encoding:NSUTF8StringEncoding];
    return strdup([jsonString UTF8String]);
}

const char* getEvent(const char* eventId) {
    if (eventStore == nil) {
        eventStore = [[EKEventStore alloc] init];
    }

    NSString *eid = [NSString stringWithUTF8String:eventId];
    EKEvent *event = [eventStore eventWithIdentifier:eid];

    if (!event) {
        return strdup("null");
    }

    NSDateFormatter *formatter = [[NSDateFormatter alloc] init];
    [formatter setDateFormat:@"yyyy-MM-dd'T'HH:mm:ssZ"];

    NSMutableArray *attendees = [NSMutableArray array];
    for (EKParticipant *participant in event.attendees) {
        if (participant.URL) {
            NSString *email = [[participant.URL absoluteString] stringByReplacingOccurrencesOfString:@"mailto:" withString:@""];
            [attendees addObject:email];
        }
    }

    NSDictionary *eventDict = @{
        @"id": event.eventIdentifier ?: @"",
        @"title": event.title ?: @"(No Title)",
        @"description": event.notes ?: @"",
        @"start_time": [formatter stringFromDate:event.startDate] ?: @"",
        @"end_time": [formatter stringFromDate:event.endDate] ?: @"",
        @"location": event.location ?: @"",
        @"all_day": @(event.allDay),
        @"calendar_id": event.calendar.calendarIdentifier ?: @"",
        @"calendar_title": event.calendar.title ?: @"",
        @"attendees": attendees,
        @"url": event.URL ? [event.URL absoluteString] : @""
    };

    NSError *error;
    NSData *jsonData = [NSJSONSerialization dataWithJSONObject:eventDict options:0 error:&error];
    if (error) {
        return strdup("null");
    }

    NSString *jsonString = [[NSString alloc] initWithData:jsonData encoding:NSUTF8StringEncoding];
    return strdup([jsonString UTF8String]);
}

void freeString(char* s) {
    free(s);
}
*/
import "C"
import (
	"encoding/json"
	"unsafe"
)

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

func RequestAccess() (string, error) {
	result := C.requestCalendarAccess()
	defer C.freeString(result)
	return C.GoString(result), nil
}

func CheckAccess() bool {
	return C.checkCalendarAccess() == 1
}

func GetCalendars() ([]CalendarInfo, error) {
	result := C.getCalendars()
	defer C.freeString(result)

	var calendars []CalendarInfo
	err := json.Unmarshal([]byte(C.GoString(result)), &calendars)
	return calendars, err
}

func GetEvents(startDate, endDate, calendarID string) ([]EventInfo, error) {
	cStart := C.CString(startDate)
	cEnd := C.CString(endDate)
	cCalID := C.CString(calendarID)
	defer C.free(unsafe.Pointer(cStart))
	defer C.free(unsafe.Pointer(cEnd))
	defer C.free(unsafe.Pointer(cCalID))

	result := C.getEvents(cStart, cEnd, cCalID)
	defer C.freeString(result)

	var events []EventInfo
	err := json.Unmarshal([]byte(C.GoString(result)), &events)
	return events, err
}

func GetEvent(eventID string) (*EventInfo, error) {
	cEventID := C.CString(eventID)
	defer C.free(unsafe.Pointer(cEventID))

	result := C.getEvent(cEventID)
	defer C.freeString(result)

	jsonStr := C.GoString(result)
	if jsonStr == "null" {
		return nil, nil
	}

	var event EventInfo
	err := json.Unmarshal([]byte(jsonStr), &event)
	return &event, err
}
