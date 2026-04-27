package dto

import (
	"time"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/utils"
)

type (
	BroadcastUpsertRequest struct {
		ID            *string          `query:"id,omitempty"`
		PhoneNumberId string           `json:"phone_number_id" validate:"required"`
		TemplateID    string           `json:"template_id" validate:"required"`
		Name          string           `json:"name" validate:"required"`
		SendAt        *time.Time       `json:"send_at" validate:"omitempty,gt"`
		Status        string           `json:"status" validate:"required,oneof=draft scheduled"`
		Recipients    []string         `json:"recipients" validate:"required,min=1,dive,required"`
		Components    []map[string]any `json:"components" validate:"omitempty,min=1,dive"`
	}

	BroadcastResponse struct {
		ID              string    `json:"id"` // uuid v7
		Name            string    `json:"name"`
		TemplateID      string    `json:"template_id"`
		PhoneNumberId   string    `json:"phone_number_id"`
		RecipientTotal  int       `json:"recipient_total"`
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
		Status *string `json:"status" query:"status" validate:"omitempty,filter_options=draft failed failed_partially cancelled success sending scheduled"`
	}

	BroadcastGetRecipientsFilteredRequest struct {
		BroadcastID string `json:"-" query:"broadcast_id" validate:"required"`
	}

	BroadcastRecipientResponse struct {
		ID            string    `json:"id"`           // uuid v7
		BroadcastID   string    `json:"broadcast_id"` // reference to broadcast document
		WamId         string    `json:"wamid,omitempty"`
		RecipientId   string    `json:"recipient_id"`
		RecipientName string    `json:"recipient_name"`
		RecipientType string    `json:"recipient_type"` // individual, group
		ReplyData     *string   `json:"reply_data"`
		Status        string    `json:"status"` // pending, sent, failed
		CreatedAt     time.Time `json:"created_at"`
		UpdatedAt     time.Time `json:"updated_at"`
		Errors        *string   `json:"errors"`
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
		RecipientTotal:  len(data.RecipientIds),
		PhoneNumberId:   data.PhoneNumberId,
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

func (r BroadcastGetRecipientsFilteredRequest) Validate() map[string]string {
	validator := utils.NewValidator()
	if err := validator.Struct(r); err != nil {
		return utils.GetValidatorErrorMessages(err)
	}
	return nil
}

func (r BroadcastRecipientResponse) FromModel(data model.BroadcastRecipient) BroadcastRecipientResponse {
	return BroadcastRecipientResponse{
		ID:            data.DocumentID,
		BroadcastID:   data.BroadcastID,
		WamId:         data.WamId,
		RecipientId:   data.RecipientId,
		RecipientName: data.RecipientName,
		RecipientType: data.RecipientType,
		ReplyData:     data.ReplyData,
		Status:        data.Status,
		CreatedAt:     data.CreatedAt,
		UpdatedAt:     data.UpdatedAt,
		Errors:        data.Errors,
	}
}
