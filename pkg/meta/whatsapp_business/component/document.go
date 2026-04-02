package whatsapp_business_component

import (
	"wa_chat_service/pkg/formatter"
	"wa_chat_service/pkg/validate_struct"
)

type Document struct {
	Media
	Caption  *string `json:"caption,omitempty" validate:"omitempty,max=1024"`
	Filename *string `json:"filename,omitempty"` // Document filename, with extension.
}

func (c Document) GetType() string {
	return "document"
}

func (c Document) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		c.GetType(): jsonData,
	}
}

func (c Document) GetPayloadString() string {
	payload := c.GetPayload()[c.GetType()]
	jsonString, err := formatter.AnyToJsonString(payload)
	if err != nil {
		panic(err)
	}
	return jsonString
}

func (c Document) Validate() error {
	validator := validate_struct.New()
	data := struct {
		Document Document `json:"document" validate:"required"`
	}{
		Document: c,
	}
	return validator.Validate(data)
}
