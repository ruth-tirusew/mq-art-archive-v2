package events

// ListFilter selects events for List. Zero value returns public discoverable upcoming events.
type ListFilter struct {
	Status       *EventStatus
	Statuses      []EventStatus
	UpcomingOnly bool
	EventType    string
	Query        string
	Limit        int
	Offset       int
}

// PublicDiscoverableFilter returns approved and pending events (excludes rejected).
func PublicDiscoverableFilter() ListFilter {
	return ListFilter{
		Statuses:      []EventStatus{EventStatusApproved, EventStatusPending},
		UpcomingOnly: true,
	}
}

// PublicUpcomingFilter returns approved events with starts_at in the future.
// Deprecated for public discovery; prefer PublicDiscoverableFilter.
func PublicUpcomingFilter() ListFilter {
	status := EventStatusApproved
	return ListFilter{
		Status:       &status,
		UpcomingOnly: true,
	}
}

// PendingFilter returns events awaiting admin review.
func PendingFilter() ListFilter {
	status := EventStatusPending
	return ListFilter{
		Status: &status,
	}
}
