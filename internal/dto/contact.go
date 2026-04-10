package dto

import (
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/utils"
)

type (
	ContactCreateRequest struct {
		TenantID    string `json:"tenant_id" validate:"required"`
		PhoneNumber string `json:"phone_number" validate:"required"` // in international format without + sign, e.g. 6281234567890
		Name        string `json:"name" validate:"required"`
	}
	ContactResponse struct {
		ID          string `json:"id"`
		PhoneNumber string `json:"phone_number"`
		Name        string `json:"name"`
	}
	ContactGetFilteredRequest struct {
		TenantID string `json:"-" query:"tenant_id" validate:"required"`
	}
	ContactUpdateRequest struct {
		ID          string `query:"id" validate:"required"`
		TenantID    string `query:"tenant_id" validate:"required"`
		Name        string `json:"name" validate:"required"`
		PhoneNumber string `json:"phone_number" validate:"required"`
	}
)

func (ContactResponse) FromModel(data model.Contact) ContactResponse {
	return ContactResponse{
		ID:          data.DocumentID,
		PhoneNumber: data.PhoneNumber,
		Name:        data.Name,
	}
}

func (r ContactGetFilteredRequest) Validate() map[string]string {
	validator := utils.NewValidator()
	err := validator.Struct(r)
	if err != nil {
		return utils.GetValidatorErrorMessages(err)
	}
	return nil
}
