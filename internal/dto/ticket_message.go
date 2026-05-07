package dto

import (
	"time"
	"wa_chat_service/pkg/utils"
)

type (
	TicketMessageGetByTicketIDRequest struct {
		TicketID string `query:"ticket_id" validate:"required"`
		Search   string `json:"-" query:"search"`
	}

	TicketMessageResponse struct {
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
)

func (r TicketMessageGetByTicketIDRequest) Validate() map[string]string {
	validator := utils.NewValidator()
	err := validator.Struct(r)
	if err != nil {
		return utils.GetValidatorErrorMessages(err)
	}
	return nil
}
