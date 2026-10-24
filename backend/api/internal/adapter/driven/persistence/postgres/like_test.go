package postgres

import "testing"

func TestContainsPattern_treatsUserWildcardsAsPlainText(t *testing.T) {
	cases := map[string]string{
		"art":     "%art%",
		"100%":    `%100\%%`,
		"a_b":     `%a\_b%`,
		`back\sl`: `%back\\sl%`,
		"":        "%%",
	}
	for in, want := range cases {
		if got := containsPattern(in); got != want {
			t.Errorf("containsPattern(%q) = %q, want %q", in, got, want)
		}
	}
}
