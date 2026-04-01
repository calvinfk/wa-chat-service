package dto

type (
	MessageSendRequest struct {
		PhoneNumberID string `json:"phone_number_id" validate:"required"`
		RecipientID   string `json:"recipient_id" validate:"required"`
		RecipientName string `json:"recipient_name" validate:"required"`
		SenderName    string `json:"sender_name" validate:"required"`
		Type          string `json:"type" validate:"required"` // text, image, video, etc
		Payload       any    `json:"payload" validate:"required"`
	}
)
