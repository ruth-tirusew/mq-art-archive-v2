package outbound

import "context"

type EventNotificationRepository interface {
	ListEventSummaryRecipients(ctx context.Context) ([]string, error)
}
