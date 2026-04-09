package message_components

import (
	"wa_chat_service/pkg/utils"
)

type Sticker struct {
	Media
}

func (c Sticker) GetType() MessageType {
	return StickerMessageType
}

func (c Sticker) GetPayload() map[string]any {
	jsonData, err := utils.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		string(c.GetType()): jsonData,
	}
}

func (c Sticker) Validate() error {
	validator := utils.NewValidator()
	data := struct {
		Sticker Sticker `json:"sticker" validate:"required"`
	}{
		Sticker: c,
	}
	return validator.Struct(data)
}

func (c Sticker) GetMessage() string {
	return "(Sticker)"
}
