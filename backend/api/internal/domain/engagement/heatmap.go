package engagement

// Coverage computes, for each character offset in a body of length bodyLen, how many of
// the given resolved spans include it. Unmatched resolutions (the underlying quote could
// no longer be found — see Anchor.Resolve) are skipped, since they have no valid position
// to contribute to the count.
func Coverage(bodyLen int, spans []Resolution) []int {
	counts := make([]int, bodyLen)
	for _, s := range spans {
		if !s.Matched {
			continue
		}
		end := s.End
		if end > bodyLen {
			end = bodyLen
		}
		for i := s.Start; i < end; i++ {
			if i >= 0 {
				counts[i]++
			}
		}
	}
	return counts
}

// MostHighlighted finds the widest contiguous run of the coverage array's peak value —
// the article's single most-highlighted passage, the way Medium surfaces one. Returns
// count 0 (and start == end == 0) if there are no highlights at all. minCount lets callers
// require some minimum number of readers before treating anything as "popular" (a single
// highlighter's own selection isn't a trend).
func MostHighlighted(counts []int, minCount int) (start, end, count int) {
	peak := 0
	for _, c := range counts {
		if c > peak {
			peak = c
		}
	}
	if peak == 0 || peak < minCount {
		return 0, 0, 0
	}

	bestLen := 0
	i := 0
	for i < len(counts) {
		if counts[i] == peak {
			j := i
			for j < len(counts) && counts[j] == peak {
				j++
			}
			if j-i > bestLen {
				bestLen = j - i
				start, end = i, j
			}
			i = j
		} else {
			i++
		}
	}
	return start, end, peak
}
