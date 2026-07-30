package dto

import (
	"time"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/institution"
)

type InstitutionContactResponse struct {
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Website  string `json:"website,omitempty"`
	Location string `json:"location,omitempty"`
}

type InstitutionResponse struct {
	ID          uuid.UUID                  `json:"id"`
	Slug        string                     `json:"slug"`
	Name        string                     `json:"name"`
	Description string                     `json:"description,omitempty"`
	Contact     InstitutionContactResponse `json:"contact,omitempty"`
	Status      string                     `json:"status"`
	CreatedAt   time.Time                  `json:"created_at"`
	UpdatedAt   time.Time                  `json:"updated_at"`
}

func ToInstitutionResponse(inst domain.Institution) InstitutionResponse {
	return InstitutionResponse{
		ID:          inst.ID,
		Slug:        inst.Slug,
		Name:        inst.Name,
		Description: inst.Description,
		Contact: InstitutionContactResponse{
			Email:    inst.Contact.Email,
			Phone:    inst.Contact.Phone,
			Website:  inst.Contact.Website,
			Location: inst.Contact.Location,
		},
		Status:    string(inst.Status),
		CreatedAt: inst.CreatedAt,
		UpdatedAt: inst.UpdatedAt,
	}
}
