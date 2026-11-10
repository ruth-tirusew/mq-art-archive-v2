package config

import "testing"

func TestSMTPAliasesAndDefaultSender(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_USER", "team@example.com")
	t.Setenv("SMTP_PASS", "s3cret")
	t.Setenv("MAIL_FROM", "")
	c := Load()
	if c.SMTPUsername != "team@example.com" || c.SMTPPassword != "s3cret" {
		t.Fatalf("SMTP_USER/SMTP_PASS must be accepted, got %q / %q", c.SMTPUsername, c.SMTPPassword)
	}
	if c.MailFrom != "Artiv <team@example.com>" {
		t.Fatalf("sender should default to the SMTP account, got %q", c.MailFrom)
	}
	if c.SMTPPort != 587 {
		t.Fatalf("port should default to 587, got %d", c.SMTPPort)
	}
}

func TestExplicitMailFromWins(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("SMTP_USERNAME", "team@example.com")
	t.Setenv("MAIL_FROM", "Artiv <hello@artiv.et>")
	if got := Load().MailFrom; got != "Artiv <hello@artiv.et>" {
		t.Fatalf("got %q", got)
	}
}
