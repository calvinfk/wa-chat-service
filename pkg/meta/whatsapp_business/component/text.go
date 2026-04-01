package whatsapp_business_component

import (
	"wa_chat_service/pkg/formatter"
	"wa_chat_service/pkg/validate_struct"
)

type Text struct {
	EnableLinkPreview *bool  `json:"enable_link_preview,omitempty"`
	Body              string `json:"body" validate:"required"`
}

func (c Text) GetType() string {
	return "text"
}

func (c Text) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		c.GetType(): jsonData,
	}
}

func (c Text) GetPayloadString() string {
	jsonData := c.GetPayload()[c.GetType()]
	jsonString, err := formatter.AnyToJsonString(jsonData)
	if err != nil {
		panic(err)
	}
	return jsonString
}

func (c Text) Validate() error {
	validator := validate_struct.New()
	return validator.Validate(c)
}
