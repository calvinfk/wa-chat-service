package dto

import (
	"time"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/utils"
)

type (
	TemplateCreateRequest struct {
		PhoneNumberID   string           `json:"phone_number_id" validate:"required"`
		Name            string           `json:"name" validate:"required"`
		Language        string           `json:"language" validate:"required"`
		Category        string           `json:"category" validate:"required,oneof=MARKETING UTILITY AUTHENTICATION"`
		ParameterFormat *string          `json:"parameter_format,omitempty"`
		Components      []map[string]any `json:"components" validate:"required"`
	}
	TemplateGetByTenantID struct {
		TenantID string  `json:"-" query:"tenant_id" validate:"required"`
		Search   string  `json:"-" query:"search"`
		Name     *string `json:"name,omitempty" query:"name" validate:"omitempty,min=1"`
		Status   *string `json:"status,omitempty" query:"status" validate:"omitempty,filter_options=APPROVED REJECTED PENDING"`
		Category *string `json:"category,omitempty" query:"category" validate:"omitempty,filter_options=MARKETING UTILITY AUTHENTICATION"`
	}
	TemplateResponse struct {
		ID                          string    `json:"id"`
		Name                        string    `json:"name"`
		Category                    string    `json:"category"`
		IsPrimaryDeviceDeliveryOnly bool      `json:"is_primary_device_delivery_only"`
		Language                    string    `json:"language"`
		MessageSendTTLSeconds       int       `json:"message_send_ttl_seconds"`
		ParameterFormat             *string   `json:"parameter_format"`
		Status                      string    `json:"status"`
		Components                  string    `json:"components"`
		CreatedAt                   time.Time `json:"created_at"`
		UpdatedAt                   time.Time `json:"updated_at"`
	}
	TemplateSyncRequest struct {
		PhoneNumberID string `json:"phone_number_id" validate:"required"`
	}

	TemplateDeleteRequest struct {
		PhoneNumberID string `query:"phone_number_id" validate:"required"`
		ID            string `query:"id" validate:"required"`
	}

	TemplateUpdateRequest struct {
		PhoneNumberID   string           `json:"phone_number_id" validate:"required"`
		ID              string           `json:"id" query:"id" validate:"required"`
		Name            string           `json:"name" validate:"required"`
		Language        string           `json:"language" validate:"required"`
		Category        string           `json:"category" validate:"required,oneof=MARKETING UTILITY AUTHENTICATION"`
		ParameterFormat *string          `json:"parameter_format,omitempty"`
		Components      []map[string]any `json:"components" validate:"required"`
	}
)

func (r TemplateGetByTenantID) Validate() map[string]string {
	validator := utils.NewValidator()
	err := validator.Struct(r)
	return utils.GetValidatorErrorMessages(err)
}

func (TemplateResponse) FromModel(data model.Template) TemplateResponse {
	return TemplateResponse{
		ID:                          data.DocumentID,
		Category:                    data.Category,
		Components:                  data.Components,
		IsPrimaryDeviceDeliveryOnly: data.IsPrimaryDeviceDeliveryOnly,
		Language:                    data.Language,
		MessageSendTTLSeconds:       data.MessageSendTTLSeconds,
		Name:                        data.Name,
		ParameterFormat:             data.ParameterFormat,
		Status:                      data.Status,
		CreatedAt:                   data.CreatedAt,
		UpdatedAt:                   data.UpdatedAt,
	}
}
