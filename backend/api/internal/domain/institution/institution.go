package institution

import (
	"time"

	"github.com/google/uuid"
)

type InstitutionStatus string

const (
	InstitutionStatusDraft    InstitutionStatus = "draft"
	InstitutionStatusPending  InstitutionStatus = "pending"
	InstitutionStatusApproved InstitutionStatus = "approved"
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
	Status      InstitutionStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
