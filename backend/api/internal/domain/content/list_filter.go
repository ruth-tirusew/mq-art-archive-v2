package content

type ListFilter struct {
	Category string
	Query    string
	Limit    int
	Offset   int
}

func PublicListFilter() ListFilter {
	return ListFilter{Limit: 50}
}
