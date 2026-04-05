package dto

import (
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/formatter"
)

type (
	ChatGetByPhoneNumberIDRequest struct {
		PhoneNumberID string `query:"phone_number_id" validate:"required"`
	}
	ChatGetByPhoneNumberIDResponse struct {
		ID          string `json:"id"`        // {recipient_id}-{phone_number_id}
		ChatType    string `json:"chat_type"` // individual or group
		DisplayName string `json:"display_name"`
		LastMessage string `json:"last_message"`
		CreatedAt   int64  `json:"created_at"`
		UpdatedAt   int64  `json:"updated_at"`
	}
)

func (r ChatGetByPhoneNumberIDRequest) Validate() map[string]string {
	validator := formatter.Validator()
	err := validator.Validate(r)
	if err != nil {
		return validator.GetErrorMessages(err)
	}
	return nil
}

func (r *ChatGetByPhoneNumberIDResponse) FromModel(data model.Chat) {
	r.ID = data.DocumentID
	r.ChatType = data.ChatType
	r.DisplayName = data.DisplayName
	r.LastMessage = data.LastMessage
	r.CreatedAt = data.CreatedAt
	r.UpdatedAt = data.UpdatedAt
}
