package whatsapp_business

import (
	"wa_chat_service/pkg/meta/whatsapp_business/message_components"
)

type UploadMediaResponse struct {
	ID string `json:"id"`
}

type GetMediaURLResponse struct {
	MessagingProduct string `json:"messaging_product"`
	URL              string `json:"url"`
	MimeType         string `json:"mime_type"`
	Sha256           string `json:"sha256"`
	FileSize         int64  `json:"file_size"`
	ID               string `json:"id"`
}

type DeleteMediaResponse struct {
	Success bool `json:"success"`
}

type MessageComponent interface {
	GetType() message_components.MessageType
	GetPayload() map[string]any
	Validate() error
	GetMessage() string
}

type MessageResponse struct {
	MessagingProduct string `json:"messaging_product"`
	Contacts         []struct {
		Input string `json:"input"`
		WaID  string `json:"wa_id"`
	} `json:"contacts"`
	Messages []struct {
		ID            string  `json:"id"`
		MessageStatus *string `json:"message_status,omitempty"` // if sending a templatete message, this field will be present in the response
	} `json:"messages"`
}

type Paging struct {
	Cursors struct {
		Before string `json:"before,omitempty"`
		After  string `json:"after,omitempty"`
	} `json:"cursors"`
	Next     string `json:"next,omitempty"`
	Previous string `json:"previous,omitempty"`
}
type TemplateResponse struct {
	Category                    string  `json:"category" validate:"required"`   // MARKETING, UTILITY, etc
	Components                  []any   `json:"components" validate:"required"` // header, body, button, etc
	ID                          string  `json:"id" validate:"required"`
	IsPrimaryDeviceDeliveryOnly bool    `json:"is_primary_device_delivery_only"`
	Language                    string  `json:"language" validate:"required"`
	MessageSendTTLSeconds       int     `json:"message_send_ttl_seconds"`
	Name                        string  `json:"name" validate:"required"`
	ParameterFormat             *string `json:"parameter_format,omitempty"`
	Status                      string  `json:"status" validate:"required"` // approved, rejected, etc
}

type TemplateCreateRequest struct {
	Name            string           `json:"name" validate:"required"`
	Category        string           `json:"category" validate:"required,oneof=marketing utility authentication"`
	Language        string           `json:"language" validate:"required"`
	ParameterFormat *string          `json:"parameter_format,omitempty" validate:"omitempty,oneof=named positional"`
	Components      []map[string]any `json:"components" validate:"required,dive"`
}

type TemplateCreateResponse struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Category string `json:"category"`
}

type TemplateDeleteRequest struct {
	ID   string `json:"id" validate:"required_without=Name"`
	Name string `json:"name" validate:"required_without=ID"`
}

type TemplateDeleteResponse struct {
	Success bool `json:"success"`
}

type StartUploadSessionRequest struct {
	FileName    string `query:"filename" validate:"required"`
	FileLength  int64  `query:"file_length" validate:"required"`
	FileType    string `query:"file_type" validate:"required"`
	AccessToken string `query:"access_token" validate:"required"`
}

type StartUploadSessionResponse struct {
	ID         string `json:"id"`
	FileOffset int64  `json:"file_offset"`
}

type UploadFileRequest struct {
	UploadSessionID string `validate:"required"`
	FileOffset      int64  `validate:"min=0"`
	FileBytes       []byte `validate:"required"`
}
type UploadFileResponse struct {
	H string `json:"h"`
}
