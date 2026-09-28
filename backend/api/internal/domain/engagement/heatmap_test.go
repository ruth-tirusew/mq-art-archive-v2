package engagement

import "testing"

func TestCoverage_countsOverlaps(t *testing.T) {
	spans := []Resolution{
		{Matched: true, Start: 0, End: 5},
		{Matched: true, Start: 3, End: 8},
		{Matched: false, Start: 100, End: 200}, // ignored — not matched
	}
	counts := Coverage(10, spans)
	want := []int{1, 1, 1, 2, 2, 1, 1, 1, 0, 0}
	for i := range want {
		if counts[i] != want[i] {
			t.Fatalf("counts[%d] = %d, want %d (full: %v)", i, counts[i], want[i], counts)
		}
	}
}

func TestCoverage_clampsToBodyLen(t *testing.T) {
	spans := []Resolution{{Matched: true, Start: 3, End: 100}}
	counts := Coverage(5, spans)
	if len(counts) != 5 {
		t.Fatalf("expected counts sized to bodyLen, got len %d", len(counts))
	}
	if counts[3] != 1 || counts[4] != 1 {
		t.Fatalf("expected the clamped span still counted, got %v", counts)
	}
}

func TestMostHighlighted_picksWidestPeakRun(t *testing.T) {
	// counts:      0 1 1 2 2 2 1 0
	counts := []int{0, 1, 1, 2, 2, 2, 1, 0}
	start, end, count := MostHighlighted(counts, 1)
	if start != 3 || end != 6 || count != 2 {
		t.Fatalf("expected widest peak run (3,6,count=2), got (%d,%d,count=%d)", start, end, count)
	}
}

func TestMostHighlighted_noHighlights_returnsZero(t *testing.T) {
	start, end, count := MostHighlighted(make([]int, 10), 1)
	if start != 0 || end != 0 || count != 0 {
		t.Fatalf("expected all zero for no highlights, got (%d,%d,%d)", start, end, count)
	}
}

func TestMostHighlighted_belowMinCount_returnsZero(t *testing.T) {
	counts := []int{0, 1, 1, 1, 0}
	start, end, count := MostHighlighted(counts, 2)
	if start != 0 || end != 0 || count != 0 {
		t.Fatalf("expected suppressed below minCount, got (%d,%d,%d)", start, end, count)
	}
}
