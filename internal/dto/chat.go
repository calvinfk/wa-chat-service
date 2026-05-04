package dto

import (
	"time"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/utils"
)

type (
	ChatGetByPhoneNumberIdRequest struct {
		PhoneNumberId string  `json:"phone_number_id" query:"phone_number_id" validate:"required"`
		AgentID       *string `json:"agent_id"`
		ChatType      *string `json:"chat_type" query:"chat_type" validate:"omitempty,filter_options=individual group ticket"`
		ChatStatus    *string `json:"chat_status" query:"chat_status" validate:"omitempty,filter_options=open in_progress closed"`
	}
	ChatGetByPhoneNumberIdResponse struct {
		ID            string    `json:"id"`          // {recipient_id}-{phone_number_id}
		ChatType      string    `json:"chat_type"`   // individual, group, ticket
		ChatStatus    string    `json:"chat_status"` // open, in_progress, closed
		RecipientName string    `json:"recipient_name"`
		LastMessage   string    `json:"last_message"`
		CreatedAt     time.Time `json:"created_at"`
		UpdatedAt     time.Time `json:"updated_at"`
	}
	ChatCloseTicketRequest struct {
		ChatID string `json:"chat_id" validate:"required"`
	}
	ChatAssignAgentRequest struct {
		ChatID  string `json:"chat_id" validate:"required"`
		AgentID string `json:"agent_id" validate:"required,uuid"`
	}
	ChatCreateRequest struct {
		PhoneNumberId string `json:"phone_number_id" validate:"required"`
		RecipientId   string `json:"recipient_id" validate:"required"`
		RecipientName string `json:"recipient_name" validate:"required"`
	}
	ChatGetTicketAnalyticsRequest struct {
		PhoneNumberIds *[]string `query:"phone_number_ids" validate:"omitempty,min=1,dive,required"` // if provided, must contain at least 1 phone number ID and each ID is required
		StartTime      time.Time `query:"start_time" validate:"required,lt"`
		EndTime        time.Time `query:"end_time" validate:"required,gtefield=StartTime"`
	}
	ChatGetTicketAnalyticsResponse struct {
		TotalCount                int `json:"total_count"`
		AverageResolutionMinutes  int `json:"average_resolution_minutes"`
		MedianResolutionMinutes   int `json:"median_resolution_minutes"`
		LongestResolutionMinutes  int `json:"longest_resolution_minutes"`
		ShortestResolutionMinutes int `json:"shortest_resolution_minutes"`
		OpenedCount               int `json:"opened_count"`
		InProgressCount           int `json:"in_progress_count"`
		ClosedCount               int `json:"closed_count"`
	}
)

func (r ChatGetByPhoneNumberIdRequest) Validate() map[string]string {
	validator := utils.NewValidator()
	err := validator.Struct(r)
	if err != nil {
		return utils.GetValidatorErrorMessages(err)
	}
	return nil
}

func (ChatGetByPhoneNumberIdResponse) FromModel(data model.Chat) ChatGetByPhoneNumberIdResponse {
	return ChatGetByPhoneNumberIdResponse{
		ID:            data.DocumentID,
		ChatType:      data.ChatType,
		ChatStatus:    string(data.ChatStatus),
		RecipientName: data.RecipientName,
		LastMessage:   data.LastMessage,
		CreatedAt:     data.CreatedAt,
		UpdatedAt:     data.UpdatedAt,
	}
}
