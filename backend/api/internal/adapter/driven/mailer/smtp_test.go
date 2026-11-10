package mailer

import (
	"bufio"
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

// fakeSMTP accepts one message on localhost and returns what the client sent.
func fakeSMTP(t *testing.T) (port int, got <-chan []string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	out := make(chan []string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		w := func(s string) { conn.Write([]byte(s + "\r\n")) }
		var lines []string
		w("220 fake ESMTP")
		inData := false
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				break
			}
			line = strings.TrimRight(line, "\r\n")
			if inData {
				if line == "." {
					inData = false
					w("250 queued")
					continue
				}
				lines = append(lines, line)
				continue
			}
			lines = append(lines, "> "+line)
			switch {
			case strings.HasPrefix(line, "EHLO"):
				w("250 fake")
			case strings.HasPrefix(line, "DATA"):
				w("354 go")
				inData = true
			case strings.HasPrefix(line, "QUIT"):
				w("221 bye")
				out <- lines
				return
			default:
				w("250 ok")
			}
		}
		out <- lines
	}()
	return ln.Addr().(*net.TCPAddr).Port, out
}

func wait(t *testing.T, got <-chan []string) string {
	t.Helper()
	select {
	case l := <-got:
		return strings.Join(l, "\n")
	case <-time.After(3 * time.Second):
		t.Fatal("fake server saw no message")
		return ""
	}
}

func TestSMTPSendsPlainText(t *testing.T) {
	port, got := fakeSMTP(t)
	m, err := NewSMTPMailer("127.0.0.1", port, "", "", "Artiv <noreply@artiv.et>")
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Send(context.Background(), "ruth@example.com", "Verify your Artiv email", "Click: http://x/verify?token=abc"); err != nil {
		t.Fatal(err)
	}
	s := wait(t, got)
	for _, want := range []string{"> MAIL FROM:<noreply@artiv.et>", "> RCPT TO:<ruth@example.com>", "Subject: Verify your Artiv email", "Content-Type: text/plain", "token=abc"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in:\n%s", want, s)
		}
	}
}

func TestSMTPSendsHTMLAsMultipartAndEncodesSubject(t *testing.T) {
	port, got := fakeSMTP(t)
	m, _ := NewSMTPMailer("127.0.0.1", port, "", "", "noreply@artiv.et")
	if err := m.SendHTML(context.Background(), "a@b.et", "ሰላም Artiv", "<p>hi</p>", "hi"); err != nil {
		t.Fatal(err)
	}
	s := wait(t, got)
	for _, want := range []string{"multipart/alternative", "text/plain", "text/html", "<p>hi</p>", "Subject: =?utf-8?q?"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in:\n%s", want, s)
		}
	}
}

func TestSMTPRejectsHeaderInjection(t *testing.T) {
	m, _ := NewSMTPMailer("127.0.0.1", 1, "", "", "noreply@artiv.et")
	if err := m.Send(context.Background(), "a@b.et", "hi\r\nBcc: evil@x.et", "body"); err == nil {
		t.Error("a subject with a line break must be refused")
	}
	if err := m.Send(context.Background(), "a@b.et\r\nBcc: evil@x.et", "hi", "body"); err == nil {
		t.Error("a recipient with a line break must be refused")
	}
}

func TestOnlyTheLocalMachineMaySkipTLS(t *testing.T) {
	for host, want := range map[string]bool{"localhost": true, "127.0.0.1": true, "smtp.gmail.com": false, "10.0.0.5": false} {
		if isLocalHost(host) != want {
			t.Errorf("isLocalHost(%q) = %v, want %v", host, !want, want)
		}
	}
}

func TestNewSMTPMailerValidatesSender(t *testing.T) {
	if _, err := NewSMTPMailer("smtp.example.com", 587, "u", "p", "not an address"); err == nil {
		t.Error("an invalid MAIL_FROM must fail at startup")
	}
	if _, err := NewSMTPMailer("", 587, "", "", "a@b.et"); err == nil {
		t.Error("an empty host must fail")
	}
}
