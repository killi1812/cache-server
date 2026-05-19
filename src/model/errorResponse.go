package model

// ErrorResponse represents a common error DTO for API responses.
type ErrorResponse struct {
	Error          string   `json:"error" example:"not found"`
	AdditionalInfo []string `json:"additional_info"`
}
