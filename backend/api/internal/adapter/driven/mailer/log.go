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
