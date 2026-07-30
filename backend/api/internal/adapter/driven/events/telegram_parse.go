package events

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/events"
)

var (
	reLocationPin   = regexp.MustCompile(`(?m)^📍\s*(.+)$`)
	reLocationAt    = regexp.MustCompile(`(?i)\b(?:at|@)\s+([A-Z][\w\s&'.-]{2,60})`)
	reHashtagPlace  = regexp.MustCompile(`#([A-Za-z][A-Za-z0-9_]{2,})`)
	reDateISO       = regexp.MustCompile(`\b(20\d{2}-\d{2}-\d{2})\b`)
	reDateSlash     = regexp.MustCompile(`\b(\d{1,2})/(\d{1,2})/(20\d{2})\b`)
	reDateDayMonth  = regexp.MustCompile(`(?i)\b(\d{1,2})\s+(Jan(?:uary)?|Feb(?:ruary)?|Mar(?:ch)?|Apr(?:il)?|May|Jun(?:e)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:t(?:ember)?)?|Oct(?:ober)?|Nov(?:ember)?|Dec(?:ember)?)(?:\s+(20\d{2}))?\b`)
	reDateMonthDay  = regexp.MustCompile(`(?i)\b(Jan(?:uary)?|Feb(?:ruary)?|Mar(?:ch)?|Apr(?:il)?|May|Jun(?:e)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:t(?:ember)?)?|Oct(?:ober)?|Nov(?:ember)?|Dec(?:ember)?)\s+(\d{1,2})(?:\s*,?\s*(20\d{2}))?\b`)
	reDateRangeDash = regexp.MustCompile(`(?i)\b(\d{1,2})\s*[–—-]\s*(\d{1,2})\s+(Jan(?:uary)?|Feb(?:ruary)?|Mar(?:ch)?|Apr(?:il)?|May|Jun(?:e)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:t(?:ember)?)?|Oct(?:ober)?|Nov(?:ember)?|Dec(?:ember)?)(?:\s+(20\d{2}))?\b`)
	reUntilDate     = regexp.MustCompile(`(?i)\buntil\s+(.+)$`)
)

var knownCities = []string{
	"Addis Ababa", "Bahir Dar", "Hawassa", "Dire Dawa", "Mekelle", "Gondar", "Jimma", "Harar",
}

type TelegramMessage struct {
	Channel   string
	MessageID int
	Text      string
	Date      time.Time
}

func MatchKeywords(text string, keywords []string) bool {
	if len(keywords) == 0 {
		return true
	}
	lower := strings.ToLower(text)
	for _, kw := range keywords {
		kw = strings.TrimSpace(strings.ToLower(kw))
		if kw != "" && strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func ParseTelegramMessage(msg TelegramMessage, keywords []string) (*domain.Event, bool) {
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return nil, false
	}
	if !MatchKeywords(text, keywords) {
		return nil, false
	}

	title := extractTitle(text)
	locationName := extractLocation(text)
	startsAt, endsAt := extractDateRange(text, msg.Date)
	now := time.Now().UTC()

	ev := &domain.Event{
		ID:          uuid.New(),
		Title:       title,
		Description: truncate(text, 2000),
		SourceURL:   TelegramPermalink(msg.Channel, msg.MessageID),
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		ScrapedAt:   now,
		Status:      domain.EventStatusPending,
		EventType:   "Opening",
		Venue:       locationName,
		City:        extractCity(text),
	}
	if locationName != "" {
		ev.Location = &domain.Location{Name: locationName}
	}
	return ev, true
}

func TelegramPermalink(channel string, messageID int) string {
	channel = strings.TrimPrefix(strings.TrimSpace(channel), "@")
	return fmt.Sprintf("https://t.me/%s/%d", channel, messageID)
}

func extractTitle(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		return truncate(line, 120)
	}
	return truncate(text, 120)
}

func extractLocation(text string) string {
	if m := reLocationPin.FindStringSubmatch(text); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	if m := reLocationAt.FindStringSubmatch(text); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	if city := extractCity(text); city != "" {
		return city
	}
	if m := reHashtagPlace.FindStringSubmatch(text); len(m) > 1 {
		return splitCamel(m[1])
	}
	return ""
}

func extractCity(text string) string {
	lower := strings.ToLower(text)
	for _, city := range knownCities {
		if strings.Contains(lower, strings.ToLower(city)) {
			return city
		}
	}
	return ""
}

func extractDateRange(text string, fallback time.Time) (time.Time, *time.Time) {
	if fallback.IsZero() {
		fallback = time.Now().UTC()
	}
	fallback = fallback.UTC()

	if m := reDateRangeDash.FindStringSubmatch(text); len(m) > 3 {
		year := fallback.Year()
		if m[4] != "" {
			fmt.Sscanf(m[4], "%d", &year)
		}
		month := parseMonth(m[3])
		startDay, endDay := 1, 1
		fmt.Sscanf(m[1], "%d", &startDay)
		fmt.Sscanf(m[2], "%d", &endDay)
		starts := time.Date(year, month, startDay, 0, 0, 0, 0, time.UTC)
		ends := time.Date(year, month, endDay, 23, 59, 59, 0, time.UTC)
		return starts, &ends
	}

	starts := extractDate(text, fallback)
	if m := reUntilDate.FindStringSubmatch(text); len(m) > 1 {
		if ends := extractDate(m[1], starts); !ends.Equal(starts) {
			e := ends
			return starts, &e
		}
	}
	return starts, nil
}

func extractDate(text string, fallback time.Time) time.Time {
	if m := reDateISO.FindStringSubmatch(text); len(m) > 1 {
		if t, err := time.Parse("2006-01-02", m[1]); err == nil {
			return t.UTC()
		}
	}
	if m := reDateSlash.FindStringSubmatch(text); len(m) > 3 {
		var d, mo, y int
		fmt.Sscanf(m[1], "%d", &d)
		fmt.Sscanf(m[2], "%d", &mo)
		fmt.Sscanf(m[3], "%d", &y)
		return time.Date(y, time.Month(mo), d, 0, 0, 0, 0, time.UTC)
	}
	if m := reDateDayMonth.FindStringSubmatch(text); len(m) > 2 {
		year := fallback.Year()
		if m[3] != "" {
			fmt.Sscanf(m[3], "%d", &year)
		}
		day := 1
		fmt.Sscanf(m[1], "%d", &day)
		return time.Date(year, parseMonth(m[2]), day, 0, 0, 0, 0, time.UTC)
	}
	if m := reDateMonthDay.FindStringSubmatch(text); len(m) > 2 {
		year := fallback.Year()
		if m[3] != "" {
			fmt.Sscanf(m[3], "%d", &year)
		}
		day := 1
		fmt.Sscanf(m[2], "%d", &day)
		return time.Date(year, parseMonth(m[1]), day, 0, 0, 0, 0, time.UTC)
	}
	return fallback
}

func parseMonth(raw string) time.Month {
	raw = strings.ToLower(raw[:3])
	months := map[string]time.Month{
		"jan": time.January, "feb": time.February, "mar": time.March,
		"apr": time.April, "may": time.May, "jun": time.June,
		"jul": time.July, "aug": time.August, "sep": time.September,
		"oct": time.October, "nov": time.November, "dec": time.December,
	}
	if m, ok := months[raw]; ok {
		return m
	}
	return time.January
}

func truncate(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max])
}

func splitCamel(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}
