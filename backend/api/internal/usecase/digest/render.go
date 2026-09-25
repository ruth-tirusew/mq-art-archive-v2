package digest

import (
	"fmt"
	"html"
	"html/template"
	"strings"

	"github.com/mq/api/internal/port/inbound"
)

// Subject returns the email subject line for a digest.
func Subject(d inbound.Digest) string {
	return fmt.Sprintf("This week at Artiv — %d event(s), %d new post(s)", len(d.Events), len(d.ArtPosts)+len(d.Articles))
}

var emailTemplate = template.Must(template.New("digest_email").Parse(`
<!doctype html>
<html>
<body style="font-family: sans-serif; color: #1a1a1a; max-width: 600px; margin: 0 auto;">
  <h1 style="font-size: 20px;">This week at Artiv</h1>

  {{if .Events}}
  <h2 style="font-size: 16px;">Upcoming events</h2>
  <ul>
    {{range .Events}}
    <li><a href="{{.URL}}">{{.Title}}</a> — {{.When}}{{if .Venue}}, {{.Venue}}{{end}}</li>
    {{end}}
  </ul>
  {{end}}

  {{if .ArtPosts}}
  <h2 style="font-size: 16px;">New art</h2>
  <ul>
    {{range .ArtPosts}}
    <li><a href="{{.URL}}">{{.Title}}</a> by {{.Artist}}</li>
    {{end}}
  </ul>
  {{end}}

  {{if .Articles}}
  <h2 style="font-size: 16px;">New wiki articles</h2>
  <ul>
    {{range .Articles}}
    <li><a href="{{.URL}}">{{.Title}}</a></li>
    {{end}}
  </ul>
  {{end}}

  <p style="margin-top: 24px; font-size: 12px; color: #666;">
    <a href="{{.UnsubscribeURL}}">Unsubscribe from this digest</a>
  </p>
</body>
</html>
`))

type emailEvent struct {
	Title, URL, When, Venue string
}

type emailItem struct {
	Title, URL, Artist string
}

type emailView struct {
	Events         []emailEvent
	ArtPosts       []emailItem
	Articles       []emailItem
	UnsubscribeURL string
}

// RenderEmail produces the HTML and plain-text bodies for a digest email.
// unsubscribeURL should already carry a signed, recipient-specific token.
func RenderEmail(d inbound.Digest, webAppURL, unsubscribeURL string) (htmlBody, textBody string, err error) {
	view := emailView{UnsubscribeURL: unsubscribeURL}
	for _, e := range d.Events {
		view.Events = append(view.Events, emailEvent{
			Title: e.Title,
			URL:   webAppURL + "/events/" + e.Slug,
			When:  e.StartsAt.Format("Mon, Jan 2"),
			Venue: e.Venue,
		})
	}
	for _, p := range d.ArtPosts {
		view.ArtPosts = append(view.ArtPosts, emailItem{
			Title:  p.Title,
			URL:    webAppURL + "/@" + p.ArtistSlug,
			Artist: p.ArtistName,
		})
	}
	for _, a := range d.Articles {
		view.Articles = append(view.Articles, emailItem{
			Title: a.Title,
			URL:   webAppURL + "/wiki/" + a.Slug,
		})
	}

	var buf strings.Builder
	if err := emailTemplate.Execute(&buf, view); err != nil {
		return "", "", fmt.Errorf("render digest email: %w", err)
	}
	return buf.String(), renderText(d, webAppURL, unsubscribeURL), nil
}

func renderText(d inbound.Digest, webAppURL, unsubscribeURL string) string {
	var b strings.Builder
	b.WriteString("This week at Artiv\n\n")

	if len(d.Events) > 0 {
		b.WriteString("Upcoming events:\n")
		for _, e := range d.Events {
			fmt.Fprintf(&b, "- %s (%s) %s/events/%s\n", e.Title, e.StartsAt.Format("Mon, Jan 2"), webAppURL, e.Slug)
		}
		b.WriteString("\n")
	}
	if len(d.ArtPosts) > 0 {
		b.WriteString("New art:\n")
		for _, p := range d.ArtPosts {
			fmt.Fprintf(&b, "- %s by %s %s/@%s\n", p.Title, p.ArtistName, webAppURL, p.ArtistSlug)
		}
		b.WriteString("\n")
	}
	if len(d.Articles) > 0 {
		b.WriteString("New wiki articles:\n")
		for _, a := range d.Articles {
			fmt.Fprintf(&b, "- %s %s/wiki/%s\n", a.Title, webAppURL, a.Slug)
		}
		b.WriteString("\n")
	}

	fmt.Fprintf(&b, "Unsubscribe: %s\n", unsubscribeURL)
	return b.String()
}

// RenderTelegram produces a message for the Telegram channel, using Telegram's HTML
// subset (a small tag allowlist, not full HTML — see
// https://core.telegram.org/bots/api#html-style). It has no unsubscribe link since
// unsubscribing from Telegram happens via the bot's /stop command, not a URL.
func RenderTelegram(d inbound.Digest, webAppURL string) string {
	var b strings.Builder
	b.WriteString("<b>This week at Artiv</b>\n\n")

	if len(d.Events) > 0 {
		b.WriteString("<b>Upcoming events</b>\n")
		for _, e := range d.Events {
			fmt.Fprintf(&b, "• <a href=\"%s/events/%s\">%s</a> — %s\n", webAppURL, e.Slug, html.EscapeString(e.Title), e.StartsAt.Format("Mon, Jan 2"))
		}
		b.WriteString("\n")
	}
	if len(d.ArtPosts) > 0 {
		b.WriteString("<b>New art</b>\n")
		for _, p := range d.ArtPosts {
			fmt.Fprintf(&b, "• <a href=\"%s/@%s\">%s</a> by %s\n", webAppURL, p.ArtistSlug, html.EscapeString(p.Title), html.EscapeString(p.ArtistName))
		}
		b.WriteString("\n")
	}
	if len(d.Articles) > 0 {
		b.WriteString("<b>New wiki articles</b>\n")
		for _, a := range d.Articles {
			fmt.Fprintf(&b, "• <a href=\"%s/wiki/%s\">%s</a>\n", webAppURL, a.Slug, html.EscapeString(a.Title))
		}
	}

	return b.String()
}
