package content

import "time"

type ListFilter struct {
	Category string
	Query    string
	// PublishedSince, when set, restricts results to articles whose PublishedAt is at or
	// after this time. Nil means no lower bound.
	PublishedSince *time.Time
	Limit          int
	Offset         int
}

func PublicListFilter() ListFilter {
	return ListFilter{Limit: 50}
}
