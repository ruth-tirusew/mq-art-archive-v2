package events

import (
	"testing"
	"time"

	"github.com/mq/api/internal/testutil/assist"
)

func TestParseTelegramMessage_extractsFields(t *testing.T) {
	msg := TelegramMessage{
		Channel:   "addisartevents",
		MessageID: 42,
		Date:      time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
		Text: `Opening Night at Modern Gallery
📍 Addis Ababa
Exhibition runs 15–20 Jul 2026
Join us for contemporary Ethiopian painting.`,
	}

	ev, ok := ParseTelegramMessage(msg, []string{"exhibition", "opening"})
	assist.Equal(t, true, ok)
	assist.NotNil(t, ev)
	assist.Equal(t, "Opening Night at Modern Gallery", ev.Title)
	assist.Equal(t, "https://t.me/addisartevents/42", ev.SourceURL)
	assist.Equal(t, "Addis Ababa", ev.Venue)
	assist.Equal(t, "Addis Ababa", ev.City)
	assist.Equal(t, 15, ev.StartsAt.Day())
	assist.Equal(t, time.July, ev.StartsAt.Month())
	assist.NotNil(t, ev.EndsAt)
	assist.Equal(t, 20, ev.EndsAt.Day())
}

func TestParseTelegramMessage_keywordFilter(t *testing.T) {
	msg := TelegramMessage{
		Channel:   "news",
		MessageID: 1,
		Date:      time.Now().UTC(),
		Text:      "Weather update for Addis Ababa tomorrow",
	}
	_, ok := ParseTelegramMessage(msg, []string{"exhibition", "opening"})
	assist.Equal(t, false, ok)
}

func TestParseTelegramMessage_skipsEmpty(t *testing.T) {
	_, ok := ParseTelegramMessage(TelegramMessage{Channel: "x", MessageID: 1, Text: "  "}, nil)
	assist.Equal(t, false, ok)
}

func TestExtractDate_formats(t *testing.T) {
	fallback := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	assist.Equal(t, 9, extractDate("See you on 2026-07-09", fallback).Day())
	assist.Equal(t, 9, extractDate("Event on 9/7/2026", fallback).Day())
	assist.Equal(t, 15, extractDate("Jan 15, 2026 opening", fallback).Day())
}

func TestTelegramPermalink(t *testing.T) {
	assist.Equal(t, "https://t.me/foo/9", TelegramPermalink("@foo", 9))
}
