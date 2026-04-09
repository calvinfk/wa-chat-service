package dto

import (
	"time"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/utils"
)

type (
	ChatGetByPhoneNumberIDRequest struct {
		PhoneNumberID string `query:"phone_number_id" validate:"required"`
	}
	ChatGetByPhoneNumberIDResponse struct {
		ID          string    `json:"id"`        // {recipient_id}-{phone_number_id}
		ChatType    string    `json:"chat_type"` // individual or group
		DisplayName string    `json:"display_name"`
		LastMessage string    `json:"last_message"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}
)

func (r ChatGetByPhoneNumberIDRequest) Validate() map[string]string {
	validator := utils.NewValidator()
	err := validator.Struct(r)
	if err != nil {
		return utils.GetValidatorErrorMessages(err)
	}
	return nil
}

func (ChatGetByPhoneNumberIDResponse) FromModel(data model.Chat) ChatGetByPhoneNumberIDResponse {
	return ChatGetByPhoneNumberIDResponse{
		ID:          data.DocumentID,
		ChatType:    data.ChatType,
		DisplayName: data.DisplayName,
		LastMessage: data.LastMessage,
		CreatedAt:   data.CreatedAt,
		UpdatedAt:   data.UpdatedAt,
	}
}
