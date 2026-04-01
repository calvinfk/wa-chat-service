package whatsapp_business_component

import (
	"wa_chat_service/pkg/formatter"
	"wa_chat_service/pkg/validate_struct"
)

type Audio struct {
	ID    *string `json:"id,omitempty" validate:"required_without=Link,excluded_with=Link"` // Only if using uploaded media
	Link  *string `json:"link,omitempty" validate:"required_without=ID,excluded_with=ID"`   // Only if using hosted media (not recommended)
	Voice *bool   `json:"voice,omitempty"`                                                  // Only include if sending voice message
}

func (c Audio) GetType() string {
	return "audio"
}

func (c Audio) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		c.GetType(): jsonData,
	}
}

func (c Audio) GetPayloadString() string {
	jsonData := c.GetPayload()[c.GetType()]
	jsonString, err := formatter.AnyToJsonString(jsonData)
	if err != nil {
		panic(err)
	}
	return jsonString
}

func (c Audio) Validate() error {
	validator := validate_struct.New()
	data := struct {
		Audio Audio `json:"audio" validate:"required"`
	}{
		Audio: c,
	}
	return validator.Validate(data)
}
