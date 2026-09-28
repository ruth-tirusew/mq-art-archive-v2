package engagement

import "testing"

func TestAnchor_Resolve_sameVersionTrustsOffsets(t *testing.T) {
	body := "The quick brown fox jumps over the lazy dog."
	a := Anchor{QuotedText: "brown fox", TextOffsetStart: 10, TextOffsetEnd: 19, ArticleVersionAtAnchor: 3}

	res := a.Resolve(body, 3)
	if !res.Matched || res.Start != 10 || res.End != 19 {
		t.Fatalf("expected trusted offsets 10-19, got %+v", res)
	}
}

func TestAnchor_Resolve_differentVersionFallsBackToSearch(t *testing.T) {
	// Article edited: a paragraph was inserted before the anchored text, shifting offsets.
	original := "The quick brown fox jumps over the lazy dog."
	edited := "An intro paragraph.\n\n" + original
	a := Anchor{QuotedText: "brown fox", TextOffsetStart: 10, TextOffsetEnd: 19, ArticleVersionAtAnchor: 3}

	res := a.Resolve(edited, 4)
	if !res.Matched {
		t.Fatal("expected quote search to still find the text after an edit")
	}
	want := len("An intro paragraph.\n\n") + 10
	if res.Start != want || res.End != want+len("brown fox") {
		t.Fatalf("expected re-anchored offsets at %d-%d, got %d-%d", want, want+9, res.Start, res.End)
	}
}

func TestAnchor_Resolve_sameVersionButOffsetsDriftedStillSearches(t *testing.T) {
	// Same version number but the stored offsets don't actually match the quoted text —
	// treat this defensively rather than trusting a version number that lied.
	body := "prefix brown fox suffix"
	a := Anchor{QuotedText: "brown fox", TextOffsetStart: 0, TextOffsetEnd: 9, ArticleVersionAtAnchor: 1}

	res := a.Resolve(body, 1)
	if !res.Matched || res.Start != 7 {
		t.Fatalf("expected fallback search to find the real position, got %+v", res)
	}
}

func TestAnchor_Resolve_quoteNoLongerPresent_notMatched(t *testing.T) {
	a := Anchor{QuotedText: "this text was removed", TextOffsetStart: 0, TextOffsetEnd: 22, ArticleVersionAtAnchor: 1}

	res := a.Resolve("completely different content now", 2)
	if res.Matched {
		t.Fatalf("expected no match when the quoted text is gone, got %+v", res)
	}
}

func TestAnchor_Resolve_emptyQuotedText_notMatched(t *testing.T) {
	a := Anchor{QuotedText: "", ArticleVersionAtAnchor: 1}

	res := a.Resolve("anything", 2)
	if res.Matched {
		t.Fatal("expected an empty quoted text to never match")
	}
}

func TestAnchor_Resolve_usesRuneOffsetsNotByteOffsets(t *testing.T) {
	// "—" (em dash, U+2014) is 3 UTF-8 bytes but 1 rune — a real anchor computed by the
	// frontend (which indexes in JS string units, equal to runes for this content) must
	// still resolve correctly here, not be thrown off by Go's byte-indexed strings.
	body := "school — the paperwork"
	quoted := "the paperwork"
	runeOffset := len([]rune("school — ")) // 9
	if runeOffset == len("school — ") {
		t.Fatal("test body must contain a multi-byte rune for this test to be meaningful")
	}

	a := Anchor{QuotedText: quoted, TextOffsetStart: runeOffset, TextOffsetEnd: runeOffset + len([]rune(quoted)), ArticleVersionAtAnchor: 1}
	res := a.Resolve(body, 1)
	if !res.Matched || res.Start != runeOffset {
		t.Fatalf("expected rune-offset anchor to resolve at %d, got %+v", runeOffset, res)
	}
}

func TestSlice_usesRuneOffsets(t *testing.T) {
	body := "school — the paperwork"
	start := len([]rune("school — "))
	got, ok := Slice(body, start, start+len([]rune("the paperwork")))
	if !ok || got != "the paperwork" {
		t.Fatalf("expected Slice to extract %q at rune offset %d, got %q (ok=%v)", "the paperwork", start, got, ok)
	}
}

func TestResolveAll_resolvesEachIndependently(t *testing.T) {
	body := "one two three"
	anchors := []Anchor{
		{QuotedText: "one", TextOffsetStart: 0, TextOffsetEnd: 3, ArticleVersionAtAnchor: 1},
		{QuotedText: "missing", ArticleVersionAtAnchor: 1},
		{QuotedText: "three", TextOffsetStart: 8, TextOffsetEnd: 13, ArticleVersionAtAnchor: 1},
	}

	results := ResolveAll(anchors, body, 1)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if !results[0].Matched || results[0].Start != 0 {
		t.Fatalf("expected first anchor matched at 0, got %+v", results[0])
	}
	if results[1].Matched {
		t.Fatalf("expected second anchor unmatched, got %+v", results[1])
	}
	if !results[2].Matched || results[2].Start != 8 {
		t.Fatalf("expected third anchor matched at 8, got %+v", results[2])
	}
}
