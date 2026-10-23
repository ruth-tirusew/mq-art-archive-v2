package apperrors

import "testing"

func TestValidateWebURL(t *testing.T) {
	ok := []string{"", "  ", "https://res.cloudinary.com/x/a.jpg", "http://t.me/channel", "https://example.com/a?b=c#d"}
	bad := []string{
		"javascript:alert(1)", "JaVaScRiPt:alert(1)", "data:text/html,<script>1</script>", "file:///etc/passwd",
		"//evil.com/x", "/relative", "example.com", "https://", "https://user:pw@example.com/", "ftp://example.com",
		"https://example.com/" + string(make([]byte, 3000)),
	}
	for _, s := range ok {
		if err := ValidateWebURL(s); err != nil {
			t.Errorf("ValidateWebURL(%q) rejected: %v", s, err)
		}
	}
	for _, s := range bad {
		if err := ValidateWebURL(s); err == nil {
			t.Errorf("ValidateWebURL(%q) should be rejected", s)
		}
	}
}
