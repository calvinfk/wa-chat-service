package dto

import (
	"time"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/utils"
)

type (
	MessageSendRequest struct {
		ChatID     string `json:"chat_id" validate:"required"` // {recipient_id}-{phone_number_id}, or uuid v7 for group and ticket
		SenderName string `json:"sender_name"`                 // optional, if not provided, will use current logged in user
		Type       string `json:"type" validate:"required"`    // text, image, video, etc
		Payload    any
		// Test       Test `json:"test" validate:"required"` // if true, will not actually send message to whatsapp, only save to database for testing purposes
	}

	Test struct {
		ID string `json:"id" validate:"required"`
	}

	MessageGetByChatIDRequest struct {
		ChatID string `query:"chat_id" validate:"required"`
		Search string `json:"-" query:"search"`
	}

	MessageResponse struct {
		ID              string                `json:"id"`      // id from whatsapp
		ChatID          string                `json:"chat_id"` // reference to chat document
		StorageMedia    *StorageMediaResponse `json:"storage_media"`
		MessageType     string                `json:"message_type"`     // text, image, video, etc
		MessageCategory string                `json:"message_category"` // marketing, authentication, utility, service, (and system_flag for system generated messages)
		SenderName      string                `json:"sender_name"`      // sender name for individual chat, group name for group chat
		Payload         string                `json:"payload"`          // raw payload from whatsapp or system, can be used for debugging or future processing
		Status          string                `json:"status"`           // -, sent, delivered, read
		CreatedAt       time.Time             `json:"created_at"`
		SentAt          *time.Time            `json:"sent_at,omitempty"`
		DeliveredAt     *time.Time            `json:"delivered_at,omitempty"`
		ReadAt          *time.Time            `json:"read_at,omitempty"`
		Error           *string               `json:"error,omitempty"` // error message if failed to send, json stringified from whatsapp error response
	}

	MessageSaveRequest struct {
		ID                *string    `json:"id"`
		ChatID            *string    `json:"chat_id"`
		Wamid             string     `json:"wamid" validate:"required"`
		PhoneNumberId     string     `json:"phone_number_id" validate:"required"`
		RecipientId       string     `json:"recipient_id" validate:"required"`
		RecipientName     string     `json:"recipient_name"`
		LastMessage       string     `json:"last_message"`
		UserLastMessageAt *time.Time `json:"user_last_message_at"`
		MessageType       string     `json:"message_type" validate:"required"`
		MessageCategory   string     `json:"message_category" validate:"required"`
		SenderName        string     `json:"sender_name" validate:"required"`
		Payload           string     `json:"payload" validate:"required"`
		StorageMediaID    *string    `json:"storage_media_id"`
		Status            string     `json:"status" validate:"required"`
		CreatedAt         time.Time  `json:"created_at" validate:"required"`
		SentAt            *time.Time `json:"sent_at"`
		DeliveredAt       *time.Time `json:"delivered_at"`
		ReadAt            *time.Time `json:"read_at"`
		Error             *string    `json:"error"`
	}
)

func (r MessageGetByChatIDRequest) Validate() map[string]string {
	validator := utils.NewValidator()
	err := validator.Struct(r)
	if err != nil {
		return utils.GetValidatorErrorMessages(err)
	}
	return nil
}

func (MessageResponse) FromModel(data model.Message, storageMedia *StorageMediaResponse) MessageResponse {
	return MessageResponse{
		ID:              data.DocumentID,
		ChatID:          data.ChatID,
		StorageMedia:    storageMedia,
		MessageType:     data.MessageType,
		MessageCategory: data.MessageCategory,
		SenderName:      data.SenderName,
		Payload:         data.Payload,
		Status:          data.Status,
		CreatedAt:       data.CreatedAt,
		SentAt:          data.SentAt,
		DeliveredAt:     data.DeliveredAt,
		ReadAt:          data.ReadAt,
		Error:           data.Error,
	}
}
