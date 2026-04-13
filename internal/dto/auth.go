package dto

type (
	AuthLoginRequest struct {
		PhoneNumberID string `json:"phone_number_id" validate:"required"`
		TenantID      string `json:"tenant_id" validate:"required"`
	}
)
