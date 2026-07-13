package events

// ListFilter selects events for List. Zero value returns approved upcoming events (public default).
type ListFilter struct {
	Status       *EventStatus
	UpcomingOnly bool
	EventType    string
	Query        string
	Limit        int
	Offset       int
}

// PublicUpcomingFilter returns approved events with starts_at in the future.
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
