package institution

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusDraft    Status = "draft"
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
)

type ContactInfo struct {
	Email    string
	Phone    string
	Website  string
	Location string
}

type Institution struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Slug        string
	Name        string
	Description string
	Contact     ContactInfo
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
