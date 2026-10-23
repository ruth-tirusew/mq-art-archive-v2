package apperrors

import (
	"fmt"
	"net/url"
	"strings"
)

const maxWebURLLength = 2048

// ValidateWebURL accepts only absolute http(s) URLs with a host and no embedded credentials. It
// stands between user-supplied links and image sources and schemes like javascript:, data: or
// file:, which would otherwise be stored and later rendered into pages and emails.
// An empty string is valid: the field is optional. (It lives here because every use case may
// import apperrors and nothing else is shared between them.)
func ValidateWebURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if len(raw) > maxWebURLLength {
		return fmt.Errorf("%w: link is too long", ErrValidation)
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil {
		return fmt.Errorf("%w: links must be full http(s) addresses", ErrValidation)
	}
	return nil
}

// IsWebURL is ValidateWebURL as a yes/no, for places that drop bad values instead of rejecting.
func IsWebURL(raw string) bool { return ValidateWebURL(raw) == nil }
