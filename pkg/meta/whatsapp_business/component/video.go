package whatsapp_business_component

import (
	"wa_chat_service/pkg/formatter"
)

type Video struct {
	ID      *string `json:"id,omitempty" validate:"required_without=Link,excluded_with=Link,omitempty,min=1"` // Only if using uploaded media, Required if using uploaded media, otherwise omit.
	Link    *string `json:"link,omitempty" validate:"required_without=ID,excluded_with=ID,omitempty,uri"`     // Only if using hosted media (not recommended), Required if using hosted media, otherwise omit.
	Caption *string `json:"caption,omitempty" validate:"omitempty,max=1024"`
}

func (c Video) GetType() MessageType {
	return VideoMessageType
}

func (c Video) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		string(c.GetType()): jsonData,
	}
}

func (c Video) Validate() error {
	validator := formatter.Validator()
	data := struct {
		Video Video `json:"video" validate:"required"`
	}{
		Video: c,
	}
	return validator.Validate(data)
}
func (c Video) GetMessage() string {
	if c.Caption != nil {
		return *c.Caption
	}
	return "(Video)"
}
