package identity

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RolePublic       Role = "public"
	RoleArtist       Role = "artist"
	RoleInstitution  Role = "institution"
	RoleContributor  Role = "contributor"
	RoleAdmin        Role = "admin"
)

type User struct {
	ID        uuid.UUID
	Email     string
	Role      Role
	CreatedAt time.Time
	UpdatedAt time.Time
}
