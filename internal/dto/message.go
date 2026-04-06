package dto

import (
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/formatter"
)

type (
	MessageSendRequest struct {
		PhoneNumberID string `json:"phone_number_id" validate:"required"`
		RecipientID   string `json:"recipient_id" validate:"required"`
		RecipientName string `json:"recipient_name" validate:"required"`
		SenderName    string `json:"sender_name" validate:"required"`
		Type          string `json:"type" validate:"required"` // text, image, video, etc
		Payload       map[string]any
	}
	TemplateListRequest struct {
		PhoneNumberID string `query:"phone_number_id" validate:"required"`
	}

	MessageGetByChatIDRequest struct {
		ChatID string `query:"chat_id" validate:"required"`
	}
	MessageGetByChatIDResponse struct {
		ID              string                      `json:"id"`               // id from whatsapp
		ChatID          string                      `json:"chat_id"`          // reference to chat document
		StorageMediaID  *string                     `json:"storage_media_id"` // reference to media document if message has media
		StorageMedia    *StorageMediaUploadResponse `json:"storage_media"`
		MessageType     string                      `json:"message_type"`     // text, image, video, etc
		MessageCategory string                      `json:"message_category"` // marketing, authentication, utility, service
		SenderName      string                      `json:"sender_name"`      // sender name for individual chat, group name for group chat
		Payload         string                      `json:"payload"`          // raw payload from whatsapp, can be used for debugging or future processing
		// Content         string `json:"content"`          // extracted content from payload, can be used for searching or displaying in UI
		Status    string `json:"status"` // -, sent, delivered, read
		CreatedAt int64  `json:"created_at"`
		UpdatedAt int64  `json:"updated_at"`
	}
)

func (r MessageGetByChatIDRequest) Validate() map[string]string {
	validator := formatter.Validator()
	err := validator.Validate(r)
	if err != nil {
		return validator.GetErrorMessages(err)
	}
	return nil
}

func (r *MessageGetByChatIDResponse) FromModel(data model.Message, storageMedia *StorageMediaUploadResponse) {
	r.ID = data.DocumentID
	r.ChatID = data.ChatID
	r.StorageMediaID = data.StorageMediaID
	r.StorageMedia = storageMedia
	r.MessageType = data.MessageType
	r.MessageCategory = data.MessageCategory
	r.SenderName = data.SenderName
	r.Payload = data.Payload
	r.Status = data.Status
	r.CreatedAt = data.CreatedAt
	r.UpdatedAt = data.UpdatedAt
}
