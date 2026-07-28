package profile

type ListFilter struct {
	Query    string
	Limit    int
	Offset   int
	Featured *bool
}

func PublicListFilter() ListFilter {
	return ListFilter{Limit: 24}
}
