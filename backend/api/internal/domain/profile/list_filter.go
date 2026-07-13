package profile

type ListFilter struct {
	Query   string
	Limit   int
	Offset  int
	Featured *bool
}

func PublicListFilter() ListFilter {
	status := ProfileStatusApproved
	_ = status
	return ListFilter{Limit: 50}
}
