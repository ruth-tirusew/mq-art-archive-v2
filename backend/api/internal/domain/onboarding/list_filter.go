package onboarding

// ListFilter selects a page of applications. A nil Status means every status.
type ListFilter struct {
	Status *ApprovalStatus
	Limit  int
	Offset int
}

// ParseStatusFilter reads the admin list's ?status= value. Empty means the review queue (pending),
// "all" means no filter. ok is false for anything else.
func ParseStatusFilter(raw string) (status *ApprovalStatus, ok bool) {
	switch raw {
	case "", string(ApprovalStatusPending):
		s := ApprovalStatusPending
		return &s, true
	case string(ApprovalStatusApproved):
		s := ApprovalStatusApproved
		return &s, true
	case string(ApprovalStatusRejected):
		s := ApprovalStatusRejected
		return &s, true
	case "all":
		return nil, true
	}
	return nil, false
}

// InviteListFilter selects a page of invites. A nil Status means every status.
type InviteListFilter struct {
	Status *InviteStatus
	Limit  int
	Offset int
}

// ParseInviteStatusFilter reads the invites list's ?status= value. Empty or "all" means no
// filter. ok is false for anything else.
func ParseInviteStatusFilter(raw string) (status *InviteStatus, ok bool) {
	switch InviteStatus(raw) {
	case InviteLive, InviteClaimed, InviteRevoked, InviteExpired:
		s := InviteStatus(raw)
		return &s, true
	}
	if raw == "" || raw == "all" {
		return nil, true
	}
	return nil, false
}
