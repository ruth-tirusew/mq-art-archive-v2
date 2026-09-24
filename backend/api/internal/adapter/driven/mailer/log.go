package mailer

import (
	"context"
	"log"
)

type LogMailer struct{}

func NewLogMailer() *LogMailer {
	return &LogMailer{}
}

func (m *LogMailer) Send(_ context.Context, to, subject, body string) error {
	log.Printf("[mailer] to=%s subject=%q body=%s", to, subject, body)
	return nil
}

func (m *LogMailer) SendHTML(_ context.Context, to, subject, htmlBody, textBody string) error {
	log.Printf("[mailer] to=%s subject=%q html_len=%d text=%s", to, subject, len(htmlBody), textBody)
	return nil
}
