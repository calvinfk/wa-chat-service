package dto

import (
	"time"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/utils"
)

type (
	BroadcastUpsertRequest struct {
		ID            *string          `query:"id,omitempty"`
		PhoneNumberID string           `json:"phone_number_id" validate:"required"`
		TemplateID    string           `json:"template_id" validate:"required"`
		Name          string           `json:"name" validate:"required"`
		SendAt        *time.Time       `json:"send_at" validate:"omitempty,gt"`
		Status        string           `json:"status" validate:"required,oneof=draft scheduled"`
		Recipients    []string         `json:"recipients" validate:"required,min=1,dive,required"`
		Components    []map[string]any `json:"components" validate:"required,min=1,dive"`
	}

	BroadcastResponse struct {
		ID              string    `json:"id"` // uuid v7
		Name            string    `json:"name"`
		TemplateID      string    `json:"template_id"`
		PhoneNumberID   string    `json:"phone_number_id"`
		ParameterFormat *string   `json:"parameter_format"`
		Payload         string    `json:"payload"` // raw json string of template
		Status          string    `json:"status"`  // scheduled, sent, cancelled
		SendAt          time.Time `json:"send_at"`
		CreatedBy       string    `json:"created_by"`
		CreatedAt       time.Time `json:"created_at"`
		UpdatedAt       time.Time `json:"updated_at"`
	}

	BroadcastScheduleRequest struct {
		ID     string     `query:"id" validate:"required"`
		SendAt *time.Time `json:"send_at" validate:"omitempty,gt"` // if empty or null, will be sent as soon as possible
	}

	BroadcastCancelRequest struct {
		ID string `query:"id" validate:"required"`
	}

	BroadcastGetFilteredRequest struct {
		PhoneNumberID string  `json:"-" query:"phone_number_id" validate:"required"`
		TenantID      string  `json:"-"`
		Status        *string `json:"status" query:"status" validate:"omitempty,filter_options=draft failed failed_partially cancelled success sending scheduled"`
	}
)

type BroadcastScheduleStatus string

const (
	BroadcastScheduleDraft           BroadcastScheduleStatus = "draft"
	BroadcastScheduleFailed          BroadcastScheduleStatus = "failed"
	BroadcastScheduleFailedPartially BroadcastScheduleStatus = "failed_partially"
	BroadcastScheduleCancelled       BroadcastScheduleStatus = "cancelled"
	BroadcastScheduleSuccess         BroadcastScheduleStatus = "success"
	BroadcastScheduleSending         BroadcastScheduleStatus = "sending"
	BroadcastScheduleScheduled       BroadcastScheduleStatus = "scheduled"
)

func (BroadcastResponse) FromModel(data model.Broadcast) BroadcastResponse {
	return BroadcastResponse{
		ID:              data.DocumentID,
		Name:            data.Name,
		TemplateID:      data.TemplateID,
		PhoneNumberID:   data.PhoneNumberID,
		ParameterFormat: data.ParameterFormat,
		Payload:         data.Payload,
		Status:          data.Status,
		SendAt:          data.SendAt,
		CreatedBy:       data.CreatedBy,
		CreatedAt:       data.CreatedAt,
		UpdatedAt:       data.UpdatedAt,
	}
}

func (r BroadcastGetFilteredRequest) Validate() map[string]string {
	validator := utils.NewValidator()
	if err := validator.Struct(r); err != nil {
		return utils.GetValidatorErrorMessages(err)
	}
	return nil
}
