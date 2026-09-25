package main

import (
	"testing"
	"time"
)

func mustParse(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time: %v", err)
	}
	return parsed.UTC()
}

func TestNextWeeklyRun_beforeTargetThisWeek(t *testing.T) {
	// Wednesday — next Monday 9am is still ahead this week... no, Monday already passed
	// this week, so it must roll to the following Monday.
	from := mustParse(t, "2026-01-14T10:00:00Z") // Wednesday
	got := nextWeeklyRun(from, time.Monday, 9)
	want := mustParse(t, "2026-01-19T09:00:00Z") // following Monday
	if !got.Equal(want) {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestNextWeeklyRun_sameDayBeforeHour(t *testing.T) {
	from := mustParse(t, "2026-01-19T05:00:00Z") // Monday, before 9am
	got := nextWeeklyRun(from, time.Monday, 9)
	want := mustParse(t, "2026-01-19T09:00:00Z") // later today
	if !got.Equal(want) {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestNextWeeklyRun_sameDayAfterHour(t *testing.T) {
	from := mustParse(t, "2026-01-19T09:00:01Z") // Monday, just after 9am
	got := nextWeeklyRun(from, time.Monday, 9)
	want := mustParse(t, "2026-01-26T09:00:00Z") // next week, not today
	if !got.Equal(want) {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestNextWeeklyRun_exactlyAtTargetRollsToNextWeek(t *testing.T) {
	from := mustParse(t, "2026-01-19T09:00:00Z") // exactly the target instant
	got := nextWeeklyRun(from, time.Monday, 9)
	want := mustParse(t, "2026-01-26T09:00:00Z")
	if !got.Equal(want) {
		t.Fatalf("got %s, want %s", got, want)
	}
}
