package profile

// IsPublic reports whether the profile may be shown on public pages. Draft and pending profiles
// (which hold contact details) are only visible to their owner and admins.
func (p *ArtistProfile) IsPublic() bool {
	return p != nil && p.Status == ProfileStatusApproved
}
