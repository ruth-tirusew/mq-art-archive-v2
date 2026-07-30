package media

import (
	"time"

	"github.com/google/uuid"
)

const MaxImageBytes = 10 * 1024 * 1024

type Asset struct {
	ID           uuid.UUID `json:"id"`
	OwnerUserID  uuid.UUID `json:"owner_user_id"`
	PublicID     string    `json:"public_id"`
	SecureURL    string    `json:"secure_url"`
	ResourceType string    `json:"resource_type"`
	Width        *int      `json:"width,omitempty"`
	Height       *int      `json:"height,omitempty"`
	Bytes        *int      `json:"bytes,omitempty"`
	Folder       string    `json:"folder"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UploadOptions struct {
	PublicID string
	Folder   string
}

type UploadSignature struct {
	Timestamp int64     `json:"timestamp"`
	Signature string    `json:"signature"`
	CloudName string    `json:"cloud_name"`
	APIKey    string    `json:"api_key"`
	Folder    string    `json:"folder"`
	PublicID  string    `json:"public_id"`
	ExpireAt  time.Time `json:"expire_at"`
}

type Completion struct {
	PublicID     string
	SecureURL    string
	ResourceType string
	Format       string
	Width        int
	Height       int
	Bytes        int
}
