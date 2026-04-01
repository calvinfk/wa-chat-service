package whatsapp_business_component

import (
	"wa_chat_service/pkg/formatter"
	"wa_chat_service/pkg/validate_struct"
)

type Text struct {
	EnableLinkPreview *bool  `json:"enable_link_preview,omitempty"`
	Body              string `json:"body" validate:"required"`
}

func (t Text) GetType() string {
	return "text"
}

func (t Text) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(t, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		t.GetType(): jsonData,
	}
}

func (t Text) GetPayloadString() string {
	jsonData := t.GetPayload()[t.GetType()]
	jsonString, err := formatter.AnyToJsonString(jsonData)
	if err != nil {
		panic(err)
	}
	return jsonString
}

func (t Text) Validate() error {
	validator := validate_struct.New()
	return validator.Validate(t)
}
