package message_components

import (
	"wa_chat_service/pkg/utils"
)

type Text struct {
	EnableLinkPreview *bool  `json:"enable_link_preview,omitempty"`
	Body              string `json:"body" validate:"required"`
}

func (c Text) GetType() MessageType {
	return TextMessageType
}

func (c Text) GetPayload() map[string]any {
	jsonData, err := utils.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		string(c.GetType()): jsonData,
	}
}

func (c Text) Validate() error {
	validator := utils.NewValidator()
	data := struct {
		Text Text `json:"text" validate:"required"`
	}{
		Text: c,
	}
	return validator.Struct(data)
}

func (c Text) GetMessage() string {
	return c.Body
}
