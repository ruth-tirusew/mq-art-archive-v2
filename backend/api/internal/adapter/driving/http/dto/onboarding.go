package dto

type ReviewApplicationRequest struct {
	Status string `json:"status" binding:"required"`
	Notes  string `json:"notes"`
}
