package whatsapp_business_component

import (
	"wa_chat_service/pkg/formatter"
)

type Reaction struct {
	MessageID string `json:"message_id" validate:"required,startswith=wamid."` // ID of the message being reacted to
	Emoji     string `json:"emoji" validate:"required"`                        // The emoji used for the reaction (e.g., "👍", "❤️", "😂")
}

func (c Reaction) GetType() MessageType {
	return ReactionMessageType
}

func (c Reaction) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		string(c.GetType()): jsonData,
	}
}

func (c Reaction) Validate() error {
	validator := formatter.Validator()
	data := struct {
		Reaction Reaction `json:"reaction" validate:"required"`
	}{
		Reaction: c,
	}
	return validator.Validate(data)
}

func (c Reaction) GetMessage() string {
	return "Reacted to message: " + c.Emoji
}
