package mailer

import (
	"context"
	"log"
)

// LogMailer is the fallback when no mail provider is configured. Messages here contain one-time
// links (password reset, email verification), so the body is only logged when verbose is set,
// which is local development. Anywhere else those links would end up in log storage that many
// people can read, and a reset link is a login.
type LogMailer struct{ verbose bool }

func NewLogMailer(verbose bool) *LogMailer {
	return &LogMailer{verbose: verbose}
}

func (m *LogMailer) Send(_ context.Context, to, subject, body string) error {
	if !m.verbose {
		log.Printf("[mailer] not sent (no mail provider configured): to=%s subject=%q", to, subject)
		return nil
	}
	log.Printf("[mailer] to=%s subject=%q body=%s", to, subject, body)
	return nil
}

func (m *LogMailer) SendHTML(_ context.Context, to, subject, htmlBody, textBody string) error {
	if !m.verbose {
		log.Printf("[mailer] not sent (no mail provider configured): to=%s subject=%q", to, subject)
		return nil
	}
	log.Printf("[mailer] to=%s subject=%q html_len=%d text=%s", to, subject, len(htmlBody), textBody)
	return nil
}
