package dto

type (
	TemplateCreateRequest struct {
		PhoneNumberID string           `json:"phone_number_id" validate:"required"`
		Name          string           `json:"name" validate:"required"`
		Language      string           `json:"language" validate:"required"`
		Category      string           `json:"category" validate:"required"`
		Components    []map[string]any `json:"components" validate:"required"`
	}
)
