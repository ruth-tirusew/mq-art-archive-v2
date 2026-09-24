package outbound

import "context"

type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
	// SendHTML sends a message with both an HTML and a plain-text body, for clients that
	// render HTML and as a fallback for those that don't. textBody should be a complete
	// plain-text rendering, not a stripped copy of the HTML.
	SendHTML(ctx context.Context, to, subject, htmlBody, textBody string) error
}
