package digest

import (
	"strings"
	"testing"
	"time"

	"github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/testutil/assist"
)

func sampleDigest() inbound.Digest {
	return inbound.Digest{
		Events: []events.Event{
			{Slug: "opening-night", Title: "Opening Night <Gala>", Venue: "Alliance", StartsAt: time.Date(2026, 3, 2, 18, 0, 0, 0, time.UTC)},
		},
		ArtPosts: []art.ArtPostWithArtist{
			{ArtPost: art.ArtPost{Title: "Blue Horizon"}, ArtistSlug: "artist-x", ArtistName: "Artist X"},
		},
		Articles: []content.Article{
			{Slug: "color-theory", Title: "Color Theory"},
		},
	}
}

func TestRenderEmail_includesAllSections(t *testing.T) {
	htmlBody, textBody, err := RenderEmail(sampleDigest(), "https://artiv.example", "https://artiv.example/unsub?t=abc")
	assist.NoError(t, err)

	assist.Contains(t, htmlBody, "https://artiv.example/events/opening-night")
	assist.Contains(t, htmlBody, "https://artiv.example/@artist-x")
	assist.Contains(t, htmlBody, "https://artiv.example/wiki/color-theory")
	assist.Contains(t, htmlBody, "https://artiv.example/unsub?t=abc")

	assist.Contains(t, textBody, "Opening Night <Gala>")
	assist.Contains(t, textBody, "https://artiv.example/unsub?t=abc")
}

func TestRenderEmail_escapesHTMLInTitles(t *testing.T) {
	htmlBody, _, err := RenderEmail(sampleDigest(), "https://artiv.example", "https://artiv.example/unsub")
	assist.NoError(t, err)

	if strings.Contains(htmlBody, "<Gala>") {
		t.Fatalf("expected event title to be HTML-escaped, got raw markup in: %s", htmlBody)
	}
	assist.Contains(t, htmlBody, "Opening Night &lt;Gala&gt;")
}

func TestRenderTelegram_escapesTitlesAndOmitsUnsubscribeLink(t *testing.T) {
	msg := RenderTelegram(sampleDigest(), "https://artiv.example")

	assist.Contains(t, msg, "Opening Night &lt;Gala&gt;")
	assist.Contains(t, msg, "https://artiv.example/events/opening-night")
	assist.Contains(t, msg, "https://artiv.example/@artist-x")
	assist.Contains(t, msg, "https://artiv.example/wiki/color-theory")

	if strings.Contains(msg, "unsub") {
		t.Fatalf("telegram message should not contain an unsubscribe link, got: %s", msg)
	}
}

func TestSubject_countsContent(t *testing.T) {
	subj := Subject(sampleDigest())
	assist.Contains(t, subj, "1 event(s)")
	assist.Contains(t, subj, "2 new post(s)")
}
