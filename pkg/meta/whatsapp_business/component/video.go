package whatsapp_business_component

import (
	"wa_chat_service/pkg/formatter"
	"wa_chat_service/pkg/validate_struct"
)

type Video struct {
	ID      *string `json:"id,omitempty" validate:"required_without=Link,excluded_with=Link,omitempty,min=1"` // Only if using uploaded media, Required if using uploaded media, otherwise omit.
	Link    *string `json:"link,omitempty" validate:"required_without=ID,excluded_with=ID,omitempty,uri"`     // Only if using hosted media (not recommended), Required if using hosted media, otherwise omit.
	Caption *string `json:"caption,omitempty" validate:"omitempty,max=1024"`
}

func (c Video) GetType() string {
	return "video"
}

func (c Video) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		c.GetType(): jsonData,
	}
}

func (c Video) GetPayloadString() string {
	jsonData := c.GetPayload()[c.GetType()]
	jsonString, err := formatter.AnyToJsonString(jsonData)
	if err != nil {
		panic(err)
	}
	return jsonString
}

func (c Video) Validate() error {
	validator := validate_struct.New()
	data := struct {
		Video Video `json:"video" validate:"required"`
	}{
		Video: c,
	}
	return validator.Validate(data)
}
