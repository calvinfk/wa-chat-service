package whatsapp_business_component

import (
	"wa_chat_service/pkg/formatter"
	"wa_chat_service/pkg/validate_struct"
)

type Reaction struct {
	MessageID string `json:"message_id" validate:"required,startswith=wamid."` // ID of the message being reacted to
	Emoji     string `json:"emoji" validate:"required"`                        // The emoji used for the reaction (e.g., "👍", "❤️", "😂")
}

func (c Reaction) GetType() string {
	return "reaction"
}

func (c Reaction) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		c.GetType(): jsonData,
	}
}

func (c Reaction) GetPayloadString() string {
	jsonData := c.GetPayload()[c.GetType()]
	jsonString, err := formatter.AnyToJsonString(jsonData)
	if err != nil {
		panic(err)
	}
	return jsonString
}

func (c Reaction) Validate() error {
	validator := validate_struct.New()
	data := struct {
		Reaction Reaction `json:"reaction" validate:"required"`
	}{
		Reaction: c,
	}
	return validator.Validate(data)
}
