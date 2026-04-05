package whatsapp_business_component

import (
	"wa_chat_service/pkg/formatter"
)

type Sticker struct {
	Media
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
