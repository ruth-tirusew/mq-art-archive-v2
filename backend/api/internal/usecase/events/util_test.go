package events

import "testing"

func TestUniqueSlug_appendsSourceSuffix(t *testing.T) {
	got := uniqueSlug("Opening Night", "https://t.me/EventsEthiopia/10588")
	if got != "opening-night-10588" {
		t.Fatalf("got %q", got)
	}
}

func TestUniqueSlug_sameTitleDifferentSources(t *testing.T) {
	a := uniqueSlug("Opening", "https://t.me/EventsEthiopia/1")
	b := uniqueSlug("Opening", "https://t.me/EventsEthiopia/2")
	if a == b {
		t.Fatalf("expected distinct slugs, both %q", a)
	}
}

func TestUniqueSlug_emptySourceFallsBack(t *testing.T) {
	got := uniqueSlug("Blue Room Opening", "")
	if got != "blue-room-opening" {
		t.Fatalf("got %q", got)
	}
}
