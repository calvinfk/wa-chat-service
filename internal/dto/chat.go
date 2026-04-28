package dto

import (
	"time"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/utils"
)

type (
	ChatGetByPhoneNumberIdRequest struct {
		PhoneNumberId string  `json:"phone_number_id" query:"phone_number_id" validate:"required"`
		ChatType      *string `json:"chat_type" query:"chat_type" validate:"omitempty,oneof=individual group ticket"`
		ChatStatus    *string `json:"chat_status" query:"chat_status" validate:"omitempty,oneof=open closed"`
	}
	ChatGetByPhoneNumberIdResponse struct {
		ID          string    `json:"id"`          // {recipient_id}-{phone_number_id}
		ChatType    string    `json:"chat_type"`   // individual, group, ticket
		ChatStatus  string    `json:"chat_status"` // open, closed
		DisplayName string    `json:"display_name"`
		LastMessage string    `json:"last_message"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}
	ChatCloseTicketRequest struct {
		ChatID string `json:"chat_id" validate:"required"`
	}
	ChatAssignAgentRequest struct {
		ChatID  string `json:"chat_id" validate:"required"`
		AgentID string `json:"agent_id" validate:"required,uuid"`
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
		ID:          data.DocumentID,
		ChatType:    data.ChatType,
		ChatStatus:  string(data.ChatStatus),
		DisplayName: data.DisplayName,
		LastMessage: data.LastMessage,
		CreatedAt:   data.CreatedAt,
		UpdatedAt:   data.UpdatedAt,
	}
}
