package richtext

import (
	"strings"
	"testing"
)

func TestSanitize(t *testing.T) {
	s := NewSanitizer()
	cases := []struct {
		name, in string
		wantHas  []string
		wantNot  []string
	}{
		{"script removed", `<p>hi</p><script>alert(1)</script>`, []string{"<p>hi</p>"}, []string{"script", "alert"}},
		{"event handler removed", `<p onclick="x()">hi</p>`, []string{"<p>hi</p>"}, []string{"onclick"}},
		{"javascript url removed", `<a href="javascript:alert(1)">x</a>`, nil, []string{"javascript"}},
		{"data url removed", `<a href="data:text/html;base64,AAAA">x</a>`, nil, []string{"data:"}},
		{"img removed", `<img src=x onerror=alert(1)>`, nil, []string{"img", "onerror"}},
		{"style and class removed", `<p style="x" class="y">hi</p>`, []string{"<p>hi</p>"}, []string{"style", "class"}},
		{"safe link kept and hardened", `<a href="https://example.com">x</a>`, []string{`href="https://example.com"`, "nofollow", "noreferrer", "noopener", `target="_blank"`}, nil},
		{"mailto allowed", `<a href="mailto:a@b.co">x</a>`, []string{"mailto:a@b.co"}, nil},
		{"iframe and form removed", `<iframe src="https://evil"></iframe><form action="/x"><input></form>`, nil, []string{"iframe", "form", "input"}},
		{"svg and math removed", `<svg onload=alert(1)></svg><math><mi xlink:href="javascript:alert(1)"></mi></math>`, nil, []string{"svg", "onload", "math"}},
		{"unclosed and nested tricks", `<p><a href="https://ok.example">x<script>alert(1)</script></a>`, []string{"ok.example"}, []string{"script", "alert"}},
		{"headings, underline and strike kept", `<h2>A</h2><h3>B</h3><p><u>u</u> <s>s</s></p>`, []string{"<h2>A</h2>", "<h3>B</h3>", "<u>u</u>", "<s>s</s>"}, nil},
		{"other heading levels lose the tag but keep the text", `<h1>Big</h1><h4>Small</h4>`, []string{"Big", "Small"}, []string{"<h1", "<h4"}},
		{"amharic preserved", `<p>ሰላም ዓለም</p>`, []string{"ሰላም ዓለም"}, nil},
	}
	for _, c := range cases {
		got, _ := s.Sanitize(c.in)
		for _, w := range c.wantHas {
			if !strings.Contains(got, w) {
				t.Errorf("%s: %q missing %q", c.name, got, w)
			}
		}
		for _, w := range c.wantNot {
			if strings.Contains(strings.ToLower(got), w) {
				t.Errorf("%s: %q should not contain %q", c.name, got, w)
			}
		}
	}
}

func TestSanitize_textAndEmpty(t *testing.T) {
	s := NewSanitizer()
	if clean, text := s.Sanitize("<p></p>"); clean != "" || text != "" {
		t.Errorf("empty markup should collapse, got %q %q", clean, text)
	}
	if clean, text := s.Sanitize("  <p> </p> <br> "); clean != "" || text != "" {
		t.Errorf("whitespace-only markup should collapse, got %q %q", clean, text)
	}
	_, text := s.Sanitize("<p>Tom &amp; Jerry <strong>live</strong></p>")
	if text != "Tom & Jerry live" {
		t.Errorf("text = %q", text)
	}
}
