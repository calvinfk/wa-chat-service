package dto

import (
	"time"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/formatter"
)

type (
	TemplateCreateRequest struct {
		PhoneNumberID   string           `json:"phone_number_id" validate:"required"`
		Name            string           `json:"name" validate:"required"`
		Language        string           `json:"language" validate:"required"`
		Category        string           `json:"category" validate:"required"`
		ParameterFormat *string          `json:"parameter_format,omitempty"`
		Components      []map[string]any `json:"components" validate:"required"`
	}
	TemplateGetByPhoneNumberID struct {
		PhoneNumberID string `query:"phone_number_id" validate:"required"`
		Fields        string `query:"fields"`
	}
	TemplateGetByPhoneNumberIDResponse struct {
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
	}
	TemplateSyncRequest struct {
		PhoneNumberID string `json:"phone_number_id" validate:"required"`
	}

	TemplateDeleteRequest struct {
		PhoneNumberID string `query:"phone_number_id" validate:"required"`
		ID            string `query:"id" validate:"required_without=Name"`
		Name          string `query:"name" validate:"required_without=ID"`
	}
)

func (r TemplateGetByPhoneNumberID) Validate() map[string]string {
	validator := formatter.Validator()
	err := validator.Validate(r)
	return validator.GetErrorMessages(err)
}

func (TemplateGetByPhoneNumberIDResponse) FromModel(data model.Template) TemplateGetByPhoneNumberIDResponse {
	return TemplateGetByPhoneNumberIDResponse{
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
	}
}
