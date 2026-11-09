package mailer

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// SMTPMailer sends through any SMTP server (Gmail, Mailgun, a company relay, ...).
//
// Port 465 uses implicit TLS. Any other port connects in plain text and upgrades with STARTTLS;
// it refuses to send credentials or mail over a connection that could not be upgraded, unless the
// server is the local machine (a dev mail catcher).
type SMTPMailer struct {
	host, port, username, password string
	from                           mail.Address
	timeout                        time.Duration
}

// NewSMTPMailer validates the sender up front, so a bad MAIL_FROM fails at startup and not on the
// first password reset.
func NewSMTPMailer(host string, port int, username, password, from string) (*SMTPMailer, error) {
	addr, err := mail.ParseAddress(from)
	if err != nil {
		return nil, fmt.Errorf("smtp: MAIL_FROM %q is not a valid address: %w", from, err)
	}
	if host == "" {
		return nil, errors.New("smtp: SMTP_HOST is empty")
	}
	if port == 0 {
		port = 587
	}
	return &SMTPMailer{
		host: host, port: strconv.Itoa(port), username: username, password: password,
		from: *addr, timeout: 20 * time.Second,
	}, nil
}

func (m *SMTPMailer) Send(ctx context.Context, to, subject, body string) error {
	return m.send(ctx, to, subject, buildMessage(m.from, to, subject, body, ""))
}

func (m *SMTPMailer) SendHTML(ctx context.Context, to, subject, htmlBody, textBody string) error {
	return m.send(ctx, to, subject, buildMessage(m.from, to, subject, textBody, htmlBody))
}

func (m *SMTPMailer) send(ctx context.Context, to, subject string, msg []byte) error {
	rcpt, err := mail.ParseAddress(to)
	if err != nil {
		return fmt.Errorf("smtp: invalid recipient: %w", err)
	}
	if strings.ContainsAny(subject, "\r\n") {
		return errors.New("smtp: subject must be a single line")
	}

	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(m.host, m.port))
	if err != nil {
		return fmt.Errorf("smtp: connect: %w", err)
	}
	if m.port == "465" {
		conn = tls.Client(conn, &tls.Config{ServerName: m.host, MinVersion: tls.VersionTLS12})
	}
	if dl, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(dl)
	}

	c, err := smtp.NewClient(conn, m.host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("smtp: handshake: %w", err)
	}
	defer c.Close()

	if m.port != "465" {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(&tls.Config{ServerName: m.host, MinVersion: tls.VersionTLS12}); err != nil {
				return fmt.Errorf("smtp: starttls: %w", err)
			}
		} else if !isLocalHost(m.host) {
			return errors.New("smtp: server does not offer STARTTLS; refusing to send over an unencrypted connection")
		}
	}
	if m.username != "" {
		if err := c.Auth(smtp.PlainAuth("", m.username, m.password, m.host)); err != nil {
			return fmt.Errorf("smtp: auth: %w", err)
		}
	}
	if err := c.Mail(m.from.Address); err != nil {
		return fmt.Errorf("smtp: sender rejected: %w", err)
	}
	if err := c.Rcpt(rcpt.Address); err != nil {
		return fmt.Errorf("smtp: recipient rejected: %w", err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp: data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp: write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp: send: %w", err)
	}
	return c.Quit()
}

func isLocalHost(h string) bool {
	return h == "localhost" || h == "127.0.0.1" || h == "::1"
}

func buildMessage(from mail.Address, to, subject, textBody, htmlBody string) []byte {
	var b strings.Builder
	h := func(k, v string) { b.WriteString(k + ": " + v + "\r\n") }
	h("From", from.String())
	h("To", to)
	h("Subject", mime.QEncoding.Encode("utf-8", subject))
	h("Date", time.Now().UTC().Format(time.RFC1123Z))
	h("Message-ID", "<"+randomID()+"@"+domainOf(from.Address)+">")
	h("MIME-Version", "1.0")

	if htmlBody == "" {
		h("Content-Type", `text/plain; charset="utf-8"`)
		h("Content-Transfer-Encoding", "8bit")
		b.WriteString("\r\n" + crlf(textBody) + "\r\n")
		return []byte(b.String())
	}
	boundary := "artiv-" + randomID()
	h("Content-Type", `multipart/alternative; boundary="`+boundary+`"`)
	b.WriteString("\r\n")
	part := func(ctype, body string) {
		b.WriteString("--" + boundary + "\r\n")
		b.WriteString("Content-Type: " + ctype + "; charset=\"utf-8\"\r\n")
		b.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		b.WriteString(crlf(body) + "\r\n")
	}
	part("text/plain", textBody)
	part("text/html", htmlBody)
	b.WriteString("--" + boundary + "--\r\n")
	return []byte(b.String())
}

// crlf normalises line endings: SMTP requires CRLF. net/smtp dot-stuffs the body itself.
func crlf(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\n", "\r\n")
}

func domainOf(addr string) string {
	if i := strings.LastIndex(addr, "@"); i >= 0 {
		return addr[i+1:]
	}
	return "localhost"
}

func randomID() string {
	var raw [12]byte
	_, _ = rand.Read(raw[:])
	return hex.EncodeToString(raw[:])
}
