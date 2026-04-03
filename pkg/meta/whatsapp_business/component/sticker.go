package whatsapp_business_component

import (
	"wa_chat_service/pkg/formatter"
)

type Sticker struct {
	ID   *string `json:"id,omitempty" validate:"required_without=Link,excluded_with=Link,omitempty,min=1"` // Only if using uploaded media, Required if using uploaded media, otherwise omit.
	Link *string `json:"link,omitempty" validate:"required_without=ID,excluded_with=ID,omitempty,uri"`     // Only if using hosted media (not recommended), Required if using hosted media, otherwise omit.
}

func (c Sticker) GetType() MessageType {
	return StickerMessageType
}

func (c Sticker) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		string(c.GetType()): jsonData,
	}
}

func (c Sticker) Validate() error {
	validator := formatter.Validator()
	data := struct {
		Sticker Sticker `json:"sticker" validate:"required"`
	}{
		Sticker: c,
	}
	return validator.Validate(data)
}

func (c Sticker) GetMessage() string {
	return "(Sticker)"
}
