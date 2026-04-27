package dto

import (
	"time"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/utils"
)

type (
	TemplateCreateRequest struct {
		WaBusinessAccountID string           `json:"wa_business_account_id" validate:"required,numeric"`
		Name                string           `json:"name" validate:"required"`
		Language            string           `json:"language" validate:"required"`
		Category            string           `json:"category" validate:"required,oneof=MARKETING UTILITY AUTHENTICATION"`
		ParameterFormat     *string          `json:"parameter_format,omitempty"`
		Components          []map[string]any `json:"components" validate:"required"`
	}
	TemplateFilterRequest struct {
		WaBusinessAccountID string  `json:"wa_business_account_id" query:"wa_business_account_id" validate:"required"`
		Search              string  `json:"-" query:"search"`
		Name                *string `json:"name,omitempty" query:"name" validate:"omitempty,min=1"`
		Status              *string `json:"status,omitempty" query:"status" validate:"omitempty,filter_options=APPROVED REJECTED PENDING"`
		Category            *string `json:"category,omitempty" query:"category" validate:"omitempty,filter_options=MARKETING UTILITY AUTHENTICATION"`
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
		WaBusinessAccountID string `json:"wa_business_account_id" validate:"required,numeric"`
	}

	TemplateDeleteRequest struct {
		WaBusinessAccountID string `query:"wa_business_account_id" validate:"required"`
		ID                  string `query:"id" validate:"required"`
	}

	TemplateUpdateRequest struct {
		WaBusinessAccountID string           `json:"wa_business_account_id" validate:"required"`
		ID                  string           `json:"id" query:"id" validate:"required"`
		Name                string           `json:"name" validate:"required"`
		Language            string           `json:"language" validate:"required"`
		Category            string           `json:"category" validate:"required,oneof=MARKETING UTILITY AUTHENTICATION"`
		ParameterFormat     *string          `json:"parameter_format,omitempty"`
		Components          []map[string]any `json:"components" validate:"required"`
	}
)

func (r TemplateFilterRequest) Validate() map[string]string {
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
